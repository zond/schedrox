package search

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"github.com/zond/schedrox/common"
	"reflect"
	"strings"
)

type Term struct {
	Kind  string
	Field string
	Value string
}

type dummy struct{}

func Clean(c context.Context) int {
	ids, err := datastore.NewQuery("Term").KeysOnly().GetAll(c, nil)
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

func Deindex(c context.Context, doc *datastore.Key) {
	ids, err := datastore.NewQuery("Term").Ancestor(doc).KeysOnly().GetAll(c, nil)
	if err != nil {
		panic(err)
	}
	if err = datastore.DeleteMulti(c, ids); err != nil {
		panic(err)
	}
}

func DeleteEncodedKeys(c context.Context) (deleted int) {
	terms := []Term{}
	termIds, err := datastore.NewQuery("Term").Filter("Kind=", "DomainUser").GetAll(c, &terms)
	common.AssertOkError(err)

	for index, id := range termIds {
		userKey, err := datastore.DecodeKey(id.Parent().StringID())
		if err == nil {
			newKey := datastore.NewKey(c, "Term", id.StringID(), id.IntID(), datastore.NewKey(c, "DomainUser", userKey.StringID(), 0, id.Parent().Parent()))
			if err := datastore.RunInTransaction(c, func(c context.Context) (err error) {
				if _, err = datastore.Put(c, newKey, &terms[index]); err != nil {
					return
				}
				err = datastore.Delete(c, id)
				log.Infof(c, "### replaced %v with %v", id, newKey)
				deleted++
				return
			}, &datastore.TransactionOptions{XG: false}); err != nil {
				panic(err)
			}
		}
	}
	return deleted
}

func IndexString(c context.Context, doc *datastore.Key, field, values string) {
	var terms []Term
	var keys []*datastore.Key
	for _, value := range strings.Split(values, " ") {
		trimmed := strings.TrimSpace(strings.ToLower(value))
		if len(trimmed) > 1 {
			terms = append(terms, Term{
				Kind:  doc.Kind(),
				Field: field,
				Value: trimmed,
			})
			keys = append(keys, datastore.NewKey(c, "Term", "", 0, doc))
		}
	}
	if _, err := datastore.PutMulti(c, keys, terms); err != nil {
		panic(err)
	}
}

func findIds(c context.Context, ancestor *datastore.Key, kind, field, badValue string, prefix bool, max int, funnel chan *datastore.Key, done1 chan bool) (done2 chan bool) {
	done2 = make(chan bool)
	go func() {
		value := strings.ToLower(badValue)
		query := datastore.NewQuery("Term").Ancestor(ancestor).Order("Value").Filter("Field=", field).Filter("Kind=", kind).Limit(max)
		if prefix {
			query = query.Filter("Value>=", value)
		} else {
			query = query.Filter("Value=", value)
		}
		var terms []Term
		var termIds []*datastore.Key
		var err error
		if termIds, err = query.GetAll(c, &terms); err != nil {
			panic(err)
		}
		for index, term := range terms {
			if strings.Index(term.Value, value) == 0 {
				funnel <- termIds[index].Parent()
			}
		}
		if done1 != nil {
			<-done1
		}
		done2 <- true
	}()
	return
}

func Search(c context.Context, ancestor *datastore.Key, kind, field, query string, prefix bool, max int, results interface{}) (ids []*datastore.Key, err error) {
	collector := make(map[string]*datastore.Key)
	funnel := make(chan *datastore.Key)
	funnelDone := make(chan bool)
	go func() {
		for id := range funnel {
			collector[id.Encode()] = id
		}
		funnelDone <- true
	}()
	var done chan bool
	var trimmed string
	for _, value := range strings.Split(query, " ") {
		trimmed = strings.TrimSpace(value)
		if len(trimmed) > 1 {
			done = findIds(c, ancestor, kind, field, trimmed, prefix, max, funnel, done)
		}
	}
	if done != nil {
		<-done
	}
	close(funnel)
	<-funnelDone
	for _, id := range collector {
		ids = append(ids, id)
	}
	resultsValue := reflect.ValueOf(results)
	resultsValue.Elem().Set(reflect.MakeSlice(reflect.TypeOf(results).Elem(), len(ids), len(ids)))
	err = datastore.GetMulti(c, ids, resultsValue.Elem().Interface())
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for index, serr := range merr {
				if serr != nil {
					log.Errorf(c, "When trying to load %v: %v", ids[index], serr)
				}
			}
		} else {
			log.Errorf(c, "When trying to load %v: %v", ids, err)
		}
	}
	return
}
