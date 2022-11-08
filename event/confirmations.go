package event

import (
	"bytes"
	"monotone/se.oort.schedrox/appuser"
	"monotone/se.oort.schedrox/common"
	"monotone/se.oort.schedrox/crm"
	"text/template"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type templateError struct {
	Message  string `json:"message"`
	Context  string `json:"context"`
	Template string `json:"template"`
}

func (self templateError) Error() string {
	return self.Message
}

func GenerateConfirmation(
	c context.Context,
	subjectTemplate, bodyTemplate string,
	participant *Participant,
	contact *crm.Contact,
	user *appuser.User,
	event *Event) (subject, body string, err error) {

	context := map[string]interface{}{
		"user":        user,
		"contact":     contact,
		"event":       event,
		"participant": participant,
	}

	var templ *template.Template
	templ, err = template.New("").Funcs(template.FuncMap{
		"AddMinutes": func(t time.Time, minutes int) time.Time {
			return t.Add(time.Minute * time.Duration(minutes))
		},
	}).Parse(bodyTemplate)
	if err != nil {
		err = templateError{
			Message:  err.Error(),
			Context:  "body",
			Template: bodyTemplate,
		}
		return
	}
	bodyBuf := new(bytes.Buffer)
	if err = templ.Execute(bodyBuf, context); err != nil {
		err = templateError{
			Message:  err.Error(),
			Context:  "body",
			Template: bodyTemplate,
		}
		return
	}
	templ, err = template.New("").Parse(subjectTemplate)
	if err != nil {
		err = templateError{
			Message:  err.Error(),
			Context:  "subject",
			Template: subjectTemplate,
		}
		return
	}
	subjectBuf := new(bytes.Buffer)
	if err = templ.Execute(subjectBuf, context); err != nil {
		err = templateError{
			Message:  err.Error(),
			Context:  "subject",
			Template: subjectTemplate,
		}
		return
	}
	body = string(bodyBuf.Bytes())
	subject = string(subjectBuf.Bytes())
	return
}

func CreateConfirmationExample(c context.Context, subjectTemplate, bodyTemplate string) (subject, body string, err error) {
	event := fakeEvent(c)
	part := fakeParticipant(c)
	user := fakeUser(c)
	contact := fakeContact(c)

	subject, body, err = GenerateConfirmation(c, subjectTemplate, bodyTemplate, part, contact, user, event)
	return
}

func fakeUser(c context.Context) *appuser.User {
	return &appuser.User{
		Id:          datastore.NewKey(c, "User", "", 0, nil),
		MobilePhone: "0701234567",
		GivenName:   "John",
		FamilyName:  "Doe",
		Email:       "mail@domain.tld",
	}
}

func fakeContact(c context.Context) *crm.Contact {
	return &crm.Contact{
		Id:                  datastore.NewKey(c, "Contact", "", 0, nil),
		Name:                "Contact Contact",
		ContactFamilyName:   "Doe",
		ContactGivenName:    "John",
		MobilePhone:         "0701234567",
		Email:               "mail@domain.tld",
		AddressLine1:        "Some address 123",
		AddressLine2:        "123456 Some city",
		AddressLine3:        "Some country",
		BillingAddressLine1: "Some billing address 123",
		BillingAddressLine2: "123456 Some billing city",
		BillingAddressLine3: "Some billing country",
		Information:         "Some information",
		Reference:           "Some reference",
		OrganizationNumber:  "Some organization number",
	}
}

func fakeEvent(c context.Context) *Event {
	start := common.MustParseJSTime("2013-03-18T18:00:00.000Z")
	end := common.MustParseJSTime("2013-03-18T19:00:00.000Z")
	return &Event{
		Id:                          datastore.NewKey(c, "Event", "", 0, nil),
		Location:                    datastore.NewKey(c, "Location", "a location", 0, nil),
		Information:                 "information",
		EventType:                   datastore.NewKey(c, "EventType", "an event type", 0, nil),
		EventKind:                   datastore.NewKey(c, "EventKind", "an event kind", 0, nil),
		AllDay:                      false,
		Title:                       "title",
		Start:                       start,
		End:                         end,
		WantedUserParticipants:      3,
		UserParticipants:            3,
		RequiredContactParticipants: 5,
		AllowedContactParticipants:  20,
		ContactParticipants:         15,
		Recurring:                   false,
		Recurrence:                  "",
		RecurrenceEnd:               end,
		RecurrenceExceptions:        "",
		RecurrenceMaster:            nil,
		RecurrenceMasterStart:       time.Time{},
		RecurrenceMasterEnd:         time.Time{},
	}
}

func fakeParticipant(c context.Context) *Participant {
	return &Participant{
		Id:              datastore.NewKey(c, "Participant", "", 0, nil),
		User:            nil,
		Contact:         datastore.NewKey(c, "Contact", "", 0, nil),
		ParticipantType: datastore.NewKey(c, "ParticipantType", "a participant type", 0, nil),
		Multiple:        5,
		Name:            "participant name",
		GivenName:       "john",
		FamilyName:      "doe",
		GravatarHash:    "7d544a8b7db6d512115fd0d2bee3aa3e",
		Email:           "user@domain.tld",
		MobilePhone:     "0701234567",
		Auths:           nil,
		Confirmations:   1,
	}
}
