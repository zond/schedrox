window.ParticipantView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#participant_underscore').html()),

	events: {
		"click .close": "removeParticipant",
		"click .open": "openParticipant",
		"change .participant-multiple": "changeMultiple",
		"keyup .participant-multiple": "changeMultiple",
		"input .participant-multiple": "changeMultiple",
		"click .send-confirmation": "sendConfirmation",
		"click .participant-paid": "switchPaid",
		"click .participant-defaulted": "switchDefaulted",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.available_participant_types = options.available_participant_types;
		this.event = options.event;
		this.event_types = options.event_types;
		this.model.bind("change", this.render);
		this.event_opener = options.event_opener;
		this.on_change = options.on_change;
	},

	switchDefaulted: function(ev) {
		ev.preventDefault();
		this.model.set('defaulted', !this.model.get('defaulted'));
		if (this.model.switched_defaulted) {
			this.model.switched_defaulted = false;
		} else {
      this.model.switched_defaulted = true;
		}
	},

	switchPaid: function(ev) {
		ev.preventDefault();
		this.model.set('paid', !this.model.get('paid'));
		this.model.modified = true;
	},

	sendConfirmation: function(ev) {
		var that = this;
		ev.preventDefault();
		if (that.event.get('event_type') != null) {
			$.ajax('/send_confirmation', {
				type: 'POST',
				dataType: 'json',
				data: JSON.stringify({
					participant: that.model,
					event: that.event,
				}),
				success: function(data) {
					if (that.model.isNew()) {
						that.model.set('confirmations', that.model.get('confirmations') + 1);
					} else {
						that.model.set('confirmations', that.model.get('confirmations') + 1);
						that.model.save();
					}
				},
			});
		}
	},

	openParticipant: function(ev) {
		var that = this;
		ev.preventDefault();
		if (this.model.get('user') == null) {
			var contact = new Contact({ id: that.model.get('contact') });
			contact.fetch({
				success: function() {
					new ContactDetailsView({ 
						model: contact,
					}).modal(that.event_opener);
				},
			});
		} else {
			if (app.hasAuth({
				auth_type: 'Users',
			})) {
				var user = new User({ id: that.model.get('user') });
				user.fetch({
					success: function() {
						new UserDetailsView({ 
							model: user, 
							hide_profile_link: true, 
						}).unprepared_modal(that.event_opener);
					},
				});
			} else {
				var user = new User({ 
					id: that.model.get('user'),
					gravatar_hash: that.model.get('gravatar_hash'),
					email: that.model.get('email'),
					given_name: that.model.get('given_name'),
					family_name: that.model.get('family_name'),
				});
				new ProfileView({
					modal: true,
					model: user,
				}).modal(that.event_opener);
			}
		}
	},

	changeMultiple: function(ev) {
		if (this.hasWriteAuth()) {
			var value = parseInt($(ev.target).val());
			if (value > 0) {
				this.model.set('multiple', value, { silent: true });
				this.model.modified = true;
				this.on_change();
			} else {
				$(ev.target).val('1');
			}
		}
	},

	hasWriteAuth: function() {
		return app.hasAuth({
			auth_type: 'Participants',
			location: this.event.get('location'),
			event_kind: this.event.get('event_kind'),
			event_type: this.event.get('event_type'),
			participant_type: this.model.get('participant_type'),
			write: true,
		});
	},

	removeParticipant: function(ev) {
		if (this.hasWriteAuth()) {
			this.collection.remove(this.model);
		}
	},

	render: function() {
		var type = this.available_participant_types.get(this.model.get('participant_type'));
		if (type == null) {
		  type = new ParticipantType();
		}
		var has_open_auth = true;
		if (this.model.get('user') == null) {
			has_open_auth = app.hasAuth({
				auth_type: 'Contacts',
			});
		}
		if (this.model.get('missing')) {
			has_open_auth = false;
		}
		var event_type_model = this.event_types.get(this.event.get('event_type'));
		var event_type_has_confirmation_email = (event_type_model.get('confirmation_email_body_template') || '').trim() != '';
		this.$el.html(this.template({ 
			model: this.model,
			event: this.event,
			event_type: this.event.get('event_type'),
			event_type_has_confirmation_email: event_type_has_confirmation_email,
			participant_type: type,
			open_auth: has_open_auth,
			write_auth: this.hasWriteAuth(),
		}));
		var messages = [];    
		if (!type.get('is_contact') && !isAuthorized(this.model.get('auths'), app.isClosed(), app.isOwner(), app.isAdmin(), {
			auth_type: 'Attend',
			location: this.event.get('location'),
			event_kind: this.event.get('event_kind'),
			event_type: this.event.get('event_type'),
			participant_type: this.model.get('participant_type'),
		})) {
			messages.push('{{.I "{0} is not authorized to be {1} in {2}." }}'.format(
				this.model.name(),
				type.get('name'),
				this.event.describe(this.event_types)
			));
		}
		if (this.model.get('email_bounce') != null && this.model.get('email_bounce') != '') {
			messages.push('{{.I "{0} has a non operational email address." }}'.format(this.model.name()));
		}
		if (this.model.get('missing')) {
			messages.push('{{.I "{0} no longer exists in the system." }}'.format(this.model.name()));
		}
		if (messages.length > 0) {
			this.$el.addClass('warning');
			this.$el.tooltip({
				placement: 'top',
				title: messages.join("<br/>"),
				html: true,
			});
		}
		return this;
	},

});
