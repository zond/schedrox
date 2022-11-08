window.DomainSettingsView = Backbone.View.extend({

	template: _.template($('#domain_settings_underscore').html()),

	events: {
		"change #new_location": "addLocation",
		"change #new_user_property": "addUserProperty",
		"change #domain_auto_disable_after": "changeAutoDisable",
		"click #domain_auto_disable": "toggleAutoDisable",
		"click #domain_limited_ics": "toggleLimitedICS",
		"change #domain_from_address": "changeFromAddress",
		"change #domain_extra_confirmation_bcc": "changeExtraBCC",
		"change #domain_tz_location": "changeTZLocation",
		"change #domain_earliest_event": "changeEarliestEvent",
		"change #domain_latest_event": "changeLatestEvent",
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'refetch');
		this.show_location = options.show_location;
		this.locations = new Locations([], { url: '/locations' });
		this.locations.bind("change", this.render);
		this.locations.bind("reset", this.render);
		this.locations.bind("add", this.render);
		this.locations.bind("remove", this.render);
		this.user_properties = new UserPropertiesForDomain([], { url: '/user_properties' });
		this.user_properties.bind("reset", this.render);
		this.user_properties.bind("add", this.render);
		this.user_properties.bind("remove", this.render);
		this.refetch();
		app.on('domainchange', this.refetch);
	},

	changeTZLocation: function(ev) {
		this.model.set('tz_location', $(ev.target).select2('val'));
		this.model.save();
	},

	changeLatestEvent: function(ev) {
		this.model.set('latest_event', anyDayTimeConverter.parse($(ev.target).val()), { silent: true });
		this.model.save();
	},

	changeEarliestEvent: function(ev) {
		this.model.set('earliest_event', anyDayTimeConverter.parse($(ev.target).val()), { silent: true });
		this.model.save();
	},

	changeFromAddress: function(ev) {
		if ($(ev.target).val() == '' || $(ev.target).val().isEmail()) {
			this.model.set('from_address', $(ev.target).val());
			this.model.save();
		} else {
			myalert('{{.I "{0} is not a valid email address." }}'.format($(ev.target).val()));
		}
	},

	changeExtraBCC: function(ev) {
		if ($(ev.target).val() == '' || $(ev.target).val().isEmail()) {
			this.model.set('extra_confirmation_bcc', $(ev.target).val());
			this.model.save();
		} else {
			myalert('{{.I "{0} is not a valid email address." }}'.format($(ev.target).val()));
		}
	},

	toggleLimitedICS: function(ev) {
		this.model.set('limited_ics', !this.model.get('limited_ics'));
		this.model.save();
	},

	toggleAutoDisable: function(ev) {
		this.model.set('auto_disable', !this.model.get('auto_disable'));
		this.model.save();
	},

	changeAutoDisable: function(ev) {
		var neu = parseInt($(ev.target).val());
		if (neu > 0) {
			this.model.set('auto_disable_after', neu, { silent: true });
		} else {
			$(ev.target).val('1');
		}
		this.model.save();
	},

	refetch: function() {
		if (app.getDomain() != null) {
		    this.model = app.getDomain();
		    this.model.fetch({ reset: true });
			this.locations.fetch({ reset: true });
			this.user_properties.fetch({ reset: true });
		}
	},

	cleanup: function() {
		app.off('domainchange', this.refetch);
	},

	addUserProperty: function(event) {
		if (app.hasAuth({
			auth_type: 'Domain',
			write: true,
		})) {
			var that = this;
			var newProperty = new UserPropertyForDomain({ name: $("#new_user_property").val() });
			newProperty.save(null, {
				success: function() {
					that.user_properties.add(newProperty);
				},
			});
		}
	},
	addLocation: function(event) {
		if (app.hasAuth({
			auth_type: 'Domain',
			write: true,
		})) {
			var that = this;
			var newLocation = new Location({ name: $("#new_location").val() });
			newLocation.save(null, {
				success: function() {
					that.locations.add(newLocation);
				},
			});
		}
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({
			model: that.model,
			write_auth: app.hasAuth({
				auth_type: 'Domain',
				write: true,
			}),
		}));
		that.locations.forEach(function(location) {
			that.$("#location_list").append(new LocationView({ model: location }).render().el);
		});
		that.user_properties.forEach(function(prop) {
			that.$("#user_property_list").append(new UserPropertyForDomainView({ model: prop }).render().el);
		});
		that.$('#domain_tz_location').select2().select2('val', that.model.get('tz_location'));
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
				format: '{{.I "any_day_time_format" }}',
			};
			that.$('#domain_earliest_event').AnyTime_noPicker().AnyTime_picker(options);
			that.$('#domain_latest_event').AnyTime_noPicker().AnyTime_picker(options);
		}, 500);
		if (this.show_location != null) {
			new LocationDetailsView({ model: this.show_location }).modal(function() {
				app.navigate('/settings/domain');
			});
			this.show_location = null;
		}
		return that;
	},

});
