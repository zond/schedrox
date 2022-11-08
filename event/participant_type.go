//go:generate msgp
//msgp:shim *datastore.Key as:string using:common.EncodeKey/common.DecodeKey
package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zond/schedrox/appuser"
	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/crm"
	"github.com/zond/schedrox/domain"
	"github.com/zond/schedrox/translation"
	"sort"
	"strings"
	"time"

	"github.com/zond/sybutils/utils/gae"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/mail"
)

func participantTypesKeyForDomain(d *datastore.Key) string {
	return fmt.Sprintf("ParticipantTypes{Domain:%v}", d)
}

func participantsKeyForEvent(e *datastore.Key) string {
	return fmt.Sprintf("Participants{Event:%v}", e)
}

func contactEventsForContactKey(k *datastore.Key) string {
	return fmt.Sprintf("Events{Contact:%v}", k)
}

func participantTypeKeyForId(i *datastore.Key) string {
	return fmt.Sprintf("ParticipantType{Id:%v}", i)
}

func participantKeyForId(i *datastore.Key) string {
	return fmt.Sprintf("Participant{Id:%v}", i)
}

func requiredParticipantTypesKeyForParent(parent *datastore.Key) string {
	return fmt.Sprintf("RequiredParticipantTypes{Parent:%v}", parent)
}

func allowedParticipantsKey(domain, location, eventKind, eventType, participantType *datastore.Key) string {
	return fmt.Sprintf(
		"PotentialParticipants{Domain:%+v,Location:%+v,EventKind:%+v,EventType:%+v,ParticipantType:%v}",
		domain,
		location,
		eventKind,
		eventType,
		participantType,
	)
}

type Participants []Participant

func (self Participants) Len() int {
	return len(self)
}

func (self Participants) Less(i, j int) bool {
	return bytes.Compare([]byte(strings.ToLower(self[i].Name)), []byte(strings.ToLower(self[j].Name))) < 0
}

func (self Participants) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type Participant struct {
	Id                  *datastore.Key `datastore:"-" json:"id"`
	User                *datastore.Key `json:"user,omitempty"`
	Contact             *datastore.Key `json:"contact,omitempty"`
	ParticipantType     *datastore.Key `json:"participant_type"`
	ParticipantTypeName string         `json:"participant_type_name" datastore:"-"`
	Multiple            int            `json:"multiple"`
	Confirmations       int            `json:"confirmations"`
	Paid                bool           `json:"paid"`
	Defaulted           bool           `json:"defaulted"`
	Name                string         `datastore:"-" json:"name"`
	GivenName           string         `datastore:"-" json:"given_name"`
	FamilyName          string         `datastore:"-" json:"family_name"`
	GravatarHash        string         `datastore:"-" json:"gravatar_hash"`
	Email               string         `datastore:"-" json:"email"`
	MobilePhone         string         `datastore:"-" json:"mobile_phone"`
	Auths               []auth.Auth    `datastore:"-" json:"auths"`
	EventStart          time.Time      `datastore:"-" json:"event_start,omitempty"`
	EventEnd            time.Time      `datastore:"-" json:"event_end,omitempty"`
	EmailBounce         string         `datastore:"-" json:"email_bounce"`
	Missing             bool           `datastore:"-" json:"missing"`

	// Just a message from the client that this participant is NOT supposed to be a member of
	// all events in a recurring series...
	AlwaysCreateException bool `datastore:"-" json:"always_create_exception"`

	// preProcess cache
	cachedUser    *appuser.User
	cachedContact *crm.Contact
	cachedAuths   map[string][]auth.Auth
}

func findParticipant(c gaecontext.HTTPContext, key *datastore.Key) *Participant {
	var t Participant
	if err := datastore.Get(c, key, &t); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil
		}
		panic(err)
	}
	t.Id = key
	return &t
}

func GetParticipant(c gaecontext.HTTPContext, key *datastore.Key) *Participant {
	if key == nil {
		return nil
	}
	var t Participant
	if common.Memoize(c, participantKeyForId(key), &t, func() interface{} {
		return findParticipant(c, key)
	}) {
		return &t
	}
	return nil
}

func (self *Participant) CopyFrom(o *Participant) *Participant {
	self.User = o.User
	self.Contact = o.Contact
	self.ParticipantType = o.ParticipantType
	self.Multiple = o.Multiple
	self.Confirmations = o.Confirmations
	self.Paid = o.Paid
	self.Defaulted = o.Defaulted
	return self
}

func (self *Participant) Delete(c gaecontext.HTTPContext, actor *datastore.Key, event *Event) {
	if self.User != nil {
		if user := appuser.GetUserByKey(c, self.User); user != nil && !user.MuteEventNotifications {
			if event == nil {
				event = GetEvent(c, self.Id.Parent(), self.Id.Parent().Parent())
			}
			if event != nil {
				var theDomain *domain.Domain
				for _, d := range user.Domains {
					if d.Id.Equal(self.Id.Parent().Parent()) {
						theDomain = &d
						break
					}
				}
				var attachment *mail.Attachment
				if theDomain.AllowICS || !theDomain.LimitedICS {
					attachment = &mail.Attachment{
						Name: "cancellation.ics",
						Data: event.getIcsAttachment(c, "cancellation.ics", user.Email),
					}
				}
				user.SendMail(c, self.Id.Parent().Parent(), translation.GetTranslation(user.LastLanguage, "Removed from event"), translation.GetEmailBody(user.LastLanguage, "delete_participant.txt", map[string]interface{}{
					"event":       event,
					"participant": self,
					"AppID":       appengine.AppID(c),
				}), "", attachment)
			}
		}
	}
	if err := datastore.Delete(c, self.Id); err != nil {
		panic(err)
	}
	CreateChange(c, self.Id.Parent(), actor, "DeleteParticipant", self)
	common.MemDel(c, participantsKeyForEvent(self.Id.Parent()), participantKeyForId(self.Id), contactEventsForContactKey(self.Contact))
}

func (self *Participant) Key(c gaecontext.HTTPContext, event_id *datastore.Key) *datastore.Key {
	if self.User != nil {
		return datastore.NewKey(c, "Participant", common.EncKey(self.User), 0, event_id)
	}
	return datastore.NewKey(c, "Participant", "", 0, event_id)
}

func (self *Participant) Save(c gaecontext.HTTPContext, event *Event, dom, actor *datastore.Key) *Participant {
	if self.User == nil && self.Contact == nil {
		panic(fmt.Errorf("%+v has to have User or Contact set", self))
	}
	if self.User != nil && IsBusyUserId(c, event.Start, event.End, event.Id, self.User) {
		panic(fmt.Errorf("%v is busy during %+v", self.User, event))
	}
	isNew := false
	if self.Id == nil {
		isNew = true
	} else {
		CreateChange(c, event.Id, actor, "UpdateParticipant", self)
	}
	var err error
	if self.Id == nil {
		self.Id, err = datastore.Put(c, self.Key(c, event.Id), self)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	if err != nil {
		panic(err)
	}
	if isNew {
		CreateChange(c, event.Id, actor, "CreateParticipant", self)
	}
	if self.User != nil {
		if user := appuser.GetUserByKey(c, self.User); user != nil {
			user.ActivateInDomain(c, dom)
			if isNew && !user.MuteEventNotifications {
				if event := GetEvent(c, self.Id.Parent(), self.Id.Parent().Parent()); event != nil {
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
							Data: event.getIcsAttachment(c, "invitation.ics", user.Email),
						}
					}
					user.SendMail(c, self.Id.Parent().Parent(), translation.GetTranslation(user.LastLanguage, "Added to event"), translation.GetEmailBody(user.LastLanguage, "add_participant.txt", map[string]interface{}{
						"event":       event,
						"participant": self,
						"AppID":       appengine.AppID(c),
					}), "", attachment)
				}
			}
		}
	}
	common.MemDel(c, participantsKeyForEvent(event.Id), participantKeyForId(self.Id), contactEventsForContactKey(self.Contact))
	return self.process(c, dom)
}

func (self *Participant) apply(user *appuser.User) {
	self.GivenName = user.GivenName
	self.FamilyName = user.FamilyName
	self.GravatarHash = user.GravatarHash
	self.MobilePhone = user.MobilePhone
	self.EmailBounce = user.EmailBounce
	self.Name = user.GivenName
}

func (self *Participant) GetContact(c gaecontext.HTTPContext) *crm.Contact {
	if self.cachedContact == nil && self.Contact != nil {
		self.cachedContact = crm.GetContact(c, self.Contact, self.Contact.Parent())
	}
	return self.cachedContact
}

func (self *Participant) GetUser(c gaecontext.HTTPContext) *appuser.User {
	if self.cachedUser == nil && self.User != nil {
		self.cachedUser = appuser.GetUserByKey(c, self.User)
	}
	return self.cachedUser
}

func (self *Participant) getUserAuthsForDomain(c gaecontext.HTTPContext, dom *datastore.Key) []auth.Auth {
	if self.cachedAuths == nil {
		self.cachedAuths = make(map[string][]auth.Auth)
	}
	if _, found := self.cachedAuths[dom.Encode()]; !found {
		self.cachedAuths[dom.Encode()] = self.GetUser(c).GetAuthsForDomain(c, dom)
	}
	return self.cachedAuths[dom.Encode()]
}

const (
	cachedUser = iota
	cachedAuths
	cachedRoles
	cachedContact
)

func preProcessParticipants(c gaecontext.HTTPContext, dom *datastore.Key, parts Participants) {
	// phase 1: fetch all users, their auths and their roles

	cacheKeys := []string{}
	funcs := []func() interface{}{}
	values := []interface{}{}
	types := []int{}

	for _, part := range parts {
		if part.User != nil {

			cacheKeys = append(cacheKeys, appuser.UserKeyForId(part.User))
			idCopy := part.User
			funcs = append(funcs, func() interface{} {
				return appuser.FindUser(c, idCopy)
			})
			var user appuser.User
			values = append(values, &user)
			types = append(types, cachedUser)

			domainUserKey := appuser.DomainUserKeyUnderDomain(c, dom, idCopy)

			cacheKeys = append(cacheKeys, auth.AuthsKeyForParent(domainUserKey))
			funcs = append(funcs, func() interface{} {
				return auth.FindAuths(c, domainUserKey)
			})
			var auths []auth.Auth
			values = append(values, &auths)
			types = append(types, cachedAuths)

			cacheKeys = append(cacheKeys, auth.RolesKeyForParent(domainUserKey))
			funcs = append(funcs, func() interface{} {
				return auth.FindRoles(c, domainUserKey)
			})
			var roles []auth.Role
			values = append(values, &roles)
			types = append(types, cachedRoles)

		} else if part.Contact != nil {

			cacheKeys = append(cacheKeys, crm.ContactKeyForId(part.Contact))
			idCopy := part.Contact
			funcs = append(funcs, func() interface{} {
				return crm.GetContact(c, idCopy, idCopy.Parent())
			})
			var contact crm.Contact
			values = append(values, &contact)
			types = append(types, cachedContact)

		}
	}

	common.MemoizeMulti(c, cacheKeys, values, funcs)

	// phase 2: build a user array, and index the auths and roles

	userMap := make(map[string]appuser.User)
	contactMap := make(map[string]crm.Contact)
	authMap := make(map[string][]auth.Auth)
	roleAuthMap := make(map[string][]auth.Auth)
	userRoleMap := make(map[string][]auth.Role)
	var lastUserId *datastore.Key
	partIndex := 0
	for index, typ := range types {
		switch typ {
		case cachedUser:
			lastUserId = parts[partIndex].User
			partIndex++
			userMap[lastUserId.Encode()] = *(values[index].(*appuser.User))
		case cachedAuths:
			authMap[lastUserId.Encode()] = *(values[index].(*[]auth.Auth))
		case cachedRoles:
			roles := *(values[index].(*[]auth.Role))
			userRoleMap[lastUserId.Encode()] = roles
			for _, role := range roles {
				roleAuthMap[role.Id.StringID()] = nil
			}
		case cachedContact:
			contactId := parts[partIndex].Contact
			partIndex++
			contactMap[contactId.Encode()] = *(values[index].(*crm.Contact))
		default:
			panic(fmt.Errorf("Do not know how to handle %v type values directly here...", types[index]))
		}
	}

	// phase 3: preProcess the users and process and update them in the map

	users := make([]appuser.User, 0, len(userMap))
	for _, user := range userMap {
		users = append(users, user)
	}
	appuser.PreProcess(c, dom, users)
	for _, user := range users {
		(&user).Process(c)
		userMap[user.Id.Encode()] = user
	}

	// phase 4: fetch the auths of each role we need
	cacheKeys = make([]string, 0, len(roleAuthMap))
	funcs = make([]func() interface{}, 0, len(roleAuthMap))
	values = make([]interface{}, 0, len(roleAuthMap))
	roleNames := make([]string, 0, len(roleAuthMap))

	for name, _ := range roleAuthMap {
		roleNames = append(roleNames, name)
		roleAuthsKey := datastore.NewKey(c, "Role", name, 0, auth.DomainRolesKey(c, dom))
		cacheKeys = append(cacheKeys, auth.AuthsKeyForParent(roleAuthsKey))
		funcs = append(funcs, func() interface{} {
			return auth.FindAuths(c, roleAuthsKey)
		})
		var auths []auth.Auth
		values = append(values, &auths)
	}

	common.MemoizeMulti(c, cacheKeys, values, funcs)

	// phase 5: index the auths of the roles

	for index, name := range roleNames {
		roleAuthMap[name] = *(values[index].(*[]auth.Auth))
	}

	// phase 6: for each participant, set find its user, set the auths of that user from the authMap and append the auths from each role of the user
	for index, _ := range parts {
		if parts[index].User != nil {
			userId := parts[index].User
			userCopy := userMap[userId.Encode()]
			parts[index].cachedUser = &userCopy
			parts[index].cachedAuths = make(map[string][]auth.Auth)
			parts[index].cachedAuths[dom.Encode()] = authMap[userId.Encode()]
			for _, role := range userRoleMap[userId.Encode()] {
				parts[index].cachedAuths[dom.Encode()] = append(parts[index].cachedAuths[dom.Encode()], roleAuthMap[role.Id.StringID()]...)
			}
		} else if parts[index].Contact != nil {
			contactId := parts[index].Contact
			contactCopy := contactMap[contactId.Encode()]
			parts[index].cachedContact = &contactCopy
		}
	}
}

func (self *Participant) quickProcess() {
	self.ParticipantTypeName = self.ParticipantType.StringID()
	if self.User != nil {
		self.Email = self.User.StringID()
	}
}

func (self *Participant) process(c gaecontext.HTTPContext, dom *datastore.Key) *Participant {
	self.quickProcess()
	if self.User != nil {
		if user := self.GetUser(c); user != nil {
			self.apply(user)
			for _, a := range self.getUserAuthsForDomain(c, dom) {
				if a.AuthType == auth.Attend {
					self.Auths = append(self.Auths, a)
				}
			}
		} else {
			self.Missing = true
		}
	} else {
		if contact := self.GetContact(c); contact != nil {
			self.Name = contact.Name
			self.GivenName = contact.ContactGivenName
			self.FamilyName = contact.ContactFamilyName
			self.GravatarHash = contact.GravatarHash
			self.MobilePhone = contact.MobilePhone
			self.Email = contact.Email
			self.EmailBounce = contact.EmailBounce
		} else {
			self.Missing = true
		}
	}
	return self
}

func RemoveUserFromFutureEvents(c gaecontext.HTTPContext, actor, user, d *datastore.Key) {
	localTime := time.Now().In(domain.GetDomain(c, d).GetLocation())
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}
	now := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), localTime.Hour(), localTime.Minute(), localTime.Second(), localTime.Nanosecond(), utc)

	// any event after now
	ids, err := datastore.NewQuery("Event").Ancestor(d).Filter("Start>", now).KeysOnly().GetAll(c, nil)
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for _, serr := range merr {
				if serr != nil {
					if _, ok := serr.(*datastore.ErrFieldMismatch); !ok {
						panic(err)
					}
				}
			}
		} else {
			panic(err)
		}
	}
	eventIds := make(map[string]bool)
	for _, id := range ids {
		eventIds[common.EncKey(id)] = true
	}

	// any participations of this user
	ids, err = datastore.NewQuery("Participant").Ancestor(d).Filter("User=", user).KeysOnly().GetAll(c, nil)
	if err != nil {
		panic(err)
	}

	// remove participants from the events found that start in the future
	var remainingParticipantIds []*datastore.Key
	for _, id := range ids {
		if eventIds[common.EncKey(id.Parent())] {
			participant := GetParticipant(c, id)
			participant.Delete(c, actor, nil)
		} else {
			// while saving the remaining participations for later
			remainingParticipantIds = append(remainingParticipantIds, id)
		}
	}

	// recurring events extending beyond now
	ids, err = datastore.NewQuery("Event").Ancestor(d).Filter("Recurring=", true).Filter("RecurrenceEnd>", now).KeysOnly().GetAll(c, nil)
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for _, serr := range merr {
				if serr != nil {
					if _, ok := serr.(*datastore.ErrFieldMismatch); !ok {
						panic(err)
					}
				}
			}
		} else {
			panic(err)
		}
	}
	eventIds = make(map[string]bool)
	for _, id := range ids {
		eventIds[common.EncKey(id)] = true
	}

	// any participation in recurring events extending beyond now will result in those events being split in two, where there is no participation in the future part
	for _, participantId := range remainingParticipantIds {
		if eventIds[common.EncKey(participantId.Parent())] {
			if event := GetEvent(c, participantId.Parent(), d); event != nil {
				event.SplitAndRemoveParticipant(c, d, now, actor, participantId, nil)
			}
		}
	}
}

func getUserIdsAndRoleNames(c gaecontext.HTTPContext, auths []auth.Auth, ignoreUserIds map[string]bool) (userIds []*datastore.Key, roleNames []string) {
	seenRoleNames := make(map[string]bool)
	var tmpString string
	for _, a := range auths {
		if a.Id.Parent().Kind() == "DomainUser" {
			tmpId := datastore.NewKey(c, "User", a.Id.Parent().StringID(), 0, nil)
			if !ignoreUserIds[common.EncKey(tmpId)] {
				ignoreUserIds[common.EncKey(tmpId)] = true
				userIds = append(userIds, tmpId)
			}
		} else if a.Id.Parent().Kind() == "Role" {
			tmpString = a.Id.Parent().StringID()
			if !seenRoleNames[tmpString] {
				roleNames = append(roleNames, tmpString)
			}
		}
	}
	return
}

func appendRoleHolders(c gaecontext.HTTPContext, domain *datastore.Key, roleNames []string, ignoreUserIds map[string]bool, userIds *[]*datastore.Key) {
	for _, name := range roleNames {
		for _, role := range auth.GetRolesByName(c, domain, name) {
			if role.Id.Parent().Kind() == "DomainUser" {
				tmpId := datastore.NewKey(c, "User", role.Id.Parent().StringID(), 0, nil)
				if !ignoreUserIds[common.EncKey(tmpId)] {
					ignoreUserIds[common.EncKey(tmpId)] = true
					*userIds = append(*userIds, tmpId)
				}
			}
		}
	}
}

func findContactEvents(c gaecontext.HTTPContext, dom, contact *datastore.Key) (result Events) {
	var participants []Participant
	ids, err := datastore.NewQuery("Participant").Ancestor(dom).Filter("Contact=", contact).GetAll(c, &participants)
	common.AssertOkError(err)
	participantTypeNames := make([]string, len(ids))
	for index, id := range ids {
		ids[index] = id.Parent()
		participantTypeNames[index] = participants[index].ParticipantType.StringID()
	}
	result = make(Events, len(ids))
	err = datastore.GetMulti(c, ids, result)
	common.AssertOkError(err)
	for index, _ := range result {
		result[index].Id = ids[index]
		result[index].ParticipantTypeName = participantTypeNames[index]
	}
	return
}

func GetAllowedContactEvents(c gaecontext.HTTPContext, dom, contact *datastore.Key, authorizer auth.Authorizer) (result Events) {
	var preResult Events
	common.Memoize(c, contactEventsForContactKey(contact), &preResult, func() interface{} {
		return findContactEvents(c, dom, contact)
	})
	for _, ev := range preResult {
		if authorizer.HasAuth(auth.Auth{
			AuthType:  auth.Events,
			Location:  ev.Location,
			EventKind: ev.EventKind,
			EventType: ev.EventType,
		}) {
			result = append(result, ev)
		}
	}
	preProcess(c, result)
	for index, _ := range result {
		(&result[index]).process(c)
	}
	if result == nil {
		result = Events{}
	}
	sort.Sort(result)
	return
}

func IsBusyUserId(c gaecontext.HTTPContext, start, end time.Time, ignoreBusyEvent, userId *datastore.Key) bool {
	for uid, _ := range getBusyUserIds(c, start, end, ignoreBusyEvent) {
		if uid == common.EncKey(userId) {
			return true
		}
	}
	return false
}

func GetBusyTimes(c gaecontext.HTTPContext, start, end time.Time, user *datastore.Key) (times [][2]time.Time) {
	for _, dom := range domain.GetAll(c) {
		for _, event := range GetEventsBetween(c, dom.Id, start, end, func(ev *Event) Events {
			return Events{*ev}
		}) {
			if common.Overlaps([2]time.Time{event.Start, event.End}, [2]time.Time{start, end}) {
				for _, participant := range event.GetParticipants(c) {
					if user.Equal(participant.User) {
						times = append(times, [2]time.Time{
							event.Start, event.End,
						})
					}
				}
			}
		}
	}
	return
}

func getBusyUserIds(c gaecontext.HTTPContext, start, end time.Time, ignoreBusyEvent *datastore.Key) (uids map[string]bool) {
	uids = make(map[string]bool)
	for _, dom := range domain.GetAll(c) {
		for _, event := range GetEventsBetween(c, dom.Id, start, end, func(ev *Event) Events {
			if !ev.Id.Equal(ignoreBusyEvent) {
				return Events{*ev}
			}
			return nil
		}) {
			if common.Overlaps([2]time.Time{event.Start, event.End}, [2]time.Time{start, end}) {
				for _, participant := range event.GetParticipants(c) {
					uids[common.EncKey(participant.User)] = true
				}
			}
		}
	}
	return
}

func injectDisabledUserIds(c gaecontext.HTTPContext, domain *datastore.Key, uids map[string]bool) map[string]bool {
	for _, uid := range appuser.GetDisabledIds(c, domain, true) {
		uids[common.EncKey(uid)] = true
	}
	return uids
}

func GetPotentialParticipants(c gaecontext.HTTPContext, domain, location, eventKind, eventType, participantType *datastore.Key, start, end time.Time, ignoreBusyEvent *datastore.Key) (result []Participant) {
	return getAllowedParticipants(c, domain, location, eventKind, eventType, participantType, injectDisabledUserIds(c, domain, getBusyUserIds(c, start, end, ignoreBusyEvent)))
}

func getAllowedParticipants(c gaecontext.HTTPContext, domain, location, eventKind, eventType, participantType *datastore.Key, ignoreUserIds map[string]bool) (result Participants) {
	var allowedAuths []auth.Auth
	common.Memoize2(c, auth.AuthsKeyForDomainAndType(domain, auth.Attend), allowedParticipantsKey(domain, location, eventKind, eventType, participantType), &allowedAuths, func() interface{} {
		return findAllowedAuths(c, domain, location, eventKind, eventType, participantType)
	})
	userIds, roleNames := getUserIdsAndRoleNames(c, allowedAuths, ignoreUserIds)
	appendRoleHolders(c, domain, roleNames, ignoreUserIds, &userIds)
	for _, userId := range userIds {
		fakeParticipant := Participant{
			User:            userId,
			Multiple:        1,
			ParticipantType: participantType,
		}
		result = append(result, fakeParticipant)
	}
	preProcessParticipants(c, domain, result)
	for index, _ := range result {
		(&result[index]).process(c, domain)
	}
	sort.Sort(result)
	return
}

func findAllowedAuths(c gaecontext.HTTPContext, domain, location, eventKind, eventType, participantType *datastore.Key) (result []auth.Auth) {
	matchAuth := auth.Auth{
		AuthType:        auth.Attend,
		Location:        location,
		EventKind:       eventKind,
		EventType:       eventType,
		ParticipantType: participantType,
	}
	for _, a := range auth.GetAuthsByType(c, domain, auth.Attend) {
		if a.Matches(c, matchAuth) {
			result = append(result, a)
		}
	}
	return
}

func (self *Event) findParticipants(c gaecontext.HTTPContext) (result []Participant) {
	keys, err := datastore.NewQuery("Participant").Ancestor(self.Id).GetAll(c, &result)
	if gae.FilterOkErrors(err) != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
		(&result[index]).quickProcess()
	}
	if result == nil {
		result = make([]Participant, 0)
	}
	return
}

func (self *Event) GetParticipant(c gaecontext.HTTPContext, participant_id *datastore.Key) *Participant {
	for _, participant := range self.GetParticipants(c) {
		if participant.Id.Equal(participant_id) {
			return &participant
		}
	}
	return nil
}

func (self *Event) GetParticipants(c gaecontext.HTTPContext) []Participant {
	if self.Id == nil {
		return nil
	}
	if self.cachedParticipants == nil {
		common.Memoize(c, participantsKeyForEvent(self.Id), &self.cachedParticipants, func() interface{} {
			return self.findParticipants(c)
		})
	}
	return self.cachedParticipants
}

func (self *EventType) GetAllowedRequiredParticipantTypes(c gaecontext.HTTPContext, location *datastore.Key, authorizer auth.Authorizer, dom *datastore.Key) (result []RequiredParticipantType) {
	if self == nil {
		return
	}
	if !self.Id.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, self.Id))
	}
	canAttend := authorizer.HasAnyAuth(auth.Auth{
		AuthType:  auth.Attend,
		Location:  location,
		EventKind: self.EventKind,
		EventType: self.Id,
	})
	hasTypesAuth := authorizer.HasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	})
	participantsAuthMatch := auth.Auth{
		AuthType:  auth.Participants,
		Location:  location,
		EventKind: self.EventKind,
		EventType: self.Id,
	}
	requiredTypes := GetRequiredParticipantTypes(c, self.Id, self.Id.Parent())
	for _, t := range requiredTypes {
		participantsAuthMatch.ParticipantType = t.Id
		if canAttend || hasTypesAuth || authorizer.HasAuth(participantsAuthMatch) {
			result = append(result, t)
		}
	}
	return
}

func (self *Event) GetAllowedRequiredParticipantTypes(c gaecontext.HTTPContext, authorizer auth.Authorizer, dom, asker *datastore.Key) (result []RequiredParticipantType) {
	isPart := self.IsParticipant(c, asker)
	requiredTypes := GetRequiredParticipantTypes(c, self.Id, dom)
	canAttend := authorizer.HasAnyAuth(auth.Auth{
		AuthType:  auth.Attend,
		Location:  self.Location,
		EventKind: self.EventKind,
		EventType: self.EventType,
	})
	participantsAuthMatch := auth.Auth{
		AuthType:  auth.Participants,
		Location:  self.Location,
		EventKind: self.EventKind,
		EventType: self.EventType,
	}
	for _, t := range requiredTypes {
		participantsAuthMatch.ParticipantType = t.Id
		if canAttend || isPart || authorizer.HasAuth(participantsAuthMatch) {
			result = append(result, t)
		}
	}
	return
}

func (self *Event) IsParticipant(c gaecontext.HTTPContext, user *datastore.Key) bool {
	for _, p := range self.GetParticipants(c) {
		if user.Equal(p.User) {
			return true
		}
	}
	return false
}

func (self *Event) GetAllowedParticipants(c gaecontext.HTTPContext, authorizer auth.Authorizer, dom, asker *datastore.Key) (result []Participant) {
	if !self.Id.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, self.Id))
	}
	isPart := self.IsParticipant(c, asker)
	canAttend := authorizer.HasAnyAuth(auth.Auth{
		AuthType:  auth.Attend,
		Location:  self.Location,
		EventKind: self.EventKind,
		EventType: self.EventType,
	})
	participantsAuthMatch := auth.Auth{
		AuthType:  auth.Participants,
		Write:     false,
		Location:  self.Location,
		EventKind: self.EventKind,
		EventType: self.EventType,
	}
	for _, p := range self.GetParticipants(c) {
		participantsAuthMatch.ParticipantType = p.ParticipantType
		if isPart || canAttend || authorizer.HasAuth(participantsAuthMatch) {
			part := p
			(&part).process(c, dom)
			result = append(result, part)
		}
	}
	return
}

func (self *Event) getRequiredParticipantTypesForType(c gaecontext.HTTPContext) []RequiredParticipantType {
	if self.cachedRequiredParticipantTypesForType == nil {
		self.cachedRequiredParticipantTypesForType = GetRequiredParticipantTypes(c, self.EventType, self.EventType.Parent())
	}
	if self.cachedRequiredParticipantTypesForType == nil {
		return make([]RequiredParticipantType, 0)
	}
	return self.cachedRequiredParticipantTypesForType
}

func (self *Event) getRequiredParticipantTypesForId(c gaecontext.HTTPContext) []RequiredParticipantType {
	if self.Id == nil {
		return make([]RequiredParticipantType, 0)
	}
	if self.cachedRequiredParticipantTypesForId == nil {
		self.cachedRequiredParticipantTypesForId = GetRequiredParticipantTypes(c, self.Id, self.Id.Parent())
	}
	if self.cachedRequiredParticipantTypesForId == nil {
		return make([]RequiredParticipantType, 0)
	}
	return self.cachedRequiredParticipantTypesForId
}

func (self *Event) getCurrentlyRequiredParticipantTypes(c gaecontext.HTTPContext, participantsByType map[string][]Participant) (currentlyRequiredTypes []RequiredParticipantType) {
	requiredTypeGrouping := make(map[string]RequiredParticipantType)
	var tmp RequiredParticipantType
	var currMin int
	var currMax int
	var currentDepNum int
	accountFor := func(t RequiredParticipantType) {
		if t.PerType == nil || t.PerNum == 0 {
			currMin = t.Min
			currMax = t.Max
		} else {
			currentDepNum = 0
			for _, part := range participantsByType[common.EncKey(t.PerType)] {
				currentDepNum += part.Multiple
			}
			currentDepNum = common.Max(0, currentDepNum-t.Min)
			currMin = currentDepNum / t.PerNum
			currMax = currentDepNum / t.PerNum
		}
		if t.ParticipantType != nil {
			tmp = requiredTypeGrouping[t.ParticipantType.Encode()]
			tmp.ParticipantType = t.ParticipantType
			tmp.Min += currMin
			tmp.Max += currMax
			requiredTypeGrouping[t.ParticipantType.Encode()] = tmp
		}
	}
	for _, t := range self.getRequiredParticipantTypesForType(c) {
		accountFor(t)
	}
	for _, t := range self.getRequiredParticipantTypesForId(c) {
		accountFor(t)
	}
	for _, grouping := range requiredTypeGrouping {
		currentlyRequiredTypes = append(currentlyRequiredTypes, grouping)
	}
	return
}

func (self *Event) GetParticipantsAndRequired(c gaecontext.HTTPContext) (participantsByType map[string][]Participant, currentlyRequiredTypes []RequiredParticipantType) {
	currentParticipants := self.GetParticipants(c)
	participantsByType = make(map[string][]Participant)
	var typeId string
	for _, p := range currentParticipants {
		typeId = common.EncKey(p.ParticipantType)
		participantsByType[typeId] = append(participantsByType[typeId], p)
	}
	currentlyRequiredTypes = self.getCurrentlyRequiredParticipantTypes(c, participantsByType)
	return
}

type RequiredParticipantType struct {
	Id              *datastore.Key `datastore:"-" json:"id"`
	ParticipantType *datastore.Key `json:"participant_type"`
	Min             int            `json:"min"`
	Max             int            `json:"max"`
	PerNum          int            `json:"per_num"`
	PerType         *datastore.Key `json:"per_type"`
}

func (self *RequiredParticipantType) Delete(c gaecontext.HTTPContext, parent, granp, actor *datastore.Key) {
	if !parent.Equal(self.Id.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", parent, self.Id))
	}
	if !granp.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", granp, parent))
	}
	if err := datastore.Delete(c, self.Id); err != nil {
		panic(err)
	}
	if self.Id.Parent().Kind() == "Event" {
		CreateChange(c, self.Id.Parent(), actor, "DeleteRequiredParticipantType", self)
	}
	common.MemDel(c, requiredParticipantTypesKeyForParent(self.Id.Parent()))
}

func findRequiredParticipantTypes(c gaecontext.HTTPContext, parent *datastore.Key) (result []RequiredParticipantType) {
	keys, err := datastore.NewQuery("RequiredParticipantType").Ancestor(parent).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	if result == nil {
		result = make([]RequiredParticipantType, 0)
	}
	return
}

func GetRequiredParticipantType(c gaecontext.HTTPContext, id, parent, dom *datastore.Key) *RequiredParticipantType {
	for _, req := range GetRequiredParticipantTypes(c, parent, dom) {
		if req.Id.Equal(id) {
			return &req
		}
	}
	return nil
}

func GetRequiredParticipantTypes(c gaecontext.HTTPContext, parent *datastore.Key, domain *datastore.Key) (result []RequiredParticipantType) {
	if parent == nil {
		return nil
	}
	if !domain.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", domain, parent))
	}
	common.Memoize(c, requiredParticipantTypesKeyForParent(parent), &result, func() interface{} {
		return findRequiredParticipantTypes(c, parent)
	})
	return
}

func (self *RequiredParticipantType) Key(c gaecontext.HTTPContext, parent *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "RequiredParticipantType", fmt.Sprintf("%+v", self), 0, parent)
}

func (self *RequiredParticipantType) Save(c gaecontext.HTTPContext, parent *datastore.Key, dom, actor *datastore.Key) *RequiredParticipantType {
	if !dom.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, parent))
	}
	if self.Min > self.Max {
		self.Max = self.Min
	}
	isNew := false
	if self.Id == nil {
		isNew = true
	} else {
		if self.Id.Parent().Kind() == "Event" {
			CreateChange(c, self.Id.Parent(), actor, "UpdateRequiredParticipantType", self)
		}
	}
	var err error
	self.Id, err = datastore.Put(c, datastore.NewKey(c, "RequiredParticipantType", "", 0, parent), self)
	if err != nil {
		panic(err)
	}
	if isNew && self.Id.Parent().Kind() == "Event" {
		CreateChange(c, self.Id.Parent(), actor, "CreateRequiredParticipantType", self)
	}
	common.MemDel(c, requiredParticipantTypesKeyForParent(parent))
	return self
}

type ParticipantType struct {
	Id        *datastore.Key `datastore:"-" json:"id"`
	IsContact bool           `json:"is_contact"`
	Name      string         `json:"name"`

	// Salary mod
	SalarySerializedProperties []byte                 `json:"-"`
	SalaryProperties           map[string]interface{} `json:"salary_properties" datastore:"-"`
}

func (self *ParticipantType) process(c gaecontext.HTTPContext) *ParticipantType {
	if len(self.SalarySerializedProperties) > 0 {
		if err := json.Unmarshal(self.SalarySerializedProperties, &self.SalaryProperties); err != nil {
			panic(err)
		}
	}
	return self
}

func findParticipantType(c gaecontext.HTTPContext, key *datastore.Key) *ParticipantType {
	var t ParticipantType
	if err := datastore.Get(c, key, &t); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil
		}
		panic(err)
	}
	t.Id = key
	return &t
}

func (self *Event) GetParticipantType(c gaecontext.HTTPContext, key, dom *datastore.Key) *ParticipantType {
	if self.cachedParticipantTypes == nil {
		self.cachedParticipantTypes = make(map[string]*ParticipantType)
	}
	if _, ok := self.cachedParticipantTypes[key.Encode()]; !ok {
		self.cachedParticipantTypes[key.Encode()] = GetParticipantType(c, key, dom)
	}
	return self.cachedParticipantTypes[key.Encode()]
}

func GetParticipantType(c gaecontext.HTTPContext, key, dom *datastore.Key) *ParticipantType {
	if !key.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, key))
	}
	if key == nil {
		return nil
	}
	var t ParticipantType
	if common.Memoize(c, participantTypeKeyForId(key), &t, func() interface{} {
		return findParticipantType(c, key)
	}) {
		return (&t).process(c)
	}
	return nil
}

func (self *ParticipantType) Save(c gaecontext.HTTPContext, dom *datastore.Key) *ParticipantType {
	var err error

	self.SalarySerializedProperties, err = json.Marshal(self.SalaryProperties)
	if err != nil {
		panic(err)
	}

	self.Id, err = datastore.Put(c, datastore.NewKey(c, "ParticipantType", self.Name, 0, dom), self)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, participantTypesKeyForDomain(dom), participantTypeKeyForId(self.Id))
	return self
}

func DeleteParticipantType(c gaecontext.HTTPContext, key, dom *datastore.Key) {
	if !dom.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, key))
	}
	if err := datastore.Delete(c, key); err != nil {
		panic(err)
	}
	common.MemDel(c, participantTypesKeyForDomain(dom), participantTypeKeyForId(key))
}

func findParticipantTypes(c gaecontext.HTTPContext, dom *datastore.Key) (result []ParticipantType) {
	keys, err := datastore.NewQuery("ParticipantType").Ancestor(dom).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	if result == nil {
		result = make([]ParticipantType, 0)
	}
	return
}

func GetParticipantTypes(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer) (result []ParticipantType) {
	var preResult []ParticipantType
	common.Memoize(c, participantTypesKeyForDomain(dom), &preResult, func() interface{} {
		return findParticipantTypes(c, dom)
	})
	hasTypeAuth := authorizer == nil || authorizer.HasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	})
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
		authMatchAttend.ParticipantType = typ.Id
		authMatchParticipants.ParticipantType = typ.Id
		authMatchAttest.ParticipantType = typ.Id
		authMatchReport.ParticipantType = typ.Id
		if authorizer == nil || (hasTypeAuth || authorizer.HasAnyAuth(authMatchParticipants) || authorizer.HasAnyAuth(authMatchAttend) || authorizer.HasAnyAuth(authMatchAttest) || authorizer.HasAnyAuth(authMatchReport)) {
			(&typ).process(c)
			result = append(result, typ)
		}
	}
	for index, _ := range result {
		(&result[index]).process(c)
	}
	return
}
