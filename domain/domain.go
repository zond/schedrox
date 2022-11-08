package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"monotone/se.oort.schedrox/auth"
	"monotone/se.oort.schedrox/common"
	"time"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
)

const (
	allDomainsKey    = "Domains"
	userPropertyKind = "UserPropertyForDomain"
)

func domainKeyForId(id *datastore.Key) string {
	return fmt.Sprintf("Domain{Id:%v}", id)
}

func userPropertiesKeyForDomain(d *datastore.Key) string {
	return fmt.Sprintf("UserPropertiesForDomain{Domain:%v}", d)
}

func locationsKeyForDomain(d *datastore.Key) string {
	return fmt.Sprintf("Locations{Domain:%v}", d)
}

func userPropertyKeyForId(d *datastore.Key) string {
	return fmt.Sprintf("%v{Id:%v}", userPropertyKind, d)
}

func locationKeyForId(d *datastore.Key) string {
	return fmt.Sprintf("Location{Id:%v}", d)
}

type UserProperty struct {
	Id        *datastore.Key `datastore:"-" json:"id"`
	Name      string         `json:"name"`
	DaysValid int            `json:"days_valid"`
}

func (self *UserProperty) CopyFrom(o *UserProperty) *UserProperty {
	self.DaysValid = o.DaysValid
	return self
}

func findUserProperty(c gaecontext.HTTPContext, key *datastore.Key) *UserProperty {
	var t UserProperty
	err := datastore.Get(c, key, &t)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	common.AssertOkError(err)
	t.Id = key
	return &t
}

func PropertyID(c gaecontext.HTTPContext, name string, dom *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, userPropertyKind, name, 0, dom)
}

func GetUserProperty(c gaecontext.HTTPContext, key, domain *datastore.Key) *UserProperty {
	if key == nil {
		return nil
	}
	if !key.Parent().Equal(domain) {
		panic(fmt.Errorf("%v is not parent of %v", domain, key))
	}
	var t UserProperty
	if common.Memoize(c, userPropertyKeyForId(key), &t, func() interface{} {
		return findUserProperty(c, key)
	}) {
		return &t
	}
	return nil
}

func DeleteUserProperty(c gaecontext.HTTPContext, key, dom *datastore.Key) {
	if !dom.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, key))
	}
	if err := datastore.Delete(c, key); err != nil {
		panic(err)
	}
	common.MemDel(c, userPropertiesKeyForDomain(dom))
	common.MemDel(c, userPropertyKeyForId(key))
}

func (self *UserProperty) Save(c gaecontext.HTTPContext, dom *datastore.Key) *UserProperty {
	var err error
	self.Id, err = datastore.Put(c, PropertyID(c, self.Name, dom), self)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, userPropertiesKeyForDomain(dom))
	common.MemDel(c, userPropertyKeyForId(self.Id))
	return self
}

func findUserProperties(c gaecontext.HTTPContext, dom *datastore.Key) (result []UserProperty) {
	keys, err := datastore.NewQuery(userPropertyKind).Ancestor(dom).GetAll(c, &result)
	common.AssertOkError(err)
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetUserProperties(c gaecontext.HTTPContext, dom *datastore.Key) (result []UserProperty) {
	common.Memoize(c, userPropertiesKeyForDomain(dom), &result, func() interface{} {
		return findUserProperties(c, dom)
	})
	if result == nil {
		result = make([]UserProperty, 0)
	}
	return
}

type Location struct {
	Id   *datastore.Key `datastore:"-" json:"id"`
	Name string         `json:"name"`

	// Salary mod
	SalarySerializedProperties []byte                 `json:"-"`
	SalaryProperties           map[string]interface{} `json:"salary_properties" datastore:"-"`
}

func (self *Location) process(c gaecontext.HTTPContext) *Location {
	if len(self.SalarySerializedProperties) > 0 {
		if err := json.Unmarshal(self.SalarySerializedProperties, &self.SalaryProperties); err != nil {
			panic(err)
		}
	}
	return self
}

func findLocation(c gaecontext.HTTPContext, key *datastore.Key) *Location {
	var t Location
	err := datastore.Get(c, key, &t)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	common.AssertOkError(err)
	t.Id = key
	return &t
}

func GetLocation(c gaecontext.HTTPContext, key *datastore.Key) *Location {
	if key == nil {
		return nil
	}
	var t Location
	if common.Memoize(c, locationKeyForId(key), &t, func() interface{} {
		return findLocation(c, key)
	}) {
		return (&t).process(c)
	}
	return nil
}

func (self *Location) Save(c gaecontext.HTTPContext, dom *datastore.Key) *Location {
	var err error

	self.SalarySerializedProperties, err = json.Marshal(self.SalaryProperties)
	if err != nil {
		panic(err)
	}

	self.Id, err = datastore.Put(c, datastore.NewKey(c, "Location", self.Name, 0, dom), self)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, locationsKeyForDomain(dom))
	common.MemDel(c, locationKeyForId(self.Id))
	return self
}

func DeleteLocation(c gaecontext.HTTPContext, key, dom *datastore.Key) {
	if !dom.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, key))
	}
	if err := datastore.Delete(c, key); err != nil {
		panic(err)
	}
	common.MemDel(c, locationsKeyForDomain(dom))
	common.MemDel(c, locationKeyForId(key))
}

func findLocations(c gaecontext.HTTPContext, dom *datastore.Key) (result []Location) {
	keys, err := datastore.NewQuery("Location").Ancestor(dom).GetAll(c, &result)
	common.AssertOkError(err)
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetLocations(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer) (result []Location) {
	var preResult []Location
	common.Memoize(c, locationsKeyForDomain(dom), &preResult, func() interface{} {
		return findLocations(c, dom)
	})
	hasDomainAuth := authorizer == nil || authorizer.HasAuth(auth.Auth{
		AuthType: auth.Domain,
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
	for _, loc := range preResult {
		authMatchEvents.Location = loc.Id
		authMatchAttend.Location = loc.Id
		authMatchParticipants.Location = loc.Id
		authMatchAttest.EventKind = loc.Id
		authMatchReport.EventKind = loc.Id
		if authorizer == nil || (hasDomainAuth || authorizer.HasAnyAuth(authMatchParticipants) || authorizer.HasAnyAuth(authMatchAttend) || authorizer.HasAnyAuth(authMatchEvents) || authorizer.HasAnyAuth(authMatchAttest) || authorizer.HasAnyAuth(authMatchReport)) {
			cpy := loc
			(&cpy).process(c)
			result = append(result, cpy)
		}
	}
	if result == nil {
		result = make([]Location, 0)
	}
	return
}

type Domain struct {
	Id                    *datastore.Key `datastore:"-" json:"id"`
	Name                  string         `json:"name"`
	TZLocation            string         `json:"tz_location"`
	AutoDisable           bool           `json:"auto_disable"`
	AutoDisableAfter      int            `json:"auto_disable_after"`
	ExtraConfirmationBCC  string         `json:"extra_confirmation_bcc"`
	FromAddress           string         `json:"from_address"`
	EarliestEvent         time.Time      `json:"earliest_event,omitempty"`
	LatestEvent           time.Time      `json:"latest_event,omitempty"`
	LimitedICS            bool           `json:"limited_ics"`
	SalaryMod             bool           `json:"salary_mod"`
	ClosedAndRedirectedTo string         `json:"closed_and_redirected_to"`

	// These two are not used for domains per se, but for the generated domains replacing a users DomainUsers
	Owner            bool                   `datastore:"-" json:"owner"`
	Disabled         bool                   `datastore:"-" json:"disabled"`
	AllowICS         bool                   `datastore:"-" json:"allow_ics"`
	Information      string                 `datastore:"-" json:"information"`
	LastActivity     time.Time              `datastore:"-" json:"last_activity"`
	SalaryProperties map[string]interface{} `datastore:"-" json:"salary_properties"`
	SalaryConfig     interface{}            `datastore:"-" json:"salary_config"`
}

func (self *Domain) CopyFrom(other *Domain) {
	self.AutoDisable = other.AutoDisable
	self.AutoDisableAfter = other.AutoDisableAfter
	self.ExtraConfirmationBCC = other.ExtraConfirmationBCC
	self.TZLocation = other.TZLocation
	self.EarliestEvent = other.EarliestEvent
	self.LatestEvent = other.LatestEvent
	self.FromAddress = other.FromAddress
}

func (self *Domain) ToJSON() string {
	buffer := new(bytes.Buffer)
	common.MustEncodeJSON(buffer, self)
	return string(buffer.Bytes())
}

func (self *Domain) GetLocation() (loc *time.Location) {
	var err error
	loc, err = time.LoadLocation(self.TZLocation)
	if err != nil {
		loc, err = time.LoadLocation("UTC")
		if err != nil {
			panic(err)
		}
	}
	return
}

func findDomain(c gaecontext.HTTPContext, key *datastore.Key) *Domain {
	var t Domain
	err := datastore.Get(c, key, &t)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	common.AssertOkError(err)
	t.Id = key
	return &t
}

func GetDomain(c gaecontext.HTTPContext, key *datastore.Key) *Domain {
	if key == nil {
		return nil
	}
	var t Domain
	if common.Memoize(c, domainKeyForId(key), &t, func() interface{} {
		return findDomain(c, key)
	}) {
		return (&t)
	}
	return nil
}

func GetDomains(c gaecontext.HTTPContext, keys []*datastore.Key) (result []Domain) {
	cacheKeys := make([]string, len(keys))
	destPs := make([]interface{}, len(keys))
	funcs := make([]func() interface{}, len(keys))
	for index, key := range keys {
		cacheKeys[index] = domainKeyForId(key)
		var dom Domain
		destPs[index] = &dom
		idCopy := key
		funcs[index] = func() interface{} {
			return findDomain(c, idCopy)
		}
	}

	common.MemoizeMulti(c, cacheKeys, destPs, funcs)

	result = make([]Domain, len(keys))

	for index, _ := range destPs {
		result[index] = *(destPs[index].(*Domain))
	}
	return

}

type tzError struct {
	TZLocation string `json:"tz_location"`
	Message    string `json:"message"`
}

func (self tzError) Error() string {
	return self.Message
}

func (self *Domain) Save(c gaecontext.HTTPContext) {
	if self.TZLocation != "" {
		if _, e := time.LoadLocation(self.TZLocation); e != nil {
			panic(e)
		}
	}

	var err error
	self.Id, err = datastore.Put(c, datastore.NewKey(c, "Domain", self.Name, 0, nil), self)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, allDomainsKey)
	common.MemDel(c, domainKeyForId(self.Id))
}

// findAll will set Owner = truefor all returned domains, since only Admins are supposed to be able to use this.
func findAll(c gaecontext.HTTPContext) (result []Domain) {
	keys, err := datastore.NewQuery("Domain").GetAll(c, &result)
	common.AssertOkError(err)
	for index, key := range keys {
		result[index].Id = key
		result[index].Owner = true
	}
	return
}

func GetAll(c gaecontext.HTTPContext) (result []Domain) {
	common.Memoize(c, allDomainsKey, &result, func() interface{} {
		return findAll(c)
	})
	if result == nil {
		result = make([]Domain, 0)
	}
	return
}

func Destroy(c gaecontext.HTTPContext, key *datastore.Key) {
	err := datastore.Delete(c, key)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, allDomainsKey)
	common.MemDel(c, domainKeyForId(key))
}
