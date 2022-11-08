package appuser

import (
	"fmt"
	"github.com/zond/schedrox/common"
	"time"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type UserProperty struct {
	Id         *datastore.Key `datastore:"-" json:"id"`
	Name       string         `json:"name"`
	AssignedAt time.Time      `json:"assigned_at"`
	ValidUntil time.Time      `json:"valid_until,omitempty"`
}

type dummy struct{}

func UserPropertyKeyForUser(c gaecontext.HTTPContext, name string, domainUser *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "UserPropertyForUser", name, 0, domainUser)
}

func CleanProperties(c gaecontext.HTTPContext) int {
	ids, err := datastore.NewQuery("UserPropertyForUser").KeysOnly().GetAll(c, nil)
	common.AssertOkError(err)
	keyMap := make(map[string]*datastore.Key)
	for _, id := range ids {
		keyMap[id.Parent().Encode()] = id
	}
	keys := make([]*datastore.Key, 0, len(keyMap))
	for keyS, _ := range keyMap {
		key, err := datastore.DecodeKey(keyS)
		if err != nil {
			panic(err)
		}
		keys = append(keys, key)
	}
	res := make([]dummy, len(keys))
	err = datastore.GetMulti(c, keys, res)
	var toDelete []*datastore.Key
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for index, serr := range merr {
				if _, ok := serr.(*datastore.ErrFieldMismatch); ok {
				} else if serr == datastore.ErrNoSuchEntity {
					toDelete = append(toDelete, keyMap[keys[index].Encode()])
				} else {
					panic(serr)
				}
			}
		} else {
			panic(err)
		}
	}
	if len(toDelete) > 0 {
		if err = datastore.DeleteMulti(c, toDelete); err != nil {
			log.Errorf(c, "When deleting %v: %v", toDelete, err)
		}
	}
	return len(toDelete)
}

func findUserProperty(c gaecontext.HTTPContext, key *datastore.Key) *UserProperty {
	var t UserProperty
	if err := datastore.Get(c, key, &t); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil
		}
		panic(err)
	}
	t.Id = key
	return &t
}

func GetUserProperty(c gaecontext.HTTPContext, key, domainUser, dom *datastore.Key) *UserProperty {
	if key == nil {
		return nil
	}
	if !dom.Equal(domainUser.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, domainUser))
	}
	if !domainUser.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", domainUser, key))
	}
	var t UserProperty
	if common.Memoize(c, userPropertyKeyForId(key), &t, func() interface{} {
		return findUserProperty(c, key)
	}) {
		return &t
	}
	return nil
}

func DeleteUserPropertyByName(c gaecontext.HTTPContext, name string, domainUser, dom *datastore.Key) {
	DeleteUserProperty(c, UserPropertyKeyForUser(c, name, domainUser), domainUser, dom)
}

func DeleteUserProperty(c gaecontext.HTTPContext, key, domainUser, dom *datastore.Key) {
	if !dom.Equal(domainUser.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, domainUser))
	}
	if !domainUser.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", domainUser, key))
	}
	if err := datastore.Delete(c, key); err != nil {
		panic(err)
	}
	common.MemDel(c, userPropertiesForUserKey(domainUser), allUserPropertiesForDomainKey(dom))
}

func (self *UserProperty) CopyFrom(o *UserProperty) *UserProperty {
	self.ValidUntil = o.ValidUntil
	self.AssignedAt = o.AssignedAt
	return self
}

func (self *UserProperty) Save(c gaecontext.HTTPContext, domainUser, dom *datastore.Key) *UserProperty {
	if !domainUser.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, domainUser))
	}
	var err error
	if self.Id == nil && self.AssignedAt.IsZero() {
		self.AssignedAt = time.Now()
	}
	self.Id, err = datastore.Put(c, UserPropertyKeyForUser(c, self.Name, domainUser), self)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, userPropertiesForUserKey(domainUser), allUserPropertiesForDomainKey(dom))
	return self
}

func findAllUserProperties(c gaecontext.HTTPContext, dom *datastore.Key) (result []UserProperty) {
	keys, err := datastore.NewQuery("UserPropertyForUser").Ancestor(dom).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetAllUserProperties(c gaecontext.HTTPContext, dom *datastore.Key) (result []UserProperty) {
	common.Memoize(c, allUserPropertiesForDomainKey(dom), &result, func() interface{} {
		return findAllUserProperties(c, dom)
	})
	if result == nil {
		result = make([]UserProperty, 0)
	}
	return
}

func findUserProperties(c gaecontext.HTTPContext, domainUser *datastore.Key) (result []UserProperty) {
	keys, err := datastore.NewQuery("UserPropertyForUser").Ancestor(domainUser).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetUserProperties(c gaecontext.HTTPContext, domainUser, dom *datastore.Key) (result []UserProperty) {
	if !domainUser.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, domainUser))
	}
	common.Memoize(c, userPropertiesForUserKey(domainUser), &result, func() interface{} {
		return findUserProperties(c, domainUser)
	})
	if result == nil {
		result = make([]UserProperty, 0)
	}
	return
}
