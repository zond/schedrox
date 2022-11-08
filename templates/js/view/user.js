window.UserView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#user_underscore').html()),

	events: {
		"click .close": "removeUser",
		"click .open": "openUser",
		"click .attest": "openAttest",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
		this.user_properties_for_domain = options.user_properties_for_domain;
		this.available_roles = options.available_roles;
		this.checked = options.checked;
		this.attestFilter = options.attestFilter;
	},

	openAttest: function(event) {
	  event.preventDefault();
		var that = this;
		new UserAttestView({
		  model: that.model,
			from: that.attestFrom(),
			to: that.attestTo(),
		}).modal(function() {
		  app.navigate('/users');
		});
	},

	openUser: function(event) {
		event.preventDefault();
		var that = this;
		new UserDetailsView({ 
			model: that.model,
			user_properties_for_domain: that.user_properties_for_domain,
			available_roles: that.available_roles,
			opener: function() {
			  that.openUser(event);
			},
		}).modal(function() {
			app.navigate('/users');
		});
	},

	removeUser: function(event) {
		if (app.hasAuth({
			auth_type: 'Users',
			write: true,
		})) {
			var that = this;
			myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(this.model.get("email").htmlEscape()), function() {
				that.model.destroy();
			});
		}
	},

	attestFrom: function() {
		if (this.attestFilter != null) {
		  var fromTo = this.attestFilter.value.split(/-/);
			return Date.fromISOTime(parseInt(fromTo[0]) * 1000);
		}
		return null;
	},

	attestTo: function() {
		if (this.attestFilter != null) {
		  var fromTo = this.attestFilter.value.split(/-/);
			return Date.fromISOTime(parseInt(fromTo[1]) * 1000);
		}
		return null;
	},

	render: function() {
		var messages = [];
		var disabled = false;
		var info = this.model.get('domains')[0].information
		if (info.length > 100) {
		  info = info.substr(0, 97) + '[...]';
		}
		var thisDom = _.find(this.model.get('domains'), function(dom) {
			return dom.id == app.getDomain().get('id');
		});
		if (thisDom != null && thisDom.disabled) {
			disabled = true;
			messages.push('{{.I "{0} is disabled."}}'.format(this.model.name()));
		}
		this.$el.html(this.template({ 
			model: this.model,
			checked: this.checked,
			disabled: disabled,
			information: info,
			attestFilter: this.attestFilter,
			attestFrom: this.attestFrom(),
			attestTo: this.attestTo(),
		}));
		this.$('.user-information').attr('title', this.model.get('domains')[0].information);
		if (this.model.get('email_bounce') != null && this.model.get('email_bounce') != '') {
			messages.push('{{.I "{0} has a non operational email address." }}'.format(this.model.name()));
		}
		if (this.model.get('has_invalid_property')) {
			messages.push('{{.I "{0} has an invalid property." }}'.format(this.model.name()));
		}
		if (messages.length > 0) {
			this.$el.addClass('warning');
			this.$el.tooltip({
				placement: 'bottom',
				title: messages.join('<br/>'),
				html: true,
			});
		}
		return this;
	},

});
