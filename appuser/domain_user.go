package appuser

import (
	"fmt"
	"monotone/se.oort.schedrox/common"
	"time"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func DomainUserKeyUnderDomain(c gaecontext.HTTPContext, dom *datastore.Key, user *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "DomainUser", user.StringID(), 0, dom)
}

func DomainUserKeyUnderUser(c gaecontext.HTTPContext, dom *datastore.Key, user *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "DomainUser", dom.StringID(), 0, user)
}

func domainUsersKeyForProperty(prop string) string {
	return fmt.Sprintf("DomainUsers{Property:%v}", prop)
}

func disabledsKeyForDomain(dom *datastore.Key, disabled bool) string {
	return fmt.Sprintf("DomainUsers{Disabled:%v,Domain:%v}", disabled, dom)
}

func allUserIdsKeyForDomain(dom *datastore.Key) string {
	return fmt.Sprintf("DomainUsers{Domain:%v}", dom)
}

func domainOwnersKeyForDomain(dom *datastore.Key) string {
	return fmt.Sprintf("DomainUsers{Owner:true,Domain:%v}", dom)
}

func domainsKeyForUser(user *datastore.Key) string {
	return fmt.Sprintf("DomainUsers{User:%v}", user)
}

func DeleteEncodedKeys(c gaecontext.HTTPContext) (deleted int) {
	domainUsers := []DomainUser{}
	domainUserIds, err := datastore.NewQuery("DomainUser").GetAll(c, &domainUsers)
	common.AssertOkError(err)

	for index, id := range domainUserIds {
		userKey, err := datastore.DecodeKey(id.StringID())
		if err == nil {
			newId := datastore.NewKey(c, "DomainUser", userKey.StringID(), 0, id.Parent())
			if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
				if _, err = datastore.Put(c, newId, &domainUsers[index]); err != nil {
					return
				}
				if err = datastore.Delete(c, id); err != nil {
					return
				}
				deleted++
				log.Infof(c, "### replaced %v with %v", id, newId)

				customFilters := []CustomFilter{}
				customFilterIds, err := datastore.NewQuery("CustomFilter").Ancestor(id).GetAll(c, &customFilters)
				common.AssertOkError(err)
				for index, id := range customFilterIds {
					newFilterId := datastore.NewKey(c, "CustomFilter", id.StringID(), id.IntID(), newId)
					if _, err = datastore.Put(c, newFilterId, &customFilters[index]); err != nil {
						return
					}
					if err = datastore.Delete(c, id); err != nil {
						return
					}
					deleted++
					log.Infof(c, "### replaced %v with %v", id, newFilterId)
				}

				properties := []UserProperty{}
				propertyIds, err := datastore.NewQuery("UserPropertyForUser").Ancestor(id).GetAll(c, &properties)
				common.AssertOkError(err)
				for index, id := range propertyIds {
					newPropertyId := datastore.NewKey(c, "UserPropertyForUser", id.StringID(), id.IntID(), newId)
					if _, err = datastore.Put(c, newPropertyId, &properties[index]); err != nil {
						return
					}
					if err = datastore.Delete(c, id); err != nil {
						return
					}
					deleted++
					log.Infof(c, "### replaced %v with %v", id, newPropertyId)
				}
				return
			}, true); err != nil {
				panic(err)
			}
		}
	}
	return
}

type DomainUser struct {
	User         *datastore.Key
	Domain       *datastore.Key
	FamilyName   string
	GivenName    string
	Owner        bool
	Disabled     bool
	Information  string
	AllowICS     bool
	LastActivity time.Time

	// Salary mod
	SalarySerializedProperties []byte
}

func (self *User) findDomainUsers(c gaecontext.HTTPContext) (result []DomainUser) {
	_, err := datastore.NewQuery("DomainUser").Ancestor(self.Id).GetAll(c, &result)
	common.AssertOkError(err)
	return
}

func (self *User) getDomainUsers(c gaecontext.HTTPContext) []DomainUser {
	if self.cachedDomainUsers == nil {
		common.Memoize(c, domainsKeyForUser(self.Id), &self.cachedDomainUsers, func() interface{} {
			return self.findDomainUsers(c)
		})
		if self.cachedDomainUsers == nil {
			self.cachedDomainUsers = make([]DomainUser, 0)
		}
	}
	return self.cachedDomainUsers
}

func findAllUserIds(c gaecontext.HTTPContext, dom *datastore.Key) (result []*datastore.Key) {
	keys, err := datastore.NewQuery("DomainUser").Ancestor(dom).KeysOnly().GetAll(c, nil)
	if err != nil {
		panic(err)
	}
	var tmpKey *datastore.Key
	for _, key := range keys {
		tmpKey = datastore.NewKey(c, "User", key.StringID(), 0, nil)
		result = append(result, tmpKey)
	}
	return
}

func GetAllUserIds(c gaecontext.HTTPContext, dom *datastore.Key) (result []*datastore.Key) {
	common.Memoize2(c, usersKeyForDomain(dom), allUserIdsKeyForDomain(dom), &result, func() interface{} {
		return findAllUserIds(c, dom)
	})
	if result == nil {
		result = make([]*datastore.Key, 0)
	}
	return
}

func findDomainOwnerIds(c gaecontext.HTTPContext, dom *datastore.Key) (result []*datastore.Key) {
	keys, err := datastore.NewQuery("DomainUser").Filter("Owner=", true).Ancestor(dom).KeysOnly().GetAll(c, nil)
	if err != nil {
		panic(err)
	}
	var tmpKey *datastore.Key
	for _, key := range keys {
		tmpKey = datastore.NewKey(c, "User", key.StringID(), 0, nil)
		result = append(result, tmpKey)
	}
	return
}

func GetDomainOwnerIds(c gaecontext.HTTPContext, dom *datastore.Key) (result []*datastore.Key) {
	common.Memoize2(c, usersKeyForDomain(dom), domainOwnersKeyForDomain(dom), &result, func() interface{} {
		return findDomainOwnerIds(c, dom)
	})
	if result == nil {
		result = make([]*datastore.Key, 0)
	}
	return
}

func findDisabledIds(c gaecontext.HTTPContext, dom *datastore.Key, disabled bool) (result []*datastore.Key) {
	keys, err := datastore.NewQuery("DomainUser").Filter("Disabled=", disabled).Ancestor(dom).KeysOnly().GetAll(c, nil)
	if err != nil {
		panic(err)
	}
	var tmpKey *datastore.Key
	for _, key := range keys {
		tmpKey = datastore.NewKey(c, "User", key.StringID(), 0, nil)
		result = append(result, tmpKey)
	}
	return
}

func GetDisabledIds(c gaecontext.HTTPContext, dom *datastore.Key, disabled bool) (result []*datastore.Key) {
	common.Memoize2(c, usersKeyForDomain(dom), disabledsKeyForDomain(dom, disabled), &result, func() interface{} {
		return findDisabledIds(c, dom, disabled)
	})
	if result == nil {
		result = make([]*datastore.Key, 0)
	}
	return
}
