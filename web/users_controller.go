package web

import (
	"bytes"
	"fmt"
	"io"
	"monotone/se.oort.schedrox/appuser"
	"monotone/se.oort.schedrox/auth"
	"monotone/se.oort.schedrox/common"
	"monotone/se.oort.schedrox/domain"
	"monotone/se.oort.schedrox/event"
	"time"

	"github.com/gorilla/mux"
	"github.com/zond/sybutils/utils/gae/gaecontext"
	"github.com/zond/sybutils/utils/web/jsoncontext"

	"google.golang.org/appengine/datastore"
)

func userBounceMessage(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	dom, err := datastore.DecodeKey(mux.Vars(r)["domain_id"])
	if err != nil {
		panic(err)
	}
	data := getBaseData(c, w, r)
	data.domain = dom
	if data.hasAuth(auth.Auth{
		AuthType: auth.Users,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		if user := appuser.GetUserFromDomain(data.context, data.domain, key); user == nil {
			w.WriteHeader(404)
		} else {
			common.SetContentType(w, "text/plain; charset=UTF-8", false)
			io.Copy(w, bytes.NewBufferString(user.EmailBounce))
		}
	}
	return
}

func searchUsers(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if authorized, authenticated := data.silentHasAuth(auth.Auth{
		AuthType: auth.Users,
	}); authorized {
		r.ParseForm()
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, appuser.GetFilteredUsersByPrefix(data.context, data.domain, data.authorizer, r.Form.Get("q"), r.Form["filter"]))
	} else if authenticated {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, appuser.Users{*data.user})
	} else {
		w.WriteHeader(401)
	}
	return
}

func deleteCustomFilter(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		appuser.DeleteCustomFilter(data.context, key, appuser.DomainUserKeyUnderDomain(data.context, data.domain, data.user.Id), data.domain)
		w.WriteHeader(204)
	}
	return
}

func getCustomFilters(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, appuser.GetCustomFilters(data.context, appuser.DomainUserKeyUnderDomain(data.context, data.domain, data.user.Id), data.domain))
	}
	return
}

func createCustomFilter(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		var filter appuser.CustomFilter
		common.MustDecodeJSON(r.Body, &filter)
		common.MustEncodeJSON(w, filter.Save(data.context, appuser.DomainUserKeyUnderDomain(data.context, data.domain, data.user.Id), data.domain))
	}
	return
}

func getUserRoles(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAnyAuth(auth.Auth{
		AuthType: auth.Roles,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, auth.GetRoles(data.context, appuser.DomainUserKeyUnderDomain(data.context, data.domain, key), data.domain, data.authorizer))
	}
	return
}

func deleteUserRole(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	role_id, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
		Role:     role_id.StringID(),
		Write:    true,
	}) {
		user_key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		auth.DeleteRole(data.context, role_id, appuser.DomainUserKeyUnderDomain(data.context, data.domain, user_key), data.domain)
		w.WriteHeader(204)
	}
	return
}

func updateUserPropertyForUser(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Users,
		Write:    true,
	}) {
		user_key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		var toUpdate appuser.UserProperty
		common.MustDecodeJSON(r.Body, &toUpdate)
		domainUserKey := appuser.DomainUserKeyUnderDomain(data.context, data.domain, user_key)
		current := appuser.GetUserProperty(data.context, key, domainUserKey, data.domain)
		current.CopyFrom(&toUpdate)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, current.Save(data.context, domainUserKey, data.domain))
	}
	return
}

func deleteUserPropertyForUser(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Users,
		Write:    true,
	}) {
		user_key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		prop_id, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		appuser.DeleteUserProperty(data.context, prop_id, appuser.DomainUserKeyUnderDomain(data.context, data.domain, user_key), data.domain)
		w.WriteHeader(204)
	}
	return
}

func createUserPropertyForUser(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Users,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		var prop appuser.UserProperty
		common.MustDecodeJSON(r.Body, &prop)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, prop.Save(data.context, appuser.DomainUserKeyUnderDomain(data.context, data.domain, key), data.domain))
	}
	return
}

type bulkUpdateRequest struct {
	Add      bool             `json:"add"`
	Users    []*datastore.Key `json:"users"`
	Role     string           `json:"role"`
	Property string           `json:"property"`
}

func bulkUserUpdate(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var req bulkUpdateRequest
	common.MustDecodeJSON(r.Body, &req)
	if req.Role != "" {
		if !data.hasAuth(auth.Auth{
			AuthType: auth.Roles,
			Role:     req.Role,
			Write:    true,
		}) {
			return
		}
	}
	now := time.Time{}
	validUntil := time.Time{}
	if req.Property != "" {
		if !data.hasAuth(auth.Auth{
			AuthType: auth.Users,
			Write:    true,
		}) {
			return
		}
		domainProperty := domain.GetUserProperty(c, domain.PropertyID(c, req.Property, data.domain), data.domain)
		now = time.Now()
		validUntil = now.AddDate(0, 0, domainProperty.DaysValid)
	}
	if req.Add {
		for _, id := range req.Users {
			domainUserKey := appuser.DomainUserKeyUnderDomain(data.context, data.domain, id)
			if req.Role != "" {
				(&auth.Role{
					Name: req.Role,
				}).Save(data.context, domainUserKey, data.domain)
			}
			if req.Property != "" {
				(&appuser.UserProperty{
					Name:       req.Property,
					AssignedAt: now,
					ValidUntil: validUntil,
				}).Save(data.context, domainUserKey, data.domain)
			}
		}
	} else {
		for _, id := range req.Users {
			domainUserKey := appuser.DomainUserKeyUnderDomain(data.context, data.domain, id)
			if req.Role != "" {
				auth.DeleteRoleByName(data.context, req.Role, domainUserKey, data.domain)
			}
			if req.Property != "" {
				appuser.DeleteUserPropertyByName(data.context, req.Property, domainUserKey, data.domain)
			}
		}
	}
	w.WriteHeader(204)
	return
}

func addUserRole(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var role auth.Role
	common.MustDecodeJSON(r.Body, &role)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
		Role:     role.Name,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, role.Save(data.context, appuser.DomainUserKeyUnderDomain(data.context, data.domain, key), data.domain))
	}
	return
}

func getUserAuths(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		auths := auth.GetAuths(data.context, appuser.DomainUserKeyUnderDomain(data.context, data.domain, key), data.domain)
		for index, auth := range auths {
			auths[index] = auth.Translate(data.translations)
		}
		common.MustEncodeJSON(w, auths)
	}
	return
}

func getUserPropertiesForUser(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Users,
	}) {
		userId, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, appuser.GetUserProperties(data.context, appuser.DomainUserKeyUnderDomain(data.context, data.domain, userId), data.domain))
	}
	return
}

func addUserAuth(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		var a auth.Auth
		common.MustDecodeJSON(r.Body, &a)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, a.Save(data.context, appuser.DomainUserKeyUnderDomain(data.context, data.domain, key), data.domain, data.domain).Translate(data.translations))
	}
	return
}

func deleteUserAuth(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Roles,
		Write:    true,
	}) {
		user_key, err := datastore.DecodeKey(mux.Vars(r)["user_id"])
		if err != nil {
			panic(err)
		}
		auth_id, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		auth.DeleteAuth(data.context, auth_id, appuser.DomainUserKeyUnderDomain(data.context, data.domain, user_key), data.domain, data.domain)
		w.WriteHeader(204)
	}
	return
}

func getProfile(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	if user := appuser.GetUserFromDomain(data.context, data.domain, key); user == nil {
		w.WriteHeader(404)
	} else {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, user.ProfileData(data.authorizer))
	}
	return
}

func logoutRedirect(c gaecontext.JSONContext) (resp jsoncontext.Resp, err error) {
	data := getBaseData(c, c.Resp(), c.Req())
	c.Resp().Header().Set("Location", fmt.Sprint(data.data["logoutUrl"]))
	resp.Status = 301
	return
}

func loginRedirect(c gaecontext.JSONContext) (resp jsoncontext.Resp, err error) {
	data := getBaseData(c, c.Resp(), c.Req())
	c.Resp().Header().Set("Location", fmt.Sprint(data.data["loginUrl"]))
	resp.Status = 301
	return
}

func usersMe(c gaecontext.JSONContext) (resp jsoncontext.Resp, err error) {
	data := getBaseData(c, c.Resp(), c.Req())
	if data.user != nil {
		resp.Body = data.user
	} else {
		resp.Status = 401
	}
	return
}

func getUser(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if authorized, authenticated := data.silentHasAuth(auth.Auth{
		AuthType: auth.Users,
	}); authorized {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		if user := appuser.GetUserFromDomain(data.context, data.domain, key); user == nil {
			w.WriteHeader(404)
		} else {
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, user)
		}
	} else if authenticated && mux.Vars(r)["id"] == data.user.Id.Encode() {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, data.user)
	} else {
		w.WriteHeader(401)
	}
	return
}

func getUsers(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Users,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, appuser.GetUsersByDomain(data.context, data.domain))
	}
	return
}

func updatePrivateUserSettings(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.isAuthed() {
		var neu appuser.User
		common.MustDecodeJSON(r.Body, &neu)
		data.user.CopySettingsFrom(&neu)
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, data.user.Save(data.context))
	}
	return
}

func updateUser(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var u appuser.User
	common.MustDecodeJSON(r.Body, &u)
	key, err := datastore.DecodeKey(mux.Vars(r)["id"])
	if err != nil {
		panic(err)
	}
	u.Id = key
	if len(u.Domains) == 1 && u.Domains[0].Id.Equal(data.domain) {
		if data.hasAuth(auth.Auth{
			AuthType: auth.Users,
			Write:    true,
		}) {
			oldUser := appuser.GetUserByKey(data.context, u.Id)
			wasDisabled := false
			isOwner, _ := data.silentIsOwner()
			for _, domain := range oldUser.Domains {
				if domain.Id.Equal(data.domain) {
					if !isOwner {
						u.Domains[0].Owner = domain.Owner
					}
					wasDisabled = domain.Disabled
				}
			}
			if u.Domains[0].Disabled && !wasDisabled {
				event.RemoveUserFromFutureEvents(data.context, data.user.Id, u.Id, data.domain)
			}
			common.SetContentType(w, "application/json; charset=UTF-8", false)
			common.MustEncodeJSON(w, u.SaveInDomain(data.context, u.Domains[0]))
		}
	} else {
		w.WriteHeader(403)
		fmt.Fprintln(w, "Unauthorized")
	}
	return
}

func createUser(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	var u appuser.User
	common.MustDecodeJSON(r.Body, &u)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Users,
		Write:    true,
	}) {
		common.SetContentType(w, "application/json; charset=UTF-8", false)
		common.MustEncodeJSON(w, u.AddToDomain(data.context, data.domain))
	}
	return
}

func deleteUser(c gaecontext.HTTPContext) (err error) {
	r := c.Req()
	w := c.Resp()
	data := getBaseData(c, w, r)
	if data.hasAuth(auth.Auth{
		AuthType: auth.Users,
		Write:    true,
	}) {
		key, err := datastore.DecodeKey(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}
		event.RemoveUserFromFutureEvents(data.context, data.user.Id, key, data.domain)
		appuser.DeleteUserFromDomain(data.context, key, data.domain)
		w.WriteHeader(204)
	}
	return
}
