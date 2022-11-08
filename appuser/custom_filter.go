package appuser

import (
	"fmt"
	"monotone/se.oort.schedrox/common"
	"strings"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
)

func customFiltersKeyForUser(k *datastore.Key) string {
	return fmt.Sprintf("CustomFilters{User:%v}", k)
}

type CustomFilter struct {
	Id              *datastore.Key `json:"id" datastore:"-"`
	Name            string         `json:"name"`
	LocationsString string         `json:"-"`
	LocationsBytes  []byte         `json:"-"`
	Locations       []string       `json:"locations" datastore:"-"`
	KindsString     string         `json:"-"`
	KindsBytes      []byte         `json:"-"`
	Kinds           []string       `json:"kinds" datastore:"-"`
	TypesString     string         `json:"-"`
	TypesBytes      []byte         `json:"-"`
	Types           []string       `json:"types" datastore:"-"`
	UsersString     string         `json:"-"`
	UsersBytes      []byte         `json:"-"`
	Users           []string       `json:"users" datastore:"-"`
}

func DeleteCustomFilter(c gaecontext.HTTPContext, key, domainUser, dom *datastore.Key) {
	if !dom.Equal(domainUser.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, domainUser))
	}
	if !domainUser.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", domainUser, key))
	}
	if err := datastore.Delete(c, key); err != nil {
		panic(err)
	}
	common.MemDel(c, customFiltersKeyForUser(domainUser))
}

func (self *CustomFilter) convertKeys(c gaecontext.HTTPContext, ary []string) {
	for index, s := range ary {
		if k, err := datastore.DecodeKey(s); err == nil {
			ary[index] = datastore.NewKey(c, k.Kind(), k.StringID(), k.IntID(), k.Parent()).Encode()
		}
	}
}

func (self *CustomFilter) process(c gaecontext.HTTPContext) *CustomFilter {
	if self.LocationsString == "" {
		self.Locations = make([]string, 0)
	} else {
		self.Locations = strings.Split(self.LocationsString, ",")
	}
	if len(self.LocationsBytes) > 0 {
		byteLocations := []string{}
		common.MustUnmarshalJSON(self.LocationsBytes, &byteLocations)
		self.Locations = append(self.Locations, byteLocations...)
	}
	if self.KindsString == "" {
		self.Kinds = make([]string, 0)
	} else {
		self.Kinds = strings.Split(self.KindsString, ",")
	}
	if len(self.KindsBytes) > 0 {
		byteKinds := []string{}
		common.MustUnmarshalJSON(self.KindsBytes, &byteKinds)
		self.Kinds = append(self.Kinds, byteKinds...)
	}
	if self.TypesString == "" {
		self.Types = make([]string, 0)
	} else {
		self.Types = strings.Split(self.TypesString, ",")
	}
	if len(self.TypesBytes) > 0 {
		byteTypes := []string{}
		common.MustUnmarshalJSON(self.TypesBytes, &byteTypes)
		self.Types = append(self.Types, byteTypes...)
	}
	if self.UsersString == "" {
		self.Users = make([]string, 0)
	} else {
		self.Users = strings.Split(self.UsersString, ",")
	}
	if len(self.UsersBytes) > 0 {
		byteUsers := []string{}
		common.MustUnmarshalJSON(self.UsersBytes, &byteUsers)
		self.Users = append(self.Users, byteUsers...)
	}
	self.convertKeys(c, self.Locations)
	self.convertKeys(c, self.Kinds)
	self.convertKeys(c, self.Types)
	self.convertKeys(c, self.Users)
	return self
}

func (self *CustomFilter) Save(c gaecontext.HTTPContext, domainUser, dom *datastore.Key) *CustomFilter {
	if !domainUser.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, domainUser))
	}
	var err error
	self.LocationsBytes = common.MustMarshalJSON(self.Locations)
	self.KindsBytes = common.MustMarshalJSON(self.Kinds)
	self.TypesBytes = common.MustMarshalJSON(self.Types)
	self.UsersBytes = common.MustMarshalJSON(self.Users)
	self.LocationsString = ""
	self.KindsString = ""
	self.TypesString = ""
	self.UsersString = ""
	self.Id, err = datastore.Put(c, datastore.NewKey(c, "CustomFilter", self.Name, 0, domainUser), self)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, customFiltersKeyForUser(domainUser))
	return self
}

func findCustomFilters(c gaecontext.HTTPContext, domainUser *datastore.Key) (result []CustomFilter) {
	keys, err := datastore.NewQuery("CustomFilter").Ancestor(domainUser).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	for index, key := range keys {
		result[index].Id = key
	}
	return
}

func GetCustomFilters(c gaecontext.HTTPContext, domainUser, dom *datastore.Key) (result []CustomFilter) {
	if !domainUser.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, domainUser))
	}
	common.Memoize(c, customFiltersKeyForUser(domainUser), &result, func() interface{} {
		return findCustomFilters(c, domainUser)
	})
	if result == nil {
		result = make([]CustomFilter, 0)
	}
	for index, _ := range result {
		(&result[index]).process(c)
	}
	return
}
