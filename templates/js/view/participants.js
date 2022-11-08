window.ParticipantsView = Backbone.View.extend({

	template: _.template($('#participants_underscore').html()),

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.event = options.event;
		this.busy_meter = options.busy_meter;
		this.event_opener = options.event_opener;
		this.event_types = options.event_types;
		this.participants = options.participants;
		this.participants.bind("change", this.render);
		this.participants.bind("reset", this.render);
		this.participants.bind("add", this.render);
		this.participants.bind("remove", this.render);
		this.required_participant_types = options.required_participant_types;
		this.required_participant_types.bind("change", this.render);
		this.required_participant_types.bind("reset", this.render);
		this.required_participant_types.bind("add", this.render);
		this.required_participant_types.bind("remove", this.render);
		this.extra_required_participant_types = options.extra_required_participant_types;
		this.extra_required_participant_types.bind("change", this.render);
		this.extra_required_participant_types.bind("reset", this.render);
		this.extra_required_participant_types.bind("add", this.render);
		this.extra_required_participant_types.bind("remove", this.render);
		this.available_participant_types = options.available_participant_types;
		this.available_participant_types.bind("change", this.render);
		this.available_participant_types.bind("reset", this.render);
		this.available_participant_types.bind("add", this.render);
		this.available_participant_types.bind("remove", this.render);
		this.missing = [];
		this.extra = [];
		this.available = {};
		this.extra_required_by_type = {};
		this.user_participants = [];
		this.contact_participants = [];
		this.changeTimer = null;
	},

	recalculate: function() {
		var that = this;
		that.missing.length = 0;
		that.extra.length = 0;
		for (var key in that.extra_required_by_type) {
			delete(that.extra_required_by_type[key]);
		}
		for (var key in that.available) {
			delete(that.available[key]);
		}
		var current_by_type = {};
		var required_by_type = {};
		
		that.user_participants = [];
		that.contact_participants = [];

		that.participants.forEach(function(part) {
			if (part.get('user') == null) {
				that.contact_participants.push(part);
			} else {
				that.user_participants.push(part);
			}
			if (current_by_type[part.get('participant_type')] == null) {
				current_by_type[part.get('participant_type')] = 0;
			}
			current_by_type[part.get('participant_type')] += part.get('multiple');
		});

		var handledTypes = {};
		var iteratorCreator = function(extra_requirement) {
			return function(type) {
				var participant_type = that.available_participant_types.get(type.get('participant_type')); 
				if (participant_type != null) {
					var current = current_by_type[type.get('participant_type')] || 0;
					var min = type.get('min');
					var max = type.get('max');
					if (type.get('per_type') != null) {
						var calcNumber = (current_by_type[type.get('per_type')] || 0) - type.get('min');
						if (calcNumber < 0) {
							calcNumber = 0;
						}
						min = parseInt(calcNumber / type.get('per_num'));
						max = min;
					}
					if (required_by_type[type.get('participant_type')] == null) {
						required_by_type[type.get('participant_type')] = 0;
					}
					if (current < min) {
						required_by_type[type.get('participant_type')] += min;
					} else if (current > max) {
						required_by_type[type.get('participant_type')] += max;
					} else {
						required_by_type[type.get('participant_type')] += current;
						if (current < max) {
							that.available[type.get('participant_type')] = participant_type;
						}
					}
					if (extra_requirement) {
						if (that.extra_required_by_type[type.get('participant_type')] == null) {
							that.extra_required_by_type[type.get('participant_type')] = 0;
						}
						if (current < min) {
							that.extra_required_by_type[type.get('participant_type')] += min;
						} else if (current > max) {
							that.extra_required_by_type[type.get('participant_type')] += max;
						} else {
							that.extra_required_by_type[type.get('participant_type')] += current;
						}
					}
					handledTypes[participant_type.get('id')] = true;
				}
			}
		}

		that.required_participant_types.forEach(iteratorCreator(false));
		that.extra_required_participant_types.forEach(iteratorCreator(true));

		for (var type_id in required_by_type) {
			var required_n = required_by_type[type_id];
			var current_n = current_by_type[type_id] || 0;
			var participant_type = that.available_participant_types.get(type_id);
			if (current_n < required_n) {
				for (var i = current_n; i < required_n; i++) {
					that.missing.push(participant_type);
				}
			} else if (current_n > required_n) {
				that.extra.push({
					participant_type: participant_type,
					n: current_n - required_n,
				});
			}
		}
	},


	render: function() {
		var that = this;
		that.recalculate();
		that.$el.html(that.template({
			has_user_participants: that.user_participants.length > 0,
			has_contact_participants: that.contact_participants.length > 0,
			event: that.event,
		}));
		_.each(that.extra, function(d) {
			var msgStub;
			if (d.n == 1) {
				msgStub = '{{.I "There is {0} {1} too many!"}}';
			} else {
				msgStub = '{{.I "There are {0} {1} too many!"}}';
			}
			that.$('#participant_alerts').append('<div class="alert"><button type="button" class="close" data-dismiss="alert">&times;</button><strong>{{.I "Warning!"}}</strong> ' + msgStub.format(d.n, d.participant_type.get('name')) + '</div>');
		});
		var availableView = new AvailableParticipantsView({
			missing: that.missing,
			event_opener: that.event_opener,
			participants: that.participants,
			busy_meter: that.busy_meter,
			event: that.event,
			available: that.available,
			available_participant_types: that.available_participant_types,
			extra_required_participant_types: that.extra_required_participant_types,
			extra_required_by_type: that.extra_required_by_type,
			el: that.$("#available_participants_container"),
		});
		_.each(that.user_participants, function(participant) {
			that.$('#user_participant_list').append(new ParticipantView({ 
				event_opener: that.event_opener,
				model: participant,
				event_types: that.event_types,
				event: that.event,
				on_change: function() {
					if (that.changeTimer != null) {
						clearTimeout(that.changeTimer);
					}
					that.changeTimer = setTimeout(that.render, 1000);
				},
				collection: that.participants,
				available_participant_types: that.available_participant_types,
			}).render().el);
		});
		_.each(that.contact_participants, function(participant) {
			that.$('#contact_participant_list').append(new ParticipantView({ 
				event_opener: that.event_opener,
				model: participant,
				event_types: that.event_types,
				event: that.event,
				on_change: function() {
					if (that.changeTimer != null) {
						clearTimeout(that.changeTimer);
					}
					that.changeTimer = setTimeout(that.render, 1000);
				},
				collection: that.participants,
				available_participant_types: that.available_participant_types,
			}).render().el);
		});
		availableView.render();
		return that;
	},

});
