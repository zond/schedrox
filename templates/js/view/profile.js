window.ProfileView = Backbone.View.extend({

	template: _.template($('#profile_underscore').html()),

	events: {
		"click .view-user-link": "showUser",
		"click #user_mute_event_notifications": "toggleMuteEventNotifications",
		"change #user_calendar_days_back": "changeDaysBack",
		"change #user_title_size": "changeTitleSize",
		"change #user_calendar_width": "changeCalendarWidth",
		"change #user_calendar_height": "changeCalendarHeight",
		"click .available-event-type": "set_default_event_type",
		"click .available-event-kind": "set_default_event_kind",
		"click .available-location": "set_default_location",
		"click .available-participant-type": "set_default_participant_type",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
		this.is_modal = options.is_modal;
		this.custom_filters = new CustomFilters();
		this.custom_filters.bind("change", this.render);
		this.custom_filters.bind("reset", this.render);
		this.custom_filters.bind("add", this.render);
		this.custom_filters.bind("remove", this.render);
		this.custom_filters.fetch({ reset: true });
		this.event_types = new EventTypes();
		this.event_kinds = new EventKinds();
		this.locations = new Locations();
		this.participant_types = new ParticipantTypes();
		if (app.getDomain() != null) {
			var that = this;
			_.each([this.event_kinds, this.event_types, this.locations, this.participant_types], function(coll) {
				coll.bind("change", that.rerender);
				coll.bind("reset", that.rerender);
				coll.bind("add", that.rerender);
				coll.bind("remove", that.rerender);
				coll.fetch({ reset: true });
			});
		}
		this.data = {
			has_gravatar_profile: false,
			missing_gravatar_profile: false,
		};
		var that = this;
		this.model.withGravatarProfileData(function(data) {
			that.data = data;
			that.render();
		});
	},

	set_default_location: function(ev) {
		var that = this;
	  ev.preventDefault();
		if ($(ev.target).attr('data-location-id') != '') {
			this.model.set('default_location', $(ev.target).attr('data-location-id'));
		} else {
			this.model.set('default_location', null);
		}
		$.ajax('/settings', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify(that.model),
		});
	},

	set_default_event_kind: function(ev) {
		var that = this;
	  ev.preventDefault();
		if ($(ev.target).attr('data-event-kind-id') != '') {
			this.model.set('default_event_kind', $(ev.target).attr('data-event-kind-id'));
		} else {
			this.model.set('default_event_kind', null);
		}
		$.ajax('/settings', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify(that.model),
		});
	},

	set_default_participant_type: function(ev) {
		var that = this;
	  ev.preventDefault();
		if ($(ev.target).attr('data-participant-type-id') != '') {
			this.model.set('default_participant_type', $(ev.target).attr('data-participant-type-id'));
		} else {
			this.model.set('default_participant_type', null);
		}
		$.ajax('/settings', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify(that.model),
		});
	},

	set_default_event_type: function(ev) {
		var that = this;
	  ev.preventDefault();
		if ($(ev.target).attr('data-event-type-id') != '') {
			this.model.set('default_event_type', $(ev.target).attr('data-event-type-id'));
		} else {
			this.model.set('default_event_type', null);
		}
		$.ajax('/settings', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify(that.model),
		});
	},

	modal: function(cb) {
		var that = this;
		$.modal.close();
		app.navigate('/profiles/' + that.model.get('id'));
		mymodal(this.render().el, {
			'onClose': cb,
		});
	},

	toggleMuteEventNotifications: function(event) {
		var that = this;
		event.preventDefault();
		this.model.set('mute_event_notifications', !this.model.get('mute_event_notifications'));
		$.ajax('/settings', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify(that.model),
		});
	},

	changeCalendarHeight: function(event) {
		var that = this;
		var new_calendar_height = parseInt($(event.target).val());
		if (new_calendar_height > 10 && new_calendar_height < 1000) {
			this.model.set('calendar_height', new_calendar_height, { silent: true });
			$.ajax('/settings', {
				type: 'POST',
				dataType: 'json',
				data: JSON.stringify(that.model),
			});
		} else {
			$(event.target).val("100");
		}
	},

	changeCalendarWidth: function(event) {
		var that = this;
		var new_calendar_width = parseInt($(event.target).val());
		if (new_calendar_width > 10 && new_calendar_width < 1000) {
			this.model.set('calendar_width', new_calendar_width, { silent: true });
			$.ajax('/settings', {
				type: 'POST',
				dataType: 'json',
				data: JSON.stringify(that.model),
			});
		} else {
			$(event.target).val("100");
		}
	},

	changeTitleSize: function(event) {
		var that = this;
		var new_title_size = parseInt($(event.target).val());
		if (new_title_size > 10 && new_title_size < 500) {
			this.model.set('title_size', new_title_size, { silent: true });
			$.ajax('/settings', {
				type: 'POST',
				dataType: 'json',
				data: JSON.stringify(that.model),
			});
		} else {
			$(event.target).val("100");
		}
	},

	changeDaysBack: function(event) {
		var that = this;
		var new_days_back = parseInt($(event.target).val());
		if (new_days_back > -1 && new_days_back < 7) {
			this.model.set('calendar_days_back', new_days_back, { silent: true });
			$.ajax('/settings', {
				type: 'POST',
				dataType: 'json',
				data: JSON.stringify(that.model),
			});
		} else {
			$(event.target).val("0");
		}
	},

	showUser: function(event) {
		event.preventDefault();
		if (this.is_modal) {
			var that = this;
			new UserDetailsView({ 
				model: that.model, 
				hide_profile_link: !app.hasAuth({ auth_type: 'Users' }), 
				opener: function() {
				  that.showUser(event);
				},
			}).unprepared_modal(function() {
			  app.navigate('/users', { trigger: true });
			});
		} else {
			app.navigate('/users/' + this.model.get('id'), { trigger: true });
		}
	},

	render: function() {
		var that = this;
		var background_color = this.model.get('background_color');
		if (background_color == '') {
			background_color = '#ffffff';
		}
		var zLocation = new Location({ name: '{{.I "None" }}' });
		var location = this.locations.get(this.model.get('default_location'));
		if (location == null) {
		  location = zLocation;
		}
		var zKind = new EventKind({ name: '{{.I "None" }}' });
		var event_kind = this.event_kinds.get(this.model.get('default_event_kind'));
		if (event_kind == null) {
		  event_kind = zKind;
		}		
		var zType = new EventType({ name: '{{.I "None" }}' });
		var event_type = this.event_types.get(this.model.get('default_event_type'));
		if (event_type == null) {
		  event_type = zType;
		}
		var zPType = new ParticipantType({ name: '{{.I "None" }}' });
		var participant_type = this.participant_types.get(this.model.get('default_participant_type'));
		if (participant_type == null) {
		  participant_type = zPType;
		}
		
		this.$el.html(this.template({
			background_color: background_color,
			modal: this.modal,
			gravatar_data: this.data,
			model: this.model,
			location: location,
			event_kind: event_kind,
			event_type: event_type,
			participant_type: participant_type,
		}));
		that.$("#available_locations").append(new AvailableLocationView({ model: zLocation }).render().el);
		this.locations.each(function(location) {
			that.$("#available_locations").append(new AvailableLocationView({ model: location }).render().el);
		});
		that.$("#available_event_kinds").append(new AvailableEventKindView({ model: zKind }).render().el);
		this.event_kinds.each(function(event_kind) {
			that.$("#available_event_kinds").append(new AvailableEventKindView({ model: event_kind }).render().el);
		});
		that.$("#available_event_types").append(new AvailableEventTypeView({ model: zType }).render().el);
		this.event_types.each(function(event_type) {
			that.$("#available_event_types").append(new AvailableEventTypeView({ model: event_type }).render().el);
		});
		that.$("#available_participant_types").append(new AvailableParticipantTypeView({ model: zPType }).render().el);
		this.participant_types.each(function(participant_type) {
			that.$("#available_participant_types").append(new AvailableParticipantTypeView({ model: participant_type }).render().el);
		});
		if (app.user.get('id') == this.model.get('id')) {
			this.$('.colorpicker').colorpicker().on('hide', function(ev) {
				that.model.set('background_color', ev.color.toHex());
				$.ajax('/settings', {
					type: 'POST',
					dataType: 'json',
					data: JSON.stringify(that.model),
				});
				$('body').css('background-color', ev.color.toHex());
			});
		}
		this.custom_filters.forEach(function(filter) {
			this.$('#custom_filters_list').append(new CustomFilterRowView({ model: filter }).render().el);
		});
		return this;
	}

});
