window.AssignableParticipantTypeView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#assignable_participant_type_underscore').html()),

	events: {
		"change .assign-participant": "assignParticipant",
		"click .attend": "attendEvent",
		"click .delete-required-participant": "deleteExtraRequirement",
    "select2-close .assign-participant": "assignParticipant",
    "click .show-sms-list": "showSMSList",
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'getPotentialParticipants', 'format', 'matchSize', 'textForParticipant');
		this.event = options.event;
		this.busy_meter = options.busy_meter;
		this.busy_meter.bind("change", this.render);
		this.is_extra = options.is_extra;
		this.event_opener = options.event_opener;
		this.participants = options.participants;
		this.model.bind("change", this.render);
		this.extra_required_participant_types = options.extra_required_participant_types;
		this.potential_participants = null;
		this.current_term = null;
	},

	showSMSList: function(ev) {
		ev.preventDefault();
		this.$('.search-col .sms-list').removeClass('hidden');
	},

	deleteExtraRequirement: function(ev) {
		ev.preventDefault();
		var that = this;
		var deleted = false;
		this.extra_required_participant_types.forEach(function(type) {
			if (!deleted) {
				if (type.get('participant_type') == that.model.get('id')) {
					that.extra_required_participant_types.remove(type);
					deleted = true;
				}
			}
		});
	},

	createContact: function(name) {
		var that = this;
		if (name != null && name.length > 0) {
			var new_contact = new Contact({ name: name });
			mymodalClose();
			mymodal(new ContactDetailsView({ model: new_contact }).render().el, {
				"{{.I "Save"}}": function() {
					new_contact.save(null, {
						success: function() {
							that.addParticipant(new_contact.attributes);
							that.event_opener();  
						}
					});
				},
				"{{.I "Cancel"}}": function() {
					that.event_opener();
				},
			});
		}
	},

	attendEvent: function(ev) {
		if (app.hasAuth({ 
			auth_type: 'Attend', 
			location: this.event.get('location'), 
			event_kind: this.event.get('event_kind'), 
			event_type: this.event.get('event_type'), 
			participant_type: this.model.get('id'),
		})) {
			this.addParticipant({
				user: app.user.get('id'),
				multiple: 1,
				participant_type: this.model.get('id'),
				given_name: app.user.get('given_name'),
				family_name: app.user.get('family_name'),
				gravatar_hash: app.user.get('gravatar_hash'),
				email: app.user.get('email'),
				mobile_phone: app.user.get('mobile_phone'),
				auths: app.auths[app.getDomain().get('id')],
				owner: app.isOwner(),
				admin: app.isAdmin(),
			});
		}
	},

	addParticipant: function(data) {
		if (this.model.get('is_contact')) {
			data.contact = data.id;
			data.confirmations = 0;
			data.defaulted = true;
			data.multiple = 1;
			data.participant_type = this.model.id;
			delete(data.id);
		}  
		this.participants.add(new Participant(data));
	},

	assignParticipant: function(ev) {
		var data = $(ev.target).select2('data');
		if (data != null) {
			if (data.id != '') {
				this.addParticipant($(ev.target).select2('data'));
			} else {
				this.createContact(data.name);
			}
		}
	},

	textForParticipant: function(part) {
		var rval = []
		if (this.model.get('is_contact')) {
			if (part.name != null && part.name != '') {
				rval.push(part.name);
			}
			if (part.email != null && part.email != '') {
				rval.push(part.email);
			}
			if (part.mobile_phone != null && part.mobile_phone != '') {
				rval.push(part.mobile_phone);
			}
		} else {
			if (part.given_name != '' && part.family_name != '') {
				rval.push("{{.I "name_order" }}".format(part.given_name, part.family_name))
			} else if (part.given_name != '') {
				rval.push(part.given_name);
			} else if (part.family_name != '') {
				rval.push(part.family_name);
			}
			rval.push(part.email);
			if (part.mobile_phone != '') {
				rval.push(part.mobile_phone);
			}
		}
		return rval.join(', ');
	},

	matchSize: function(part) {
		if (part.match_size != null) {
			return part.match_size;
		}
		var text = this.textForParticipant(part);
		var res = 0;
		if (this.current_term != null && this.current_term != '') {
			var values = this.current_term.split(" ")
			for (var i = 0; i < values.length; i++) {
				var value = values[i];
				if (value.length > 1) {
					var matchStart = text.toLowerCase().indexOf(value.toLowerCase());
					if (matchStart != -1) {
						res += value.length;
					}
				}
			}
		}
		part.match_size = res;
		return res;
	},

	format: function(part) {
		var text = this.textForParticipant(part);
		if (this.current_term != null && this.current_term != '') {
			var values = this.current_term.split(" ")
			for (var i = 0; i < values.length; i++) {
				var value = values[i];
				if (value.length > 1) {
					var matchStart = text.toLowerCase().indexOf(value.toLowerCase());
					if (matchStart != -1) {
						var matchEnd = matchStart + value.length;
						text = text.substr(0, matchStart) + '<strong>' + text.substr(matchStart, value.length) + '</strong>' + text.substr(matchEnd);
					}
				}
			}
		}
		return '<img class="gravatar-small" src="' + gravatarImage(part.gravatar_hash, {s: 20}) + '">' + text;
	},

	getPotentialParticipants: function(options) {
		this.current_term = options.term;
		var that = this;
		if (that.model.get('is_contact')) {
			$.ajax('/contacts/search?q=' + encodeURIComponent(that.current_term), {
				type: 'GET',
				dataType: 'json',
				success: function(data) {
					options.callback({
						results: data,
					});
				},
			});
		} else {
			var callback = function(data) {
				options.callback({ 
					results: _.filter(data, function(part) {
						var rval = that.textForParticipant(part).toUpperCase().indexOf(options.term.toUpperCase()) >= 0;
						if (rval) {
							if (that.participants.any(function(p) {
								return p.get('user') == part.user;
							})) {
								rval = false;
							}
						}
						return rval;
					}), 
				});
			};
			if (this.potential_participants == null) {
				$.ajax('/potential_participants', {
					type: 'POST',
					dataType: 'json',
					data: JSON.stringify({
						location: that.event.get('location'),
						event_kind: that.event.get('event_kind'),
						event_type: that.event.get('event_type'),
						participant_type: that.model.get('id'),
						start: that.event.get('start'),
						end: that.event.get('end'),
						ignore_busy_event: that.event.get('id'),
					}),
					success: function(data) {
						that.potential_participants = data;
						var email_addresses = [];
						var mobile_numbers = [];
						_.each(data, function(p) {
							if (p.mobile_phone != null && p.mobile_phone != '') {
								mobile_numbers.push(p.mobile_phone);
							}
							if (p.email != null && p.email != '') {
								email_addresses.push(p.email);
							}
						});
						that.$('.search-col .email-list').attr('href', 'mailto:' + email_addresses.join(','));
						that.$('.search-col .sms-list').attr('value', mobile_numbers.join(','))
						that.$('.search-col .contact-links').removeClass('hidden');
						callback(data);
					},
				});
			} else {
				callback(that.potential_participants);
			}
		}
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({ 
			create_contact_auth: app.hasAuth({
				auth_type: 'Contacts',
				write: true,
			}) && app.hasAuth({
				auth_type: 'Participants',
				location: that.event.get('location'), 
				event_kind: that.event.get('event_kind'), 
				event_type: that.event.get('event_type'), 
				participant_type: that.model.get('id'),
				write: true,
			}),
			model: that.model,
			is_extra: that.is_extra,
			participant_write_auth: app.hasAuth({ 
				write: true, 
				auth_type: 'Participants', 
				location: that.event.get('location'), 
				event_kind: that.event.get('event_kind'), 
				event_type: that.event.get('event_type'), 
				participant_type: that.model.get('id'),
			}),
			event: that.event,
			busy_meter: that.busy_meter,
			in_event: that.participants.any(function(p) {
				return p.get('user') == app.user.get('id');
			}),
		}));
		var opts = {
			query: that.getPotentialParticipants,
			sortResults: function(res, cont, quer) {
				res.sort(function(b, a) {
					return that.matchSize(a) - that.matchSize(b);
				});
        return res;
			},
			formatResult: that.format,
			formatSelection: that.format,
		};
		if (that.model.get('is_contact')) {
			opts.minimumInputLength = 2;
			opts.createSearchChoice = function(term) {
				return {
					id: '',
					name: term,
				};
			};
		}
		that.$('.assign-participant').select2(opts);
		return that;
	},

});
