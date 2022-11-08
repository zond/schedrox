package web

import (
	"bytes"
	"io"
	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/crm"
	"github.com/zond/schedrox/event"

	"github.com/gorilla/mux"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
)

func deleteContact(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Contacts,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		crm.DeleteContact(data.context, key, data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func getContactEvents(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Events,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, event.GetAllowedContactEvents(data.context, data.domain, key, data.authorizer))
	}
	return
}

func updateContact(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Contacts,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		if old := crm.GetContact(data.context, key, data.domain); old == nil {
			w.WriteHeader(404)
		} else {
			var neu crm.Contact
			common.MustDecodeJSON(r.Body, &neu)
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, old.CopyFrom(&neu).Save(data.context, data.domain))
		}
	}
	return
}

func contactBounceMessage(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	dom, err := datastore.DecodeKey(mux.Vars(r)["domain_id"])
	if err != nil {
		panic(err)
	}
	data := getBaseData(c, w, r)
	data.domain = dom
	if data.hasAuth(auth.Auth{
		AuthType: auth.Contacts,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		if contact := crm.GetContact(data.context, key, data.domain); contact == nil {
			w.WriteHeader(404)
		} else {
			common.SetContentType(w, "text/plain; charset=UTF-8", false)
			io.Copy(w, bytes.NewBufferString(contact.EmailBounce))
		}
	}
	return
}

func getContact(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Contacts,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		if contact := crm.GetContact(data.context, key, data.domain); contact == nil {
			w.WriteHeader(404)
		} else {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, contact)
		}
	}
	return
}

func getContacts(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Contacts,
	}) {
		r.ParseForm()
		offset := 0
		limit := 20
		if s := r.Form.Get("offset"); s != "" {
			offset = common.MustParseInt(s)
		}
		if s := r.Form.Get("limit"); s != "" {
			limit = common.MustParseInt(s)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		results, total := crm.GetContactsByDomain(data.context, data.domain, offset, limit)
		common.MustEncodeJSON(w, common.Page{
			Total:   total,
			Results: results,
		})
	}
	return
}

func searchContacts(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Contacts,
	}) {
		r.ParseForm()
		q := r.Form.Get("q")
		var result []crm.Contact
		if q != "" {
			result = crm.GetContactsByPrefix(data.context, data.domain, q)
		} else {
			result = make([]crm.Contact, 0)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, result)
	}
	return
}

func createContact(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var con crm.Contact
	common.MustDecodeJSON(r.Body, &con)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Contacts,
		Write:    true,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, con.Save(data.context, data.domain))
	}
	return
}
