window.EventTypeDetailsView = Backbone.View.extend({

	template: _.template($('#event_type_details_underscore').html()),
	help_template: _.template($('#help_with_confirmation_email_underscore').html()),

	className: 'event-type-details',

	events: {
		"click .nav a": "changeTab",
		"change #event_type_default_minutes": "changeDefaultMinutes",
		"change #event_type_title_size": "changeTitleSize",
		"change #event_type_participants_format": "changeParticipantsFormat",
		"click #event_type_signal_colors_when_0_contacts": "toggleSignal0Contacts",
		"click #event_type_signal_colors_when_more_possible_contacts": "toggleSignalMoreContacts",
		"click #event_type_signal_colors_when_more_possible_users": "toggleSignalMoreUsers",
		"click #event_type_name_hidden_in_calendar": "toggleNameHidden",
		"click #event_type_unique": "toggleUnique",
		"click #event_type_display_users_in_calendar": "toggleDisplayUsers",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
		this.available_participant_types = new ParticipantTypes();
		this.available_participant_types.fetch({ reset: true });

		this.required_participant_types = new EventTypeRequiredParticipantTypes([], { available_types: this.available_participant_types, url: '/event_types/' + this.model.id + '/participant_types' });
		this.required_participant_types.fetch({ reset: true });

		_.bindAll(this.required_participant_types, 'sort');
		this.available_participant_types.bind('reset', this.required_participant_types.sort);
		this.mode = 'edit';
	},

	toggleDisplayUsers: function(event) {
		this.model.set('display_users_in_calendar', !this.model.get('display_users_in_calendar'));
	},

	toggleUnique: function(event) {
		this.model.set('unique', !this.model.get('unique'));
	},

	toggleSignal0Contacts: function(event) {
		this.model.set('signal_colors_when_0_contacts', !this.model.get('signal_colors_when_0_contacts'));
	},

	toggleSignalMoreUsers: function(event) {
		this.model.set('signal_colors_when_more_possible_users', !this.model.get('signal_colors_when_more_possible_users'));
	},

	toggleSignalMoreContacts: function(event) {
		this.model.set('signal_colors_when_more_possible_contacts', !this.model.get('signal_colors_when_more_possible_contacts'));
	},

	toggleNameHidden: function(event) {
		this.model.set('name_hidden_in_calendar', !this.model.get('name_hidden_in_calendar'));
	},

	changeParticipantsFormat: function(ev) {
		this.model.set('participants_format', $(ev.target).val(), { silent: true });
	},

	changeTitleSize: function(ev) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var new_size = parseInt($(ev.target).val());
			if (new_size > 10) {
				if (new_size < 1000) {
					this.model.set('title_size', new_size, { silent: true });
				} else {
					$(ev.target).val("1000");
				}
			} else {
				$(ev.target).val("10");
			}
		}
	},

	changeDefaultMinutes: function(ev) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var new_minutes = parseInt($(ev.target).val());
			if (new_minutes > -1) {
				this.model.set('default_minutes', new_minutes, { silent: true });
			} else {
				$(ev.target).val("0");
			}
		}
	},

	modal: function(cb) {
		var that = this;
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var errorHandler = null;
			var typeOpener = function() {
				$.modal.close();
				app.navigate('/events/types/' + that.model.get('id'));
				mymodal(that.render().el, {
					'{{.I "Save"}}': function() {
						that.model.save(null, {
							error: errorHandler,
							success: function() {
								that.model.errors = {};
								cb();
							},
						});
						that.required_participant_types.save();
					},
					'onCancel': cb,
				}, {
					min_height: '80%',
				});
			};
			errorHandler = function(model, xhr, options) {
				var error = JSON.parse(xhr.responseText);
				that.model.errors = {};
				that.model.errors[error.context] = error.message;
				typeOpener();
				that.delegateEvents();
			};
			typeOpener();
		} else {
			mymodal(that.render().el, { 'onClose': cb });
		}
	},

	changeTab: function(ev) {
		ev.preventDefault();
		this.mode = $(ev.target).attr('data-mode');
		this.render();
	},

	render: function() {
	  var that = this;
		var write_auth = app.hasAuth({
			auth_type: 'Event types',
			write: true,
		});
		that.$el.html(that.template({ 
			model: that.model,
			mode: that.mode,
			write_auth: write_auth,
		}));
		if (window.app != null && app.getDomain() != null && app.getDomain().get('salary_mod') && (app.getDomain().get('salary_config').salary_event_type_properties || []).length > 0) {
			new SalaryPropertiesView({
				el: that.$('#salary_properties'),
				set_name: 'salary_event_type_properties',
				model: that.model,
				getter: function() {
	        return that.model.get('salary_properties');			  
				},
				setter: (write_auth ? function(props) {
				  that.model.set('salary_properties', props, { silent: true });
				} : null),
			}).render();
		}
		new RequiredParticipantTypesView({
			el: that.$("#participant_types"),
			collection: that.required_participant_types,
			available_participant_types: that.available_participant_types,
		}).render();
		if (that.mode == 'edit') {
			that.$('#confirmation_container').append(new EditConfirmationEmailView({ model: that.model }).render().el);
		} else if (that.mode == 'preview') {
			that.$('#confirmation_container').append(new PreviewConfirmationEmailView({ model: that.model }).render().el);
		} else if (that.mode == 'help') {
			that.$('#confirmation_container').html(that.help_template({}));
		}
		return that;
	},

});
