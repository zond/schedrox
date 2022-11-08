package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"github.com/zond/schedrox/appuser"
	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/crm"
	"github.com/zond/schedrox/domain"
	"github.com/zond/schedrox/event"
	"github.com/zond/schedrox/salary"
	"github.com/zond/schedrox/search"
	"github.com/zond/schedrox/translation"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var userBouncePattern = regexp.MustCompile("^ub\\+([^@]+)@[^.]+\\.appspotmail\\.com$")
var contactBouncePattern = regexp.MustCompile("cb\\+([^@]+)@[^.]+\\.appspotmail\\.com$")

type cleanResponse struct {
	Deleted int `json:"deleted"`
}

func convertRecurrenceExceptions(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, cleanResponse{
			Deleted: event.ConvertRecurrenceExceptions(c),
		})
	}
	return
}

type debugChange struct {
	event.Change
	At        time.Time
	Formatted interface{}
}

func getChanges(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		from, err := time.Parse("20060102", r.URL.Query().Get("from"))
		if err != nil {
			return err
		}
		to, err := time.Parse("20060102", r.URL.Query().Get("to"))
		if err != nil {
			return err
		}
		event_date := r.URL.Query().Get("event_date")
		changes := event.FindChangesFrom(c, data.domain, from, to)
		debugs := make([]debugChange, 0, len(changes))
		matches := func(formatted map[string]interface{}) bool {
			if event_date == "" {
				return true
			}
			if event_start, found := formatted["event_start"]; found {
				if start_string, ok := event_start.(string); ok {
					if start_time, err := time.Parse(time.RFC3339, start_string); err == nil {
						if start_time.Format("20060102") == event_date {
							return true
						}
					} else {
						log.Infof(c, "event_start was unparseable in %+v?", formatted)
					}
				} else {
					log.Infof(c, "event_start was not a string in %+v?", formatted)
				}
			} else {
				log.Infof(c, "Found no event_start in %+v?", formatted)
			}
			return false
		}
		for _, change := range changes {
			formatted := map[string]interface{}{}
			json.Unmarshal(change.DataBytes, &formatted)
			if matches(formatted) {
				debug := debugChange{
					Change:    change,
					At:        change.At,
					Formatted: formatted,
				}
				debugs = append(debugs, debug)
			}
		}
		encoded, err := json.MarshalIndent(debugs, "  ", "  ")
		if err != nil {
			return err
		}
		w.Write(encoded)
	}
	return nil
}

func removeEncodedKeys(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		deleted := 0

		deleted += appuser.DeleteEncodedKeys(c)

		deleted += salary.DeleteEncodedKeys(c)

		deleted += search.DeleteEncodedKeys(c)

		deleted += auth.DeleteEncodedKeys(c)

		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, cleanResponse{
			Deleted: deleted,
		})
	}
	return
}

type detachedProblem struct {
	Event       string
	Error       error
	EventWeekId string
	EventWeek   event.EventWeek
}

func cleanEventWeeksWithoutEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		result := []detachedProblem{}
		eventWeeks := []event.EventWeek{}
		eventWeekIds, err := datastore.NewQuery("EventWeek").GetAll(c, &eventWeeks)
		if err != nil {
			panic(err)
		}
		eventIds := make([]*datastore.Key, len(eventWeeks))
		for index, deta := range eventWeeks {
			eventIds[index] = deta.Event
		}
		missing := []*datastore.Key{}
		offset := 0
		for len(eventIds) > 0 {
			l := len(eventIds)
			if l > 1000 {
				l = 1000
			}
			toUse := eventIds[:l]
			eventIds = eventIds[l:]
			events := make([]event.Event, len(toUse))
			if err = datastore.GetMulti(c, toUse, events); err != nil {
				if merr, ok := err.(appengine.MultiError); ok {
					for index, serr := range merr {
						if serr != nil {
							if serr == datastore.ErrNoSuchEntity {
								result = append(result, detachedProblem{
									Event:       toUse[index].Encode(),
									Error:       fmt.Errorf(serr.Error()),
									EventWeekId: eventWeekIds[offset+index].String(),
									EventWeek:   eventWeeks[offset+index],
								})
								missing = append(missing, eventWeekIds[offset+index])
							}
						}
					}
				} else {
					panic(err)
				}
			}
			offset += l
		}
		if err = datastore.DeleteMulti(c, missing); err != nil {
			panic(err)
		}
		if err = json.NewEncoder(w).Encode(result); err != nil {
			panic(err)
		}
	}
	return
}

func cleanDetachedEventWeeks(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		result := []detachedProblem{}
		eventWeeks := []event.EventWeek{}
		eventWeekIds, err := datastore.NewQuery("EventWeek").GetAll(c, &eventWeeks)
		if err != nil {
			panic(err)
		}
		eventIds := make([]*datastore.Key, len(eventWeeks))
		for index, deta := range eventWeeks {
			eventIds[index] = deta.Event
		}
		missing := []*datastore.Key{}
		offset := 0
		for len(eventIds) > 0 {
			l := len(eventIds)
			if l > 1000 {
				l = 1000
			}
			toUse := eventIds[:l]
			eventIds = eventIds[l:]
			events := make([]event.Event, len(toUse))
			if err = datastore.GetMulti(c, toUse, events); err != nil {
				if merr, ok := err.(appengine.MultiError); ok {
					for index, serr := range merr {
						if serr != nil {
							if _, ok := serr.(*datastore.ErrFieldMismatch); !ok {
								result = append(result, detachedProblem{
									Event:       toUse[index].Encode(),
									Error:       serr,
									EventWeekId: eventWeekIds[offset+index].String(),
									EventWeek:   eventWeeks[offset+index],
								})
								missing = append(missing, eventWeekIds[offset+index])
							}
						}
					}
				} else {
					panic(err)
				}
			}
			offset += l
		}
		if err = datastore.DeleteMulti(c, missing); err != nil {
			panic(err)
		}
		if err = json.NewEncoder(w).Encode(result); err != nil {
			panic(err)
		}
	}
	return
}

func cleanSearch(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, cleanResponse{
			Deleted: search.Clean(data.context),
		})
	}
	return
}

func cleanProperties(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, cleanResponse{
			Deleted: appuser.CleanProperties(data.context),
		})
	}
	return
}

func incomingMail(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	recipient := mux.Vars(r)["recipient"]
	if match := userBouncePattern.FindStringSubmatch(recipient); match != nil {
		user := appuser.GetUser(c, common.MustDecodeBase64(match[1]))
		buf := new(bytes.Buffer)
		io.Copy(buf, r.Body)
		user.EmailBounce = string(buf.Bytes())
		user.Save(c)
	} else if match = contactBouncePattern.FindStringSubmatch(recipient); match != nil {
		parts := strings.Split(common.MustDecodeBase64(match[1]), ":")
		domainKey := datastore.NewKey(c, "Domain", parts[0], 0, nil)
		contact := crm.GetContact(c, datastore.NewKey(c, "Contact", parts[1], 0, domainKey), domainKey)
		buf := new(bytes.Buffer)
		io.Copy(buf, r.Body)
		contact.EmailBounce = string(buf.Bytes())
		contact.Save(c, domainKey)
	} else {
		message, err := mail.ReadMessage(r.Body)
		if err != nil {
			panic(err)
		}
		b := new(bytes.Buffer)
		io.Copy(b, message.Body)
		log.Infof(c, "got email to %v", recipient)
		log.Infof(c, "%v", message)
		log.Infof(c, "%v", string(b.Bytes()))
	}
	return
}

func mobileAllCss(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	common.SetContentType(w, "text/css; charset=UTF-8", true)
	renderText(w, r, mobileCssTemplates, "bootstrap-3.2.min.css", data)
	renderText(w, r, mobileCssTemplates, "common.css", data)
	return
}

func allCss(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	common.SetContentType(w, "text/css; charset=UTF-8", true)
	renderText(w, r, cssTemplates, "bootstrap.min.css", data)
	renderText(w, r, cssTemplates, "bootstrap-responsive.min.css", data)
	renderText(w, r, cssTemplates, "fullcalendar.css", data)
	renderText(w, r, cssTemplates, "anytime.c.css", data)
	renderText(w, r, cssTemplates, "select2.css", data)
	renderText(w, r, cssTemplates, "colorpicker.css", data)
	renderText(w, r, cssTemplates, "common.css", data)
	return
}

func appJs(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	common.SetContentType(w, "application/javascript; charset=UTF-8", false)
	renderText(w, r, jsTemplates, "app.js", data)
	return
}

func mobileAllJs(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	common.SetContentType(w, "application/javascript; charset=UTF-8", true)
	renderJs(w, r, mobileJsTemplates, "underscore-min.js", data)
	renderJs(w, r, mobileJsTemplates, "jquery-2.1.1.min.js", data)
	renderJs(w, r, jsTemplates, "jquery.timeago.js", data)
	renderJs(w, r, jsTemplates, "backbone-min.js", data)
	renderJs(w, r, mobileJsTemplates, "bootstrap-3.2.min.js", data)
	renderJs(w, r, jsTemplates, "anytime.c.js", data)
	renderJs(w, r, jsTemplates, "util.js", data)
	renderJs(w, r, jsTemplates, "google_analytics.js", data)
	render_Templates(mobile_Templates, data)
	for _, templ := range jsModelTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
	for _, templ := range jsCollectionTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
	for _, templ := range mobileJsViewTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
	renderJs(w, r, mobileJsTemplates, "app.js", data)
	return
}

func allJs(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	common.SetContentType(w, "application/javascript; charset=UTF-8", true)
	renderJs(w, r, jsTemplates, "underscore-min.js", data)
	renderJs(w, r, jsTemplates, "jquery-1.8.1.min.js", data)
	renderJs(w, r, jsTemplates, "jquery.timeago.js", data)
	renderJs(w, r, jsTemplates, "select2.min.js", data)
	renderJs(w, r, jsTemplates, "anytime.c.js", data)
	renderJs(w, r, jsTemplates, "jquery.simplemodal.1.4.3.min.js", data)
	renderJs(w, r, jsTemplates, "backbone-min.js", data)
	renderJs(w, r, jsTemplates, "fullcalendar.min.js", data)
	renderJs(w, r, jsTemplates, "bootstrap.min.js", data)
	renderJs(w, r, jsTemplates, "bootstrap-colorpicker.js", data)
	renderJs(w, r, jsTemplates, "sha1.js", data)
	renderJs(w, r, jsTemplates, "util.js", data)
	renderJs(w, r, jsTemplates, "modals.js", data)
	renderJs(w, r, jsTemplates, "google_analytics.js", data)
	renderJs(w, r, jsTemplates, "shake.js", data)
	render_Templates(_Templates, data)
	for _, templ := range jsModelTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
	for _, templ := range jsCollectionTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
	for _, templ := range jsViewTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
	}
	return
}

func renderEditor(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	for _, dom := range data.user.Domains {
		if dom.Id.Equal(data.domain) {
			data.user.Domains = []domain.Domain{dom}
		}
	}
	if data.hasAuth(auth.Auth{
		AuthType: auth.SalaryConfiguration,
	}) {
		renderHtml(w, r, htmlTemplates, "editor.html", data)
	}
	return
}

func desktopToMobile(c gaecontext.HTTPContext) (err error) {
	sess, err := sessionStore.Get(c.Req(), "schedrox-session")
	if err != nil {
		return
	}
	delete(sess.Values, "mobile-off")
	if err = sess.Save(c.Req(), c.Resp()); err != nil {
		return
	}
	c.Resp().Header().Set("Location", "/")
	c.Resp().WriteHeader(303)
	return
}

func mobileToDesktop(c gaecontext.HTTPContext) (err error) {
	sess, err := sessionStore.Get(c.Req(), "schedrox-session")
	if err != nil {
		return
	}
	sess.Values["mobile-off"] = true
	if err = sess.Save(c.Req(), c.Resp()); err != nil {
		return
	}
	c.Resp().Header().Set("Location", "/")
	c.Resp().WriteHeader(303)
	return
}

func mobileIndex(c gaecontext.HTTPContext) (err error) {
	renderHtml(c.Resp(), c.Req(), mobileHtmlTemplates, "index.html", getBaseData(c, c.Resp(), c.Req()))
	return
}

func index(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if user := data.User(); user != nil {
		user.MarkActive(data.context, translation.GetLanguage(r))
	}
	renderHtml(w, r, htmlTemplates, "index.html", getBaseData(c, w, r))
	return
}
