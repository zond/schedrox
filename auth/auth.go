package auth

import (
	"fmt"
	"monotone/se.oort.schedrox/common"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const (
	Roles               = "Roles"
	Attend              = "Attend"
	Participants        = "Participants"
	Users               = "Users"
	Events              = "Events"
	EventTypes          = "Event types"
	Domain              = "Domain"
	Contacts            = "Contacts"
	Attest              = "Attest"
	SalaryReport        = "Salary report"
	ReportHours         = "Report hours"
	SalaryConfiguration = "Salary configuration"
	roleRootName        = "Role{}"
)

func RoleKeyByName(c gaecontext.HTTPContext, name string, parent *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "Role", name, 0, parent)
}

func DeleteEncodedKeys(c gaecontext.HTTPContext) (deleted int) {
	auths := []Auth{}
	authIds, err := datastore.NewQuery("Auth").GetAll(c, &auths)
	common.AssertOkError(err)

	for index, id := range authIds {
		if id.Parent().Kind() == "DomainUser" {
			userKey, err := datastore.DecodeKey(id.Parent().StringID())
			if err == nil {
				newKey := datastore.NewKey(c, "Auth", id.StringID(), id.IntID(), datastore.NewKey(c, "DomainUser", userKey.StringID(), 0, id.Parent().Parent()))
				if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
					if _, err = datastore.Put(c, newKey, &auths[index]); err != nil {
						return
					}
					err = datastore.Delete(c, id)
					log.Infof(c, "### replaced %v with %v", id, newKey)
					deleted++
					return
				}, false); err != nil {
					panic(err)
				}
			}
		}
	}

	roles := []Role{}
	roleIds, err := datastore.NewQuery("Role").GetAll(c, &roles)
	common.AssertOkError(err)

	for index, id := range roleIds {
		if id.Parent().Kind() == "DomainUser" {
			userKey, err := datastore.DecodeKey(id.Parent().StringID())
			if err == nil {
				newKey := datastore.NewKey(c, "Role", id.StringID(), id.IntID(), datastore.NewKey(c, "DomainUser", userKey.StringID(), 0, id.Parent().Parent()))
				if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
					if _, err = datastore.Put(c, newKey, &roles[index]); err != nil {
						return
					}
					err = datastore.Delete(c, id)
					log.Infof(c, "### replaced %v with %v", id, newKey)
					deleted++
					return
				}, false); err != nil {
					panic(err)
				}
			}
		}
	}
	return deleted
}

type Authorizer interface {
	HasAnyAuth(match Auth) bool
	HasAuth(match Auth) bool
}

var authTypes = make(map[string]AuthType)

func init() {
	for _, code := range []AuthType{
		AuthType{
			Name:     Contacts,
			HasWrite: true,
		},
		AuthType{
			Name:     Roles,
			HasRole:  true,
			HasWrite: true,
		},
		AuthType{
			Name:               Attend,
			HasEventType:       true,
			HasEventKind:       true,
			HasParticipantType: true,
			HasLocation:        true,
		},
		AuthType{
			Name:     EventTypes,
			HasWrite: true,
		},
		AuthType{
			Name:               Participants,
			HasWrite:           true,
			HasEventType:       true,
			HasEventKind:       true,
			HasParticipantType: true,
			HasLocation:        true,
		},
		AuthType{
			Name:     Domain,
			HasWrite: true,
		},
		AuthType{
			Name:     Users,
			HasWrite: true,
		},
		AuthType{
			Name:         Events,
			HasLocation:  true,
			HasEventType: true,
			HasEventKind: true,
			HasWrite:     true,
		},
		AuthType{
			Name:     SalaryConfiguration,
			HasWrite: true,
		},
		AuthType{
			Name:               SalaryReport,
			HasEventType:       true,
			HasEventKind:       true,
			HasParticipantType: true,
			HasLocation:        true,
		},
		AuthType{
			Name:               ReportHours,
			HasEventType:       true,
			HasEventKind:       true,
			HasParticipantType: true,
			HasLocation:        true,
		},
		AuthType{
			Name:               Attest,
			HasEventType:       true,
			HasEventKind:       true,
			HasParticipantType: true,
			HasLocation:        true,
		},
	} {
		authTypes[code.String()] = code
	}
}

func AuthTypes(salaryMod bool) (result map[string]AuthType) {
	result = make(map[string]AuthType)
	for code, typ := range authTypes {
		if (code != Attest && code != SalaryReport) || salaryMod {
			result[code] = typ
		}
	}
	return
}

func DomainRolesKey(c gaecontext.HTTPContext, d *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "Roles", "Roles", 0, d)
}

func RolesKeyForParent(pa *datastore.Key) string {
	return fmt.Sprintf("Roles{Parent:%v}", pa)
}
func AuthsKeyForParent(p *datastore.Key) string {
	return fmt.Sprintf("Auths{Parent:%v}", p)
}

func AuthsKeyForDomainAndType(d *datastore.Key, typ string) string {
	return fmt.Sprintf("Auths{Domain:%v,AuthType:%v}", d, typ)
}

func rolesKeyForDomainAndName(d *datastore.Key, name string) string {
	return fmt.Sprintf("Roles{Domain:%v,Name:%v}", d, name)
}

type AuthType struct {
	Name               string `json:"name"`
	Translation        string `datastore:"-" json:"translation"`
	HasWrite           bool   `json:"has_write"`
	HasLocation        bool   `json:"has_location"`
	HasEventKind       bool   `json:"has_event_kind"`
	HasEventType       bool   `json:"has_event_type"`
	HasParticipantType bool   `json:"has_participant_type"`
	HasRole            bool   `json:"has_role"`
}

func (self AuthType) String() string {
	return self.Name
}

type Auth struct {
	Id              *datastore.Key `json:"id"`
	AuthType        string         `json:"auth_type"`
	Translation     string         `datastore:"-" json:"translation"`
	Write           bool           `json:"write"`
	Location        *datastore.Key `json:"location"`
	EventKind       *datastore.Key `json:"event_kind"`
	EventType       *datastore.Key `json:"event_type"`
	ParticipantType *datastore.Key `json:"participant_type"`
	Role            string         `json:"role"`
}

func (self Auth) Translate(translations map[string]string) (result Auth) {
	result = self
	result.Translation = translations[self.AuthType]
	return
}

func (self Auth) MatchesAny(c gaecontext.HTTPContext, match Auth) bool {
	if self.AuthType == match.AuthType {
		authType := authTypes[self.AuthType]
		if !authType.HasWrite || self.Write || !match.Write {
			if !authType.HasLocation || self.Location == nil || match.Location == nil || self.Location.Equal(match.Location) {
				if !authType.HasEventKind || self.EventKind == nil || match.EventKind == nil || self.EventKind.Equal(match.EventKind) {
					if !authType.HasEventType || self.EventType == nil || match.EventType == nil || self.EventType.Equal(match.EventType) {
						if !authType.HasParticipantType || self.ParticipantType == nil || match.ParticipantType == nil || self.ParticipantType.Equal(match.ParticipantType) {
							if !authType.HasRole || self.Role == "" || match.Role == "" || self.Role == match.Role {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (self Auth) Matches(c gaecontext.HTTPContext, match Auth) bool {
	if self.AuthType == match.AuthType {
		authType := authTypes[self.AuthType]
		if !authType.HasWrite || self.Write || !match.Write {
			if !authType.HasLocation || self.Location == nil || self.Location.Equal(match.Location) {
				if !authType.HasEventKind || self.EventKind == nil || self.EventKind.Equal(match.EventKind) {
					if !authType.HasEventType || self.EventType == nil || self.EventType.Equal(match.EventType) {
						if !authType.HasParticipantType || self.ParticipantType == nil || self.ParticipantType.Equal(match.ParticipantType) {
							if !authType.HasRole || self.Role == "" || self.Role == match.Role {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (self Auth) Key(c gaecontext.HTTPContext, parent *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "Auth", fmt.Sprintf("%+v", self), 0, parent)
}

func (self Auth) Save(c gaecontext.HTTPContext, parent, granp, dom *datastore.Key) Auth {
	if !granp.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", granp, parent))
	}
	var err error
	if self.Id, err = datastore.Put(c, self.Key(c, parent), &self); err != nil {
		panic(err)
	}
	common.MemDel(c, AuthsKeyForDomainAndType(dom, self.AuthType))
	common.MemDel(c, AuthsKeyForParent(parent))
	return self
}

func DeleteAuth(c gaecontext.HTTPContext, key, parent, granp, dom *datastore.Key) {
	if !parent.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", parent, key))
	}
	if !granp.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", granp, parent))
	}
	var original Auth
	err := datastore.Get(c, key, &original)
	if err != nil {
		if err != datastore.ErrNoSuchEntity {
			panic(err)
		}
	} else {
		if err = datastore.Delete(c, key); err != nil {
			panic(err)
		}
		common.MemDel(c, AuthsKeyForDomainAndType(dom, original.AuthType))
		common.MemDel(c, AuthsKeyForParent(parent))
	}
}

func FindAuths(c gaecontext.HTTPContext, parent *datastore.Key) (result []Auth) {
	keys, err := datastore.NewQuery("Auth").Ancestor(parent).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetAuths(c gaecontext.HTTPContext, parent *datastore.Key, granp *datastore.Key) (result []Auth) {
	if !granp.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", granp, parent))
	}
	common.Memoize(c, AuthsKeyForParent(parent), &result, func() interface{} {
		return FindAuths(c, parent)
	})
	if result == nil {
		result = make([]Auth, 0)
	}
	return
}

func findAuthsByType(c gaecontext.HTTPContext, domain *datastore.Key, typ string) (result []Auth) {
	keys, err := datastore.NewQuery("Auth").Ancestor(domain).Filter("AuthType=", typ).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetAuthsByType(c gaecontext.HTTPContext, domain *datastore.Key, typ string) (result []Auth) {
	common.Memoize2(c, AuthsKeyForDomainAndType(domain, typ), "All", &result, func() interface{} {
		return findAuthsByType(c, domain, typ)
	})
	if result == nil {
		result = make([]Auth, 0)
	}
	return
}

type Role struct {
	Id   *datastore.Key `datastore:"-" json:"id"`
	Name string         `json:"name"`
}

func (self *Role) memClear(c gaecontext.HTTPContext, parent, dom *datastore.Key) {
	common.MemDel(c, RolesKeyForParent(parent))
	common.MemDel(c, rolesKeyForDomainAndName(dom, self.Name))
	for _, auth := range GetAuths(c, self.Id, parent) {
		common.MemDel(c, AuthsKeyForDomainAndType(dom, auth.AuthType))
	}
}

func DeleteRoleByName(c gaecontext.HTTPContext, name string, parent, dom *datastore.Key) {
	DeleteRole(c, RoleKeyByName(c, name, parent), parent, dom)
}

func DeleteRole(c gaecontext.HTTPContext, r, parent, dom *datastore.Key) {
	if !dom.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, parent))
	}
	if !parent.Equal(r.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", parent, r))
	}
	var original Role
	err := datastore.Get(c, r, &original)
	if err != nil {
		if err != datastore.ErrNoSuchEntity {
			panic(err)
		}
	} else {
		original.Id = r
		if err = datastore.Delete(c, r); err != nil {
			panic(err)
		}
		(&original).memClear(c, parent, dom)
	}
}

func (self *Role) Save(c gaecontext.HTTPContext, parent, dom *datastore.Key) *Role {
	if !dom.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, parent))
	}
	var err error
	if self.Id, err = datastore.Put(c, RoleKeyByName(c, self.Name, parent), self); err != nil {
		panic(err)
	}
	self.memClear(c, parent, dom)
	return self
}

func FindRoles(c gaecontext.HTTPContext, parent *datastore.Key) (result []Role) {
	keys, err := datastore.NewQuery("Role").Ancestor(parent).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetRoles(c gaecontext.HTTPContext, parent, do *datastore.Key, authorizer Authorizer) (result []Role) {
	if !do.Equal(parent.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", do, parent))
	}
	preResult := []Role{}
	common.Memoize(c, RolesKeyForParent(parent), &preResult, func() interface{} {
		return FindRoles(c, parent)
	})
	for _, role := range preResult {
		if authorizer == nil || authorizer.HasAuth(Auth{
			AuthType: Roles,
			Role:     role.Id.StringID(),
		}) {
			result = append(result, role)
		}
	}
	if result == nil {
		result = make([]Role, 0)
	}
	return
}

func findRolesByName(c gaecontext.HTTPContext, dom *datastore.Key, name string) (result []Role) {
	keys, err := datastore.NewQuery("Role").Ancestor(dom).Filter("Name=", name).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetRolesByName(c gaecontext.HTTPContext, dom *datastore.Key, name string) (result []Role) {
	common.Memoize(c, rolesKeyForDomainAndName(dom, name), &result, func() interface{} {
		return findRolesByName(c, dom, name)
	})
	if result == nil {
		result = make([]Role, 0)
	}
	return
}
