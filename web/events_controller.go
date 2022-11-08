package web

import (
	"fmt"
	"github.com/zond/schedrox/appuser"
	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/crm"
	"github.com/zond/schedrox/domain"
	"github.com/zond/schedrox/event"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
)

type isBusyRequest struct {
	Start       time.Time      `json:"start"`
	End         time.Time      `json:"end"`
	IgnoreEvent *datastore.Key `json:"ignore_event"`
}

func getIsBusy(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var request isBusyRequest
	common.MustDecodeJSON(r.Body, &request)
	if data.User() != nil {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, map[string]interface{}{
			"busy": event.IsBusyUserId(data.context, request.Start, request.End, request.IgnoreEvent, data.User().Id),
		})
	} else {
		w.WriteHeader(401)
		fmt.Fprintln(w, "Unauthenticated")
	}
	return
}

type confirmationRequest struct {
	Participant event.Participant `json:"participant"`
	Event       event.Event       `json:"event"`
}

func sendConfirmation(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var request confirmationRequest
	common.MustDecodeJSON(r.Body, &request)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Participants,
		Location:        request.Event.Location,
		EventKind:       request.Event.EventKind,
		EventType:       request.Event.EventType,
		ParticipantType: request.Participant.ParticipantType,
	}) {
		event_type := event.GetEventType(data.context, request.Event.EventType)
		if request.Participant.User != nil && !request.Participant.User.Parent().Equal(data.domain) {
			panic(fmt.Errorf("%v is not parent of %v", data.domain, request.Participant.User))
		}
		if request.Participant.Contact != nil && !request.Participant.Contact.Parent().Equal(data.domain) {
			panic(fmt.Errorf("%v is not parent of %v", data.domain, request.Participant.Contact))
		}
		contact := (&request.Participant).GetContact(c)
		user := (&request.Participant).GetUser(c)
		subject, body, err := event.GenerateConfirmation(
			data.context,
			event_type.ConfirmationEmailSubjectTemplate, event_type.ConfirmationEmailBodyTemplate,
			&request.Participant,
			contact,
			user,
			&request.Event,
		)
		if err != nil {
			panic(err)
		}
		dom := domain.GetDomain(data.context, data.domain)
		if request.Participant.User != nil {
			user := appuser.GetUserByKey(data.context, request.Participant.User)
			user.SendMail(data.context, data.domain, subject, body, dom.ExtraConfirmationBCC, nil)
		} else {
			contact := crm.GetContact(data.context, request.Participant.Contact, data.domain)
			contact.SendMail(data.context, subject, body, dom.ExtraConfirmationBCC)
		}
	}
	return
}

type exampleConfirmationRequest struct {
	SubjectTemplate string `json:"subject_template"`
	BodyTemplate    string `json:"body_template"`
}

func exampleConfirmation(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var request exampleConfirmationRequest
	common.MustDecodeJSON(r.Body, &request)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		subject, body, _ := event.CreateConfirmationExample(data.context, request.SubjectTemplate, request.BodyTemplate)
		common.MustEncodeJSON(w, map[string]interface{}{
			"subject": subject,
			"body":    body,
		})
	}
	return
}

type potentialParticipantsRequest struct {
	Location        *datastore.Key `json:"location"`
	EventKind       *datastore.Key `json:"event_kind"`
	EventType       *datastore.Key `json:"event_type"`
	ParticipantType *datastore.Key `json:"participant_type"`
	Start           string         `json:"start"`
	End             string         `json:"end"`
	IgnoreBusyEvent *datastore.Key `json:"ignore_busy_event"`
}

func getPotentialParticipants(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var request potentialParticipantsRequest
	common.MustDecodeJSON(r.Body, &request)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Participants,
		Location:        request.Location,
		EventKind:       request.EventKind,
		EventType:       request.EventType,
		ParticipantType: request.ParticipantType,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetPotentialParticipants(
			data.context,
			data.domain,
			request.Location,
			request.EventKind,
			request.EventType,
			request.ParticipantType,
			common.MustParseJSTime(request.Start),
			common.MustParseJSTime(request.End),
			request.IgnoreBusyEvent),
		)
	}
	return
}

func getLatestChanges(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Events,
	}) && data.hasAuth(auth.Auth{
		AuthType: auth.Participants,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetLatestChanges(data.context, data.domain))
	}
	return
}

func getEventChanges(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	original := event.GetEvent(data.context, key, data.domain)
	if data.hasAuth(auth.Auth{
		AuthType:  auth.Events,
		Location:  original.Location,
		EventKind: original.EventKind,
		EventType: original.EventType,
		Write:     true,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetChanges(data.context, key, data.domain))
	}
	return
}

func getCurrentAlerts(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		r.ParseForm()
		at := common.MustParseJSTime(r.Form.Get("at"))
		alertMap := make(map[string][]event.Event)
		for _, dom := range data.user.Domains {
			alertMap[common.EncKey(dom.Id)] = event.CurrentAlerts(data.context, dom.Id, at, data.user.Id, data.authorizer)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, alertMap)
	}
	return
}

type uniquenessRequest struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func checkUniqueEvent(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		var req uniquenessRequest
		common.MustDecodeJSON(r.Body, &req)
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		event_type := event.GetEventType(data.context, key)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		if event_type.Unique {
			common.MustEncodeJSON(w, map[string]interface{}{
				"ids": event_type.GetIdsBetween(data.context, data.domain, req.Start, req.End),
			})
		} else {
			common.MustEncodeJSON(w, map[string]interface{}{
				"ids": make([]*datastore.Key, 0),
			})
		}
	}
	return
}

func getEventKind(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		event_kind := event.GetEventKind(data.context, key)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event_kind)
	}
	return
}

func getEventType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		event_type := event.GetEventType(data.context, key)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event_type)
	}
	return
}

func getEvent(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	original := event.GetEvent(data.context, key, data.domain)
	if original == nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "Unknown event %v", key)
		return
	}
	readEvent, authenticated := data.silentHasAuth(auth.Auth{
		AuthType:  auth.Events,
		EventKind: original.EventKind,
		EventType: original.EventType,
		Location:  original.Location,
	})
	attend, _ := data.silentHasAnyAuth(auth.Auth{
		AuthType:  auth.Attend,
		EventKind: original.EventKind,
		EventType: original.EventType,
		Location:  original.Location,
	})
	isParticipant := original.IsParticipant(data.context, data.user.Id)
	if !authenticated || (!readEvent && !attend && !isParticipant) {
		w.WriteHeader(401)
		fmt.Fprintln(w, "Unauthenticated")
	} else {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, original)
	}
	return
}

type eventUserData struct {
	eventID  *datastore.Key
	users    []*appuser.User
	panicErr interface{}
}

type eventUserField struct {
	name      string
	generator func(ev event.Event, u *appuser.User) string
}

type eventContactData struct {
	eventID  *datastore.Key
	contacts []*crm.Contact
	panicErr interface{}
}

type eventContactField struct {
	name      string
	generator func(ev event.Event, c *crm.Contact) string
}

type exportEventData struct {
	eventID          *datastore.Key
	participantTypes []string
	nContacts        int
	nUsers           int
	userProperties   []string
	panicErr         interface{}
}

type exportEventField struct {
	name      string
	generator func(ev event.Event) string
}

func userEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Events,
	}) && data.hasAnyAuth(auth.Auth{
		AuthType: auth.Participants,
	}) && data.hasAnyAuth(auth.Auth{
		AuthType: auth.Users,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("unix_start")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("unix_to")), 0)
		to = to.AddDate(0, 0, 1)
		key, err := datastore.DecodeKey(r.URL.Query().Get("user_id"))
		if err != nil {
			key = nil
		}
		common.SetContentType(w, "text/csv; charset=UTF-8", false)

		events := event.GetFilteredEventsBetween(c, data.domain, from, to, nil, func(ev *event.Event) event.Events {
			if key == nil {
				return event.Events{*ev}
			}
			for _, part := range ev.GetParticipants(c) {
				if part.User != nil && part.User.Equal(key) {
					return event.Events{*ev}
				}
			}
			return nil
		})

		dataChan := make(chan eventUserData)
		for _, ev := range events {
			go func(ev event.Event) {
				d := eventUserData{
					eventID: ev.Id,
				}
				defer func() {
					d.panicErr = recover()
					dataChan <- d
				}()
				for _, participant := range ev.GetParticipants(c) {
					if participant.User != nil && (key == nil || participant.User.Equal(key)) {
						if user := appuser.GetUserByKey(c, participant.User); user != nil {
							user.Process(c)
							d.users = append(d.users, user)
						}
					}
				}
			}(ev)
		}
		dataByEvent := map[string]eventUserData{}
		for _ = range events {
			d := <-dataChan
			if d.panicErr != nil {
				panic(d.panicErr)
			}
			dataByEvent[d.eventID.Encode()] = d
		}

		fields := []eventUserField{
			{
				name: "type",
				generator: func(ev event.Event, u *appuser.User) string {
					return ev.EventType.StringID()
				},
			},
			{
				name: "location",
				generator: func(ev event.Event, u *appuser.User) string {
					return ev.Location.StringID()
				},
			},
			{
				name: "date",
				generator: func(ev event.Event, u *appuser.User) string {
					return ev.Start.Format("20060102")
				},
			},
			{
				name: "time",
				generator: func(ev event.Event, u *appuser.User) string {
					return ev.Start.Format("15:04")
				},
			},
			{
				name: "minutes",
				generator: func(ev event.Event, u *appuser.User) string {
					return fmt.Sprint(int64(ev.End.Sub(ev.Start) / time.Minute))
				},
			},
			{
				name: "userFamilyName",
				generator: func(ev event.Event, u *appuser.User) string {
					return u.FamilyName
				},
			},
			{
				name: "userGivenName",
				generator: func(ev event.Event, u *appuser.User) string {
					return u.GivenName
				},
			},
		}
		for _, prop_ := range domain.GetUserProperties(c, data.domain) {
			prop := prop_
			fields = append(fields, eventUserField{
				name: "userProperty-" + prop.Name,
				generator: func(ev event.Event, u *appuser.User) string {
					hasProp := false
					for _, ownedProp := range u.GetPreCachedProperties(c, data.domain) {
						if prop.Name == ownedProp.Name && ownedProp.AssignedAt.Before(ev.Start) && ownedProp.ValidUntil.After(ev.Start) {
							hasProp = true
							break
						}
					}
					if hasProp {
						return "True"
					}
					return "False"
				},
			})
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")

		fieldNames := []string{}
		for _, field := range fields {
			fieldNames = append(fieldNames, field.name)
		}
		fmt.Fprintln(w, strings.Join(fieldNames, ","))

		for _, ev := range events {
			for _, u := range dataByEvent[ev.Id.Encode()].users {
				line := []string{}
				for _, field := range fields {
					line = append(line, field.generator(ev, u))
				}
				fmt.Fprintln(w, strings.Join(line, ","))
			}
		}
	}
	return
}

func contactEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Events,
	}) && data.hasAnyAuth(auth.Auth{
		AuthType: auth.Participants,
	}) && data.hasAnyAuth(auth.Auth{
		AuthType: auth.Contacts,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("unix_start")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("unix_to")), 0)
		to = to.AddDate(0, 0, 1)
		key, err := datastore.DecodeKey(r.URL.Query().Get("event_type_id"))
		if err != nil {
			key = nil
		}
		common.SetContentType(w, "text/csv; charset=UTF-8", false)

		events := event.GetEventsBetween(c, data.domain, from, to, func(ev *event.Event) event.Events {
			if key != nil && !key.Equal(ev.EventType) {
				return nil
			}
			return event.Events{*ev}
		})

		dataChan := make(chan eventContactData)
		for _, ev := range events {
			go func(ev event.Event) {
				d := eventContactData{
					eventID: ev.Id,
				}
				defer func() {
					d.panicErr = recover()
					dataChan <- d
				}()
				for _, participant := range ev.GetParticipants(c) {
					if participant.Contact != nil {
						d.contacts = append(d.contacts, crm.GetContact(c, participant.Contact, data.domain))
					}
				}
			}(ev)
		}
		dataByEvent := map[string]eventContactData{}
		for _ = range events {
			d := <-dataChan
			if d.panicErr != nil {
				panic(d.panicErr)
			}
			dataByEvent[d.eventID.Encode()] = d
		}

		fields := []eventContactField{
			{
				name: "type",
				generator: func(ev event.Event, c *crm.Contact) string {
					return ev.EventType.StringID()
				},
			},
			{
				name: "location",
				generator: func(ev event.Event, c *crm.Contact) string {
					return ev.Location.StringID()
				},
			},
			{
				name: "date",
				generator: func(ev event.Event, c *crm.Contact) string {
					return ev.Start.Format("20060102")
				},
			},
			{
				name: "time",
				generator: func(ev event.Event, c *crm.Contact) string {
					return ev.Start.Format("15:04")
				},
			},
			{
				name: "minutes",
				generator: func(ev event.Event, c *crm.Contact) string {
					return fmt.Sprint(int64(ev.End.Sub(ev.Start) / time.Minute))
				},
			},
			{
				name: "contactName",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.Name
				},
			},
			{
				name: "contactFamilyName",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.ContactFamilyName
				},
			},
			{
				name: "contactGivenName",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.ContactGivenName
				},
			},
			{
				name: "contactOrgNr",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.OrganizationNumber
				},
			},
			{
				name: "contactEmail",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.Email
				},
			},
			{
				name: "contactMobilePhone",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.MobilePhone
				},
			},
			{
				name: "contactAddressLine1",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.AddressLine1
				},
			},
			{
				name: "contactAddressLine2",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.AddressLine2
				},
			},
			{
				name: "contactAddressLine3",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.AddressLine3
				},
			},
			{
				name: "contactBillingAddressLine1",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.BillingAddressLine1
				},
			},
			{
				name: "contactBillingAddressLine2",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.BillingAddressLine2
				},
			},
			{
				name: "contactBillingAddressLine3",
				generator: func(ev event.Event, c *crm.Contact) string {
					return c.BillingAddressLine3
				},
			},
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")

		fieldNames := []string{}
		for _, field := range fields {
			fieldNames = append(fieldNames, field.name)
		}
		fmt.Fprintln(w, strings.Join(fieldNames, ","))

		for _, ev := range events {
			for _, c := range dataByEvent[ev.Id.Encode()].contacts {
				line := []string{}
				for _, field := range fields {
					line = append(line, field.generator(ev, c))
				}
				fmt.Fprintln(w, strings.Join(line, ","))
			}
		}
	}
	return
}

func exportEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Events,
	}) && data.hasAnyAuth(auth.Auth{
		AuthType: auth.Participants,
	}) && data.hasAnyAuth(auth.Auth{
		AuthType: auth.Contacts,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("unix_start")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("unix_to")), 0)
		to = to.AddDate(0, 0, 1)
		key, err := datastore.DecodeKey(r.URL.Query().Get("event_type_id"))
		if err != nil {
			key = nil
		}
		common.SetContentType(w, "text/csv; charset=UTF-8", false)

		events := event.GetEventsBetween(c, data.domain, from, to, func(ev *event.Event) event.Events {
			if key != nil && !key.Equal(ev.EventType) {
				return nil
			}
			return event.Events{*ev}
		})

		dataChan := make(chan exportEventData)
		for _, ev := range events {
			go func(ev event.Event) {
				d := exportEventData{
					eventID: ev.Id,
				}
				defer func() {
					d.panicErr = recover()
					dataChan <- d
				}()
				for _, participant := range ev.GetParticipants(c) {
					if participant.User != nil {
						d.participantTypes = append(d.participantTypes, participant.ParticipantType.StringID())
						for _, userProperty := range appuser.GetUserProperties(c, appuser.DomainUserKeyUnderDomain(c, data.domain, participant.User), data.domain) {
							if userProperty.AssignedAt.Before(ev.Start) && userProperty.ValidUntil.After(ev.Start) {
								d.userProperties = append(d.userProperties, userProperty.Name)
							}
						}
						d.nUsers++
					}
					if participant.Contact != nil {
						d.nContacts += participant.Multiple
					}
				}
			}(ev)
		}
		exportDataByEvent := map[string]exportEventData{}
		for _ = range events {
			d := <-dataChan
			if d.panicErr != nil {
				panic(d.panicErr)
			}
			exportDataByEvent[d.eventID.Encode()] = d
		}

		fields := []exportEventField{
			{
				name: "type",
				generator: func(ev event.Event) string {
					return ev.EventType.StringID()
				},
			},
			{
				name: "location",
				generator: func(ev event.Event) string {
					return ev.Location.StringID()
				},
			},
			{
				name: "date",
				generator: func(ev event.Event) string {
					return ev.Start.Format("20060102")
				},
			},
			{
				name: "time",
				generator: func(ev event.Event) string {
					return ev.Start.Format("15:04")
				},
			},
			{
				name: "minutes",
				generator: func(ev event.Event) string {
					return fmt.Sprint(int64(ev.End.Sub(ev.Start) / time.Minute))
				},
			},
			{
				name: "contacts",
				generator: func(ev event.Event) string {
					return fmt.Sprint(exportDataByEvent[ev.Id.Encode()].nContacts)
				},
			},
			{
				name: "users",
				generator: func(ev event.Event) string {
					return fmt.Sprint(exportDataByEvent[ev.Id.Encode()].nUsers)
				},
			},
		}

		participantTypes := event.GetParticipantTypes(c, data.domain, data.authorizer)
		for _, wantedParticipantType := range participantTypes {
			func(wantedParticipantType event.ParticipantType) {
				fields = append(fields, exportEventField{
					name: fmt.Sprintf("%s-participants", wantedParticipantType.Id.StringID()),
					generator: func(ev event.Event) string {
						count := 0
						for _, participantType := range exportDataByEvent[ev.Id.Encode()].participantTypes {
							if participantType == wantedParticipantType.Id.StringID() {
								count++
							}
						}
						return fmt.Sprint(count)
					},
				})
			}(wantedParticipantType)
		}

		userProperties := domain.GetUserProperties(c, data.domain)
		for _, wantedProperty := range userProperties {
			func(wantedProperty domain.UserProperty) {
				fields = append(fields, exportEventField{
					name: fmt.Sprintf("%s-properties", wantedProperty.Name),
					generator: func(ev event.Event) string {
						count := 0
						for _, userProperty := range exportDataByEvent[ev.Id.Encode()].userProperties {
							if userProperty == wantedProperty.Name {
								count++
							}
						}
						return fmt.Sprint(count)
					},
				})
			}(wantedProperty)
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")

		fieldNames := []string{}
		for _, field := range fields {
			fieldNames = append(fieldNames, field.name)
		}
		fmt.Fprintln(w, strings.Join(fieldNames, ","))

		for _, ev := range events {
			line := []string{}
			for _, field := range fields {
				line = append(line, field.generator(ev))
			}
			fmt.Fprintln(w, strings.Join(line, ","))
		}
	}
	return
}

func getUnpaidEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Events,
	}) {
		from := time.Unix(common.MustParseInt64(r.URL.Query().Get("from")), 0)
		to := time.Unix(common.MustParseInt64(r.URL.Query().Get("to")), 0)
		to = to.AddDate(0, 0, 1)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetUnpaidEventsBetween(c, data.domain, data.authorizer, from, to))
	}
	return
}

func deleteEvent(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	original := event.GetEvent(data.context, key, data.domain)
	if data.hasAuth(auth.Auth{
		AuthType:  auth.Events,
		Write:     true,
		EventKind: original.EventKind,
		EventType: original.EventType,
		Location:  original.Location,
	}) {
		original.Delete(data.context, data.domain, data.user.Id)
		w.WriteHeader(204)
	}
	return
}

func createEvent(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var e event.Event
	common.MustDecodeJSON(r.Body, &e)
	if data.hasAuth(auth.Auth{
		AuthType:  auth.Events,
		Location:  e.Location,
		EventKind: e.EventKind,
		EventType: e.EventType,
		Write:     true,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, e.Save(data.context, data.domain, data.user.Id))
	}
	return
}

func updateEvent(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	original := event.GetEvent(data.context, key, data.domain)
	var toUpdate event.Event
	common.MustDecodeJSON(r.Body, &toUpdate)
	toUpdate.Id = key
	if data.hasAuth(auth.Auth{
		AuthType:  auth.Events,
		Write:     true,
		EventKind: original.EventKind,
		EventType: original.EventType,
		Location:  original.Location,
	}) {
		if data.hasAuth(auth.Auth{
			AuthType:  auth.Events,
			Write:     true,
			EventKind: toUpdate.EventKind,
			EventType: toUpdate.EventType,
			Location:  toUpdate.Location,
		}) {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, toUpdate.Save(data.context, data.domain, data.user.Id))
		}
	}
	return
}

func getOpenEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		if data.domain == nil {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, []string{})
		} else {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			r.ParseForm()
			common.MustEncodeJSON(w, event.GetOpenEventsBetween(data.context, data.domain, data.user.Id, time.Now(), time.Now().AddDate(0, 1, 0), data.authorizer, r.Form["locations"], r.Form["kinds"], r.Form["types"]))
		}
	}
	return
}

func getMyEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		if data.domain == nil {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, []string{})
		} else {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			r.ParseForm()
			common.MustEncodeJSON(w, event.GetMyEventsBetween(data.context, data.domain, data.user.Id, time.Now(), time.Now().AddDate(0, 1, 0), nil, nil, nil))
		}
	}
	return
}

func getEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		if data.domain == nil {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, []string{})
		} else {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			r.ParseForm()
			start := time.Unix(common.MustParseInt64(r.Form.Get("start")), 0)
			end := time.Unix(common.MustParseInt64(r.Form.Get("end")), 0)
			common.MustEncodeJSON(w, event.GetAllowedEventsBetween(data.context, data.user.Id, data.authorizer, data.domain, start, end, r.Form["locations"], r.Form["kinds"], r.Form["types"], r.Form["users"]))
		}
	}
	return
}

func updateEventKind(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		var toUpdate event.EventKind
		common.MustDecodeJSON(r.Body, &toUpdate)
		toUpdate.Id = key
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, toUpdate.Save(data.context, data.domain))
	}
	return
}

func updateEventType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		var toUpdate event.EventType
		common.MustDecodeJSON(r.Body, &toUpdate)
		toUpdate.Id = key
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		if saved, err := toUpdate.Save(data.context, data.domain); err != nil {
			w.WriteHeader(417)
			common.MustEncodeJSON(w, err)
		} else {
			common.MustEncodeJSON(w, saved)
		}
	}
	return
}

func setParticipantPaid(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	event_key, err := datastore.DecodeKey(mux.Vars(r)["event_id"])
	if err != nil {
		panic(err)
	}
	ev := event.GetEvent(data.context, event_key, data.domain)
	part_key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	part := ev.GetParticipant(data.context, part_key)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Participants,
		Write:           true,
		Location:        ev.Location,
		EventKind:       ev.EventKind,
		EventType:       ev.EventType,
		ParticipantType: part.ParticipantType,
	}) {
		part.Paid = r.FormValue("paid") == "true"
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, part.Save(data.context, ev, data.domain, data.user.Id))
	}
	return
}

func updateParticipant(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	event_key, err := datastore.DecodeKey(mux.Vars(r)["event_id"])
	if err != nil {
		panic(err)
	}
	ev := event.GetEvent(data.context, event_key, data.domain)
	part_key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	part := ev.GetParticipant(data.context, part_key)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Participants,
		Write:           true,
		Location:        ev.Location,
		EventKind:       ev.EventKind,
		EventType:       ev.EventType,
		ParticipantType: part.ParticipantType,
	}) {
		var neu event.Participant
		common.MustDecodeJSON(r.Body, &neu)
		if data.hasAuth(auth.Auth{
			AuthType:        auth.Participants,
			Write:           true,
			Location:        ev.Location,
			EventKind:       ev.EventKind,
			EventType:       ev.EventType,
			ParticipantType: neu.ParticipantType,
		}) {
			if part.Defaulted == neu.Defaulted || !ev.Recurring {
				common.SetContentType(w, "application/json; charset=UTF-8", false)
				common.MustEncodeJSON(w, part.CopyFrom(&neu).Save(data.context, ev, data.domain, data.user.Id))
			} else {
				if generated := ev.ExpandRecurrences(data.context, data.domain, neu.EventStart, neu.EventEnd); len(generated) == 1 {
					newEvent := ev.AddRecurrenceException(
						data.context,
						data.domain,
						generated[0].Start, generated[0].End,
						nil,
						data.user.Id,
						true,
					)
					common.SetContentType(w, "application/json; charset=UTF-8", false)
					part.Id = nil
					common.MustEncodeJSON(w, part.CopyFrom(&neu).Save(data.context, newEvent, data.domain, data.user.Id))
				} else {
					panic(fmt.Errorf("Participant times (%+v) expanded into != 1 events (%v)!", neu, generated))
				}
			}
		}
	}
	return
}

func deleteParticipant(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	event_key, err := datastore.DecodeKey(mux.Vars(r)["event_id"])
	if err != nil {
		panic(err)
	}
	ev := event.GetEvent(data.context, event_key, data.domain)
	part_key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	part := ev.GetParticipant(data.context, part_key)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Participants,
		Write:           true,
		Location:        ev.Location,
		EventKind:       ev.EventKind,
		EventType:       ev.EventType,
		ParticipantType: part.ParticipantType,
	}) {
		part.Delete(data.context, data.user.Id, nil)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func deleteEventTypeRequiredParticipantType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		event_type_key, err := datastore.DecodeKey(mux.Vars(r)["event_type_id"])
		if err != nil {
			panic(err)
		}
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		original := event.GetRequiredParticipantType(data.context, key, event_type_key, data.domain)
		original.Delete(data.context, event_type_key, data.domain, data.user.Id)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func deleteEventType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		event.DeleteEventType(data.context, key, data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func deleteEventKind(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		event.DeleteEventKind(data.context, key, data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func updateParticipantType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		var toUpdate event.ParticipantType
		common.MustDecodeJSON(r.Body, &toUpdate)
		toUpdate.Id = key
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, toUpdate.Save(data.context, data.domain))
	}
	return
}

func getParticipantType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetParticipantType(data.context, key, data.domain))
	}
	return
}

func deleteParticipantType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		event.DeleteParticipantType(data.context, key, data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

type splitRequest struct {
	At time.Time `json:"at"`
}

func splitRecurringEvent(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	event_key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	ev := event.GetEvent(data.context, event_key, data.domain)
	var req splitRequest
	common.MustDecodeJSON(r.Body, &req)
	partsAuth, authenticated := data.silentHasAuth(auth.Auth{
		AuthType:  auth.Participants,
		Write:     true,
		Location:  ev.Location,
		EventKind: ev.EventKind,
		EventType: ev.EventType,
	})
	eventsAuth, _ := data.silentHasAuth(auth.Auth{
		AuthType: auth.Events,
		Write:    true,
	})
	if authenticated && (partsAuth || eventsAuth) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		dayStart, err := time.Parse(common.ISO8601Format, req.At.Format(common.ISO8601Format))
		if err != nil {
			panic(err)
		}
		common.MustEncodeJSON(w, ev.SplitAndRemoveParticipant(data.context, data.domain, dayStart, data.user.Id, nil, data.authorizer))
	} else {
		w.WriteHeader(401)
		fmt.Fprintln(w, "Unauthenticated")
	}
	return
}

type exceptionRequest struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func addEventRecurrenceException(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	event_key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	ev := event.GetEvent(data.context, event_key, data.domain)
	var req exceptionRequest
	common.MustDecodeJSON(r.Body, &req)
	partsAuth, authenticated := data.silentHasAuth(auth.Auth{
		AuthType:  auth.Participants,
		Write:     true,
		Location:  ev.Location,
		EventKind: ev.EventKind,
		EventType: ev.EventType,
	})
	eventsAuth, _ := data.silentHasAuth(auth.Auth{
		AuthType: auth.Events,
		Write:    true,
	})
	if authenticated && (partsAuth || eventsAuth) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, ev.AddRecurrenceException(data.context, data.domain, req.Start, req.End, data.authorizer, data.user.Id, false))
	} else {
		w.WriteHeader(401)
		fmt.Fprintln(w, "Unauthenticated")
	}
	return
}

func createParticipant(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		event_key, err := datastore.DecodeKey(mux.Vars(r)["event_id"])
		if err != nil {
			panic(err)
		}
		ev := event.GetEvent(data.context, event_key, data.domain)
		var participant event.Participant
		common.MustDecodeJSON(r.Body, &participant)
		if participant.User.Equal(data.user.Id) {
			canAttend, authenticated := data.silentHasAuth(auth.Auth{
				AuthType:        auth.Attend,
				Location:        ev.Location,
				EventKind:       ev.EventKind,
				EventType:       ev.EventType,
				ParticipantType: participant.ParticipantType,
			})
			hasAnyPartAuth, _ := data.silentHasAnyAuth(auth.Auth{
				AuthType:  auth.Participants,
				Write:     true,
				Location:  ev.Location,
				EventKind: ev.EventKind,
				EventType: ev.EventType,
			})
			hasPartAuth, _ := data.silentHasAuth(auth.Auth{
				AuthType:        auth.Participants,
				Write:           true,
				Location:        ev.Location,
				EventKind:       ev.EventKind,
				EventType:       ev.EventType,
				ParticipantType: participant.ParticipantType,
			})
			if !authenticated {
				w.WriteHeader(401)
				fmt.Fprintln(w, "Unauthenticated")
			} else if canAttend && (!hasAnyPartAuth || participant.AlwaysCreateException) {
				if ev.Recurring {
					if generated := ev.ExpandRecurrences(data.context, data.domain, participant.EventStart, participant.EventEnd); len(generated) == 1 {
						newEvent := ev.AddRecurrenceException(
							data.context,
							data.domain,
							generated[0].Start, generated[0].End,
							data.authorizer,
							data.user.Id,
							true,
						)
						common.SetContentType(w, "application/json; charset=UTF-8", false)
						common.MustEncodeJSON(w, (&participant).Save(data.context, newEvent, data.domain, data.user.Id))
					} else {
						panic(fmt.Errorf("Participant times (%+v) expanded into != 1 events (%v)!", participant, generated))
					}
				} else {
					common.SetContentType(w, "application/json; charset=UTF-8", false)
					common.MustEncodeJSON(w, (&participant).Save(data.context, ev, data.domain, data.user.Id))
				}
			} else if hasPartAuth {
				common.SetContentType(w, "application/json; charset=UTF-8", false)
				common.MustEncodeJSON(w, (&participant).Save(data.context, ev, data.domain, data.user.Id))
			} else {
				w.WriteHeader(403)
				fmt.Fprintln(w, "Unauthorized")
			}
		} else {
			if data.hasAuth(auth.Auth{
				AuthType:        auth.Participants,
				Write:           true,
				Location:        ev.Location,
				EventKind:       ev.EventKind,
				EventType:       ev.EventType,
				ParticipantType: participant.ParticipantType,
			}) {
				common.SetContentType(w, "application/json; charset=UTF-8", false)
				common.MustEncodeJSON(w, (&participant).Save(data.context, ev, data.domain, data.user.Id))
			}
		}
	}
	return
}

func createEventType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		var toCreate event.EventType
		common.MustDecodeJSON(r.Body, &toCreate)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		if saved, err := toCreate.Save(data.context, data.domain); err != nil {
			common.MustEncodeJSON(w, err)
			w.WriteHeader(417)
		} else {
			common.MustEncodeJSON(w, saved)
		}
	}
	return
}

func createEventKind(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		var toCreate event.EventKind
		common.MustDecodeJSON(r.Body, &toCreate)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, toCreate.Save(data.context, data.domain))
	}
	return
}

func createParticipantType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		var toCreate event.ParticipantType
		common.MustDecodeJSON(r.Body, &toCreate)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, toCreate.Save(data.context, data.domain))
	}
	return
}

func createEventTypeRequiredParticipantType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["event_type_id"])
		if err != nil {
			panic(err)
		}
		var toCreate event.RequiredParticipantType
		common.MustDecodeJSON(r.Body, &toCreate)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, toCreate.Save(data.context, key, data.domain, data.user.Id))
	}
	return
}

func getParticipantTypes(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetParticipantTypes(data.context, data.domain, data.authorizer))
	}
	return
}

func getEventKinds(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetEventKinds(data.context, data.domain, data.authorizer))
	}
	return
}

func getEventTypeAllowedRequiredParticipantTypesWithLocation(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		location_id, err := datastore.DecodeKey(mux.Vars(r)["location_id"])
		if err != nil {
			panic(err)
		}
		event_type := event.GetEventType(data.context, key)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event_type.GetAllowedRequiredParticipantTypes(data.context, location_id, data.authorizer, data.domain))
	}
	return
}

func getEventTypeAllowedRequiredParticipantTypesWithoutLocation(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		event_type := event.GetEventType(data.context, key)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event_type.GetAllowedRequiredParticipantTypes(data.context, nil, data.authorizer, data.domain))
	}
	return
}

func createEventRequiredParticipantType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	event_key, err := datastore.DecodeKey(mux.Vars(r)["event_id"])
	if err != nil {
		panic(err)
	}
	ev := event.GetEvent(data.context, event_key, data.domain)
	if ev == nil {
		panic(err)
	}
	var toCreate event.RequiredParticipantType
	common.MustDecodeJSON(r.Body, &toCreate)
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Participants,
		Location:        ev.Location,
		EventKind:       ev.EventKind,
		EventType:       ev.EventType,
		ParticipantType: toCreate.ParticipantType,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, toCreate.Save(data.context, event_key, data.domain, data.user.Id))
	}
	return
}

func deleteEventRequiredParticipantType(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	event_key, err := datastore.DecodeKey(mux.Vars(r)["event_id"])
	if err != nil {
		panic(err)
	}
	ev := event.GetEvent(data.context, event_key, data.domain)
	if ev == nil {
		panic(err)
	}
	key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	req := event.GetRequiredParticipantType(data.context, key, event_key, data.domain)
	if req == nil {
		panic(fmt.Errorf("No required participant type for %v with id %v found", event_key, key))
	}
	if data.hasAuth(auth.Auth{
		AuthType:        auth.Participants,
		Location:        ev.Location,
		EventType:       ev.EventType,
		EventKind:       ev.EventKind,
		ParticipantType: req.ParticipantType,
		Write:           true,
	}) {
		original := event.GetRequiredParticipantType(data.context, key, event_key, data.domain)
		original.Delete(data.context, event_key, data.domain, data.user.Id)
		w.WriteHeader(204)
	}
	return
}

func getEventAllowedRequiredParticipantTypes(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		key, err := datastore.DecodeKey(mux.Vars(r)["event_id"])
		if err != nil {
			panic(err)
		}
		ev := event.GetEvent(data.context, key, data.domain)
		if ev == nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, ev.GetAllowedRequiredParticipantTypes(data.context, data.authorizer, data.domain, data.user.Id))
	}
	return
}

func getParticipants(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		key, err := datastore.DecodeKey(mux.Vars(r)["event_id"])
		if err != nil {
			panic(err)
		}
		ev := event.GetEvent(data.context, key, data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, ev.GetAllowedParticipants(data.context, data.authorizer, data.domain, data.user.Id))
	}
	return
}

func getEventTypes(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetEventTypes(data.context, data.domain, data.authorizer))
	}
	return
}

func getEventTypeRequiredParticipantTypes(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.EventTypes,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["event_type_id"])
		if err != nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetRequiredParticipantTypes(data.context, key, data.domain))
	}
	return
}
