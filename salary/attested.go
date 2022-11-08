package salary

import (
	"fmt"
	"monotone/se.oort.schedrox/auth"
	"monotone/se.oort.schedrox/common"
	"monotone/se.oort.schedrox/event"
	"sort"
	"time"

	"github.com/zond/sybutils/utils/gae/gaecontext"
	"google.golang.org/appengine/datastore"
)

func attestedIdForDomainAndPeriod(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time) *datastore.Key {
	return datastore.NewKey(c, "PeriodAttests", fmt.Sprintf("%v-%v", from, to), 0, dom)
}

func attestedIdForDomainPeriodAndUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "Attests", user.StringID(), 0, attestedIdForDomainAndPeriod(c, dom, from, to))
}

func attestedKeyForDomainAndPeriod(dom *datastore.Key, from, to time.Time) string {
	return fmt.Sprintf("PeriodAttests{Domain:%v,From:%v,To:%v}", dom, from, to)
}

func attestedKeyForDomainPeriodAndUser(dom *datastore.Key, from, to time.Time, user *datastore.Key) string {
	return fmt.Sprintf("Attests{Domain:%v,From:%v,To:%v,User:%v}", dom, from, to, user)
}

func attestedKeyForEvent(k *datastore.Key) string {
	return fmt.Sprintf("Attests{SalaryAttestedEvent:%v}", k)
}

func SetAttestedForUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user, attester *datastore.Key, events []event.Event) {
	keys := make([]*datastore.Key, len(events))
	for index, ev := range events {
		keys[index] = datastore.NewKey(c, "Attest", ev.Id.StringID(), ev.Id.IntID(), attestedIdForDomainPeriodAndUser(c, dom, from, to, user))
		(&events[index]).InformationBytes = []byte(events[index].Information)
		(&events[index]).Information = ""
		(&events[index]).SalaryAttestedEvent = ev.Id
		(&events[index]).SalaryAttester = attester
		(&events[index]).SalaryAttestedUser = user
		(&events[index]).SalaryAttestedAt = time.Now()
		(&events[index]).RecurrenceExceptionsBytes = []byte(events[index].RecurrenceExceptions)
		(&events[index]).RecurrenceExceptions = ""
		common.MemDel(c, attestedKeyForEvent(ev.Id))
	}
	_, err := datastore.PutMulti(c, keys, events)
	common.AssertOkError(err)
	common.MemDel(c, attestedKeyForDomainAndPeriod(dom, from, to))
	common.MemDel(c, attestedKeyForDomainPeriodAndUser(dom, from, to, user))
}

func DeleteAttestedForUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key, authorizer auth.Authorizer) {
	attested := GetAttestedForUser(c, dom, from, to, user, authorizer)
	keys := make([]*datastore.Key, 0, len(attested))
	for _, att := range attested {
		keys = append(keys, att.Id)
		common.MemDel(c, attestedKeyForEvent(att.SalaryAttestedEvent))
	}
	err := datastore.DeleteMulti(c, keys)
	common.AssertOkError(err)
	common.MemDel(c, attestedKeyForDomainAndPeriod(dom, from, to))
	common.MemDel(c, attestedKeyForDomainPeriodAndUser(dom, from, to, user))
}

func findAttestedForEvent(c gaecontext.HTTPContext, ev *datastore.Key) (result event.Events) {
	ids, err := datastore.NewQuery("Attest").Ancestor(ev.Parent()).Filter("SalaryAttestedEvent=", ev).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		(&result[index]).Id = id
	}
	return
}

func GetAttestedForEvent(c gaecontext.HTTPContext, ev *datastore.Key) (result event.Events) {
	common.Memoize(c, attestedKeyForEvent(ev), &result, func() interface{} {
		return findAttestedForEvent(c, ev)
	})
	if result == nil {
		result = event.Events{}
	}
	return
}

func findAttestedForUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) (result event.Events) {
	ids, err := datastore.NewQuery("Attest").Ancestor(attestedIdForDomainPeriodAndUser(c, dom, from, to, user)).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		(&result[index]).Id = id
	}
	return
}

func GetAttestedForUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key, authorizer auth.Authorizer) (result event.Events) {
	var preResult event.Events
	common.Memoize(c, attestedKeyForDomainPeriodAndUser(dom, from, to, user), &preResult, func() interface{} {
		return findAttestedForUser(c, dom, from, to, user)
	})
	matchAuth := auth.Auth{
		AuthType: auth.Attest,
	}
	for _, ev := range preResult {
		if !ev.Start.After(to) && !ev.End.Before(from) {
			matchAuth.Location = ev.Location
			matchAuth.EventKind = ev.EventKind
			matchAuth.EventType = ev.EventType
			matchAuth.ParticipantType = ev.SalaryAttestedParticipantType
			if authorizer == nil || authorizer.HasAuth(matchAuth) {
				cpy := ev
				if cpy.SalaryAttester != nil {
					cpy.SalaryAttesterEmail = cpy.SalaryAttester.StringID()
				}
				(&cpy).QuickProcess(c)
				result = append(result, cpy)
			}
		}
	}
	if result == nil {
		result = make(event.Events, 0)
	}
	sort.Sort(result)
	return
}

func findAttested(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time) (result event.Events) {
	ids, err := datastore.NewQuery("Attest").Ancestor(attestedIdForDomainAndPeriod(c, dom, from, to)).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		(&result[index]).Id = id
	}
	return
}

func GetAttested(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time) (result event.Events) {
	var preResult event.Events
	common.Memoize(c, attestedKeyForDomainAndPeriod(dom, from, to), &preResult, func() interface{} {
		return findAttested(c, dom, from, to)
	})
	for index, _ := range preResult {
		(&preResult[index]).QuickProcess(c)
	}
	if preResult == nil {
		preResult = make(event.Events, 0)
	}
	for _, ev := range preResult {
		if !ev.Start.After(to) && !ev.End.Before(from) {
			result = append(result, ev)
		}
	}
	sort.Sort(result)
	return
}
