package event

import (
	"encoding/json"
	"fmt"
	"monotone/se.oort.schedrox/common"
	"sort"
	"time"

	"github.com/zond/sybutils/utils/gae/gaecontext"
	"github.com/zond/sybutils/utils/gae/memcache"

	"google.golang.org/appengine/datastore"
)

func changesKeyForDomain(k *datastore.Key) string {
	return fmt.Sprintf("Changes{Domain:%v}", k)
}

func changesKeyForParent(k *datastore.Key) string {
	return fmt.Sprintf("Changes{Parent:%v}", k)
}

type Changes []Change

func (c Changes) Len() int {
	return len(c)
}

func (c Changes) Less(a, b int) bool {
	return c[a].At.After(c[b].At)
}

func (c Changes) Swap(a, b int) {
	c[a], c[b] = c[b], c[a]
}

type Change struct {
	At        time.Time      `json:"-"`
	Ago       int            `json:"ago" datastore:"-"`
	User      *datastore.Key `json:"user"`
	UserEmail string         `json:"user_email" datastore:"-"`
	Action    string         `json:"action"`
	DataBytes []byte         `json:"-"`
	Data      string         `json:"data" datastore:"-"`
}

func (self *Change) process(c gaecontext.HTTPContext) *Change {
	dataMap := map[string]interface{}{}
	if err := json.Unmarshal(self.DataBytes, &dataMap); err != nil {
		panic(err)
	}
	for k, v := range dataMap {
		if s, ok := v.(string); ok {
			if key, err := datastore.DecodeKey(s); err == nil {
				if key.StringID() != "" {
					dataMap[k+"_name"] = key.StringID()
				}
			}
		}
	}
	refinedData, err := json.MarshalIndent(dataMap, "", "  ")
	if err != nil {
		panic(err)
	}
	self.Data = string(refinedData)
	self.Ago = int(time.Now().Sub(self.At) / time.Second)
	self.UserEmail = self.User.StringID()
	return self
}

func CreateChange(c gaecontext.HTTPContext, parent, actor *datastore.Key, action string, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	entry := &Change{
		At:        time.Now(),
		Action:    action,
		User:      actor,
		DataBytes: bytes,
	}
	_, err = datastore.Put(c, datastore.NewKey(c, "Change", "", 0, parent), entry)
	if err != nil {
		panic(err)
	}
	memcache.Del(c, changesKeyForDomain(parent.Parent()))
	memcache.Del(c, changesKeyForParent(parent))
}

func findLatestChanges(c gaecontext.HTTPContext, domain *datastore.Key) (result []Change) {
	_, err := datastore.NewQuery("Change").Ancestor(domain).Order("-At").Limit(256).GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	return
}

func FindChangesFrom(c gaecontext.HTTPContext, domain *datastore.Key, from, to time.Time) (result []Change) {
	_, err := datastore.NewQuery("Change").Ancestor(domain).Filter("At>", from).Filter("At<", to).Order("-At").GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	return
}

func GetLatestChanges(c gaecontext.HTTPContext, domain *datastore.Key) (result Changes) {
	common.Memoize(c, changesKeyForDomain(domain), &result, func() interface{} {
		return findLatestChanges(c, domain)
	})
	for index, _ := range result {
		(&result[index]).process(c)
	}
	if result == nil {
		result = make(Changes, 0)
	}
	sort.Sort(result)
	return
}

func findChanges(c gaecontext.HTTPContext, parent *datastore.Key) (result []Change) {
	_, err := datastore.NewQuery("Change").Ancestor(parent).Order("At").GetAll(c, &result)
	if err != nil {
		panic(err)
	}
	return
}

func GetChanges(c gaecontext.HTTPContext, parent, domain *datastore.Key) (result []Change) {
	if !parent.Parent().Equal(domain) {
		panic(fmt.Errorf("%v is not parent of %v", domain, parent))
	}
	common.Memoize(c, changesKeyForParent(parent), &result, func() interface{} {
		return findChanges(c, parent)
	})
	for index, _ := range result {
		(&result[index]).process(c)
	}
	if result == nil {
		result = make([]Change, 0)
	}
	return
}
