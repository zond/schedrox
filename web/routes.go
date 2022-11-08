package web

import (
	"monotone/se.oort.schedrox/appuser"
	"monotone/se.oort.schedrox/auth"
	"monotone/se.oort.schedrox/common"
	"monotone/se.oort.schedrox/event"
	"monotone/se.oort.schedrox/salary"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
	_ "google.golang.org/appengine/remote_api"
)

func wantsJSON(r *http.Request, m *mux.RouteMatch) bool {
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return true
	}
	mostAccepted := common.MostAcceptedMap(r, "text/html", "Accept")
	for k, _ := range mostAccepted {
		if strings.Contains(k, "json") || strings.Contains(k, "javascript") {
			return true
		}
	}
	return false
}

var mobileUserAgentRegexp = regexp.MustCompile("(?i)(iphone|ipod|blackberry|android|palm|windows\\s+ce)")
var desktopUserAgentRegexp = regexp.MustCompile("(?i)(windows|linux|os\\s+[x9]|solaris|bsd)")
var botUserAgentRegexp = regexp.MustCompile("(?i)(spider|crawl|slurp|bot)")

func getUA(r *http.Request) string {
	if h := r.Header.Get("X-OperaMini-Phone-UA"); h != "" {
		return h
	}
	if h := r.Header.Get("X-Skyfire-Phone"); h != "" {
		return h
	}
	return r.Header.Get("User-Agent")
}

func isMobile(r *http.Request, m *mux.RouteMatch) bool {
	sessMobileTurnedOff := false
	sess, err := sessionStore.Get(r, "schedrox-session")
	if err == nil {
		if val, found := sess.Values["mobile-off"]; found {
			if bo, ok := val.(bool); ok {
				sessMobileTurnedOff = bo
			}
		}
	}
	if sessMobileTurnedOff {
		return false
	}
	ua := getUA(r)
	return !botUserAgentRegexp.MatchString(ua) && (mobileUserAgentRegexp.MatchString(ua) || !desktopUserAgentRegexp.MatchString(ua))
}

func wantsHTML(r *http.Request, m *mux.RouteMatch) bool {
	return common.MostAccepted(r, "text/html", "Accept") == "text/html"
}

func reverseDependencies() {
	appuser.GetSalaryConfigs = func(c gaecontext.HTTPContext, domains []*datastore.Key) (result []interface{}) {
		configs := salary.GetConfigs(c, domains)
		result = make([]interface{}, len(configs))
		for index, config := range configs {
			config.SalaryCode = ""
			result[index] = config
		}
		return
	}
	appuser.GetAttestedUids = func(c gaecontext.HTTPContext, d *datastore.Key, authorizer auth.Authorizer, from, to time.Time, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool) {
		done2 = make(chan bool)
		go func() {
			attestedEvents := make(map[string]bool)
			for _, attested := range salary.GetAttested(c, d, from, to) {
				attestedEvents[attested.AttestUUID()] = true
			}
			attestableUsers := make(map[string]bool)
			unattestedUsers := make(map[string]bool)
			for _, ev := range event.GetAttestableEventsForUserBetween(c, d, authorizer, nil, from, to, true) {
				attestableUsers[ev.SalaryAttestedUser.Encode()] = true
				if !attestedEvents[ev.AttestUUID()] {
					unattestedUsers[ev.SalaryAttestedUser.Encode()] = true
				}
			}
			for _, ev := range salary.GetAllowedReported(c, d, from, to, auth.Attest, authorizer) {
				attestableUsers[ev.SalaryAttestedUser.Encode()] = true
				if !attestedEvents[ev.AttestUUID()] {
					unattestedUsers[ev.SalaryAttestedUser.Encode()] = true
				}
			}
			var result []*datastore.Key
			var err error
			for k, _ := range attestableUsers {
				if !unattestedUsers[k] {
					var key *datastore.Key
					if key, err = datastore.DecodeKey(k); err != nil {
						panic(err)
					}
					result = append(result, key)
				}
			}
			keys <- result
			if done1 != nil {
				<-done1
			}
			close(done2)
		}()
		return
	}
	appuser.GetUnattestedUids = func(c gaecontext.HTTPContext, d *datastore.Key, authorizer auth.Authorizer, from, to time.Time, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool) {
		done2 = make(chan bool)
		go func() {
			var result []*datastore.Key
			attestedEvents := make(map[string]bool)
			for _, attested := range salary.GetAttested(c, d, from, to) {
				attestedEvents[attested.AttestUUID()] = true
			}
			for _, ev := range event.GetAttestableEventsForUserBetween(c, d, authorizer, nil, from, to, true) {
				if !attestedEvents[ev.AttestUUID()] {
					result = append(result, ev.SalaryAttestedUser)
				}
			}
			for _, ev := range salary.GetAllowedReported(c, d, from, to, auth.Attest, authorizer) {
				if !attestedEvents[ev.AttestUUID()] {
					result = append(result, ev.SalaryAttestedUser)
				}
			}
			keys <- result
			if done1 != nil {
				<-done1
			}
			close(done2)
		}()
		return
	}
}

func init() {
	reverseDependencies()

	router := mux.NewRouter()

	router.Path("/js/{ver}/all.js").Handler(gaecontext.HTTPHandlerFunc(allJs))
	router.Path("/js/app.js").Handler(gaecontext.HTTPHandlerFunc(appJs))
	router.Path("/css/{ver}/all.css").Handler(gaecontext.HTTPHandlerFunc(allCss))

	router.Path("/mobile/js/{ver}/all.js").Handler(gaecontext.HTTPHandlerFunc(mobileAllJs))
	router.Path("/mobile/css/{ver}/all.css").Handler(gaecontext.HTTPHandlerFunc(mobileAllCss))

	router.Path("/mobile/to-desktop").Handler(gaecontext.HTTPHandlerFunc(mobileToDesktop))
	router.Path("/desktop/to-mobile").Handler(gaecontext.HTTPHandlerFunc(desktopToMobile))

	router.Path("/download/events/reports/export.csv").Handler(gaecontext.HTTPHandlerFunc(exportEvents))
	router.Path("/download/events/reports/contacts.csv").Handler(gaecontext.HTTPHandlerFunc(contactEvents))
	router.Path("/download/events/reports/users.csv").Handler(gaecontext.HTTPHandlerFunc(userEvents))

	// Maintenance

	router.Path("/maintenance/changes").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getChanges))
	router.Path("/maintenance/clean_detached_event_weeks").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(cleanDetachedEventWeeks))
	router.Path("/maintenance/clean_event_weeks_without_events").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(cleanEventWeeksWithoutEvents))
	router.Path("/maintenance/cleansearch").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(cleanSearch))
	router.Path("/maintenance/cleanproperties").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(cleanProperties))
	router.Path("/maintenance/remove_encoded_keys").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(removeEncodedKeys))
	router.Path("/maintenance/convert_recurrence_exceptions").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(convertRecurrenceExceptions))

	// Login redirection
	router.Path("/login/redirect").Methods("GET").Handler(gaecontext.JSONHandlerFunc(loginRedirect, 0, 0))
	router.Path("/logout/redirect").Methods("GET").Handler(gaecontext.JSONHandlerFunc(logoutRedirect, 0, 0))

	// Editor

	router.Path("/editor").Methods("GET").MatcherFunc(wantsHTML).Handler(gaecontext.HTTPHandlerFunc(renderEditor))

	// Incoming mail

	router.Path("/_ah/mail/{recipient}").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(incomingMail))

	// Confirmation mail

	router.Path("/send_confirmation").MatcherFunc(wantsJSON).Methods("POST").Handler(gaecontext.HTTPHandlerFunc(sendConfirmation))
	router.Path("/example_confirmation").MatcherFunc(wantsJSON).Methods("POST").Handler(gaecontext.HTTPHandlerFunc(exampleConfirmation))
	router.Path("/{domain_id}/contacts/{id}/bounce_message").MatcherFunc(wantsHTML).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(contactBounceMessage))
	router.Path("/{domain_id}/users/{id}/bounce_message").MatcherFunc(wantsHTML).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(userBounceMessage))

	// Settings update

	router.Path("/settings").MatcherFunc(wantsJSON).Methods("POST").Handler(gaecontext.HTTPHandlerFunc(updatePrivateUserSettings))

	// Reported hours

	reportedHoursRouter := router.PathPrefix("/reported").Subrouter()
	reportedHoursRouter.Path("/finished").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(setMyReportFinished))
	reportedHoursRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getReportedHours))
	reportedHoursRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(addMyReportedHours))
	reportedHoursRouter.Path("/{id}").Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(removeMyReportedHours))

	// Latest changes
	router.Path("/changes/latest").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getLatestChanges))

	// Is busy

	router.Path("/isbusy").MatcherFunc(wantsJSON).Methods("POST").Handler(gaecontext.HTTPHandlerFunc(getIsBusy))

	// Alerts

	router.Path("/alerts").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getCurrentAlerts))

	// Custom filters

	customFiltersRouter := router.PathPrefix("/custom_filters").MatcherFunc(wantsJSON).Subrouter()
	customFiltersRouter.Path("/{id}").Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteCustomFilter))
	customFiltersRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getCustomFilters))
	customFiltersRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createCustomFilter))

	// Participant types

	participantTypesRouter := router.PathPrefix("/participant_types").MatcherFunc(wantsJSON).Subrouter()

	participantTypeRouter := participantTypesRouter.Path("/{id}").Subrouter()
	participantTypeRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getParticipantType))
	participantTypeRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateParticipantType))
	participantTypeRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteParticipantType))

	participantTypesRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getParticipantTypes))
	participantTypesRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createParticipantType))

	// Potential participants

	router.Path("/potential_participants").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(getPotentialParticipants))

	// Locations

	locationsRouter := router.PathPrefix("/locations").MatcherFunc(wantsJSON).Subrouter()

	locationRouter := locationsRouter.Path("/{id}").Subrouter()
	locationRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteLocation))
	locationRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getLocation))
	locationRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateLocation))

	locationsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getLocations))
	locationsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createLocation))

	// Event kinds

	eventKindsRouter := router.PathPrefix("/event_kinds").MatcherFunc(wantsJSON).Subrouter()

	eventKindRouter := eventKindsRouter.Path("/{id}").Subrouter()
	eventKindRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateEventKind))
	eventKindRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteEventKind))
	eventKindRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventKind))

	eventKindsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventKinds))
	eventKindsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createEventKind))

	// User properties

	userPropertiesForDomainRouter := router.PathPrefix("/user_properties").MatcherFunc(wantsJSON).Subrouter()

	userPropertyForDomainRouter := userPropertiesForDomainRouter.Path("/{id}").Subrouter()
	userPropertyForDomainRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateUserPropertyForDomain))
	userPropertyForDomainRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteUserPropertyForDomain))

	userPropertiesForDomainRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createUserPropertyForDomain))
	userPropertiesForDomainRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUserPropertiesForDomain))

	// Event types

	eventTypesRouter := router.PathPrefix("/event_types").MatcherFunc(wantsJSON).Subrouter()

	eventTypeRequiredParticipantTypesRouter := eventTypesRouter.PathPrefix("/{event_type_id}/participant_types").Subrouter()

	eventTypeRequiredParticipantTypeRouter := eventTypeRequiredParticipantTypesRouter.Path("/{id}").Subrouter()
	eventTypeRequiredParticipantTypeRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteEventTypeRequiredParticipantType))

	eventTypeRequiredParticipantTypesRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventTypeRequiredParticipantTypes))
	eventTypeRequiredParticipantTypesRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createEventTypeRequiredParticipantType))

	eventTypeRouter := eventTypesRouter.PathPrefix("/{id}").Subrouter()

	eventTypeRouter.Path("/unique").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(checkUniqueEvent))

	eventTypeAllowedRequiredParticipantTypeRouter := eventTypeRouter.PathPrefix("/required_participant_types").Subrouter()
	eventTypeAllowedRequiredParticipantTypeRouter.Path("/{location_id}").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventTypeAllowedRequiredParticipantTypesWithLocation))
	eventTypeAllowedRequiredParticipantTypeRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventTypeAllowedRequiredParticipantTypesWithoutLocation))

	eventTypeRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateEventType))
	eventTypeRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventType))
	eventTypeRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteEventType))

	eventTypesRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventTypes))
	eventTypesRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createEventType))

	// Events

	eventsRouter := router.PathPrefix("/events").MatcherFunc(wantsJSON).Subrouter()

	eventsReportsRouter := eventsRouter.PathPrefix("/reports").Subrouter()
	eventsReportsRouter.Path("/unpaid").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUnpaidEvents))

	eventsRouter.Path("/mine").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getMyEvents))
	eventsRouter.Path("/open").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getOpenEvents))

	participantsRouter := eventsRouter.PathPrefix("/{event_id}/participants").Subrouter()

	participantRouter := participantsRouter.PathPrefix("/{id}").Subrouter()
	participantRouter.Path("/set_paid").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(setParticipantPaid))
	participantRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateParticipant))
	participantRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteParticipant))

	participantsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createParticipant))
	participantsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getParticipants))

	eventRequiredParticipantTypesRouter := eventsRouter.PathPrefix("/{event_id}/required_participant_types").Subrouter()

	eventRequiredParticipantTypeRouter := eventRequiredParticipantTypesRouter.Path("/{id}").Subrouter()
	eventRequiredParticipantTypeRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteEventRequiredParticipantType))

	eventRequiredParticipantTypesRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventAllowedRequiredParticipantTypes))
	eventRequiredParticipantTypesRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createEventRequiredParticipantType))

	eventRouter := eventsRouter.PathPrefix("/{id}").Subrouter()
	eventRouter.Path("/changes").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEventChanges))
	eventRouter.Path("/exceptions").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(addEventRecurrenceException))
	eventRouter.Path("/splits").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(splitRecurringEvent))
	eventRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateEvent))
	eventRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteEvent))
	eventRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEvent))

	eventsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createEvent))
	eventsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getEvents))

	// Contacts

	contactsRouter := router.PathPrefix("/contacts").MatcherFunc(wantsJSON).Subrouter()
	contactsRouter.Path("/search").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(searchContacts))

	contactRouter := contactsRouter.PathPrefix("/{id}").Subrouter()

	contactRouter.Path("/events").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getContactEvents))

	contactRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteContact))
	contactRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getContact))
	contactRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateContact))

	contactsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getContacts))
	contactsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createContact))

	// Profiles

	router.Path("/profiles/{id}").MatcherFunc(wantsJSON).Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getProfile))

	// Users

	usersRouter := router.PathPrefix("/users").MatcherFunc(wantsJSON).Subrouter()
	usersRouter.Path("/me").Methods("GET").Handler(gaecontext.JSONHandlerFunc(usersMe, 0, 0))
	usersRouter.Path("/search").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(searchUsers))
	usersRouter.Path("/bulk/update").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(bulkUserUpdate))

	//   user roles

	userRolesRouter := usersRouter.PathPrefix("/{user_id}/roles").Subrouter()

	userRoleRouter := userRolesRouter.Path("/{id}").Subrouter()
	userRoleRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteUserRole))

	userRolesRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUserRoles))
	userRolesRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(addUserRole))

	//   user salaries

	userAttestableEventsRouter := usersRouter.PathPrefix("/{user_id}/attestable_events").Subrouter()
	userAttestableEventsRouter.Path("/{reported_id}").Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(removeReportedHours))
	userAttestableEventsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(addReportedHours))
	userAttestableEventsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUserAttestableHours))

	userAttestedEventsRouter := usersRouter.PathPrefix("/{user_id}/attested_events").Subrouter()
	userAttestedEventsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUserAttestedHours))
	userAttestedEventsRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(setUserAttestedHours))
	userAttestedEventsRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteUserAttestedHours))

	reportedFinishedRouter := usersRouter.Path("/{user_id}/reported/finished").Subrouter()
	reportedFinishedRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(unsetReportFinished))
	reportedFinishedRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(setReportFinished))

	//    user auths

	userAuthsRouter := usersRouter.PathPrefix("/{user_id}/auths").Subrouter()

	userAuthRouter := userAuthsRouter.Path("/{id}").Subrouter()
	userAuthRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteUserAuth))

	userAuthsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(addUserAuth))
	userAuthsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUserAuths))

	//    user properties

	userPropertiesForUserRouter := usersRouter.PathPrefix("/{user_id}/properties").Subrouter()

	userPropertyForUserRouter := userPropertiesForUserRouter.Path("/{id}").Subrouter()
	userPropertyForUserRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateUserPropertyForUser))
	userPropertyForUserRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteUserPropertyForUser))

	userPropertiesForUserRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUserPropertiesForUser))
	userPropertiesForUserRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createUserPropertyForUser))

	userRouter := usersRouter.Path("/{id}").Subrouter()
	userRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteUser))
	userRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUser))
	userRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateUser))

	usersRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createUser))
	usersRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getUsers))

	// Domains

	domainsRouter := router.PathPrefix("/domains").MatcherFunc(wantsJSON).Subrouter()

	domainRouter := domainsRouter.Path("/{id}").Subrouter()
	domainRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getDomain))
	domainRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateDomain))

	domainRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteDomain))
	domainsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createDomain))
	domainsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getDomains))

	// Salary

	salaryRouter := router.PathPrefix("/salary").MatcherFunc(wantsJSON).Subrouter()

	salaryConfigRouter := salaryRouter.Path("/config").Subrouter()
	salaryConfigRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getSalaryConfig))
	salaryConfigRouter.Methods("PUT").Handler(gaecontext.HTTPHandlerFunc(updateSalaryConfig))

	salaryReportRouter := salaryRouter.Path("/report").Subrouter()
	salaryReportRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getSalaryReport))

	router.Path("/salary/{version}/code.js").Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getSalaryCode))
	router.Path("/salary/code.js").Methods("POST").Handler(gaecontext.HTTPHandlerFunc(setSalaryCode))

	// Roles

	rolesRouter := router.PathPrefix("/roles").MatcherFunc(wantsJSON).Subrouter()

	roleAuthsRouter := rolesRouter.PathPrefix("/{role_id}/auths").Subrouter()

	roleAuthRouter := roleAuthsRouter.Path("/{id}").Subrouter()
	roleAuthRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteRoleAuth))

	roleAuthsRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getRoleAuths))
	roleAuthsRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(addRoleAuth))

	roleRouter := rolesRouter.Path("/{id}").Subrouter()
	roleRouter.Methods("DELETE").Handler(gaecontext.HTTPHandlerFunc(deleteRole))

	rolesRouter.Methods("POST").Handler(gaecontext.HTTPHandlerFunc(createRole))
	rolesRouter.Methods("GET").Handler(gaecontext.HTTPHandlerFunc(getRoles))

	indexRouter := router.PathPrefix("/").MatcherFunc(wantsHTML).Subrouter()
	indexRouter.MatcherFunc(isMobile).Handler(gaecontext.HTTPHandlerFunc(mobileIndex))
	indexRouter.NewRoute().Handler(gaecontext.HTTPHandlerFunc(index))

	http.Handle("/", router)

}
