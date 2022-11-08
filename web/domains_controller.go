package web

import (
	"fmt"
	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/domain"
	"github.com/zond/schedrox/salary"

	"github.com/gorilla/mux"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
)

func updateUserPropertyForDomain(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Domain,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		var toUpdate domain.UserProperty
		common.MustDecodeJSON(r.Body, &toUpdate)
		current := domain.GetUserProperty(data.context, key, data.domain)
		current.CopyFrom(&toUpdate)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, current.Save(data.context, data.domain))
	}
	return
}

func updateLocation(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Domain,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		var toUpdate domain.Location
		common.MustDecodeJSON(r.Body, &toUpdate)
		toUpdate.Id = key
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, toUpdate.Save(data.context, data.domain))
	}
	return
}

func updateDomain(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()

	var toUpdate domain.Domain
	common.MustDecodeJSON(r.Body, &toUpdate)
	key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}

	data := getBaseData(c, w, r)
	if authenticated, authorized := data.silentIsAdmin(); authenticated && authorized {
		toUpdate.Id = key
		toUpdate.Save(data.context)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, toUpdate)
	} else {
		if data.hasAuth(auth.Auth{
			AuthType: auth.Domain,
			Write:    true,
		}) {
			current := domain.GetDomain(data.context, key)
			current.CopyFrom(&toUpdate)
			current.Save(data.context)
			current.SalaryConfig = salary.GetConfig(data.context, current.Id)
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, current)
		}
	}
	return
}

func getLocation(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Domain,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		current := domain.GetLocation(data.context, key)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, current)
	}
	return
}

func getDomain(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Domain,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		current := domain.GetDomain(data.context, key)
		current.SalaryConfig = salary.GetConfig(data.context, current.Id)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, current)
	}
	return
}

func deleteDomain(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		domain.Destroy(data.context, key)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func createDomain(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		var d domain.Domain
		common.MustDecodeJSON(r.Body, &d)
		d.Save(data.context)
		common.MustEncodeJSON(w, d)
	}
	return
}

func getDomains(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAdmin() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, domain.GetAll(data.context))
	}
	return
}

func deleteUserPropertyForDomain(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Domain,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		domain.DeleteUserProperty(data.context, key, data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func deleteLocation(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Domain,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		domain.DeleteLocation(data.context, key, data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func createUserPropertyForDomain(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Domain,
		Write:    true,
	}) {
		var toCreate domain.UserProperty
		common.MustDecodeJSON(r.Body, &toCreate)
		common.MustEncodeJSON(w, toCreate.Save(data.context, data.domain))
	}
	return
}

func createLocation(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Domain,
		Write:    true,
	}) {
		var toCreate domain.Location
		common.MustDecodeJSON(r.Body, &toCreate)
		common.MustEncodeJSON(w, toCreate.Save(data.context, data.domain))
	}
	return
}

func getLocations(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, domain.GetLocations(data.context, data.domain, data.user.GetAuthorizer(data.context, data.domain)))
	}
	return
}

func getUserPropertiesForDomain(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	domainAuth, authenticated := data.silentHasAuth(auth.Auth{
		AuthType: auth.Domain,
	})
	usersAuth, _ := data.silentHasAuth(auth.Auth{
		AuthType: auth.Users,
		Write:    true,
	})
	if !authenticated || (!domainAuth && !usersAuth) {
		w.WriteHeader(401)
		fmt.Fprintln(w, "Unauthenticated")
	} else {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, domain.GetUserProperties(data.context, data.domain))
	}
	return
}
