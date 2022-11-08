
window.clobberEventIds = false;

var AppRouter = Backbone.Router.extend({

	routes: {
		"": "myEvents",
		"events/mine": "myEvents",
		"events/open": "openEvents",
	},
	openEvents: function() {
		this.display(OpenEventsView, {el: $("#content")}, "events/open");
	},
	myEvents: function() {
		this.display(MyEventsView, {el: $("#content")}, "events/mine");
	},
	login: function() {
		this.display(LoginView, {el: $("#content")}, "login");
	},
	
	display: function(view, opts, active) {
		var that = this;
		this.reloader = function() {
			that.display(view, opts, active);
		};
		if (this.views[opts.el] != null) {
			this.views[opts.el].undelegateEvents();
			this.views[opts.el].unbind();
			if (this.views[opts.el].cleanup) {
				this.views[opts.el].cleanup();
			}
		}
		window.session.menu.set('active', active);
		var instance = new view(opts).render();
		this.views[opts.el] = instance;
		instance.delegateEvents();
	},
	views: {},
	reloader: function() {},

	displayLoader: function(ev, jqXHR, opts) {
		this.loads++;
		if (this.loads > 0) {
			$("#loader").show();
		}
	},

	hideLoader: function(ev, jqXHR, opts) {
		var that = this;
		that.loads--;
		if (that.loads < 1) {
			$("#loader").hide();
		}
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

	ajaxFilter: function(options, originalOptions, jqXHR) {
		if (options.headers == null) {
			options.headers = {};
		}
		if (options.headers['Authorization'] == null && window.session.menu.get('domain') != null) {
			options.headers['Authorization'] = 'realm=' + window.session.menu.get('domain').get('id');
		}
	},

	initialize: function() {
		// because user is already started loading
		this.loads = 1;
		_.extend(this, Backbone.Event);
		_.bindAll(this, 'ajaxFilter');
		_.bindAll(this, 'handleAjaxError');
		_.bindAll(this, 'displayLoader');
		_.bindAll(this, 'hideLoader');
		$(document).ajaxSend(this.displayLoader);
		$(document).ajaxError(this.handleAjaxError);
		$(document).ajaxComplete(this.hideLoader);
		$.ajaxPrefilter(this.ajaxFilter);
	},
});


$(function() {
	window.session = {};

	window.session.user = new User({}, { url: "/users/me" });
	window.session.user.fetch({
	  error: function(model, response, error) {
			if (response.status == 401) {
        window.location.href = '/login/redirect';
			}
		},
		success: function(model, response, error) {
			window.session.menu = new Menu({
				'active': Backbone.history.fragment || 'events/mine',
			  'domain': new Domain(model.get('domains')[0] || {}),
			  'active_filter': new CustomFilter({
					name: '{{.I "Filter" }}',
				}),
			});
			window.session.app = new AppRouter();
			window.session.participant_types = new ParticipantTypes();
			window.session.participant_types.fetch({ reset: true });
			window.session.menu.bind('change', function() {
				window.session.participant_types.fetch({ reset: true });
			});
			window.session.custom_filters = new CustomFilters();
			window.session.custom_filters.fetch({ reset: true });
		  Backbone.history.start({ 
				pushState: true,
			});
			$(document).on('click', '.navigate', function(ev) {
				ev.preventDefault();
				window.session.app.navigate($(ev.target).attr('href'), { trigger: true });
				window.session.menu.set('active', $(ev.target).attr('href').substring(1));
			});
			new TopNavView({ 
				el: $('#top-nav'),
				model: window.session.menu,
			}).render();
		},
	});


});							
