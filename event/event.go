package event

import (
	"bytes"
	"fmt"
	"github.com/zond/schedrox/appuser"
	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/crm"
	"github.com/zond/schedrox/domain"
	"github.com/zond/schedrox/translation"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
)

var icsTemplates = template.Must(template.New("icsTemplates").ParseGlob("templates/ics/*.ics"))

func eventKeyForWeek(d *datastore.Key, week common.Week) string {
	return fmt.Sprintf("Events{Week:%v,Domain:%v}", week, d)
}

func eventKeyForId(d *datastore.Key) string {
	return fmt.Sprintf("Event{Id:%v}", d)
}

type EventWeek struct {
	Event *datastore.Key
}

func ConvertRecurrenceExceptions(c gaecontext.HTTPContext) (converted int) {
	events := []Event{}
	ids, err := datastore.NewQuery("Event").GetAll(c, &events)
	common.AssertOkError(err)
	for index, id := range ids {
		if events[index].RecurrenceExceptions != "" && len(events[index].RecurrenceExceptionsBytes) == 0 {
			events[index].RecurrenceExceptionsBytes = []byte(events[index].RecurrenceExceptions)
			events[index].RecurrenceExceptions = ""
			if _, err = datastore.Put(c, id, &events[index]); err != nil {
				panic(err)
			}
			converted++
		}
	}
	return
}

type Events []Event

func (self Events) Len() int {
	return len(self)
}

func (self Events) Less(a, b int) bool {
	if self[a].Start.Equal(self[b].Start) {
		if self[a].CreatedAt.Equal(self[b].CreatedAt) {
			return bytes.Compare([]byte(common.EncKey(self[a].Id)), []byte(common.EncKey(self[b].Id))) < 0
		}
		return self[a].CreatedAt.Before(self[b].CreatedAt)
	}
	return self[a].Start.Before(self[b].Start)
}

func (self Events) Swap(a, b int) {
	self[a], self[b] = self[b], self[a]
}

type Event struct {
	Id *datastore.Key `json:"id" datastore:"-"`

	// just plain comments
	Information      string `json:"information"`
	InformationBytes []byte `json:"-"`

	// interconnections and auth system
	Location  *datastore.Key `json:"location"`
	EventType *datastore.Key `json:"event_type"`
	EventKind *datastore.Key `json:"event_kind"`

	// core fullCalendar
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	AllDay bool      `json:"allDay"`
	Title  string    `json:"title"`

	// .ics stuff
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Caching of names for when stuff gets deleted (we can still fetch names from the ids)
	LocationName  string `json:"location_name,omitempty" datastore:"-"`
	EventTypeName string `json:"event_type_name,omitempty" datastore:"-"`
	EventKindName string `json:"event_kind_name,omitempty" datastore:"-"`

	// calendar colors and participant indications
	WantedUserParticipants   int `datastore:"-" json:"wanted_user_participants"`
	AllowedUserParticipants  int `datastore:"-" json:"allowed_user_participants"`
	RequiredUserParticipants int `datastore:"-" json:"required_user_participants"`
	UserParticipants         int `datastore:"-" json:"user_participants"`

	RequiredContactParticipants       int                       `datastore:"-" json:"required_contact_participants"`
	AllowedContactParticipants        int                       `datastore:"-" json:"allowed_contact_participants"`
	ContactParticipants               int                       `datastore:"-" json:"contact_participants"`
	Participants                      []Participant             `datastore:"-" json:"participants"`
	CurrentlyRequiredParticipantTypes []RequiredParticipantType `datastore:"-" json:"currently_required_participant_types"`

	// recurrence crap
	Recurring  bool   `json:"recurring"`
	Recurrence string `json:"recurrence,omitempty"`

	RecurrenceEnd time.Time `json:"recurrence_end,omitempty"`

	RecurrenceExceptions      string `json:"recurrence_exceptions,omitempty"`
	RecurrenceExceptionsBytes []byte `json:"-"`

	RecurrenceMaster      *datastore.Key `json:"recurrence_master,omitempty" datastore:"-"`
	RecurrenceMasterStart time.Time      `json:"recurrence_master_start,omitempty" datastore:"-"`
	RecurrenceMasterEnd   time.Time      `json:"recurrence_master_end,omitempty" datastore:"-"`

	// temporary in process caching of lots of fields, mainly to speed up calendar fetching through preProcess
	cachedParticipants                    []Participant
	cachedEventType                       *EventType
	cachedRequiredParticipantTypesForType []RequiredParticipantType
	cachedRequiredParticipantTypesForId   []RequiredParticipantType
	cachedParticipantTypes                map[string]*ParticipantType
	cachedRecurrenceExceptions            map[string]bool
	cachedRecurrenceParser                *recurrenceParser

	// utility for when fetching events for given participant
	ParticipantTypeName string `json:"participant_type_name,omitempty" datastore:"-"`

	// utility for when fetching events by contact
	ReportContact *crm.Contact   `json:"contact,omitempty" datastore:"-"`
	ReportedEvent *datastore.Key `json:"reported_event,omitempty" datastore:"-"`

	// utility for when salary mod
	SalaryAttestedParticipantType     *datastore.Key `json:"salary_attested_participant_type,omitempty"`
	SalaryAttestedParticipantTypeName string         `json:"salary_attested_participant_type_name,omitempty" datastore:"-"`

	SalaryAttestedEvent *datastore.Key `json:"salary_attested_event,omitempty"`
	SalaryAttestedUser  *datastore.Key `json:"salary_attested_user,omitempty"`
	SalaryAttester      *datastore.Key `json:"salary_attester,omitempty"`
	SalaryAttesterEmail string         `json:"salary_attester_string,omitempty" datastore:"-"`
	SalaryAttestedAt    time.Time      `json:"salary_attested_at,omitempty"`

	// to tell the client what kind of event this is
	SalaryTimeReported bool `json:"salary_time_reported,omitempty"`
}

func (self Event) AttestUUID() string {
	return fmt.Sprintf("%v:%v:%v", self.SalaryAttestedEvent.Encode(), self.SalaryAttestedUser.Encode(), self.UpdatedAt.UnixNano())
}

func (self *Event) getIcsAttachment(c gaecontext.HTTPContext, name, attendee string) []byte {
	replyTo := fmt.Sprintf("noreply@%v.appspotmail.com", appengine.AppID(c))
	dom := domain.GetDomain(c, self.Id.Parent())
	if dom.FromAddress != "" {
		replyTo = dom.FromAddress
	}
	loc := dom.GetLocation()
	location := ""
	if self.Location != nil {
		location = self.Location.StringID()
	}
	summary := self.Title
	if summary == "" && self.EventType != nil {
		summary = self.EventType.StringID()
	}
	if summary == "" && self.Information != "" {
		summary = self.Information
	}
	rrule := ""
	var exdates []time.Time
	if parser := theRecurrenceTypes.find(self.Recurrence); parser != nil {
		rrule = parser.rrule(self.RecurrenceEnd)
		for _, isoDate := range strings.Split(self.RecurrenceExceptions, ",") {
			if exDate, err := time.Parse(common.ISO8601Format, isoDate); err == nil {
				localExDate := time.Date(exDate.Year(), exDate.Month(), exDate.Day(), self.Start.Hour(), self.Start.Minute(), self.Start.Second(), self.Start.Nanosecond(), loc).UTC()
				exdates = append(exdates, localExDate)
			}
		}
	}
	start := time.Date(self.Start.Year(), self.Start.Month(), self.Start.Day(), self.Start.Hour(), self.Start.Minute(), self.Start.Second(), self.Start.Nanosecond(), loc).UTC()
	end := time.Date(self.End.Year(), self.End.Month(), self.End.Day(), self.End.Hour(), self.End.Minute(), self.End.Second(), self.End.Nanosecond(), loc).UTC()
	context := map[string]interface{}{
		"version":   appengine.VersionID(c),
		"start":     start,
		"end":       end,
		"now":       time.Now(),
		"id":        self.Id.Encode(),
		"createdAt": self.CreatedAt,
		"updatedAt": self.UpdatedAt,
		"location":  location,
		"sequence":  time.Now().UnixNano() / int64(time.Second),
		"summary":   summary,
		"organizer": replyTo,
		"attendee":  attendee,
		"recurring": self.Recurring,
		"rrule":     rrule,
		"exdates":   exdates,
	}
	buf := new(bytes.Buffer)
	if err := icsTemplates.ExecuteTemplate(buf, name, context); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func findEvent(c gaecontext.HTTPContext, key *datastore.Key) (result *Event) {
	var e Event
	err := datastore.Get(c, key, &e)
	if err == nil {
		result = &e
		result.Id = key
	} else {
		if _, ok := err.(*datastore.ErrFieldMismatch); ok {
			result = &e
			result.Id = key
		}
	}
	return
}

func (self Event) Minutes() int {
	return int(self.End.Sub(self.Start) / time.Minute)
}

func CurrentAlerts(c gaecontext.HTTPContext, dom *datastore.Key, at time.Time, asker *datastore.Key, authorizer auth.Authorizer) (result []Event) {
	result = make([]Event, 0)
	if dom == nil {
		return
	}
	alerting := make(map[string]bool)
	for _, kind := range getAlertEventKinds(c, dom) {
		alerting[common.EncKey(kind.Id)] = true
	}
	for _, event := range GetAllowedEventsBetween(c, asker, authorizer, dom, at, at, nil, nil, nil, nil) {
		if alerting[common.EncKey(event.EventKind)] && at.After(event.Start) && (at.Before(event.End) || (event.AllDay && at.Before(event.End.Add(time.Hour*24)))) {
			result = append(result, event)
		}
	}
	return
}

func GetEvent(c gaecontext.HTTPContext, k *datastore.Key, dom *datastore.Key) *Event {
	if !k.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, k))
	}
	var event Event
	if common.Memoize(c, eventKeyForId(k), &event, func() interface{} {
		return findEvent(c, k)
	}) {
		if event.Recurring {
			(&event).RecurrenceMaster = event.Id
			(&event).RecurrenceMasterStart = event.Start
			(&event).RecurrenceMasterEnd = event.End
		}
		return (&event).process(c)
	}
	return nil
}

func preProcess(c gaecontext.HTTPContext, events []Event) {
	if len(events) == 0 {
		return
	}

	// find unique event types to fetch
	typeIdMap := make(map[string]*datastore.Key)
	for index, event := range events {
		(&events[index]).QuickProcess(c)
		typeIdMap[common.EncKey(event.EventType)] = event.EventType
	}

	// request all unique participant types and present event types and their required particiapnts, and participants + required participants for each event
	cacheKeys := make([]string, 0, 1+len(typeIdMap)*2+len(events)*2)
	funcs := make([]func() interface{}, 0, 1+len(typeIdMap)*2+len(events)*2)
	values := make([]interface{}, 0, 1+len(typeIdMap)*2+len(events)*2)

	// build the event type fetching
	// build the required participants for event type bits
	for _, typeId := range typeIdMap {
		cacheKeys = append(cacheKeys, eventTypeKeyForId(typeId))
		cacheKeys = append(cacheKeys, requiredParticipantTypesKeyForParent(typeId))
		idCopy := typeId
		funcs = append(funcs, func() interface{} {
			return findEventType(c, idCopy)
		})
		funcs = append(funcs, func() interface{} {
			return findRequiredParticipantTypes(c, idCopy)
		})
		var typ EventType
		values = append(values, &typ)
		var reqs []RequiredParticipantType
		values = append(values, &reqs)
	}
	// build the participants bits
	// build the required participants for unique events bits
	for _, event := range events {
		cacheKeys = append(cacheKeys, participantsKeyForEvent(event.Id))
		cacheKeys = append(cacheKeys, requiredParticipantTypesKeyForParent(event.Id))
		eventCopy := event
		funcs = append(funcs, func() interface{} {
			return eventCopy.findParticipants(c)
		})
		funcs = append(funcs, func() interface{} {
			return findRequiredParticipantTypes(c, eventCopy.Id)
		})
		var parts []Participant
		values = append(values, &parts)
		var reqs []RequiredParticipantType
		values = append(values, &reqs)
	}
	// participant types
	cacheKeys = append(cacheKeys, participantTypesKeyForDomain(events[0].Id.Parent()))
	funcs = append(funcs, func() interface{} {
		return findParticipantTypes(c, events[0].Id.Parent())
	})
	var participantTypes []ParticipantType
	values = append(values, &participantTypes)

	// fetch
	exists := common.MemoizeMulti(c, cacheKeys, values, funcs)

	participantTypeMap := make(map[string]*ParticipantType)
	for index, participantType := range participantTypes {
		participantTypeMap[participantType.Id.Encode()] = &participantTypes[index]
	}

	// build a map of fetched types
	typeMap := make(map[string]*EventType)
	typeReqMap := make(map[string][]RequiredParticipantType)
	i := 0
	var typ *EventType
	var typReqs *[]RequiredParticipantType
	for i < len(typeIdMap) {
		if exists[i*2] && values[i*2] != nil {
			typ = values[i*2].(*EventType)
			typeMap[common.EncKey(typ.Id)] = typ
			if exists[i*2+1] && values[i*2+1] != nil {
				typReqs = values[i*2+1].(*[]RequiredParticipantType)
				typeReqMap[common.EncKey(typ.Id)] = *typReqs
			}
		}
		i += 1
	}

	i *= 2

	userIdMap := make(map[string]*datastore.Key)

	// set cached values for each event
	for index, _ := range events {
		events[index].cachedParticipantTypes = participantTypeMap
		events[index].cachedEventType = typeMap[common.EncKey(events[index].EventType)]
		events[index].cachedRequiredParticipantTypesForType = typeReqMap[common.EncKey(events[index].EventType)]
		events[index].cachedParticipants = *(values[(index*2)+i].(*[]Participant))
		if events[index].cachedEventType != nil {
			for _, part := range events[index].cachedParticipants {
				if part.User != nil {
					idCopy := part.User
					userIdMap[part.User.Encode()] = idCopy
				}
			}
		}
		events[index].cachedRequiredParticipantTypesForId = *(values[(index*2+1)+i].(*[]RequiredParticipantType))
	}

	if len(userIdMap) > 0 {
		values = nil
		funcs = nil
		cacheKeys = nil

		// build the user fetching bits
		for _, userId := range userIdMap {
			cacheKeys = append(cacheKeys, appuser.UserKeyForId(userId))
			idCopy := userId
			funcs = append(funcs, func() interface{} {
				return appuser.FindUser(c, idCopy)
			})
			var user appuser.User
			values = append(values, &user)
		}

		// fetch
		exists = common.MemoizeMulti(c, cacheKeys, values, funcs)

		// build a map of fetched users
		userMap := make(map[string]*appuser.User)
		i = 0
		var user *appuser.User
		for i < len(userIdMap) {
			if exists[i] {
				user = values[i].(*appuser.User)
				userMap[common.EncKey(user.Id)] = user
			}
			i += 1
		}

		// apply the fetched users
		for index, _ := range events {
			if events[index].cachedEventType != nil {
				for index2, _ := range events[index].cachedParticipants {
					if events[index].cachedParticipants[index2].User != nil {
						if user, ok := userMap[events[index].cachedParticipants[index2].User.Encode()]; ok {
							(&events[index].cachedParticipants[index2]).apply(user)
						}
					}
				}
			}
		}
	}
}

func (self *Event) QuickProcess(c gaecontext.HTTPContext) *Event {
	if self.EventKind != nil {
		self.EventKindName = self.EventKind.StringID()
	}
	if self.EventType != nil {
		self.EventTypeName = self.EventType.StringID()
	}
	if self.Location != nil {
		self.LocationName = self.Location.StringID()
	}
	if self.SalaryAttestedParticipantType != nil {
		self.SalaryAttestedParticipantTypeName = self.SalaryAttestedParticipantType.StringID()
	}
	self.RecurrenceExceptions = string(self.RecurrenceExceptionsBytes)
	if len(self.InformationBytes) > 0 {
		self.Information = string(self.InformationBytes)
	}
	return self
}

func (self *Event) process(c gaecontext.HTTPContext) *Event {
	if eventType := self.GetEventType(c); eventType != nil {
		self.Participants = self.GetParticipants(c)
		self.EventKind = eventType.EventKind
		participantsByType, currentlyRequiredTypes := self.GetParticipantsAndRequired(c)
		self.CurrentlyRequiredParticipantTypes = currentlyRequiredTypes
		self.UserParticipants = 0
		self.WantedUserParticipants = 0
		self.ContactParticipants = 0
		self.AllowedContactParticipants = 0
		self.RequiredContactParticipants = 0
		var current int
		for _, t := range currentlyRequiredTypes {
			if typ := self.GetParticipantType(c, t.ParticipantType, t.ParticipantType.Parent()); typ != nil {
				current = 0
				for _, participant := range participantsByType[common.EncKey(t.ParticipantType)] {
					current += participant.Multiple
				}
				if typ.IsContact {
					self.ContactParticipants += current
					self.AllowedContactParticipants += t.Max
					self.RequiredContactParticipants += t.Min
				} else {
					self.UserParticipants += current
					self.AllowedUserParticipants += t.Max
					self.RequiredUserParticipants += t.Min
					if current < t.Min {
						self.WantedUserParticipants += t.Min
					} else if current > t.Max {
						self.WantedUserParticipants += t.Max
					} else {
						self.WantedUserParticipants += current
					}
				}
			}
		}
	}
	self.QuickProcess(c)
	return self
}

func (self *Event) Delete(c gaecontext.HTTPContext, dom, actor *datastore.Key) {
	if !self.Id.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, self.Id))
	}
	for _, part := range self.GetParticipants(c) {
		part.Delete(c, actor, self)
	}
	if err := c.Transaction(func(c gaecontext.HTTPContext) error {
		CreateChange(c, self.Id, actor, "DeleteEvent", self)
		self.UpdatedAt = time.Now()
		for _, req := range GetRequiredParticipantTypes(c, self.Id, dom) {
			req.Delete(c, self.Id, dom, actor)
		}
		self.clearWeeks(c, false)
		err := datastore.Delete(c, self.Id)
		if err != nil {
			panic(err)
		}
		common.MemDel(c, eventKeyForId(self.Id))
		common.MemDel(c, eventTypeKeyForId(self.Id))
		return nil
	}, true); err != nil {
		panic(err)
	}
}

func (self *Event) eventWeekKey(c gaecontext.HTTPContext, week common.Week) *datastore.Key {
	return datastore.NewKey(c, "EventWeek", fmt.Sprintf("%v", common.EncKey(self.Id)), 0, datastore.NewKey(c, "Week", fmt.Sprintf("%v", week), 0, self.Id.Parent()))
}

func (self *Event) clearWeeks(c gaecontext.HTTPContext, cacheOnly bool) {
	var eventWeekKeys []*datastore.Key
	var cacheKeys []string
	var err error
	to := self.End
	if self.Recurring {
		to = self.RecurrenceEnd
	}
	for _, week := range common.WeeksBetween(self.Start, to) {
		if !cacheOnly {
			eventWeekKeys = append(eventWeekKeys, self.eventWeekKey(c, week))
		}
		cacheKeys = append(cacheKeys, eventKeyForWeek(self.Id.Parent(), week))
	}
	common.MemDel(c, cacheKeys...)
	if !cacheOnly {
		err = datastore.DeleteMulti(c, eventWeekKeys)
		if err != nil {
			panic(err)
		}
	}
}

func (self *Event) saveWeeks(c gaecontext.HTTPContext) {
	var eventWeekKeys []*datastore.Key
	var cacheKeys []string
	var eventWeeks []EventWeek
	to := self.End
	if self.Recurring {
		to = self.RecurrenceEnd
	}
	for _, week := range common.WeeksBetween(self.Start, to) {
		eventWeekKeys = append(eventWeekKeys, self.eventWeekKey(c, week))
		eventWeeks = append(eventWeeks, EventWeek{
			Event: self.Id,
		})
		cacheKeys = append(cacheKeys, eventKeyForWeek(self.Id.Parent(), week))
	}
	common.MemDel(c, cacheKeys...)
	_, err := datastore.PutMulti(c, eventWeekKeys, eventWeeks)
	if err != nil {
		panic(err)
	}
}

func (self *Event) NotifyEqual(o *Event) bool {
	return (self.Start.Equal(o.Start) &&
		self.End.Equal(o.End) &&
		self.AllDay == o.AllDay &&
		self.Recurring == o.Recurring &&
		self.Recurrence == o.Recurrence &&
		self.RecurrenceEnd.Equal(o.RecurrenceEnd) &&
		self.RecurrenceExceptions == o.RecurrenceExceptions &&
		((self.EventType != nil && self.EventType.Equal(o.EventType)) || (self.EventType == nil && o.EventType == nil)) &&
		((self.Location != nil && self.Location.Equal(o.Location)) || (self.Location == nil && o.Location == nil)))
}

func (self *Event) Save(c gaecontext.HTTPContext, dom, actor *datastore.Key) *Event {
	if self.Id != nil && !self.Id.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, self.Id))
	}
	postTrans := []func(gaecontext.HTTPContext){}
	if err := c.Transaction(func(c gaecontext.HTTPContext) error {
		var old *Event
		sameSpan := false
		if self.Id == nil {
			self.Id = datastore.NewKey(c, "Event", "", 0, dom)
			self.CreatedAt = time.Now()
		} else {
			CreateChange(c, self.Id, actor, "UpdateEvent", self)
			old = GetEvent(c, self.Id, dom)
			if old != nil {
				if !old.NotifyEqual(self) {
					for _, participant := range old.GetParticipants(c) {
						if participant.User != nil {
							oldCpy := *old
							selfCpy := *self
							partCpy := participant
							postTrans = append(postTrans, func(c gaecontext.HTTPContext) {
								if user := appuser.GetUserByKey(c, partCpy.User); user != nil && !user.MuteEventNotifications {
									var theDomain *domain.Domain
									for _, d := range user.Domains {
										if d.Id.Equal(dom) {
											theDomain = &d
											break
										}
									}
									var attachment *mail.Attachment
									if theDomain.AllowICS || !theDomain.LimitedICS {
										attachment = &mail.Attachment{
											Name: "invitation.ics",
											Data: self.getIcsAttachment(c, "invitation.ics", user.Email),
										}
									}
									user.SendMail(c, dom, translation.GetTranslation(user.LastLanguage, "Event updated"), translation.GetEmailBody(user.LastLanguage, "event_updated.txt", map[string]interface{}{
										"old_event":   &oldCpy,
										"new_event":   &selfCpy,
										"participant": partCpy,
										"AppID":       appengine.AppID(c),
									}), "", attachment)
								}
							})
						}
					}
				}
				sameSpan = old.Start.Equal(self.Start) && old.End.Equal(self.End) && old.Recurring == self.Recurring && old.RecurrenceEnd == self.RecurrenceEnd
				old.clearWeeks(c, sameSpan)
			}
		}
		self.UpdatedAt = time.Now()
		self.RecurrenceExceptionsBytes = []byte(self.RecurrenceExceptions)
		self.RecurrenceExceptions = ""
		self.InformationBytes = []byte(self.Information)
		self.Information = ""
		if self.AllDay {
			sy, sm, sd := self.Start.Date()
			ey, em, ed := self.End.Date()
			self.Start = time.Date(sy, sm, sd, 0, 0, 0, 0, self.Start.Location())
			self.End = time.Date(ey, em, ed, 23, 59, 59, 0, self.End.Location())
		}
		var err error
		self.Id, err = datastore.Put(c, self.Id, self)
		if err != nil {
			panic(err)
		}
		self.Information = string(self.InformationBytes)
		self.RecurrenceExceptions = string(self.RecurrenceExceptionsBytes)
		common.MemDel(c, eventTypeKeyForId(self.Id))
		common.MemDel(c, eventKeyForId(self.Id))
		if old == nil {
			CreateChange(c, self.Id, actor, "CreateEvent", self)
		}
		if old == nil || !sameSpan {
			self.saveWeeks(c)
		}
		return nil
	}, true); err != nil {
		panic(err)
	}
	for _, post := range postTrans {
		post(c)
	}
	return self.process(c)
}

func findForWeek(c gaecontext.HTTPContext, d *datastore.Key, week common.Week) (result []Event) {
	var eventKeys []*datastore.Key
	var eventWeeks []EventWeek
	_, err := datastore.NewQuery("EventWeek").Ancestor(datastore.NewKey(c, "Week", fmt.Sprintf("%v", week), 0, d)).GetAll(c, &eventWeeks)
	if err != nil {
		panic(err)
	}
	for _, eventWeek := range eventWeeks {
		eventKeys = append(eventKeys, eventWeek.Event)
	}
	result = make([]Event, len(eventKeys))
	err = datastore.GetMulti(c, eventKeys, result)
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for index, serr := range merr {
				if serr != nil {
					if _, ok := serr.(*datastore.ErrFieldMismatch); !ok {
						if serr != datastore.ErrNoSuchEntity {
							panic(fmt.Errorf("Error doing GetMulti for %v: %v", eventKeys[index], serr))
						} else {
							log.Infof(c, "### %v (%v) is in EventWeek but doesn't have an event!", eventKeys[index], eventKeys[index].Encode())
						}
					}
				}
			}
		} else {
			panic(err)
		}
	}
	for index, _ := range result {
		result[index].Id = eventKeys[index]
	}
	return
}

func getForWeek(c gaecontext.HTTPContext, d *datastore.Key, week common.Week, funnel chan Event, done1 chan bool) (done2 chan bool) {
	done2 = make(chan bool)
	go func() {
		var events []Event
		common.Memoize(c, eventKeyForWeek(d, week), &events, func() interface{} {
			return findForWeek(c, d, week)
		})
		for _, event := range events {
			funnel <- event
		}
		if done1 != nil {
			<-done1
		}
		done2 <- true
	}()
	return
}

func (self Event) recurrenceClone() (result Event) {
	result = self
	return
}

func (self *Event) AddRecurrenceException(c gaecontext.HTTPContext, dom *datastore.Key, start, end time.Time, authorizer auth.Authorizer, actor *datastore.Key, copyParticipants bool) *Event {
	exceptions := strings.Split(self.RecurrenceExceptions, ",")
	exceptions = append(exceptions, start.Format(common.ISO8601Format))
	self.RecurrenceExceptions = strings.Join(exceptions, ",")
	self.Save(c, dom, actor)

	return self.cpy(c, dom, false, "", authorizer, actor, nil, start, end, copyParticipants)
}

func (self *Event) SplitAndRemoveParticipant(c gaecontext.HTTPContext, dom *datastore.Key, at time.Time, actor, toRemove *datastore.Key, authorizer auth.Authorizer) *Event {
	clone := self.recurrenceClone()

	self.RecurrenceEnd = at
	self.Save(c, dom, actor)

	if nextOcc := clone.nextOccurence(c, clone.Start, at); nextOcc != nil {
		return clone.cpy(c, dom, clone.Recurring, clone.Recurrence, authorizer, actor, toRemove, nextOcc.Start, nextOcc.End, false)
	}
	return nil
}

func (self *Event) cpy(c gaecontext.HTTPContext, dom *datastore.Key, recurring bool, recurrence string, authorizer auth.Authorizer, actor *datastore.Key, participantToSkip *datastore.Key, start, end time.Time, copyParticipants bool) *Event {
	newEvent := *self
	newEvent.Id = nil
	newEvent.RecurrenceMaster = nil
	newEvent.Recurring = recurring
	newEvent.Recurrence = recurrence
	newEvent.Start = start
	newEvent.End = end

	(&newEvent).process(c).Save(c, dom, actor)

	if copyParticipants {
		for _, participant := range self.GetParticipants(c) {
			if !participant.Id.Equal(participantToSkip) {

				if authorizer == nil || !authorizer.HasAuth(auth.Auth{
					AuthType:        auth.Participants,
					EventKind:       self.EventKind,
					EventType:       self.EventType,
					ParticipantType: participant.ParticipantType,
					Write:           true,
				}) {
					participant.Id = nil
					participant.Save(c, &newEvent, dom, actor)
				}
			}
		}

		for _, req := range GetRequiredParticipantTypes(c, self.Id, dom) {
			if authorizer == nil || !authorizer.HasAuth(auth.Auth{
				AuthType:        auth.Participants,
				EventKind:       self.EventKind,
				EventType:       self.EventType,
				ParticipantType: req.ParticipantType,
				Write:           true,
			}) {
				req.Id = nil
				req.Save(c, newEvent.Id, dom, actor)
			}
		}
	}

	return (&newEvent).process(c)
}

func (self Event) ExpandRecurrences(c gaecontext.HTTPContext, dom *datastore.Key, start, end time.Time) (result []Event) {
	originalStart := self.Start
	self.RecurrenceMasterStart = self.Start
	self.RecurrenceMasterEnd = self.End

	if !self.Start.After(end) && !self.End.Before(start) && !(&self).getRecurrenceExceptions()[self.Start.Format(common.ISO8601Format)] {
		result = append(result, self.recurrenceClone())
	}
	self.Start = self.Start.AddDate(0, 0, 1)
	self.End = self.End.AddDate(0, 0, 1)

	for nextOcc := (&self).nextOccurence(c, originalStart, start); nextOcc != nil && !nextOcc.Start.After(end); nextOcc = (&self).nextOccurence(c, originalStart, start) {
		result = append(result, *nextOcc)
	}
	return
}

func (self *Event) getRecurrenceParser() *recurrenceParser {
	if self.cachedRecurrenceParser == nil {
		self.cachedRecurrenceParser = theRecurrenceTypes.find(self.Recurrence)
	}
	return self.cachedRecurrenceParser
}

func (self *Event) getRecurrenceExceptions() map[string]bool {
	if self.cachedRecurrenceExceptions == nil {
		self.cachedRecurrenceExceptions = make(map[string]bool)
		for _, exception := range strings.Split(self.RecurrenceExceptions, ",") {
			self.cachedRecurrenceExceptions[exception] = true
		}
	}
	return self.cachedRecurrenceExceptions
}

func (self *Event) nextOccurence(c gaecontext.HTTPContext, originalStart, start time.Time) (result *Event) {
	if parser := self.getRecurrenceParser(); parser != nil {
		for !self.Start.After(self.RecurrenceEnd) {
			if !self.End.Before(start) {
				if parser.matchesTime(originalStart, self.Start) {
					if !self.getRecurrenceExceptions()[self.Start.Format(common.ISO8601Format)] {
						clone := self.recurrenceClone()
						self.Start = self.Start.AddDate(0, 0, 1)
						self.End = self.End.AddDate(0, 0, 1)
						return &clone
					}
				}
			}
			self.Start = self.Start.AddDate(0, 0, 1)
			self.End = self.End.AddDate(0, 0, 1)
		}
	}
	return nil
}

func GetEventsBetween(c gaecontext.HTTPContext, d *datastore.Key, start, end time.Time, preFilter func(ev *Event) Events) (result Events) {
	return GetFilteredEventsBetween(c, d, start, end, preFilter, nil)
}

func GetMyEventsBetween(c gaecontext.HTTPContext, d *datastore.Key, user *datastore.Key, start, end time.Time, locations, kinds, types []string) (result Events) {
	allowedLocations := make(map[string]bool)
	allowedKinds := make(map[string]bool)
	allowedTypes := make(map[string]bool)
	for _, s := range locations {
		allowedLocations[s] = true
	}
	for _, s := range kinds {
		allowedKinds[s] = true
	}
	for _, s := range types {
		allowedTypes[s] = true
	}
	return GetFilteredEventsBetween(c, d, start, end, func(ev *Event) (result Events) {
		if (len(locations) == 0 || ev.Location == nil || allowedLocations[ev.Location.Encode()]) &&
			(len(kinds) == 0 || ev.EventKind == nil || allowedKinds[ev.EventKind.Encode()]) &&
			(len(types) == 0 || ev.EventType == nil || allowedTypes[ev.EventType.Encode()]) {
			result = Events{*ev}
		}
		return
	}, func(ev *Event) (result Events) {
		participants := ev.GetParticipants(c)
		allowed := false
		for _, participant := range participants {
			if participant.User.Equal(user) {
				allowed = true
				break
			}
		}
		if allowed {
			result = Events{*ev}
		}
		return
	})
}

func GetOpenEventsBetween(c gaecontext.HTTPContext, d *datastore.Key, user *datastore.Key, start, end time.Time, authorizer auth.Authorizer, locations, kinds, types []string) (result Events) {
	userTypes := map[string]struct{}{}
	for _, typ := range GetParticipantTypes(c, d, nil) {
		if !typ.IsContact {
			userTypes[typ.Id.Encode()] = struct{}{}
		}
	}
	allowedLocations := make(map[string]bool)
	allowedKinds := make(map[string]bool)
	allowedTypes := make(map[string]bool)
	for _, s := range locations {
		allowedLocations[s] = true
	}
	for _, s := range kinds {
		allowedKinds[s] = true
	}
	for _, s := range types {
		allowedTypes[s] = true
	}
	preResult := GetFilteredEventsBetween(c, d, start, end, func(ev *Event) (result Events) {
		if (len(locations) == 0 || ev.Location == nil || allowedLocations[ev.Location.Encode()]) &&
			(len(kinds) == 0 || ev.EventKind == nil || allowedKinds[ev.EventKind.Encode()]) &&
			(len(types) == 0 || ev.EventType == nil || allowedTypes[ev.EventType.Encode()]) {
			if authorizer.HasAnyAuth(auth.Auth{
				AuthType:  auth.Attend,
				EventKind: ev.EventKind,
				EventType: ev.EventType,
				Location:  ev.Location,
			}) {
				result = Events{*ev}
			}
		}
		return
	}, func(ev *Event) (result Events) {
		allowed := true
		for _, participant := range ev.GetParticipants(c) {
			if participant.User.Equal(user) {
				allowed = false
				break
			}
		}
		if allowed {
			baseAuth := auth.Auth{
				AuthType:  auth.Attend,
				EventKind: ev.EventKind,
				EventType: ev.EventType,
				Location:  ev.Location,
			}
			participantsByType, requiredTypes := ev.GetParticipantsAndRequired(c)
			allowed = false
			for _, requiredType := range requiredTypes {
				if _, found := userTypes[requiredType.ParticipantType.Encode()]; found {
					allowed = false
					currentNumber := 0
					for _, participant := range participantsByType[requiredType.ParticipantType.Encode()] {
						currentNumber += participant.Multiple
					}
					if currentNumber < requiredType.Max {
						baseAuth.ParticipantType = requiredType.ParticipantType
						if authorizer.HasAuth(baseAuth) {
							allowed = true
							break
						}
					}
				}
			}
		}
		if allowed {
			result = Events{*ev}
		}
		return
	})
	busyTimes := GetBusyTimes(c, start, end, user)
	for _, event := range preResult {
		ok := true
		for _, times := range busyTimes {
			if common.Overlaps([2]time.Time{event.Start, event.End}, times) {
				ok = false
				break
			}
		}
		if ok {
			result = append(result, event)
		}
	}
	return
}

func GetFilteredEventsBetween(c gaecontext.HTTPContext, d *datastore.Key, start, end time.Time, preFilter, postFilter func(ev *Event) Events) (result Events) {
	collector := make(map[string]Event)
	funnel := make(chan Event)
	var fetchDone chan bool
	collectDone := make(chan bool)
	go func() {
		for event := range funnel {
			collector[common.EncKey(event.Id)] = event
		}
		collectDone <- true
	}()
	for _, week := range common.WeeksBetween(start, end) {
		fetchDone = getForWeek(c, d, week, funnel, fetchDone)
	}
	if fetchDone != nil {
		<-fetchDone
	}
	close(funnel)
	<-collectDone
	preResult := make([]Event, 0, len(collector))
	for _, event := range collector {
		if event.Recurring || (!event.Start.After(end) && !event.End.Before(start)) {
			if preFilter == nil {
				preResult = append(preResult, event)
			} else {
				for _, filtered := range preFilter(&event) {
					preResult = append(preResult, filtered)
				}
			}
		}
	}
	preProcess(c, preResult)
	if postFilter != nil {
		filtered := make([]Event, 0, len(preResult))
		for _, event := range preResult {
			for _, filteredEvent := range postFilter(&event) {
				filtered = append(filtered, filteredEvent)
			}
		}
		preResult = filtered
	}
	for _, event := range preResult {
		if event.Recurring {
			for _, generatedEvent := range event.ExpandRecurrences(c, d, start, end) {
				(&generatedEvent).process(c)
				generatedEvent.RecurrenceMaster = event.Id
				generatedEvent.Id = datastore.NewKey(c, "Event", fmt.Sprintf("%v.%v", event.Id, generatedEvent.Start.Unix()), 0, d)
				result = append(result, generatedEvent)
			}
		} else {
			(&event).process(c)
			result = append(result, event)
		}
	}
	sort.Sort(result)
	if result == nil {
		result = Events{}
	}
	return
}

func GetReportableEventsBetween(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer, from, to time.Time) (result Events) {
	requiredAuth := auth.Auth{
		AuthType: auth.SalaryReport,
	}
	ok := false
	return GetFilteredEventsBetween(c, dom, from, to, func(ev *Event) (result Events) {
		requiredAuth.Location = ev.Location
		requiredAuth.EventKind = ev.EventKind
		requiredAuth.EventType = ev.EventType
		if authorizer == nil || authorizer.HasAnyAuth(requiredAuth) {
			result = append(result, *ev)
		}
		return
	}, func(ev *Event) (result Events) {
		requiredAuth.Location = ev.Location
		requiredAuth.EventKind = ev.EventKind
		requiredAuth.EventType = ev.EventType
		ok = false
		for _, thisPart := range ev.GetParticipants(c) {
			if thisPart.User != nil {
				requiredAuth.ParticipantType = thisPart.ParticipantType
				if authorizer == nil || authorizer.HasAuth(requiredAuth) {
					ok = true
					break
				}
			}
		}
		if ok {
			cpy := *ev
			result = append(result, cpy)
		}
		return
	})
}

func GetUnpaidEventsBetween(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer, from, to time.Time) (result Events) {
	requiredEventAuth := auth.Auth{
		AuthType: auth.Events,
	}
	requiredPartAuth := auth.Auth{
		AuthType: auth.Participants,
	}
	preResult := GetFilteredEventsBetween(c, dom, from, to, func(ev *Event) (result Events) {
		if !ev.Recurring {
			requiredEventAuth.Location, requiredPartAuth.Location = ev.Location, ev.Location
			requiredEventAuth.EventKind, requiredPartAuth.EventKind = ev.EventKind, ev.EventKind
			requiredEventAuth.EventType, requiredPartAuth.EventType = ev.EventType, ev.EventType
			if authorizer == nil || (authorizer.HasAnyAuth(requiredEventAuth) && authorizer.HasAnyAuth(requiredPartAuth)) {
				result = append(result, *ev)
			}
		}
		return
	}, nil)
	for _, ev := range preResult {
		for _, thisPart := range ev.GetParticipants(c) {
			if thisPart.Contact != nil && !thisPart.Paid {
				cpy := ev
				cpy.ReportContact = thisPart.GetContact(c)
				cpy.Participants = []Participant{thisPart}
				cpy.ReportedEvent = cpy.Id
				cpy.Id = datastore.NewKey(c, "Event", thisPart.Id.Encode(), 0, cpy.Id)
				requiredEventAuth.Location, requiredPartAuth.Location = cpy.Location, cpy.Location
				requiredEventAuth.EventKind, requiredPartAuth.EventKind = cpy.EventKind, cpy.EventKind
				requiredEventAuth.EventType, requiredPartAuth.EventType = cpy.EventType, cpy.EventType
				requiredPartAuth.ParticipantType = thisPart.ParticipantType
				if authorizer == nil || (authorizer.HasAuth(requiredEventAuth) && authorizer.HasAuth(requiredPartAuth)) {
					result = append(result, cpy)
				}
			}
		}
	}
	return
}

func GetAttestableEventsForUserBetween(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer, user *datastore.Key, from, to time.Time, onlyAttestable bool) (result Events) {
	requiredAuth := auth.Auth{
		AuthType: auth.Attest,
	}
	result = GetFilteredEventsBetween(c, dom, from, to, func(ev *Event) (result Events) {
		requiredAuth.Location = ev.Location
		requiredAuth.EventKind = ev.EventKind
		requiredAuth.EventType = ev.EventType
		viewAuth := requiredAuth
		viewAuth.AuthType = auth.Events
		if authorizer == nil || authorizer.HasAnyAuth(requiredAuth) || (!onlyAttestable && authorizer.HasAnyAuth(viewAuth)) {
			result = append(result, *ev)
		}
		return
	}, func(ev *Event) (result Events) {
		for _, thisPart := range ev.GetParticipants(c) {
			if thisPart.User != nil && !thisPart.Defaulted && (user == nil || user.Equal(thisPart.User)) {
				cpy := *ev
				cpy.SalaryAttestedParticipantType = thisPart.ParticipantType
				cpy.SalaryAttestedParticipantTypeName = thisPart.ParticipantType.StringID()
				cpy.SalaryAttestedEvent = ev.Id
				cpy.SalaryAttestedUser = thisPart.User
				requiredAuth.Location = cpy.Location
				requiredAuth.EventKind = cpy.EventKind
				requiredAuth.EventType = cpy.EventType
				requiredAuth.ParticipantType = thisPart.ParticipantType
				viewAuth := requiredAuth
				viewAuth.AuthType = auth.Events
				viewAuth.ParticipantType = nil
				if authorizer == nil || authorizer.HasAuth(requiredAuth) || (!onlyAttestable && authorizer.HasAuth(viewAuth)) {
					result = append(result, cpy)
				}
			}
		}
		return
	})
	for index, _ := range result {
		(&result[index]).SalaryAttestedEvent = result[index].Id
	}
	return
}

func fixAuths(ev *Event, auths ...*auth.Auth) {
	for index, _ := range auths {
		auths[index].Location = ev.Location
		auths[index].EventKind = ev.EventKind
		auths[index].EventType = ev.EventType
	}
}

func GetAllowedEventsBetween(c gaecontext.HTTPContext, asker *datastore.Key, authorizer auth.Authorizer, d *datastore.Key, start, end time.Time, locations, kinds, types, users []string) (result []Event) {
	blocking := make(map[string]bool)
	for _, kind := range getBlockEventKinds(c, d) {
		blocking[common.EncKey(kind.Id)] = true
	}
	allowedLocations := make(map[string]bool)
	allowedKinds := make(map[string]bool)
	allowedTypes := make(map[string]bool)
	allowedUsers := make(map[string]bool)
	for _, s := range users {
		allowedUsers[s] = true
	}
	for _, s := range locations {
		allowedLocations[s] = true
	}
	for _, s := range kinds {
		allowedKinds[s] = true
	}
	for _, s := range types {
		allowedTypes[s] = true
	}
	eventsAuth := auth.Auth{
		AuthType: auth.Events,
	}
	attendAuth := auth.Auth{
		AuthType: auth.Attend,
	}
	partAuth := auth.Auth{
		AuthType: auth.Participants,
	}
	attestAuth := auth.Auth{
		AuthType: auth.Attest,
	}
	reportAuth := auth.Auth{
		AuthType: auth.SalaryReport,
	}
	return GetFilteredEventsBetween(c, d, start, end, func(ev *Event) Events {
		fixAuths(ev, &eventsAuth, &attendAuth, &partAuth, &attestAuth, &reportAuth)
		if blocking[common.EncKey(ev.EventKind)] ||
			((authorizer.HasAuth(eventsAuth) || authorizer.HasAnyAuth(attendAuth) || authorizer.HasAnyAuth(partAuth) || authorizer.HasAnyAuth(attestAuth) || authorizer.HasAnyAuth(reportAuth) || ev.IsParticipant(c, asker)) &&
				(len(locations) == 0 || ev.Location == nil || allowedLocations[ev.Location.Encode()]) &&
				(len(kinds) == 0 || ev.EventKind == nil || allowedKinds[ev.EventKind.Encode()]) &&
				(len(types) == 0 || ev.EventType == nil || allowedTypes[ev.EventType.Encode()])) {
			return Events{*ev}
		}
		return nil
	}, func(ev *Event) Events {
		if len(users) == 0 {
			return Events{*ev}
		}
		for _, part := range ev.GetParticipants(c) {
			if part.User != nil && allowedUsers[part.User.Encode()] {
				return Events{*ev}
			}
		}
		return nil
	})
}
