window.UserDetailsView = Backbone.View.extend({

	template: _.template($('#user_details_underscore').html()),

	className: 'user-details',

	events: {
		"click .view-profile-link": "showProfile",
		"click #user_owner": "changeOwner",
		"click #user_allow_ics": "changeAllowICS",
		"click #user_disabled": "changeDisabled",
		"change #user_given_name": "changeGivenName",
		"change #user_family_name": "changeFamilyName",
		"change #user_mobile_phone": "changeMobilePhone",
		"change #user_information": "changeInformation",
		"click #clear_email_bounce": "clearEmailBounce",
		"click .available-user-property": "addUserProperty",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
		this.hide_profile_link = options.hide_profile_link;
		this.user_properties_for_domain = options.user_properties_for_domain;
		this.available_roles = options.available_roles;
		this.opener = options.opener;
		this.user_properties_for_user = new UserPropertiesForUser([], { url: '/users/' + this.model.get('id') + '/properties' });
		this.user_properties_for_user.bind("change", this.render);
		this.user_properties_for_user.bind("reset", this.render);
		this.user_properties_for_user.bind("add", this.render);
		this.user_properties_for_user.bind("remove", this.render);
		this.user_properties_for_user.fetch({ reset: true });
		this.roles = new Roles([], { url: '/users/' + this.model.id + '/roles' });
		this.auths = new Auths([], { url: '/users/' + this.model.id + '/auths' });
		if (app.hasAnyAuth({
		  auth_type: 'Roles',
		})) {
			this.roles.fetch({ reset: true });
		}
		if (app.hasAuth({
			auth_type: 'Roles',
		})) {
			this.auths.fetch({ reset: true });
		}

		this.locations = new Locations();
		this.locations.fetch({ reset: true });
		this.event_types = new EventTypes();
		this.event_types.fetch({ reset: true });
		this.event_kinds = new EventKinds();
		this.event_kinds.fetch({ reset: true });
		this.participant_types = new ParticipantTypes();
		this.participant_types.fetch({ reset: true });
		this.domain_roles = new Roles([], { url: '/roles' });
		this.domain_roles.fetch({ reset: true });
	},

	unprepared_modal: function(cb) {
		this.user_properties_for_domain = new UserPropertiesForDomain([], { url: '/user_properties' });
		this.available_roles = new Roles([], { url: '/roles' });
		var that = this;
		var after = new cbCounter(2, function() {
			that.modal(cb);
		});
		if (app.hasAuth({
			auth_type: 'Domain',
		})) {
			this.user_properties_for_domain.fetch({
				reset: true,
				success: function() {
					after.call();
				},
			});	
		} else {
			after.call();
		}
		this.available_roles.fetch({
		  reset: true,
			success: after.call,
		});
	},

	modal: function(cb) {
		var that = this;
		if (app.hasAuth({
			auth_type: 'Users',
			write: true,
		}) || app.hasAnyAuth({
			auth_type: 'Roles',
		  write: true,
		})) {
			$.modal.close();
			app.navigate('/users/' + that.model.get('id'));
			mymodal(that.render().el, {
				"{{.I "Save"}}": function() {
					that.save(cb);
				},
				'onCancel': cb,
			},
			{
        'min_height': '80%',
			});
		} else {
			$.modal.close();
			app.navigate('/users/' + that.model.get('id'));
			mymodal(that.render().el, { 'onClose': cb });
		}
	},

	clearEmailBounce: function(event) {
		event.preventDefault();
		this.model.set('email_bounce', '');
	},

	save: function(cb) {
		var after = new cbCounter(4, cb);
		if (app.hasAuth({
			auth_type: 'Roles',
			write: true,
		})) {
			this.auths.save(after.call);
		} else {
		  after.call();
		}
		if (app.hasAnyAuth({
      auth_type: 'Roles',
			write: true,
		})) {
			this.roles.save(after.call);
		} else {
			after.call();
		}
		if (app.hasAuth({
			auth_type: 'Users',
			write: true,
		})) {
			this.model.save(null, {
				success: after.call,
			});
			this.user_properties_for_user.save(after.call);
		} else {
			after.call();
			after.call();
		}
	},

	showProfile: function(event) {
		event.preventDefault();
		new ProfileView({
			is_modal: true,
			model: this.model,
		}).modal(function() {
		  app.navigate('/users');
		});
	},

	addUserProperty: function(event) {
		event.preventDefault();
		var daysValid = $(event.target).attr('data-user-property-days-valid');
		var attributes = {
			name: $(event.target).attr('data-user-property-name'),
			assigned_at: new Date(),
		};
		if (daysValid != 0) {
		  var today = new Date();
			attributes.valid_until = new Date(today.getFullYear(), today.getMonth(), today.getDate() + parseInt($(event.target).attr('data-user-property-days-valid')));
		}
		this.user_properties_for_user.add(new UserPropertyForUser(attributes));
	},

	changeMobilePhone: function(event) {
		this.model.set('mobile_phone', $(event.target).val(), { silent: true });
	},

	changeInformation: function(event) {
		this.model.get('domains')[0].information = $(event.target).val();
	},

	changeGivenName: function(event) {
		this.model.set('given_name', $(event.target).val(), { silent: true });
	},

	changeFamilyName: function(event) {
		this.model.set('family_name', $(event.target).val(), { silent: true });
	},

	changeOwner: function(event) {
		this.model.get('domains')[0].owner = !!$(event.target).attr('checked');
	},

	changeAllowICS: function(event) {
		this.model.get('domains')[0].allow_ics = !!$(event.target).attr('checked');
	},

	changeDisabled: function(event) {
		this.model.get('domains')[0].disabled = !!$(event.target).attr('checked');
	},

	render: function() {
		var that = this;
		var write_auth = app.hasAuth({ 
			auth_type: 'Users', 
			write: true,
		});
		this.$el.html(this.template({ 
			model: this.model,
			write_auth: write_auth,
			hide_profile_link: that.hide_profile_link,
		}));
		if (app.getDomain() != null && app.hasAnyAuth({auth_type: 'Roles'})) {
			new UserRolesView({
				el: this.$("#roles"),
				available_roles: this.available_roles,
				collection: this.roles,
			}).render();
		}
		if (app.getDomain() != null && app.hasAuth({auth_type: 'Roles'})) {
			new AuthsView({
				el: this.$("#auths"),
				collection: this.auths,
				locations: this.locations,
				event_types: this.event_types,
				event_kinds: this.event_kinds,
				participant_types: this.participant_types,
				roles: this.domain_roles,
			}).render();
		}
		if (window.app != null && app.getDomain() != null && app.getDomain().get('salary_mod')) {
		  if ((app.getDomain().get('salary_config').salary_user_properties || []).length > 0) {
			  new SalaryPropertiesView({
				  el: this.$('#salary_properties'),
					set_name: 'salary_user_properties',
					model: this.model,
					getter: function() {
						return that.model.get('domains')[0].salary_properties;			  
					},
					setter: (write_auth ? function(props) {
						that.model.get('domains')[0].salary_properties = props;
					} : null),
				}).render();
			}
			if (app.hasAnyAuth({ auth_type: 'Attest' })) {
				new UserSalariesView({
					el: this.$('#salaries'),
					model: this.model,
					userOpener: this.opener,
				}).render();
			}
		}
		this.user_properties_for_user.forEach(function(prop) {
			that.$("#user_property_list").append(new UserPropertyForUserView({ 
				model: prop,
				user_properties_for_domain: that.user_properties_for_domain,
				collection: that.user_properties_for_user,
			}).render().el);
		});
		this.user_properties_for_domain.forEach(function(prop) {
			that.$("#available_user_properties").append(new AvailableUserPropertyView({ model: prop }).render().el);
		});
		setTimeout(function() {
			var options = {
				askSecond: false,
				dayAbbreviations: {{.I "day_names_short"}},
				dayNames: {{.I "day_names"}},
				firstDOW: {{.I "firstDOW"}},
				labelDayOfMonth: '{{.I "labelDayOfMonth"}}',
				labelHour: '{{.I "labelHour"}}',
				labelMinute: '{{.I "labelMinute"}}',
				labelMonth: '{{.I "labelMonth"}}',
				labelTitle: '{{.I "labelTitle"}}',
				labelYear: '{{.I "labelYear"}}',
				monthAbbreviations: {{.I "month_names_short"}},
				monthNames: {{.I "month_names"}},
				format: '{{.I "any_date_format" }}',
			};
			that.user_properties_for_user.forEach(function(prop) {
			  if (prop.get('assigned_at') != null) {
					that.$('#user_property_' + prop.get('name').hash() + '_assigned_at').AnyTime_noPicker().AnyTime_picker(options);
				}
			  if (prop.get('valid_until') != null) {
					that.$('#user_property_' + prop.get('name').hash() + '_valid_until').AnyTime_noPicker().AnyTime_picker(options);
				}
			});
		}, 500);
		return this;
	},

});
