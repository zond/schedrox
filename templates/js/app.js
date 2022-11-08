
var AppRouter = Backbone.Router.extend({

	{{if .User}}
	routes: {
		"": "calendar",
		"login": "login",
		"calendar/:id": "showEvent",
		"calendar": "calendar",
		"users/:id/attest/:period": "attestUser",
		"users/:id": "showUser",
		"users": "users",
		"contacts/:id": "showContact",
		"contacts": "contacts",
		"events/properties": "eventProperties",
		"events/types/:id": "showEventType",
		"events/kinds/:id": "showEventKind",
		"events/types": "eventTypes",
		"events/reports/changed": "changedEvents",
		"events/reports/unpaid": "unpaidEvents",
		"events/reports/export": "exportEvents",
		"events/reports/contacts": "eventContacts",
		"events/reports/users": "eventUsers",
		"events/reports": "eventReports",
		"events/participants/:id": "showParticipantType",
		"events/participants": "participantTypes",
		"locations/:id": "showLocation",
		"settings/global": "globalSettings",
		"settings/domain": "domainSettings",
		"settings/roles/:id": "showRole",
		"settings/roles": "rolesSettings",
		"profile": "showMyProfile",
		"profiles/:id": "showProfile",
		"salaries/report": "reportSalaries",
		"salaries/configuration": "configureSalaries",
		"salaries/hours": "reportHours",
	},
	{{else}}
	routes: {
		"*var": "login",
	},
	{{end}}
	login: function() {
		this.display(LoginView, {el: $("#content")});
		$("#menu").hide();
		this.menu.set('active', null);
	},
	attestUser: function(id, period) {
		var that = this;
		var user = new User({ id: id });
		var fromTo = period.split(/-/);
		var from = new Date(parseInt(fromTo[0]));
		var to = new Date(parseInt(fromTo[1]));
		user.fetch({
			success: function() {
				that.display(UsersView, {
					el: $('#content'),
					attest_user: user,
					attest_from: from,
					attest_to: to,
				});
				that.menu.set('active', 'users');
			},
		});
	},
	showUser: function(id) {
		var that = this;
		var user = new User({ id: id });
		user.fetch({
			success: function() {
				that.display(UsersView, {
					el: $('#content'),
					show_user: user,
				});
				that.menu.set('active', 'users');
			},
		});
	},
	configureSalaries: function() {
	  this.display(SalaryConfigureView, {el: $('#content')});
		this.menu.set('active', 'salaries');
	},
	reportHours: function() {
	  this.display(ReportHoursView, {el: $('#content')});
		this.menu.set('active', 'salaries');
	},
	reportSalaries: function() {
	  this.display(SalaryReportView, {el: $('#content')});
		this.menu.set('active', 'salaries');
	},
	users: function() {
		this.display(UsersView, {el: $("#content")});
		this.menu.set('active', 'users');
	},
	showContact: function(id) {
		var that = this;
		var contact = new Contact({ id: id });
		contact.fetch({
			success: function() {
				that.display(ContactsView, { 
					el: $('#content'),
					show_contact: contact,
				});
				that.menu.set('active', 'contacts');
			},
		});
	},
	contacts: function() {
		this.display(ContactsView, {el: $("#content")});
		this.menu.set('active', 'contacts');
	},
	showEvent: function(id) {
		var that = this;
		var ev = new Event({ id: id });
		ev.fetch({
			success: function() {
				that.display(CalendarView, {
					el: $("#content"),
					show_event: ev,
				});
				that.menu.set('active', 'calendar');
			},
		});
	},
	calendar: function() {
		this.display(CalendarView, {el: $("#content")});
		this.menu.set('active', 'calendar');
	},
	globalSettings: function() {
		this.display(GlobalSettingsView, {el: $("#content")});
		this.menu.set('active', 'settings');
	},
	showRole: function(id) {
		var that = this;
		var role = new Role({ 
			id: id,
		});
		role.url = '/roles/' + id;
		role.fetch({
			success: function() {
				that.display(RolesSettingsView, {
					el: $('#content'),
					show_role: role,
				});
				that.menu.set('active', 'settings');
			},
		});
	},
	rolesSettings: function() {
		this.display(RolesSettingsView, {el: $("#content")});
		this.menu.set('active', 'settings');
	},
	domainSettings: function() {
		this.display(DomainSettingsView, {el: $("#content")});
		this.menu.set('active', 'settings');
	},
	showMyProfile: function(id) {
	  this.showProfile({{if .User}}"{{.User.Id.Encode}}"{{else}}null{{end}});
	},
	showProfile: function(id) {
		var that = this;
		if (id == that.user.get('id')) {
			that.display(ProfileView, {
				el: $("#content"),
				is_modal: false,
				model: that.user,
			});
			that.menu.set('active', 'settings');
		} else {
			var user = new User();
			user.url = '/profiles/' + id;
			user.fetch({
				success: function() {
					that.display(ProfileView, {
						el: $("#content"),
						is_modal: false,
						model: user,
					});
					that.menu.set('active', 'settings');
				},
			});
		}
	},
	eventProperties: function() {
		this.display(EventPropertiesView, {el: $("#content")});
		this.menu.set('active', 'events');
	},
	participantTypes: function() {
		this.display(ParticipantTypesView, {el: $("#content")});
		this.menu.set('active', 'events');
	},
	showLocation: function(id) {
    var that = this;
		var location = new Location({ id: id });
		location.fetch({
		  success: function() {
			  that.display(DomainSettingsView, {
				  el: $('#content'),
					show_location: location,
				}),
				that.menu.set('active', 'settings');
			},
		});
	},
	showParticipantType: function(id) {
		var that = this;
		var type = new ParticipantType({ id: id });
	  type.fetch({
			success: function() {
				that.display(ParticipantTypesView, { 
					el: $('#content'),
					show_type: type,
				}),
				that.menu.set('active', 'events');
			},
		});
	},
	showEventKind: function(id) {
		var that = this;
		var kind = new EventKind({ id: id });
		kind.fetch({
			success: function() {
				that.display(EventTypesView, { 
					el: $('#content'),
					show_kind: kind,
				}),
				that.menu.set('active', 'events');
			},
		});
	},
	showEventType: function(id) {
		var that = this;
		var type = new EventType({ id: id });
		type.fetch({
			success: function() {
				that.display(EventTypesView, { 
					el: $('#content'),
					show_type: type,
				}),
				that.menu.set('active', 'events');
			},
		});
	},
	changedEvents: function() {
		this.display(ChangedEventsView, {el: $("#content")});
		this.menu.set('active', 'events');
	},
	unpaidEvents: function() {
		this.display(UnpaidEventsView, {el: $("#content")});
		this.menu.set('active', 'events');
	},
	exportEvents: function() {
		this.display(ExportEventsView, {el: $("#content")});
		this.menu.set('active', 'events');
	},
	eventContacts: function() {
		this.display(EventContactsView, {el: $("#content")});
		this.menu.set('active', 'events');
	},
	eventUsers: function() {
		this.display(EventUsersView, {el: $("#content")});
		this.menu.set('active', 'events');
	},
	eventReports: function() {
		this.display(EventReportsView, {el: $("#content")});
		this.menu.set('active', 'events');
	},
	eventTypes: function() {
		this.display(EventTypesView, {el: $("#content")});
		this.menu.set('active', 'events');
	},

	user: {{if .User}}new User({{.User.ToJSON}}){{else}}null{{end}},

	email: {{if .Data "email"}}'{{.Data "email"}}'{{else}}null{{end}},

	_domain: {{if .User.FirstDomain}}new Domain({{.User.FirstDomain.ToJSON}}){{else}}null{{end}},

	auths: {{ .Auths }},

  authTypes: {{ .AuthTypes }},

	getDomain: function() {
		return this._domain;
	},

	domainAuths: function() {
		var dom = this.getDomain();
		if (dom == null) {
			return [];
		} else {
			return this.auths[dom.id];
		}
	},

	hasAnyAuth: function(match) {
		return isAuthorizedAny(this.domainAuths(), this.isClosed(), this.isOwner(), this.isAdmin(), match);
	},

	hasAuth: function(match) {
		return isAuthorized(this.domainAuths(), this.isClosed(), this.isOwner(), this.isAdmin(), match);
	},

	isAdmin: function() {
		if (this.user == null) {
			return false;
		}
		return this.user.get('admin');
	},

	isClosed: function() {
	  if (this.getDomain() == null) {
			return false;
		}
	  var redirect = this.getDomain().get('closed_and_redirected_to');
		return redirect != null && redirect != '';
	},

	isOwner: function() {
		if (this.getDomain() == null) {
			return false;
		}
		if (this.user == null) {
			return false;
		}
		var that = this;
		return _.any(this.user.get('domains'), function(domain) {
			return domain.id == that.getDomain().id && domain.owner;
		});
	},

	setDomain: function(d) {
		this._domain = d;
		this.trigger('domainchange');
	},

	views: {},

	loads: 0,

	displayLoader: function(ev, jqXHR, opts) {
		this.loads++;
		if (this.loads > 0) {
			$("#loader").show();
		}
	},

	hideLoader: function(ev, jqXHR, opts) {
		this.loads--;
		if (this.loads < 1) {
			$("#loader").hide();
		}
	},

	display: function(view, opts) {
		if (this.views[opts.el] != null) {
			this.views[opts.el].undelegateEvents();
			this.views[opts.el].unbind();
			if (this.views[opts.el].cleanup) {
				this.views[opts.el].cleanup();
			}
		}
		var instance = new view(opts).render();
		this.views[opts.el] = instance;
		instance.delegateEvents();
	},

	handleError: function(e) {
		console.log(e);
		{{if .IsDev}}
		alert(e);
		{{end}}
	},

	handleAjaxError: function(ev, jqXHR, ajaxSettings, thrownError) {
		this.hideLoader();
		if (jqXHR.status == 401) {
		  window.location.href = '/login';
		} else if (jqXHR.status == 412 && jqXHR.getResponseHeader('Location') != null) {
			window.location.href = jqXHR.getResponseHeader('Location');
		} else if (jqXHR.status == 403) {
			myalert("{{.I "You do not have the required permissions."}}");
		} else if (jqXHR.status != 417) {
			this.handleError([ev, jqXHR, ajaxSettings, thrownError]);
		}
	},

	menu: new Menu(),

	ajaxFilter: function(options, originalOptions, jqXHR) {
		if (options.headers == null) {
			options.headers = {};
		}
		if (options.headers['Authorization'] == null && this._domain != null) {
			options.headers['Authorization'] = 'realm=' + this._domain.id;
		}
	},

	initialize: function() {
		_.extend(this, Backbone.Event);
		_.bindAll(this, 'ajaxFilter');
		_.bindAll(this, 'displayLoader');
		_.bindAll(this, 'hideLoader');
		_.bindAll(this, 'handleAjaxError');
		$(document).ajaxSend(this.displayLoader);
		$(document).ajaxComplete(this.hideLoader);
		$(document).ajaxError(this.handleAjaxError);
		$.ajaxPrefilter(this.ajaxFilter);

		if (this.user != null) {
			this.menu.set('domains', JSON.parse(JSON.stringify(_.filter(this.user.get('domains'), function(dom) {
				return !dom.disabled;
			}))));
			if (this.menu.get('domains').length == 0 && !this.isAdmin()) {
				this.user = null;
			}
		}
		new MenuView({
			model: this.menu, 
			el: $("#menu"),
			app: this,
		}).render();

		$("#loader").hide();

		if (this.user != null) {
			var that = this;
			$.ajax('/alerts?at=' + new Date().toISOString(), {
				type: 'GET',
				dataType: 'json',
				success: function(data) {
					var hasAlerts = false;
					for (var dom in data) {
						if (data[dom].length > 0) {
							hasAlerts = true;
						}
					};
					if (hasAlerts) {
						mymodal(new DomainAlertsView({ 
							alerts: data,
							domains: that.user.get('domains'),
						}).render().el, null, {
							min_width: '80%',
							min_height: '80%',
						});
					}
				},
			});
			$('body').css('background-color', this.user.get('background_color'));
		}
	},
});


$(function() {

	window.app = new AppRouter();

	var oldOnError = window.onerror;
	window.onerror = function(e) {
		app.handleError(e);
		if (typeof(oldOnError) == 'function') {
			oldOnError(e);
		}
	};

	Backbone.history.start({ 
		pushState: true,
	});
	// This is here instead of inside AppRouter.initialize because it's a simple way 
	// to trigger a change in the MenuView while window.app is defined (and the menu gets properly rendered...
		app.menu.set('active', Backbone.history.fragment || 'calendar');

});							
