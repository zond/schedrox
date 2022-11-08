package web

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"github.com/zond/schedrox/appuser"
	"github.com/zond/schedrox/auth"
	"github.com/zond/schedrox/common"
	"github.com/zond/schedrox/domain"
	"github.com/zond/schedrox/translation"
	"net/http"
	"strings"
	textTemplate "text/template"

	"github.com/gorilla/sessions"
	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

var mobileHtmlTemplates = htmlTemplate.Must(htmlTemplate.New("mobileHtmlTemplates").ParseGlob("templates/mobile/html/*.html"))
var htmlTemplates = htmlTemplate.Must(htmlTemplate.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var _Templates = textTemplate.Must(textTemplate.New("_Templates").ParseGlob("templates/_/*.html"))
var jsModelTemplates = textTemplate.Must(textTemplate.New("jsModelTemplates").ParseGlob("templates/js/model/*.js"))
var jsCollectionTemplates = textTemplate.Must(textTemplate.New("jsCollectionTemplates").ParseGlob("templates/js/collection/*.js"))
var jsViewTemplates = textTemplate.Must(textTemplate.New("jsViewTemplates").ParseGlob("templates/js/view/*.js"))
var jsTemplates = textTemplate.Must(textTemplate.New("jsTemplates").ParseGlob("templates/js/*.js"))
var cssTemplates = textTemplate.Must(textTemplate.New("cssTemplates").ParseGlob("templates/css/*.css"))

var mobileCssTemplates = textTemplate.Must(textTemplate.New("mobileCssTemplates").ParseGlob("templates/mobile/css/*.css"))
var mobileJsTemplates = textTemplate.Must(textTemplate.New("mobileJsTemplates").ParseGlob("templates/mobile/js/*.js"))
var mobileJsViewTemplates = textTemplate.Must(textTemplate.New("mobileJsViewTemplates").ParseGlob("templates/mobile/js/view/*.js"))
var mobile_Templates = textTemplate.Must(textTemplate.New("mobile_Templates").ParseGlob("templates/mobile/_/*.html"))

var sessionStore = sessions.NewCookieStore([]byte("da39a3ee5e6b4b0d3255bfef95601890afd80709"))

type baseData struct {
	translations map[string]string
	response     http.ResponseWriter
	request      *http.Request
	data         map[string]interface{}
	user         *appuser.User
	authorizer   auth.Authorizer
	email        *string
	context      gaecontext.HTTPContext
	domain       *datastore.Key
	salaryMod    bool
	isMobile     bool
}

func getBaseData(c gaecontext.HTTPContext, w http.ResponseWriter, r *http.Request) (result baseData) {
	translations, err := translation.GetTranslations(r)
	if err != nil {
		panic(err)
	}
	u, email := appuser.GetCurrentUser(c)
	if email != nil {
		log.Infof(c, "Request by %#v", *email)
	}
	result = baseData{
		translations: translations,
		request:      r,
		response:     w,
		data:         make(map[string]interface{}),
		user:         u,
		context:      c,
		isMobile:     mobileUserAgentRegexp.MatchString(r.Header.Get("User-Agent")),
	}
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.Split(authHeader, ",")
		for _, part := range parts {
			keyval := strings.Split(part, "=")
			if len(keyval) == 2 {
				if keyval[0] == "realm" {
					result.domain, err = datastore.DecodeKey(keyval[1])
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
	if result.domain == nil {
		if domid := r.URL.Query().Get("domain"); domid != "" {
			result.domain, err = datastore.DecodeKey(domid)
			if err != nil {
				panic(err)
			}
		}
	}
	if email != nil {
		result.data["email"] = *email
	}
	if u != nil {
		result.authorizer = u.GetAuthorizer(result.context, result.domain)
	}
	loginUrl, err := user.LoginURL(c, common.GetHostURL(r))
	if err != nil {
		panic(err)
	}
	result.data["loginUrl"] = loginUrl
	logoutUrl, err := user.LogoutURL(c, common.GetHostURL(r))
	if err != nil {
		panic(err)
	}
	result.data["logoutUrl"] = logoutUrl
	if dom := domain.GetDomain(result.context, result.domain); dom != nil {
		result.salaryMod = dom.SalaryMod
	}
	return
}

func (self baseData) silentHasAnyAuth(a auth.Auth) (authorized, authenticated bool) {
	if self.user != nil {
		if self.authorizer.HasAnyAuth(a) {
			return true, true
		}
		return false, true
	}
	return false, false
}

func (self baseData) silentHasAuth(a auth.Auth) (authorized, authenticated bool) {
	if self.user != nil {
		if self.authorizer.HasAuth(a) {
			return true, true
		}
		return false, true
	}
	return false, false
}

func (self baseData) CustomerLogo() string {
	switch appengine.AppID(self.context) {
	case "kc-sched":
		return "http://www.klattercentret.se/wp-content/themes/brandson-wordpress-theme_3.0/images/logo.png"
	}
	return ""
}

func (self baseData) CustomerMessage() string {
	return ""
}

func (self baseData) isAuthed() (result bool) {
	if self.user != nil {
		return true
	}
	self.response.WriteHeader(401)
	fmt.Fprintln(self.response, "Unauthenticated")
	return false
}

func (self baseData) hasAnyAuth(a auth.Auth) (result bool) {
	authorized, authenticated := self.silentHasAnyAuth(a)
	if authenticated {
		if authorized {
			return true
		}
		self.response.WriteHeader(403)
		fmt.Fprintln(self.response, "Unauthorized")
		return false
	}
	self.response.WriteHeader(401)
	fmt.Fprintln(self.response, "Unauthenticated")
	return false
}

func (self baseData) hasAuth(a auth.Auth) (result bool) {
	authorized, authenticated := self.silentHasAuth(a)
	if authenticated {
		if authorized {
			return true
		}
		self.response.WriteHeader(403)
		fmt.Fprintln(self.response, "Unauthorized")
		return false
	}
	self.response.WriteHeader(401)
	fmt.Fprintln(self.response, "Unauthenticated")
	return false
}

func (self baseData) IsMobile() bool {
	return self.isMobile
}

func (self baseData) IsDev() bool {
	return appengine.IsDevAppServer()
}

func (self baseData) Auths() string {
	if self.user == nil {
		return "{}"
	}
	buffer := new(bytes.Buffer)
	common.MustEncodeJSON(buffer, self.user.GetAuths(self.context))
	return string(buffer.Bytes())
}

func (self baseData) AuthTypes() string {
	translatedAuthTypes := make(map[string]auth.AuthType)
	for name, authType := range auth.AuthTypes(true) {
		var cpy auth.AuthType
		cpy = authType
		var err error
		cpy.Translation, err = self.I(cpy.Name)
		if err != nil {
			panic(err)
		}
		translatedAuthTypes[name] = cpy
	}
	buffer := new(bytes.Buffer)
	common.MustEncodeJSON(buffer, translatedAuthTypes)
	return string(buffer.Bytes())
}

func (self baseData) Version() string {
	return appengine.VersionID(self.context)
}

func (self baseData) User() *appuser.User {
	return self.user
}

func (self baseData) Email() *string {
	return self.email
}

func (self baseData) Data(s string) interface{} {
	return self.data[s]
}
func (self baseData) I(phrase string, args ...string) (result string, err error) {
	pattern, ok := self.translations[phrase]
	if !ok {
		err = fmt.Errorf("Unable to find translation for %v", phrase)
		result = err.Error()
		return
	}
	if len(args) > 0 {
		result = fmt.Sprintf(pattern, args)
	} else {
		result = pattern
	}
	return
}

func (self baseData) silentIsAdmin() (authenticated, authorized bool) {
	if self.user != nil {
		if user.IsAdmin(self.context) {
			return true, true
		}
		return true, false
	}
	return false, false
}

func (self baseData) isAdmin() bool {
	authenticated, authorized := self.silentIsAdmin()
	if !authenticated {
		self.response.WriteHeader(401)
		fmt.Fprintln(self.response, "Unauthenticated")
		return false
	}
	if !authorized {
		self.response.WriteHeader(403)
		fmt.Fprintln(self.response, "Unauthorized")
		return false
	}
	return true
}

func (self baseData) silentIsOwner() (result, authenticated bool) {
	if self.user != nil {
		if self.user.Admin {
			return true, true
		}
		if self.domain != nil {
			for _, domain := range self.user.Domains {
				if domain.Id.Equal(self.domain) && domain.Owner {
					return true, true
				}
			}
		}
		return false, true
	}
	return false, false
}

func (self baseData) isOwner() bool {
	result, authenticated := self.silentIsOwner()
	if result {
		return true
	}
	if authenticated {
		self.response.WriteHeader(403)
		fmt.Fprintln(self.response, "Unauthorized")
		return false
	}
	self.response.WriteHeader(401)
	fmt.Fprintln(self.response, "Unauthenticated")
	return false
}

func (self baseData) hasDomain(domainId *datastore.Key) bool {
	if self.user != nil {
		if self.user.Admin {
			return true
		}
		for _, domain := range self.user.Domains {
			if domain.Id.Equal(domainId) {
				return true
			}
		}
		self.response.WriteHeader(403)
		fmt.Fprintln(self.response, "Forbidden")
		return false
	}
	self.response.WriteHeader(401)
	fmt.Fprintln(self.response, "Unauthenticated")
	return false
}

func render_Templates(templates *textTemplate.Template, data baseData) {
	fmt.Fprintln(data.response, "(function() {")
	fmt.Fprintln(data.response, "  var n;")
	var buf *bytes.Buffer
	var rendered string
	for _, templ := range templates.Templates() {
		fmt.Fprintf(data.response, "  n = $('<script type=\"text/template\" id=\"%v_underscore\"></script>');\n", strings.Split(templ.Name(), ".")[0])
		fmt.Fprintf(data.response, "  n.text('")
		buf = new(bytes.Buffer)
		if err := templ.Execute(buf, data); err != nil {
			panic(err)
		}
		rendered = string(buf.Bytes())
		rendered = strings.Replace(rendered, "\\", "\\\\", -1)
		rendered = strings.Replace(rendered, "'", "\\'", -1)
		rendered = strings.Replace(rendered, "\n", "\\n", -1)
		fmt.Fprint(data.response, rendered)
		fmt.Fprintln(data.response, "');")
		fmt.Fprintln(data.response, "  $('head').append(n);")
	}
	fmt.Fprintln(data.response, "})();")
}

func renderHtml(w http.ResponseWriter, r *http.Request, templates *htmlTemplate.Template, template string, data interface{}) {
	common.SetContentType(w, "text/html; charset=UTF-8", false)
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		panic(fmt.Errorf("While rendering HTML: %v", err))
	}
}

func renderText(w http.ResponseWriter, r *http.Request, templates *textTemplate.Template, template string, data interface{}) {
	if err := templates.ExecuteTemplate(w, template, data); err != nil {
		panic(fmt.Errorf("While rendering text: %v", err))
	}
}

func renderJs(w http.ResponseWriter, r *http.Request, templates *textTemplate.Template, template string, data interface{}) {
	renderText(w, r, templates, template, data)
}
