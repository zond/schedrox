package web

import (
	"fmt"
	"io/ioutil"
	"github.com/zond/schedrox/appuser"
	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/domain"
	"github.com/zond/schedrox/event"
	"github.com/zond/schedrox/salary"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
)

func updateSalaryConfig(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.SalaryConfiguration,
		Write:    true,
	}) {
		var newConfig salary.Config
		common.MustDecodeJSON(r.Body, &newConfig)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, (&newConfig).Save(data.context, data.domain))
	}
	return
}

func setSalaryCode(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	if err := r.ParseMultipartForm(1024 * 1024); err != nil {
		panic(err)
	}
	multipart := r.MultipartForm
	defer func() {
		multipart.RemoveAll()
	}()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.SalaryReport,
	}) {
		if len(multipart.File["datafile"]) > 0 {
			file, err := multipart.File["datafile"][0].Open()
			if err != nil {
				panic(err)
			}
			defer func() {
				file.Close()
			}()
			bytes, err := ioutil.ReadAll(file)
			if err != nil {
				panic(err)
			}
			if len(bytes) > 0 {
				conf := salary.GetConfig(data.context, data.domain)
				conf.SalaryCode = string(bytes)
				conf.Save(data.context, data.domain)
			}
		}
		w.Header().Set("Location", "/salaries/configuration")
		w.WriteHeader(303)
	}
	return
}

func getSalaryCode(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.SalaryReport,
	}) {
		common.SetContentType(w, "application/javascript; charset=UTF-8", false)
		fmt.Fprint(w, salary.GetConfig(data.context, data.domain).SalaryCode)
	}
	return
}

func getSalaryConfig(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.SalaryConfiguration,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, salary.GetConfig(data.context, data.domain))
	}
	return
}

func setReportFinished(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Attest,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		salary.SetPeriodReported(data.context, data.domain, from, to, key)
		w.WriteHeader(204)
	}
	return
}

func unsetReportFinished(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Attest,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		salary.DeletePeriodReported(data.context, data.domain, from, to, key)
		w.WriteHeader(204)
	}
	return
}

func setMyReportFinished(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		salary.SetPeriodReported(data.context, data.domain, from, to, data.user.Id)
		w.WriteHeader(204)
	}
	return
}

type reportedHoursResponse struct {
	Finished bool          `json:"finished"`
	Events   []event.Event `json:"events"`
}

func getReportedHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, reportedHoursResponse{
			Events:   append(salary.GetReportedForUser(data.context, data.domain, from, to, data.user.Id, nil), event.GetAttestableEventsForUserBetween(data.context, data.domain, nil, data.user.Id, from, to, true)...),
			Finished: salary.GetPeriodReported(data.context, data.domain, from, to, data.user.Id),
		})
	}
	return
}

func removeReportedHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	user_id, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
	if err != nil {
		panic(err)
	}
	key, err := datastore.DecodeKey(mux.Vars(r)["reported_id"])
	if err != nil {
		panic(err)
	}
	reported := salary.GetReportedById(c, data.domain, user_id, key)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Attest,
		Location:        reported.Location,
		EventKind:       reported.EventKind,
		EventType:       reported.EventType,
		ParticipantType: reported.SalaryAttestedParticipantType,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		salary.DeleteReportedForUser(data.context, data.domain, user_id, key, from, to)
		w.WriteHeader(204)
	}
	return
}

func removeMyReportedHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		salary.DeleteReportedForUser(data.context, data.domain, data.user.Id, key, from, to)
		w.WriteHeader(204)
	}
	return
}

func addReportedHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var ev event.Event
	common.MustDecodeJSON(r.Body, &ev)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Attest,
		Location:        ev.Location,
		EventKind:       ev.EventKind,
		EventType:       ev.EventType,
		ParticipantType: ev.SalaryAttestedParticipantType,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		ev.CreatedAt = time.Now()
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, salary.AddReported(data.context, data.domain, from, to, key, &ev))
	}
	return
}

func addMyReportedHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var ev event.Event
	common.MustDecodeJSON(r.Body, &ev)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.ReportHours,
		Location:        ev.Location,
		EventKind:       ev.EventKind,
		EventType:       ev.EventType,
		ParticipantType: ev.SalaryAttestedParticipantType,
	}) {
		ev.CreatedAt = time.Now()
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, salary.AddReported(data.context, data.domain, from, to, data.user.Id, &ev))
	}
	return
}

func setUserAttestedHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Attest,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		var events event.Events
		common.MustDecodeJSON(r.Body, &events)
		authorizedEvents := event.Events{}
		for _, event := range events {
			if data.authorizer.HasAuth(auth.Auth{
				AuthType:        auth.Attest,
				Location:        event.Location,
				EventKind:       event.EventKind,
				EventType:       event.EventType,
				ParticipantType: event.SalaryAttestedParticipantType,
			}) {
				authorizedEvents = append(authorizedEvents, event)
			}
		}
		salary.SetAttestedForUser(data.context, data.domain, from, to, key, data.user.Id, authorizedEvents)
		w.WriteHeader(204)
	}
	return
}

type userAttestableResponse struct {
	Finished bool         `json:"finished"`
	Events   event.Events `json:"events"`
}

func getUserAttestedHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Attest,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		attested := salary.GetAttestedForUser(data.context, data.domain, from, to, key, data.authorizer)
		sort.Sort(attested)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, attested)
	}
	return
}

func getUserAttestableHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Attest,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		resp := userAttestableResponse{
			Finished: salary.GetPeriodReported(data.context, data.domain, from, to, key),
		}
		attestedEvents := map[string]bool{}
		for _, ev := range salary.GetAttestedForUser(data.context, data.domain, from, to, key, data.authorizer) {
			attestedEvents[ev.AttestUUID()] = true
		}
		unattested := event.Events{}
		for _, ev := range salary.GetReportedForUser(data.context, data.domain, from, to, key, data.authorizer) {
			if !attestedEvents[ev.AttestUUID()] {
				unattested = append(unattested, ev)
			}
		}
		for _, ev := range event.GetAttestableEventsForUserBetween(data.context, data.domain, data.authorizer, key, from, to, false) {
			if !attestedEvents[ev.AttestUUID()] {
				unattested = append(unattested, ev)
			}
		}
		sort.Sort(unattested)
		resp.Events = unattested
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, resp)
	}
	return
}

func deleteUserAttestedHours(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Attest,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		salary.DeleteAttestedForUser(data.context, data.domain, from, to, key, data.authorizer)
		w.WriteHeader(204)
	}
	return
}

type salaryReport struct {
	Users            map[string]*appuser.User         `json:"users"`
	EventTypes       map[string]event.EventType       `json:"event_types"`
	EventKinds       map[string]event.EventKind       `json:"event_kinds"`
	Locations        map[string]domain.Location       `json:"locations"`
	ParticipantTypes map[string]event.ParticipantType `json:"participant_types"`
	Events           event.Events                     `json:"events"`
	Missing          event.Events                     `json:"missing"`
	Removed          event.Events                     `json:"removed"`
	Changed          event.Events                     `json:"changed"`
	From             time.Time                        `json:"from"`
	To               time.Time                        `json:"to"`
}

func safeEncode(k *datastore.Key) string {
	if k == nil {
		return ""
	}
	return k.Encode()
}

func getSalaryReport(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.SalaryReport,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)

		resp := salaryReport{
			Users:            make(map[string]*appuser.User),
			EventTypes:       make(map[string]event.EventType),
			EventKinds:       make(map[string]event.EventKind),
			Locations:        make(map[string]domain.Location),
			ParticipantTypes: make(map[string]event.ParticipantType),
			Events:           make(event.Events, 0),
			Missing:          make(event.Events, 0),
			Removed:          make(event.Events, 0),
			Changed:          make(event.Events, 0),
			From:             from,
			To:               to,
		}

		attestedEvents := make(map[string]time.Time)

		userIds := make(map[string]bool)
		eventTypeIds := make(map[string]bool)
		eventKindIds := make(map[string]bool)
		locationIds := make(map[string]bool)
		participantTypeIds := make(map[string]bool)

		for _, ev := range salary.GetAttested(data.context, data.domain, from, to) {
			if data.authorizer.HasAuth(auth.Auth{
				AuthType:        auth.SalaryReport,
				Location:        ev.Location,
				EventKind:       ev.EventKind,
				EventType:       ev.EventType,
				ParticipantType: ev.SalaryAttestedParticipantType,
			}) {
				resp.Events = append(resp.Events, ev)
				userIds[safeEncode(ev.SalaryAttestedUser)] = true
				participantTypeIds[safeEncode(ev.SalaryAttestedParticipantType)] = true
				eventTypeIds[safeEncode(ev.EventType)] = true
				eventKindIds[safeEncode(ev.EventKind)] = true
				locationIds[safeEncode(ev.Location)] = true

				attestedEvents[safeEncode(ev.SalaryAttestedEvent)] = ev.UpdatedAt
			}
		}

		foundEvents := make(map[string]bool)

		for _, ev := range event.GetReportableEventsBetween(data.context, data.domain, data.authorizer, from, to) {
			foundEvents[ev.Id.Encode()] = true
			if attestedAt, found := attestedEvents[ev.Id.Encode()]; found {
				if !attestedAt.Equal(ev.UpdatedAt) {
					resp.Changed = append(resp.Changed, ev)
				}
			} else {
				resp.Missing = append(resp.Missing, ev)
			}
		}

		for _, ev := range salary.GetAllowedReported(data.context, data.domain, from, to, auth.SalaryReport, data.authorizer) {
			foundEvents[ev.Id.Encode()] = true
			if attestedAt, found := attestedEvents[ev.Id.Encode()]; found {
				if !attestedAt.Equal(ev.UpdatedAt) {
					resp.Changed = append(resp.Changed, ev)
				}
			} else {
				resp.Missing = append(resp.Missing, ev)
			}
		}

		for _, ev := range resp.Events {
			if !foundEvents[ev.SalaryAttestedEvent.Encode()] {
				resp.Removed = append(resp.Removed, ev)
			}
		}

		for _, user := range appuser.GetUsersByDomain(data.context, data.domain) {
			userCopy := user
			if userIds[user.Id.Encode()] {
				resp.Users[user.Id.Encode()] = &userCopy
			}
		}
		for _, eventType := range event.GetEventTypes(data.context, data.domain, nil) {
			if eventTypeIds[eventType.Id.Encode()] {
				resp.EventTypes[eventType.Id.Encode()] = eventType
			}
		}
		for _, eventKind := range event.GetEventKinds(data.context, data.domain, nil) {
			if eventKindIds[eventKind.Id.Encode()] {
				resp.EventKinds[eventKind.Id.Encode()] = eventKind
			}
		}
		for _, location := range domain.GetLocations(data.context, data.domain, nil) {
			if locationIds[location.Id.Encode()] {
				resp.Locations[location.Id.Encode()] = location
			}
		}
		for _, partType := range event.GetParticipantTypes(data.context, data.domain, nil) {
			if participantTypeIds[partType.Id.Encode()] {
				resp.ParticipantTypes[partType.Id.Encode()] = partType
			}
		}
		for _, prop := range appuser.GetAllUserProperties(data.context, data.domain) {
			if user, existed := resp.Users[datastore.NewKey(c, "User", prop.Id.Parent().StringID(), 0, nil).Encode()]; existed {
				if user.CachedProperties == nil {
					user.CachedProperties = make(map[string]interface{})
				}
				user.CachedProperties[prop.Name] = prop
			}
		}

		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, resp)
	}
	return
}
