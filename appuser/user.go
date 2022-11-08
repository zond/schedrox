package appuser

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/domain"
	"github.com/zond/schedrox/search"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/mail"
	"google.golang.org/appengine/user"
)

var GetSalaryConfigs func(c gaecontext.HTTPContext, domains []*datastore.Key) (result []interface{})
var GetUnattestedUids func(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer, from, to time.Time, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool)
var GetAttestedUids func(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer, from, to time.Time, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool)

type userAuthorizer struct {
	user     *User
	auths    []auth.Auth
	owner    bool
	closed   bool
	disabled bool
	context  gaecontext.HTTPContext
}

func (self userAuthorizer) HasAuth(match auth.Auth) bool {
	if self.user.Admin {
		return true
	}
	if self.closed {
		return false
	}
	if self.owner {
		return true
	}
	if self.disabled {
		return false
	}
	for _, a := range self.auths {
		if a.Matches(self.context, match) {
			return true
		}
	}
	return false
}

func (self userAuthorizer) HasAnyAuth(match auth.Auth) bool {
	if self.user.Admin {
		return true
	}
	if self.closed {
		return false
	}
	if self.owner {
		return true
	}
	if self.disabled {
		return false
	}
	for _, a := range self.auths {
		if a.MatchesAny(self.context, match) {
			return true
		}
	}
	return false
}

func usersKeyForOffsetLimit(offset, limit int) string {
	return fmt.Sprintf("Users{Offset:%v,Limit:%v}", offset, limit)
}

func usersKeyForDomain(domain *datastore.Key) string {
	return fmt.Sprintf("Users{Domain:%v}", domain)
}

func usersKeyForPrefix(prefix string) string {
	return fmt.Sprintf("Users{Prefix:%v}", prefix)
}

func UserKeyForId(id *datastore.Key) string {
	return fmt.Sprintf("User{Id:%v}", id)
}

func userKeyForEmailLowercase(email string) string {
	return fmt.Sprintf("User{EmailLowercase:%v}", email)
}

func allUserPropertiesForDomainKey(k *datastore.Key) string {
	return fmt.Sprintf("UserPropertiesForUser{Domain:%v}", k)
}

func userPropertiesForUserKey(k *datastore.Key) string {
	return fmt.Sprintf("UserPropertiesForUser{User:%v}", k)
}

func userPropertyKeyForId(k *datastore.Key) string {
	return fmt.Sprintf("UserPropertyForUser{Id:%v}", k)
}

func emailsKeyForDomain(k *datastore.Key) string {
	return fmt.Sprintf("User.Emails{Domain:%v}", k)
}

type Users []User

func (self Users) Len() int {
	return len(self)
}
func (self Users) Less(i, j int) bool {
	if cmp := bytes.Compare([]byte(self[i].GivenName), []byte(self[j].GivenName)); cmp < 0 {
		return true
	} else if cmp > 0 {
		return false
	}
	return bytes.Compare([]byte(self[i].GivenName), []byte(self[j].GivenName)) < 0
}
func (self Users) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type User struct {
	Id                     *datastore.Key  `datastore:"-" json:"id"`
	MobilePhone            string          `json:"mobile_phone"`
	GivenName              string          `json:"given_name"`
	FamilyName             string          `json:"family_name"`
	Email                  string          `json:"email"`
	EmailLowercase         string          `json:"-"`
	Admin                  bool            `json:"admin"`
	Domains                []domain.Domain `datastore:"-" json:"domains"`
	GravatarHash           string          `datastore:"-" json:"gravatar_hash"`
	EmailBounceBytes       []byte          `json:"-"`
	LastLanguage           string          `json:"-"`
	EmailBounce            string          `json:"email_bounce" datastore:"-"`
	MuteEventNotifications bool            `json:"mute_event_notifications"`
	BackgroundColor        string          `json:"background_color"`
	CalendarDaysBack       int             `json:"calendar_days_back"`
	CalendarWidth          int             `json:"calendar_width"`
	CalendarHeight         int             `json:"calendar_height"`
	HasInvalidProperty     bool            `json:"has_invalid_property" datastore:"-"`
	DefaultLocation        *datastore.Key  `json:"default_location,omitempty"`
	DefaultEventKind       *datastore.Key  `json:"default_event_kind,omitempty"`
	DefaultEventType       *datastore.Key  `json:"default_event_type,omitempty"`
	DefaultParticipantType *datastore.Key  `json:"default_participant_type,omitempty"`

	// Salary cache
	CachedProperties map[string]interface{} `json:"properties,omitempty" datastore:"-"`

	// PreProcess cache
	cachedProperties    map[string][]UserProperty
	cachedDomainModels  map[string]domain.Domain
	cachedDomainUsers   []DomainUser
	cachedSalaryConfigs map[string]interface{}
	cachedAllDomains    []domain.Domain
}

type profileStruct struct {
	Id           *datastore.Key `json:"id"`
	GravatarHash string         `json:"gravatar_hash"`
	FamilyName   string         `json:"family_name,omitempty"`
	GivenName    string         `json:"given_name,omitempty"`
}

func (self *User) ProfileData(authorizer auth.Authorizer) (result profileStruct) {
	result = profileStruct{
		Id:           self.Id,
		GravatarHash: self.GravatarHash,
	}
	if authorizer.HasAuth(auth.Auth{
		AuthType: auth.Users,
	}) {
		result.FamilyName = self.FamilyName
		result.GivenName = self.GivenName
	}
	return
}

func (self *User) CopySettingsFrom(o *User) *User {
	self.MuteEventNotifications = o.MuteEventNotifications
	self.BackgroundColor = o.BackgroundColor
	self.CalendarDaysBack = o.CalendarDaysBack
	self.CalendarWidth = o.CalendarWidth
	self.CalendarHeight = o.CalendarHeight
	self.DefaultLocation = o.DefaultLocation
	self.DefaultEventKind = o.DefaultEventKind
	self.DefaultEventType = o.DefaultEventType
	self.DefaultParticipantType = o.DefaultParticipantType
	return self
}

func (self *User) SendMail(c gaecontext.HTTPContext, domainKey *datastore.Key, subject, body, extra_bcc string, attachment *mail.Attachment) {
	replyTo := fmt.Sprintf("noreply@%v.appspotmail.com", appengine.AppID(c))
	dom := domain.GetDomain(c, domainKey)
	if dom.FromAddress != "" {
		replyTo = dom.FromAddress
	}
	var bcc []string
	if extra_bcc != "" {
		bcc = []string{extra_bcc}
	}
	var attachments []mail.Attachment
	if attachment != nil {
		attachments = append(attachments, *attachment)
	}
	msg := &mail.Message{
		Sender:      fmt.Sprintf("%v <ub+%v@%v.appspotmail.com>", common.SenderUser(c), common.EncodeBase64(self.Email), appengine.AppID(c)),
		ReplyTo:     replyTo,
		To:          []string{self.Email},
		Subject:     subject,
		Body:        body,
		Bcc:         bcc,
		Attachments: attachments,
	}
	if appengine.AppID(c) != "schedev" {
		if err := mail.Send(c, msg); err != nil {
			buf := new(bytes.Buffer)
			fmt.Fprintf(buf, "%v while sending %v", err, string(buf.Bytes()))
			self.EmailBounce = string(buf.Bytes())
			self.Save(c)
		}
	} else {
		log.Printf("Would have sent\n%v\nif appengine.AppID(c) [%+v] == schedrox", msg, appengine.AppID(c))
	}
	if appengine.IsDevAppServer() {
		log.Printf("Body in above mail: %v", body)
		for _, a := range attachments {
			log.Printf("Attachment in above mail: %v => %v", a.Name, string(a.Data))
		}
	}
}

func findUserIdsByPrefix(c gaecontext.HTTPContext, dom *datastore.Key, prefix string) (result []*datastore.Key) {
	var domainUsers []DomainUser
	_, err := search.Search(c, dom, "DomainUser", "default", prefix, true, 128, &domainUsers)
	if err != nil {
		panic(err)
	}
	for _, domainUser := range domainUsers {
		result = append(result, domainUser.User)
	}
	return
}

func getUserIdsByPrefix(c gaecontext.HTTPContext, dom *datastore.Key, prefix string, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool) {
	done2 = make(chan bool)
	go func() {
		var result []*datastore.Key
		common.Memoize2(c, usersKeyForDomain(dom), usersKeyForPrefix(prefix), &result, func() interface{} {
			return findUserIdsByPrefix(c, dom, prefix)
		})
		if result == nil {
			result = make([]*datastore.Key, 0)
		}
		keys <- result
		if done1 != nil {
			<-done1
		}
		close(done2)
	}()
	return
}

func getUserIdsByNoRoleName(c gaecontext.HTTPContext, name string, domain *datastore.Key, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool) {
	done2 = make(chan bool)
	go func() {
		hasRoleMap := map[string]struct{}{}
		for _, role := range auth.GetRolesByName(c, domain, name) {
			if role.Id.Parent().Kind() == "DomainUser" {
				tmpId := datastore.NewKey(c, "User", role.Id.Parent().StringID(), 0, nil)
				hasRoleMap[tmpId.Encode()] = struct{}{}
			}
		}
		var result []*datastore.Key
		for _, uid := range GetAllUserIds(c, domain) {
			if _, found := hasRoleMap[uid.Encode()]; !found {
				result = append(result, uid)
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

func getUserIdsByRoleName(c gaecontext.HTTPContext, name string, domain *datastore.Key, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool) {
	done2 = make(chan bool)
	go func() {
		var result []*datastore.Key
		for _, role := range auth.GetRolesByName(c, domain, name) {
			if role.Id.Parent().Kind() == "DomainUser" {
				tmpId := datastore.NewKey(c, "User", role.Id.Parent().StringID(), 0, nil)
				result = append(result, tmpId)
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

func findUserIdsByProperty(c gaecontext.HTTPContext, name string, domain *datastore.Key) (result []*datastore.Key) {
	ids, err := datastore.NewQuery("UserPropertyForUser").Ancestor(domain).Filter("Name=", name).KeysOnly().GetAll(c, nil)
	if err != nil {
		panic(err)
	}
	for _, propId := range ids {
		if propId.Parent().Kind() == "DomainUser" {
			tmpId := datastore.NewKey(c, "User", propId.Parent().StringID(), 0, nil)
			result = append(result, tmpId)
		}
	}
	return
}

func getUserIdsByProperty(c gaecontext.HTTPContext, name string, domain *datastore.Key, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool) {
	done2 = make(chan bool)
	go func() {
		var result []*datastore.Key
		common.Memoize2(c, usersKeyForDomain(domain), domainUsersKeyForProperty(name), &result, func() interface{} {
			return findUserIdsByProperty(c, name, domain)
		})
		if result == nil {
			result = make([]*datastore.Key, 0)
		}
		keys <- result
		if done1 != nil {
			<-done1
		}
		close(done2)
	}()
	return
}

func getUserIdsByNoProperty(c gaecontext.HTTPContext, name string, domain *datastore.Key, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool) {
	done2 = make(chan bool)
	go func() {
		var hasProp []*datastore.Key
		common.Memoize2(c, usersKeyForDomain(domain), domainUsersKeyForProperty(name), &hasProp, func() interface{} {
			return findUserIdsByProperty(c, name, domain)
		})
		hasPropMap := map[string]struct{}{}
		for _, uid := range hasProp {
			hasPropMap[uid.Encode()] = struct{}{}
		}
		var result []*datastore.Key
		for _, uid := range GetAllUserIds(c, domain) {
			if _, found := hasPropMap[uid.Encode()]; !found {
				result = append(result, uid)
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

func getUserIdsByDisabled(c gaecontext.HTTPContext, wantDisabled bool, domain *datastore.Key, done1 chan bool, keys chan []*datastore.Key) (done2 chan bool) {
	done2 = make(chan bool)
	go func() {
		keys <- GetDisabledIds(c, domain, wantDisabled)
		if done1 != nil {
			<-done1
		}
		close(done2)
	}()
	return
}

func GetFilteredUsersByPrefix(c gaecontext.HTTPContext, dom *datastore.Key, authorizer auth.Authorizer, prefix string, filters []string) (result Users) {
	byPrefix := make(chan []*datastore.Key)
	byProp := make(chan []*datastore.Key)
	byRole := make(chan []*datastore.Key)
	byDisabled := make(chan []*datastore.Key)
	byUnattested := make(chan []*datastore.Key)
	byAttested := make(chan []*datastore.Key)
	byNoRole := make(chan []*datastore.Key)
	byNoProp := make(chan []*datastore.Key)

	prefixIds := make(map[string]bool)
	propIds := make(map[string]bool)
	roleIds := make(map[string]bool)
	disabledIds := make(map[string]bool)
	unattestedIds := make(map[string]bool)
	attestedIds := make(map[string]bool)
	noRoleIds := make(map[string]bool)
	noPropIds := make(map[string]bool)

	byPrefixDone := make(chan bool)
	byRoleDone := make(chan bool)
	byPropDone := make(chan bool)
	byDisabledDone := make(chan bool)
	byUnattestedDone := make(chan bool)
	byAttestedDone := make(chan bool)
	byNoRoleDone := make(chan bool)
	byNoPropDone := make(chan bool)

	go func() {
		for ids := range byPrefix {
			for _, id := range ids {
				prefixIds[id.Encode()] = true
			}
		}
		close(byPrefixDone)
	}()

	go func() {
		for ids := range byProp {
			for _, id := range ids {
				propIds[id.Encode()] = true
			}
		}
		close(byPropDone)
	}()

	go func() {
		for ids := range byRole {
			for _, id := range ids {
				roleIds[id.Encode()] = true
			}
		}
		close(byRoleDone)
	}()

	go func() {
		for ids := range byDisabled {
			for _, id := range ids {
				disabledIds[id.Encode()] = true
			}
		}
		close(byDisabledDone)
	}()

	go func() {
		for ids := range byUnattested {
			for _, id := range ids {
				unattestedIds[id.Encode()] = true
			}
		}
		close(byUnattestedDone)
	}()

	go func() {
		for ids := range byAttested {
			for _, id := range ids {
				attestedIds[id.Encode()] = true
			}
		}
		close(byAttestedDone)
	}()

	go func() {
		for ids := range byNoRole {
			for _, id := range ids {
				noRoleIds[id.Encode()] = true
			}
		}
		close(byNoRoleDone)
	}()

	go func() {
		for ids := range byNoProp {
			for _, id := range ids {
				noPropIds[id.Encode()] = true
			}
		}
		close(byNoPropDone)
	}()

	var done chan bool
	if prefix != "" {
		done = getUserIdsByPrefix(c, dom, prefix, done, byPrefix)
	}
	hasRoleFilter := false
	hasPropFilter := false
	hasDisabledFilter := false
	hasUnattestedFilter := false
	hasAttestedFilter := false
	hasNoRoleFilter := false
	hasNoPropFilter := false
	for _, filter := range filters {
		parts := strings.Split(filter, ":")
		if len(parts) != 2 {
			panic(fmt.Errorf("Weird filter: %v", filter))
		}
		if parts[0] == "role" {
			hasRoleFilter = true
			done = getUserIdsByRoleName(c, parts[1], dom, done, byRole)
		} else if parts[0] == "property" {
			hasPropFilter = true
			done = getUserIdsByProperty(c, parts[1], dom, done, byProp)
		} else if parts[0] == "disabled" {
			hasDisabledFilter = true
			done = getUserIdsByDisabled(c, parts[1] == "true", dom, done, byDisabled)
		} else if parts[0] == "unattested" {
			fromTo := strings.Split(parts[1], "-")
			hasUnattestedFilter = true
			from := time.Unix(common.MustParseInt64(fromTo[0]), 0)
			to := time.Unix(common.MustParseInt64(fromTo[1]), 0)
			log.Printf("Finding unattested users between %v and %v", from, to)
			done = GetUnattestedUids(c, dom, authorizer, from, to, done, byUnattested)
		} else if parts[0] == "attested" {
			fromTo := strings.Split(parts[1], "-")
			hasAttestedFilter = true
			from := time.Unix(common.MustParseInt64(fromTo[0]), 0)
			to := time.Unix(common.MustParseInt64(fromTo[1]), 0)
			log.Printf("Finding attested users between %v and %v", from, to)
			done = GetAttestedUids(c, dom, authorizer, from, to, done, byAttested)
		} else if parts[0] == "norole" {
			hasNoRoleFilter = true
			done = getUserIdsByNoRoleName(c, parts[1], dom, done, byNoRole)
		} else if parts[0] == "noprop" {
			hasNoPropFilter = true
			done = getUserIdsByNoProperty(c, parts[1], dom, done, byNoProp)
		}
	}

	if done != nil {
		<-done
	}
	close(byPrefix)
	close(byProp)
	close(byRole)
	close(byDisabled)
	close(byUnattested)
	close(byAttested)
	close(byNoRole)
	close(byNoProp)
	<-byPrefixDone
	<-byRoleDone
	<-byPropDone
	<-byDisabledDone
	<-byUnattestedDone
	<-byAttestedDone
	<-byNoRoleDone
	<-byNoPropDone

	// Collect the sets of users we want to look at
	var relevant []map[string]bool
	if prefix != "" {
		relevant = append(relevant, prefixIds)
	}
	if hasPropFilter {
		relevant = append(relevant, propIds)
	}
	if hasRoleFilter {
		relevant = append(relevant, roleIds)
	}
	if hasDisabledFilter {
		relevant = append(relevant, disabledIds)
	}
	if hasUnattestedFilter {
		relevant = append(relevant, unattestedIds)
	}
	if hasAttestedFilter {
		relevant = append(relevant, attestedIds)
	}
	if hasNoRoleFilter {
		relevant = append(relevant, noRoleIds)
	}
	if hasNoPropFilter {
		relevant = append(relevant, noPropIds)
	}

	// If there were any
	if len(relevant) > 0 {
		var matches []*datastore.Key

		intersects := true
		var key *datastore.Key
		var err error
		// Find their intersection
		for id, _ := range relevant[0] {
			intersects = true
			for _, ids := range relevant[1:] {
				if !ids[id] {
					intersects = false
					break
				}
			}
			if intersects {
				// Get the pointed at user
				key, err = datastore.DecodeKey(id)
				if err != nil {
					panic(err)
				}
				matches = append(matches, key)
			}
		}

		preResult := make(Users, len(matches))
		err = datastore.GetMulti(c, matches, preResult)
		common.AssertOkError(err, datastore.ErrNoSuchEntity)
		for index, id := range matches {
			if err == nil {
				preResult[index].Id = id
				result = append(result, preResult[index])
			} else if merr, ok := err.(appengine.MultiError); ok {
				if merr[index] == nil {
					preResult[index].Id = id
					result = append(result, preResult[index])
				} else if _, ok := merr[index].(*datastore.ErrFieldMismatch); ok {
					preResult[index].Id = id
					result = append(result, preResult[index])
				} else if merr[index] != datastore.ErrNoSuchEntity {
					panic(merr[index])
				}
			} else {
				panic(err)
			}
		}
		PreProcess(c, dom, result)
		for index, _ := range result {
			(&result[index]).Process(c).filter(dom)
		}

		sort.Sort(result)
	}
	if result == nil {
		result = make(Users, 0)
	}
	return
}

func (self *User) ToJSON() string {
	buffer := new(bytes.Buffer)
	common.MustEncodeJSON(buffer, self)
	return string(buffer.Bytes())
}

func (self *User) GetAuthorizer(c gaecontext.HTTPContext, dom *datastore.Key) auth.Authorizer {
	result := userAuthorizer{
		user:    self,
		auths:   self.GetAuths(c)[common.EncKey(dom)],
		closed:  false,
		context: c,
	}
	for _, domain := range self.Domains {
		if domain.Id.Equal(dom) {
			result.owner = domain.Owner
			result.disabled = domain.Disabled
			result.closed = domain.ClosedAndRedirectedTo != ""
		}
	}
	return result
}

func (self *User) GetAuthsForDomain(c gaecontext.HTTPContext, dom *datastore.Key) (auths []auth.Auth) {
	if self.Id == nil {
		return nil
	}
	for _, domain := range self.Domains {
		if (!domain.Owner && domain.Disabled) || (!self.Admin && domain.ClosedAndRedirectedTo != "") {
			return nil
		}
	}
	domainUserKey := DomainUserKeyUnderDomain(c, dom, self.Id)
	auths = auth.GetAuths(c, domainUserKey, dom)
	for _, role := range auth.GetRoles(c, domainUserKey, dom, nil) {
		domainRoles := auth.DomainRolesKey(c, dom)
		auths = append(auths, auth.GetAuths(c, datastore.NewKey(c, "Role", role.Name, 0, domainRoles), domainRoles)...)
	}
	return
}

func (self *User) GetAuths(c gaecontext.HTTPContext) (result map[string][]auth.Auth) {
	result = make(map[string][]auth.Auth)
	for _, domain := range self.Domains {
		result[common.EncKey(domain.Id)] = self.GetAuthsForDomain(c, domain.Id)
	}
	return
}

func (self *User) flushMemcache(c gaecontext.HTTPContext) *User {
	common.MemDel(c, UserKeyForId(self.Id))
	common.MemDel(c, domainsKeyForUser(self.Id))
	common.MemDel(c, userKeyForEmailLowercase(self.EmailLowercase))
	return self
}

// filter removes all except one domain from a user
func (self *User) filter(d *datastore.Key) *User {
	newDomains := make([]domain.Domain, 0, len(self.Domains))
	for _, domain := range self.Domains {
		if domain.Id.Equal(d) {
			newDomains = append(newDomains, domain)
		}
	}
	self.Domains = newDomains
	return self
}

func (self *User) GetPreCachedProperties(c gaecontext.HTTPContext, dom *datastore.Key) []UserProperty {
	return self.getProperties(c, dom)
}

func (self *User) getProperties(c gaecontext.HTTPContext, dom *datastore.Key) []UserProperty {
	if self.cachedProperties == nil {
		self.cachedProperties = make(map[string][]UserProperty)
	}
	if _, found := self.cachedProperties[dom.Encode()]; !found {
		self.cachedProperties[dom.Encode()] = GetUserProperties(c, DomainUserKeyUnderDomain(c, dom, self.Id), dom)
	}
	return self.cachedProperties[dom.Encode()]
}

func PreProcess(c gaecontext.HTTPContext, dom *datastore.Key, users Users) {
	domainMap := make(map[string]domain.Domain)
	var allDomains []domain.Domain
	for _, d := range domain.GetAll(c) {
		domainMap[d.Id.Encode()] = d
		allDomains = append(allDomains, d)
	}
	config := GetSalaryConfigs(c, []*datastore.Key{dom})[0]

	cacheKeys := make([]string, 0, len(users)*2)
	funcs := make([]func() interface{}, 0, len(users)*2)
	values := make([]interface{}, 0, len(users)*2)

	for index, _ := range users {
		users[index].cachedDomainModels = domainMap
		users[index].cachedAllDomains = allDomains
		users[index].cachedSalaryConfigs = make(map[string]interface{})
		users[index].cachedSalaryConfigs[dom.Encode()] = config

		userCopy := users[index]
		domainUserKey := DomainUserKeyUnderDomain(c, dom, userCopy.Id)
		cacheKeys = append(cacheKeys, userPropertiesForUserKey(domainUserKey))
		funcs = append(funcs, func() interface{} {
			return findUserProperties(c, domainUserKey)
		})
		var props []UserProperty
		values = append(values, &props)

		cacheKeys = append(cacheKeys, domainsKeyForUser(users[index].Id))
		funcs = append(funcs, func() interface{} {
			return (&userCopy).findDomainUsers(c)
		})
		var domainUsers []DomainUser
		values = append(values, &domainUsers)
	}

	common.MemoizeMulti(c, cacheKeys, values, funcs)

	for index, _ := range users {
		users[index].cachedProperties = make(map[string][]UserProperty)
		users[index].cachedProperties[dom.Encode()] = *(values[index*2].(*[]UserProperty))
		users[index].cachedDomainUsers = *(values[(index*2)+1].(*[]DomainUser))
	}
}

// Process decorates the user with all domains it belongs to (which is ALL if it is an Admin)
func (self *User) Process(c gaecontext.HTTPContext) *User {
	self.Domains = self.getDomains(c)
	self.EmailBounce = string(self.EmailBounceBytes)
	m := md5.New()
	io.WriteString(m, strings.ToLower(strings.TrimSpace(self.Email)))
	self.GravatarHash = fmt.Sprintf("%x", m.Sum(nil))
	for _, dom := range self.Domains {
		for _, prop := range self.getProperties(c, dom.Id) {
			if !prop.ValidUntil.IsZero() && time.Now().After(prop.ValidUntil) {
				self.HasInvalidProperty = true
			}
		}
	}
	return self
}

func (self *User) ActivateInDomain(c gaecontext.HTTPContext, dom *datastore.Key) {
	for _, d := range self.getDomains(c) {
		if d.Id.Equal(dom) {
			self.addToDomain(c, d.Id, d.Owner, false, d.AllowICS, d.Information, time.Now(), d.SalaryProperties)
		}
	}
}

func (self *User) MarkActive(c gaecontext.HTTPContext, language string) {
	for _, d := range self.getDomains(c) {
		self.addToDomain(c, d.Id, d.Owner, d.Disabled, d.AllowICS, d.Information, time.Now(), d.SalaryProperties)
	}
	self.LastLanguage = language
	self.Save(c)
}

// addToDomain adds the user to a domain, and flushes the key for the user and the domain
func (self *User) addToDomain(c gaecontext.HTTPContext, d *datastore.Key, owner, disabled, allowICS bool, information string, lastActivity time.Time, salaryPropertiesForUser map[string]interface{}) {
	serializedProps, err := json.Marshal(salaryPropertiesForUser)
	if err != nil {
		panic(err)
	}
	if err = c.Transaction(func(c gaecontext.HTTPContext) error {
		domainUser := &DomainUser{
			User:                       self.Id,
			Domain:                     d,
			FamilyName:                 self.FamilyName,
			GivenName:                  self.GivenName,
			Owner:                      owner,
			Disabled:                   disabled,
			AllowICS:                   allowICS,
			Information:                information,
			LastActivity:               lastActivity,
			SalarySerializedProperties: serializedProps,
		}
		id, err := datastore.Put(c, DomainUserKeyUnderDomain(c, d, self.Id), domainUser)
		if err != nil {
			panic(err)
		}
		_, err = datastore.Put(c, DomainUserKeyUnderUser(c, d, self.Id), domainUser)
		if err != nil {
			panic(err)
		}
		search.Deindex(c, id)
		search.IndexString(c, id, "default", strings.Join([]string{self.GivenName, self.FamilyName, self.Email, self.MobilePhone, information}, " "))
		common.MemDel(c, usersKeyForDomain(d))
		self.flushMemcache(c)
		return nil
	}, true); err != nil {
		panic(err)
	}
}

// removeFromDomain removes the user from a domain, and flushes the key for the user and the domain
func (self *User) removeFromDomain(c gaecontext.HTTPContext, d *datastore.Key) {
	if err := c.Transaction(func(c gaecontext.HTTPContext) error {
		domainUserKey := DomainUserKeyUnderDomain(c, d, self.Id)
		for _, a := range auth.GetAuths(c, domainUserKey, d) {
			auth.DeleteAuth(c, a.Id, domainUserKey, d, d)
		}
		for _, role := range auth.GetRoles(c, domainUserKey, d, nil) {
			auth.DeleteRole(c, role.Id, domainUserKey, d)
		}
		search.Deindex(c, domainUserKey)
		for _, prop := range GetUserProperties(c, domainUserKey, d) {
			DeleteUserProperty(c, prop.Id, domainUserKey, d)
		}
		if err := datastore.Delete(c, domainUserKey); err != nil {
			panic(err)
		}
		if err := datastore.Delete(c, DomainUserKeyUnderUser(c, d, self.Id)); err != nil {
			panic(err)
		}
		common.MemDel(c, usersKeyForDomain(d))
		self.flushMemcache(c)
		return nil
	}, true); err != nil {
		panic(err)
	}
}

func FindUser(c gaecontext.HTTPContext, key *datastore.Key) (result *User) {
	var u User
	err := datastore.Get(c, key, &u)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	common.AssertOkError(err)
	result = &u
	result.Id = key
	return
}

func findUsersByDomain(c gaecontext.HTTPContext, d *datastore.Key) (result Users) {
	var domainUsers []DomainUser
	_, err := datastore.NewQuery("DomainUser").Ancestor(d).Order("GivenName").Order("FamilyName").GetAll(c, &domainUsers)
	common.AssertOkError(err)

	var userKeys []*datastore.Key
	for _, domainUser := range domainUsers {
		userKeys = append(userKeys, domainUser.User)
	}
	result = make(Users, len(userKeys))

	err = datastore.GetMulti(c, userKeys, result)
	common.AssertOkError(err)
	for index, _ := range result {
		result[index].Id = userKeys[index]
	}
	return
}

func GetUserByKey(c gaecontext.HTTPContext, k *datastore.Key) *User {
	var user User
	if common.Memoize(c, UserKeyForId(k), &user, func() interface{} {
		rval := FindUser(c, k)
		log.Printf("Found %+v", rval)
		return rval
	}) {
		log.Printf("Processing %+v", user)
		return (&user).Process(c)
	}
	return nil
}

func findUserByEmailLowercase(c gaecontext.HTTPContext, email string) *User {
	users := Users{}
	ids, err := datastore.NewQuery("User").Filter("EmailLowercase=", strings.ToLower(email)).GetAll(c, &users)
	if err != nil {
		panic(err)
	}
	if len(users) != 1 {
		return nil
	}
	users[0].Id = ids[0]
	return (&users[0]).Process(c)
}

func GetUserByEmailLowercase(c gaecontext.HTTPContext, email string) *User {
	var user User
	if common.Memoize(c, userKeyForEmailLowercase(email), &user, func() interface{} {
		return findUserByEmailLowercase(c, email)
	}) {
		return (&user).Process(c)
	}
	return nil
}

func GetUser(c gaecontext.HTTPContext, email string) *User {
	if result := GetUserByKey(c, datastore.NewKey(c, "User", email, 0, nil)); result != nil {
		return result
	}
	return GetUserByEmailLowercase(c, strings.ToLower(email))
}

func (self *User) getDomainModels(c gaecontext.HTTPContext, domainIds []*datastore.Key) (result []domain.Domain) {
	if self.cachedDomainModels == nil {
		self.cachedDomainModels = make(map[string]domain.Domain)
	}
	result = make([]domain.Domain, len(domainIds))
	var missing []*datastore.Key
	for index, id := range domainIds {
		if d, found := self.cachedDomainModels[id.Encode()]; found {
			d.SalaryProperties = map[string]interface{}{}
			result[index] = d
		} else {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		loaded := make(map[string]domain.Domain)
		for _, d := range domain.GetDomains(c, missing) {
			d.SalaryProperties = map[string]interface{}{}
			self.cachedDomainModels[d.Id.Encode()] = d
			loaded[d.Id.Encode()] = d
		}
		for index, id := range domainIds {
			if d, found := loaded[id.Encode()]; found {
				result[index] = d
			}
		}
	}
	return
}

func (self *User) getSalaryConfigs(c gaecontext.HTTPContext, domainIds []*datastore.Key) (result []interface{}) {
	if self.cachedSalaryConfigs == nil {
		self.cachedSalaryConfigs = make(map[string]interface{})
	}
	result = make([]interface{}, len(domainIds))
	var missing []*datastore.Key
	for index, id := range domainIds {
		if conf, found := self.cachedSalaryConfigs[id.Encode()]; found {
			result[index] = conf
		} else {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		loaded := make(map[string]interface{})
		for index, conf := range GetSalaryConfigs(c, missing) {
			self.cachedSalaryConfigs[missing[index].Encode()] = conf
			loaded[missing[index].Encode()] = conf
		}
		for index, id := range domainIds {
			if conf, found := loaded[id.Encode()]; found {
				result[index] = conf
			}
		}
	}
	return
}

func (self *User) getAllDomains(c gaecontext.HTTPContext) []domain.Domain {
	if self.cachedAllDomains == nil {
		self.cachedAllDomains = domain.GetAll(c)
	}
	return self.cachedAllDomains
}

func (self *User) getDomains(c gaecontext.HTTPContext) (result []domain.Domain) {
	// Get all domain - user connections for user
	domainUsers := self.getDomainUsers(c)
	// Get all domain ids
	domainIds := make([]*datastore.Key, len(domainUsers))
	for index, _ := range domainUsers {
		domainIds[index] = domainUsers[index].Domain
	}
	// Get the domains
	result = self.getDomainModels(c, domainIds)
	// Get the domain salary configs
	configs := self.getSalaryConfigs(c, domainIds)
	// Decorate the domains with the configs and the domain - user connections
	for index, domainUser := range domainUsers {
		if configs[index] == nil {
			result[index].SalaryConfig = make(map[string]interface{})
		} else {
			result[index].SalaryConfig = configs[index]
		}
		result[index].Owner = domainUser.Owner
		result[index].Disabled = domainUser.Disabled
		result[index].LastActivity = domainUser.LastActivity
		result[index].Information = domainUser.Information
		result[index].AllowICS = domainUser.AllowICS
		if domainUser.SalarySerializedProperties != nil {
			if err := json.Unmarshal(domainUser.SalarySerializedProperties, &(result[index].SalaryProperties)); err != nil {
				log.Printf("Unable to unserialize %#v", string(domainUser.SalarySerializedProperties))
			}
		}
	}
	// Admins must have all domains..
	if self.Admin {
		for _, dom := range self.getAllDomains(c) {
			already := false
			for _, resDom := range result {
				if resDom.Id.Equal(dom.Id) {
					already = true
					break
				}
			}
			if !already {
				dom.Owner = true
				dom.LastActivity = time.Now()
				result = append(result, dom)
			}
		}
	}
	// Deactivate if not active enough
	for index, dom := range result {
		if !dom.Owner && !dom.Disabled && dom.AutoDisable && dom.AutoDisableAfter > 0 && time.Now().Sub(dom.LastActivity) > time.Hour*24*time.Duration(dom.AutoDisableAfter) {
			result[index].Disabled = true
			self.addToDomain(c, dom.Id, dom.Owner, true, dom.AllowICS, dom.Information, dom.LastActivity, dom.SalaryProperties)
		}
	}
	if result == nil {
		result = make([]domain.Domain, 0)
	}
	return
}

func (self *User) FirstDomain() *domain.Domain {
	if self == nil {
		return nil
	}
	if len(self.Domains) > 0 {
		return &self.Domains[0]
	}
	return nil
}

func (self *User) refreshDomains(c gaecontext.HTTPContext, except *datastore.Key) {
	for _, domain := range self.getDomains(c) {
		if !domain.Id.Equal(except) {
			self.addToDomain(c, domain.Id, domain.Owner, domain.Disabled, domain.AllowICS, domain.Information, domain.LastActivity, domain.SalaryProperties)
		}
	}
}

// SaveInDomain updates a user and adds it to all its domains (possible as owner) (to update ownership and family name for all domains)
// Only allows updating, does not allow creation.
func (self *User) SaveInDomain(c gaecontext.HTTPContext, dom domain.Domain) *User {
	if newKey := datastore.NewKey(c, "User", self.Email, 0, nil); newKey.Equal(self.Id) {
		var err error
		self.EmailBounceBytes = []byte(self.EmailBounce)
		self.EmailLowercase = strings.ToLower(self.Email)
		if self.Id, err = datastore.Put(c, self.Id, self); err != nil {
			panic(err)
		}
		var lastActivity time.Time
		var oldDomain *domain.Domain
		for _, od := range self.getDomains(c) {
			if od.Id.Equal(dom.Id) {
				oldDomain = &od
			}
		}
		if oldDomain != nil {
			lastActivity = oldDomain.LastActivity
			if oldDomain.Disabled && !dom.Disabled {
				lastActivity = time.Now()
			}
		}
		self.addToDomain(c, dom.Id, dom.Owner, dom.Disabled, dom.AllowICS, dom.Information, lastActivity, dom.SalaryProperties)
		self.refreshDomains(c, dom.Id)
		self.Process(c).filter(dom.Id)
	}
	return self
}

// save updates a user and cleans it from cache
func (self *User) Save(c gaecontext.HTTPContext) *User {
	var err error
	self.EmailLowercase = strings.ToLower(self.Email)
	self.EmailBounceBytes = []byte(self.EmailBounce)
	if self.Id, err = datastore.Put(c, datastore.NewKey(c, "User", self.Email, 0, nil), self); err != nil {
		panic(err)
	}
	self.flushMemcache(c)
	return self.Process(c).filter(nil)
}

// AddToDomain creates the user if it didn't exist, then adds it to the domain. Never as owner or disabled.
func (self *User) AddToDomain(c gaecontext.HTTPContext, domainKey *datastore.Key) *User {
	self.Id = datastore.NewKey(c, "User", self.Email, 0, nil)
	if current := GetUserByKey(c, self.Id); current == nil {
		var err error
		self.EmailBounceBytes = []byte(self.EmailBounce)
		self.EmailLowercase = strings.ToLower(self.Email)
		if self.Id, err = datastore.Put(c, self.Id, self); err != nil {
			panic(err)
		}
	}
	self.addToDomain(c, domainKey, false, false, false, "", time.Now(), nil)
	self.refreshDomains(c, domainKey)
	return self.Process(c).filter(domainKey)
}

// DeleteUserFromDomain deletes a user from a domain, and then deletes the user if that was the last domain
func DeleteUserFromDomain(c gaecontext.HTTPContext, userKey, domainKey *datastore.Key) {
	if user := GetUserByKey(c, userKey); user != nil {
		for _, domain := range user.Domains {
			if domain.Id.Equal(domainKey) {
				user.removeFromDomain(c, domainKey)
				if len(user.Domains) == 1 {
					if err := datastore.Delete(c, user.Id); err != nil {
						panic(err)
					}
					return
				}
			}
		}
	}
}

func GetUsersByDomain(c gaecontext.HTTPContext, d *datastore.Key) (result Users) {
	common.Memoize2(c, usersKeyForDomain(d), "all", &result, func() interface{} {
		return findUsersByDomain(c, d)
	})
	PreProcess(c, d, result)
	for i, _ := range result {
		(&result[i]).Process(c).filter(d)
	}
	return
}

type APIUser struct {
	AppUserEmail string
}

func GetCurrentUser(c gaecontext.HTTPContext) (appuser *User, email *string) {
	log.Printf("getcurrentuser")
	if apiUserSecret := c.Req().Header.Get("x-api-user-secret"); apiUserSecret != "" {
		log.Printf("apiUserSecret %v", apiUserSecret)
		apiUser := &APIUser{}
		if err := datastore.Get(c, datastore.NewKey(c, "APIUser", apiUserSecret, 0, nil), apiUser); err != nil {
			log.Printf("Trying to load APIUser for %q: %v", apiUserSecret, err)
			return nil, nil
		}
		appuser = GetUser(c, apiUser.AppUserEmail)
		if appuser != nil {
			appuser.Process(c)
			email = &appuser.Email
			newDomains := make([]domain.Domain, 0)
			for _, dom := range appuser.Domains {
				if !dom.Disabled {
					newDomains = append(newDomains, dom)
				}
			}
			appuser.Domains = newDomains
			if len(appuser.Domains) == 0 {
				appuser = nil
			}
		}
		log.Printf("returning %v, %v", appuser, email)
		return appuser, email
	}

	u := user.Current(c)
	log.Printf("u is %+v", u)
	if u != nil {
		email = &u.Email
		appuser = GetUser(c, u.Email)
		if appuser != nil {
			if u.Admin != appuser.Admin {
				appuser.Admin = u.Admin
				appuser.Save(c)
			}
			appuser.Process(c)
		} else {
			if u.Admin {
				appuser = &User{
					Email: u.Email,
					Admin: true,
				}
				appuser.Save(c).Process(c)
			}
		}
	}

	if appuser != nil {
		newDomains := make([]domain.Domain, 0)
		for _, dom := range appuser.Domains {
			if !dom.Disabled {
				newDomains = append(newDomains, dom)
			}
		}
		appuser.Domains = newDomains
		if !u.Admin && len(appuser.Domains) == 0 {
			appuser = nil
		}
	}
	return
}

func findEmailsForDomain(c gaecontext.HTTPContext, d *datastore.Key) (result []string) {
	ids, err := datastore.NewQuery("DomainUser").Ancestor(d).KeysOnly().GetAll(c, nil)
	common.AssertOkError(err)
	var tmpKey *datastore.Key
	for _, id := range ids {
		tmpKey = datastore.NewKey(c, "User", id.StringID(), 0, nil)
		result = append(result, GetUserByKey(c, tmpKey).Email)
	}
	return
}

// GetUserFromDomain gets a user from a given domain, filtering out info about it existing in other domains
func GetUserFromDomain(c gaecontext.HTTPContext, d *datastore.Key, u *datastore.Key) *User {
	if user := GetUserByKey(c, u); user == nil {
		return nil
	} else {
		if user.Admin {
			return user.filter(d)
		}
		for _, domain := range user.Domains {
			if domain.Id.Equal(d) {
				return user.filter(d)
			}
		}
	}
	return nil
}
