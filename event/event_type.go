package event

import (
	"encoding/json"
	"fmt"
	"monotone/se.oort.schedrox/auth"
	"monotone/se.oort.schedrox/common"
	"time"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func eventKindsKeyForDomain(d *datastore.Key) string {
	return fmt.Sprintf("EventKinds{Domain:%v}", d)
}

func eventKindKeyForId(i *datastore.Key) string {
	return fmt.Sprintf("EventKind{Id:%v}", i)
}

func eventTypesKeyForDomain(d *datastore.Key) string {
	return fmt.Sprintf("EventTypes{Domain:%v}", d)
}

func eventTypeKeyForId(i *datastore.Key) string {
	return fmt.Sprintf("EventType{Id:%v}", i)
}

type EventType struct {
	Id                                   *datastore.Key `datastore:"-" json:"id"`
	EventKind                            *datastore.Key `json:"event_kind"`
	Color                                string         `json:"color"`
	Name                                 string         `json:"name"`
	DefaultMinutes                       int            `json:"default_minutes"`
	TitleSize                            int            `json:"title_size"`
	NameHiddenInCalendar                 bool           `json:"name_hidden_in_calendar"`
	ParticipantsFormat                   string         `json:"participants_format,omitempty"`
	Unique                               bool           `json:"unique"`
	DisplayUsersInCalendar               bool           `json:"display_users_in_calendar"`
	SignalColorsWhen0Contacts            bool           `json:"signal_colors_when_0_contacts"`
	SignalColorsWhenMorePossibleContacts bool           `json:"signal_colors_when_more_possible_contacts"`
	SignalColorsWhenMorePossibleUsers    bool           `json:"signal_colors_when_more_possible_users"`

	ConfirmationEmailSubjectTemplateBytes []byte `json:"-"`
	ConfirmationEmailSubjectTemplate      string `json:"confirmation_email_subject_template" datastore:"-"`
	ConfirmationEmailBodyTemplateBytes    []byte `json:"-"`
	ConfirmationEmailBodyTemplate         string `json:"confirmation_email_body_template" datastore:"-"`

	// Salary mod
	SalarySerializedProperties []byte                 `json:"-"`
	SalaryProperties           map[string]interface{} `json:"salary_properties" datastore:"-"`
}

func (self *EventType) process(c gaecontext.HTTPContext) *EventType {
	self.ConfirmationEmailBodyTemplate = string(self.ConfirmationEmailBodyTemplateBytes)
	self.ConfirmationEmailSubjectTemplate = string(self.ConfirmationEmailSubjectTemplateBytes)
	if len(self.SalarySerializedProperties) > 0 {
		if err := json.Unmarshal(self.SalarySerializedProperties, &self.SalaryProperties); err != nil {
			panic(err)
		}
	}
	return self
}

func findEventType(c gaecontext.HTTPContext, key *datastore.Key) *EventType {
	var t EventType
	if err := datastore.Get(c, key, &t); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil
		}
		panic(err)
	}
	t.Id = key
	return &t
}

func (self *Event) GetEventType(c gaecontext.HTTPContext) *EventType {
	if self.cachedEventType == nil {
		self.cachedEventType = GetEventType(c, self.EventType)
	}
	return self.cachedEventType
}

func GetEventType(c gaecontext.HTTPContext, key *datastore.Key) *EventType {
	if key == nil {
		return nil
	}
	var t EventType
	if common.Memoize(c, eventTypeKeyForId(key), &t, func() interface{} {
		return findEventType(c, key)
	}) {
		return (&t).process(c)
	}
	return nil
}

func (self *EventType) GetIdsBetween(c gaecontext.HTTPContext, d *datastore.Key, start, end time.Time) (result []*datastore.Key) {
	for _, ev := range GetEventsBetween(c, d, start, end, func(ev *Event) Events {
		if self.Id.Equal(ev.EventType) {
			return Events{*ev}
		}
		return nil
	}) {
		if ev.Recurring {
			result = append(result, ev.RecurrenceMaster)
		} else {
			result = append(result, ev.Id)
		}
	}
	return
}

func (self *EventType) flushMemcache(c gaecontext.HTTPContext, dom *datastore.Key) *EventType {
	common.MemDel(c, eventTypesKeyForDomain(dom))
	common.MemDel(c, eventTypeKeyForId(self.Id))
	return self
}

func (self *EventType) Save(c gaecontext.HTTPContext, dom *datastore.Key) (result *EventType, err error) {

	self.SalarySerializedProperties, err = json.Marshal(self.SalaryProperties)
	if err != nil {
		panic(err)
	}

	self.ConfirmationEmailBodyTemplateBytes = []byte(self.ConfirmationEmailBodyTemplate)
	self.ConfirmationEmailSubjectTemplateBytes = []byte(self.ConfirmationEmailSubjectTemplate)

	if _, _, err = CreateConfirmationExample(c, self.ConfirmationEmailSubjectTemplate, self.ConfirmationEmailBodyTemplate); err != nil {
		return
	}

	self.Id, err = datastore.Put(c, datastore.NewKey(c, "EventType", self.Name, 0, dom), self)
	if err != nil {
		panic(err)
	}
	self.flushMemcache(c, dom)
	result = self.process(c)
	return
}

func DeleteEventType(c gaecontext.HTTPContext, key, dom *datastore.Key) {
	if !dom.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, key))
	}
	if err := datastore.Delete(c, key); err != nil {
		panic(err)
	}
	common.MemDel(c, eventTypesKeyForDomain(dom))
}

func findEventTypes(c gaecontext.HTTPContext, dom *datastore.Key) (result []EventType) {
	keys, err := datastore.NewQuery("EventType").Ancestor(dom).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetEventTypes(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer) (result []EventType) {
	var preResult []EventType
	common.Memoize(c, eventTypesKeyForDomain(dom), &preResult, func() interface{} {
		return findEventTypes(c, dom)
	})
	hasTypeAuth := authorizer == nil || authorizer.HasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	})
	authMatchEvents := auth.Auth{
		AuthType: auth.Events,
	}
	authMatchAttend := auth.Auth{
		AuthType: auth.Attend,
	}
	authMatchParticipants := auth.Auth{
		AuthType: auth.Participants,
	}
	authMatchAttest := auth.Auth{
		AuthType: auth.Attest,
	}
	authMatchReport := auth.Auth{
		AuthType: auth.ReportHours,
	}
	for index, typ := range preResult {
		authMatchEvents.EventType = typ.Id
		authMatchAttend.EventType = typ.Id
		authMatchParticipants.EventType = typ.Id
		authMatchAttest.EventType = typ.Id
		authMatchReport.EventType = typ.Id
		if authorizer == nil || (hasTypeAuth || authorizer.HasAnyAuth(authMatchParticipants) || authorizer.HasAnyAuth(authMatchAttend) || authorizer.HasAnyAuth(authMatchEvents) || authorizer.HasAnyAuth(authMatchAttest) || authorizer.HasAnyAuth(authMatchReport)) {
			result = append(result, *((&preResult[index]).process(c)))
		}
	}
	if result == nil {
		result = make([]EventType, 0)
	}
	return
}

type EventKind struct {
	Id             *datastore.Key `datastore:"-" json:"id"`
	Alert          bool           `json:"alert"`
	Block          bool           `json:"block"`
	Color          string         `json:"color"`
	Name           string         `json:"name"`
	SeriesEditable bool           `json:"series_editable"`

	// Salary mod
	SalarySerializedProperties []byte                 `json:"-"`
	SalaryProperties           map[string]interface{} `json:"salary_properties" datastore:"-"`
}

func (self *EventKind) process(c gaecontext.HTTPContext) *EventKind {
	if len(self.SalarySerializedProperties) > 0 {
		if err := json.Unmarshal(self.SalarySerializedProperties, &self.SalaryProperties); err != nil {
			panic(err)
		}
	}
	return self
}

func findEventKind(c gaecontext.HTTPContext, key *datastore.Key) *EventKind {
	var t EventKind
	if err := datastore.Get(c, key, &t); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil
		}
		panic(err)
	}
	t.Id = key
	return &t
}

func GetEventKind(c gaecontext.HTTPContext, key *datastore.Key) *EventKind {
	if key == nil {
		return nil
	}
	var t EventKind
	if common.Memoize(c, eventKindKeyForId(key), &t, func() interface{} {
		return findEventKind(c, key)
	}) {
		return (&t).process(c)
	}
	return nil
}

func (self *EventKind) Save(c gaecontext.HTTPContext, dom *datastore.Key) *EventKind {
	log.Infof(c, "gonna save %+v", self)
	var err error

	self.SalarySerializedProperties, err = json.Marshal(self.SalaryProperties)
	if err != nil {
		panic(err)
	}

	self.Id, err = datastore.Put(c, datastore.NewKey(c, "EventKind", self.Name, 0, dom), self)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, eventKindsKeyForDomain(dom))
	common.MemDel(c, eventKindKeyForId(self.Id))
	return self
}

func DeleteEventKind(c gaecontext.HTTPContext, key, dom *datastore.Key) {
	if !dom.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, key))
	}
	if err := datastore.Delete(c, key); err != nil {
		panic(err)
	}
	common.MemDel(c, eventKindsKeyForDomain(dom))
	common.MemDel(c, eventKindKeyForId(key))
}

func findEventKinds(c gaecontext.HTTPContext, dom *datastore.Key) (result []EventKind) {
	keys, err := datastore.NewQuery("EventKind").Ancestor(dom).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func getBlockEventKinds(c gaecontext.HTTPContext, dom *datastore.Key) (result []EventKind) {
	var preResult []EventKind
	common.Memoize(c, eventKindsKeyForDomain(dom), &preResult, func() interface{} {
		return findEventKinds(c, dom)
	})
	for _, typ := range preResult {
		if typ.Block {
			(&typ).process(c)
			result = append(result, typ)
		}
	}
	return
}

func getAlertEventKinds(c gaecontext.HTTPContext, dom *datastore.Key) (result []EventKind) {
	var preResult []EventKind
	common.Memoize(c, eventKindsKeyForDomain(dom), &preResult, func() interface{} {
		return findEventKinds(c, dom)
	})
	for _, typ := range preResult {
		if typ.Alert {
			(&typ).process(c)
			result = append(result, typ)
		}
	}
	return
}

func GetEventKinds(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer) (result []EventKind) {
	var preResult []EventKind
	common.Memoize(c, eventKindsKeyForDomain(dom), &preResult, func() interface{} {
		return findEventKinds(c, dom)
	})
	hasTypeAuth := authorizer == nil || authorizer.HasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	})
	authMatchEvents := auth.Auth{
		AuthType: auth.Events,
	}
	authMatchAttend := auth.Auth{
		AuthType: auth.Attend,
	}
	authMatchParticipants := auth.Auth{
		AuthType: auth.Participants,
	}
	authMatchAttest := auth.Auth{
		AuthType: auth.Attest,
	}
	authMatchReport := auth.Auth{
		AuthType: auth.ReportHours,
	}
	for _, typ := range preResult {
		authMatchEvents.EventKind = typ.Id
		authMatchAttend.EventKind = typ.Id
		authMatchParticipants.EventKind = typ.Id
		authMatchAttest.EventKind = typ.Id
		authMatchReport.EventKind = typ.Id
		if authorizer == nil || (hasTypeAuth || authorizer.HasAnyAuth(authMatchParticipants) || authorizer.HasAnyAuth(authMatchAttend) || authorizer.HasAnyAuth(authMatchEvents) || authorizer.HasAnyAuth(authMatchAttest) || authorizer.HasAnyAuth(authMatchReport)) {
			(&typ).process(c)
			result = append(result, typ)
		}
	}
	if result == nil {
		result = make([]EventKind, 0)
	}
	return
}
