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
	"google.golang.org/appengine/log"
)

func periodReportedIdForDomainAndPeriod(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time) *datastore.Key {
	return datastore.NewKey(c, "PeriodIsReported", fmt.Sprintf("%v-%v", from, to), 0, dom)
}

func reportedIdForDomainAndPeriod(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time) *datastore.Key {
	return datastore.NewKey(c, "PeriodTimeReports", fmt.Sprintf("%v-%v", from, to), 0, dom)
}

func periodReportedIdForDomainPeriodAndUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "IsReported", user.StringID(), 0, periodReportedIdForDomainAndPeriod(c, dom, from, to))
}

func reportedIdForDomainPeriodAndUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "TimeReports", user.StringID(), 0, reportedIdForDomainAndPeriod(c, dom, from, to))
}

func reportedKeyForDomainAndPeriod(dom *datastore.Key, from, to time.Time) string {
	return fmt.Sprintf("PeriodTimeReports{Domain:%v,From:%v,To:%v}", dom, from, to)
}

func periodReportedKeyForDomainPeriodAndUser(dom *datastore.Key, from, to time.Time, user *datastore.Key) string {
	return fmt.Sprintf("IsReported{Domain:%v,From:%v,To:%v,User:%v}", dom, from, to, user)
}

func reportedKeyForDomainPeriodAndUser(dom *datastore.Key, from, to time.Time, user *datastore.Key) string {
	return fmt.Sprintf("TimeReports{Domain:%v,From:%v,To:%v,User:%v}", dom, from, to, user)
}

func reportedKeyForId(k *datastore.Key) string {
	return fmt.Sprintf("TimeReport{Id:%v}", k)
}

func DeleteEncodedKeys(c gaecontext.HTTPContext) (deleted int) {
	timeReports := []event.Event{}
	timeReportIds, err := datastore.NewQuery("TimeReport").GetAll(c, &timeReports)
	common.AssertOkError(err)

	for index, id := range timeReportIds {
		userKey, err := datastore.DecodeKey(id.Parent().StringID())
		if err == nil {
			newId := datastore.NewKey(c, "TimeReport", id.StringID(), id.IntID(), datastore.NewKey(c, "TimeReports", userKey.StringID(), 0, id.Parent().Parent()))
			if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
				if _, err = datastore.Put(c, newId, &timeReports[index]); err != nil {
					return
				}
				if err = datastore.Delete(c, id); err != nil {
					return
				}
				deleted++
				log.Infof(c, "### replaced %v with %v", id, newId)
				return nil
			}, true); err != nil {
				panic(err)
			}
		}
	}

	attests := []event.Event{}
	attestIds, err := datastore.NewQuery("Attest").GetAll(c, &attests)
	common.AssertOkError(err)

	for index, id := range attestIds {
		userKey, err := datastore.DecodeKey(id.Parent().StringID())
		if err == nil {
			newId := datastore.NewKey(c, "Attest", id.StringID(), id.IntID(), datastore.NewKey(c, "Attests", userKey.StringID(), 0, id.Parent().Parent()))
			if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
				if _, err = datastore.Put(c, newId, &attests[index]); err != nil {
					return
				}
				if err = datastore.Delete(c, id); err != nil {
					return
				}
				deleted++
				log.Infof(c, "### replaced %v with %v", id, newId)
				return nil
			}, true); err != nil {
				panic(err)
			}
		}
	}

	attests = []event.Event{}
	attestIds, err = datastore.NewQuery("Attest").GetAll(c, &attests)
	common.AssertOkError(err)

	for index, id := range attestIds {
		eventKey, err := datastore.DecodeKey(id.StringID())
		if err == nil {
			newId := datastore.NewKey(c, "Attest", eventKey.StringID(), eventKey.IntID(), id.Parent())
			if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
				if _, err = datastore.Put(c, newId, &attests[index]); err != nil {
					return
				}
				if err = datastore.Delete(c, id); err != nil {
					return
				}
				deleted++
				log.Infof(c, "### replaced %v with %v", id, newId)
				return nil
			}, true); err != nil {
				panic(err)
			}
		}
	}

	isReporteds := []struct{}{}
	isReportedIds, err := datastore.NewQuery("IsReported").GetAll(c, &isReporteds)
	common.AssertOkError(err)

	for index, id := range isReportedIds {
		userKey, err := datastore.DecodeKey(id.StringID())
		if err == nil {
			newId := datastore.NewKey(c, "IsReported", userKey.StringID(), 0, id.Parent())
			if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
				if _, err = datastore.Put(c, newId, &isReporteds[index]); err != nil {
					return
				}
				if err = datastore.Delete(c, id); err != nil {
					return
				}
				deleted++
				c.Infof("### replaced %v with %v", id, newId)
				return nil
			}, true); err != nil {
				panic(err)
			}
		}
	}
	return
}

func findPeriodReported(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) *struct{} {
	if err := datastore.Get(c, periodReportedIdForDomainPeriodAndUser(c, dom, from, to, user), &struct{}{}); err == nil {
		return &struct{}{}
	} else if err == datastore.ErrNoSuchEntity {
		return nil
	} else {
		common.AssertOkError(err)
	}
	return nil
}

func GetPeriodReported(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) bool {
	if common.Memoize(c, periodReportedKeyForDomainPeriodAndUser(dom, from, to, user), &struct{}{}, func() interface{} {
		return findPeriodReported(c, dom, from, to, user)
	}) {
		return true
	}
	return false
}

func SetPeriodReported(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) {
	_, err := datastore.Put(c, periodReportedIdForDomainPeriodAndUser(c, dom, from, to, user), &struct{}{})
	common.AssertOkError(err)
	common.MemDel(c, periodReportedKeyForDomainPeriodAndUser(dom, from, to, user))
}

func DeletePeriodReported(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) {
	err := datastore.Delete(c, periodReportedIdForDomainPeriodAndUser(c, dom, from, to, user))
	common.AssertOkError(err)
	common.MemDel(c, periodReportedKeyForDomainPeriodAndUser(dom, from, to, user))
}

func AddReported(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key, event *event.Event) *event.Event {
	var err error
	event.SalaryTimeReported = true
	event.InformationBytes = []byte(event.Information)
	event.Information = ""
	// Set the user so that we know who this belonged to when fetching in bulk
	event.SalaryAttestedUser = user
	event.RecurrenceExceptionsBytes = []byte(event.RecurrenceExceptions)
	event.RecurrenceExceptions = ""
	event.Id, err = datastore.Put(c, datastore.NewKey(c, "TimeReport", "", 0, reportedIdForDomainPeriodAndUser(c, dom, from, to, user)), event)
	common.AssertOkError(err)
	common.MemDel(c, reportedKeyForDomainPeriodAndUser(dom, from, to, user), reportedKeyForDomainAndPeriod(dom, from, to), reportedKeyForId(event.Id))
	return event.QuickProcess(c)
}

func findReportedById(c gaecontext.HTTPContext, key *datastore.Key) *event.Event {
	var result event.Event
	err := datastore.Get(c, key, &result)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	result.Id = key
	return &result
}

func GetReportedById(c gaecontext.HTTPContext, dom, user, key *datastore.Key) *event.Event {
	if user.StringID() != key.Parent().StringID() {
		panic(fmt.Errorf("%v does not belong to %v", key, user))
	}
	if !dom.Equal(key.Parent().Parent().Parent()) {
		panic(fmt.Errorf("%v does not belong to %v", key, dom))
	}
	var result event.Event
	if common.Memoize(c, reportedKeyForId(key), &result, func() interface{} {
		return findReportedById(c, key)
	}) {
		return (&result).QuickProcess(c)
	}
	return nil
}

func DeleteReportedForUser(c gaecontext.HTTPContext, dom, user, key *datastore.Key, from, to time.Time) {
	if user.StringID() != key.Parent().StringID() {
		panic(fmt.Errorf("%v does not belong to %v", key, user))
	}
	if !dom.Equal(key.Parent().Parent().Parent()) {
		panic(fmt.Errorf("%v does not belong to %v", key, dom))
	}
	err := datastore.Delete(c, key)
	common.AssertOkError(err)
	common.MemDel(c, reportedKeyForDomainPeriodAndUser(dom, from, to, user), reportedKeyForDomainAndPeriod(dom, from, to), reportedKeyForId(key))
}

func findReportedForUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key) (result event.Events) {
	ids, err := datastore.NewQuery("TimeReport").Ancestor(reportedIdForDomainPeriodAndUser(c, dom, from, to, user)).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		(&result[index]).Id = id
		(&result[index]).SalaryTimeReported = true
	}
	return
}

type reportedEvents event.Events

func (self reportedEvents) Len() int {
	return len(self)
}

func (self reportedEvents) Less(j, i int) bool {
	return self[i].CreatedAt.Before(self[j].CreatedAt)
}

func (self reportedEvents) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func GetReportedForUser(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, user *datastore.Key, authorizer auth.Authorizer) (result event.Events) {
	c.Infof("Getting attested events between %v and %v", from, to)
	var preResult event.Events
	common.Memoize(c, reportedKeyForDomainPeriodAndUser(dom, from, to, user), &preResult, func() interface{} {
		return findReportedForUser(c, dom, from, to, user)
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
			viewAuth := matchAuth
			viewAuth.AuthType = auth.Events
			viewAuth.ParticipantType = nil
			if authorizer == nil || authorizer.HasAuth(matchAuth) || authorizer.HasAuth(viewAuth) {
				cpy := ev
				cpy.SalaryTimeReported = true
				cpy.SalaryAttestedUser = user
				cpy.SalaryAttestedEvent = ev.Id
				(&cpy).QuickProcess(c)
				result = append(result, cpy)
			}
		}
	}
	if result == nil {
		result = make(event.Events, 0)
	}
	sort.Sort(reportedEvents(result))
	return
}

func findReported(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time) (result event.Events) {
	ids, err := datastore.NewQuery("TimeReport").Ancestor(reportedIdForDomainAndPeriod(c, dom, from, to)).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		(&result[index]).Id = id
	}
	return
}

func GetAllowedReported(c gaecontext.HTTPContext, dom *datastore.Key, from, to time.Time, authType string, authorizer auth.Authorizer) (result reportedEvents) {
	var preResult event.Events
	common.Memoize(c, reportedKeyForDomainAndPeriod(dom, from, to), &preResult, func() interface{} {
		return findReported(c, dom, from, to)
	})
	matchAuth := auth.Auth{
		AuthType: authType,
	}
	var user *datastore.Key
	for _, ev := range preResult {
		if !ev.Start.After(to) && !ev.End.Before(from) {
			matchAuth.Location = ev.Location
			matchAuth.EventKind = ev.EventKind
			matchAuth.EventType = ev.EventType
			matchAuth.ParticipantType = ev.SalaryAttestedParticipantType
			if authorizer == nil || authorizer.HasAuth(matchAuth) {
				cpy := ev
				user = datastore.NewKey(c, "User", cpy.Id.Parent().StringID(), 0, nil)
				cpy.SalaryAttestedUser = user
				cpy.SalaryAttestedEvent = ev.Id
				(&cpy).QuickProcess(c)
				result = append(result, cpy)
			}
		}
	}
	if result == nil {
		result = make(reportedEvents, 0)
	}
	sort.Sort(result)
	return
}
