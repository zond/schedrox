package web

import (
	"monotone/se.oort.schedrox/auth"
	"monotone/se.oort.schedrox/common"

	"github.com/gorilla/mux"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
)

func getRoles(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	common.SetContentType(w, "application/json; charset=UTF-8", false)
	common.MustEncodeJSON(w, auth.GetRoles(data.context, auth.DomainRolesKey(data.context, data.domain), data.domain, data.authorizer))
	return
}

func deleteRole(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		auth.DeleteRole(data.context, key, auth.DomainRolesKey(data.context, data.domain), data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}

func createRole(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var role auth.Role
	common.MustDecodeJSON(r.Body, &role)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
		Write:    true,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, role.Save(data.context, auth.DomainRolesKey(data.context, data.domain), data.domain))
	}
	return
}

func getRoleAuths(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["role_id"])
		if err != nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		auths := auth.GetAuths(data.context, key, auth.DomainRolesKey(data.context, data.domain))
		for index, auth := range auths {
			auths[index] = auth.Translate(data.translations)
		}
		common.MustEncodeJSON(w, auths)
	}
	return
}

func addRoleAuth(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["role_id"])
		if err != nil {
			panic(err)
		}
		var a auth.Auth
		common.MustDecodeJSON(r.Body, &a)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, a.Save(data.context, key, auth.DomainRolesKey(data.context, data.domain), data.domain).Translate(data.translations))
	}
	return
}

func deleteRoleAuth(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
		Write:    true,
	}) {
		role_key, err := datastore.DecodeKey(mux.Vars(r)["role_id"])
		if err != nil {
			panic(err)
		}
		auth_id, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		auth.DeleteAuth(data.context, auth_id, role_key, auth.DomainRolesKey(data.context, data.domain), data.domain)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		w.WriteHeader(204)
	}
	return
}
