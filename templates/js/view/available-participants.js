window.AvailableParticipantsView = Backbone.View.extend({

  template: _.template($('#available_participants_underscore').html()),

  events: {
    "click .available-extra-required-participant-type": "addRequired",
  },

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.missing = options.missing;
		this.busy_meter = options.busy_meter;
		this.participants = options.participants;
		this.event = options.event;
		this.event_opener = options.event_opener;
		this.available = options.available;
		this.available_participant_types = options.available_participant_types;
		this.extra_required_participant_types = options.extra_required_participant_types;
		this.extra_required_by_type = options.extra_required_by_type;
	},

	addRequired: function(ev) {
	  ev.preventDefault();
		this.extra_required_participant_types.add(new RequiredParticipantType({
			participant_type: $(ev.target).attr('data-participant-type-id'),
			min: 1,
			max: 1,
		}));
	},

	render: function() {
		var that = this;
		this.$el.html(this.template({ 
			event: that.event,
		}));
		this.available_participant_types.forEach(function(type) {
			if (app.hasAuth({
				auth_type: 'Participants',
				location: that.event.get('location'),
				event_kind: that.event.get('event_kind'),
				event_type: that.event.get('event_type'),
				participant_type: type.get('id'),
			})) {
				that.$('#available_extra_required_participant_types').append(new AvailableParticipantTypeView({
					klass: 'available-extra-required-participant-type',
					model: type,
				}).render().el);
			}
		});
		_.each(that.missing, function(req) {
			that.extra_required_by_type[req.get('id')]--;
			that.$('#available_participant_list').append(new AssignableParticipantTypeView({
				model: req,
				busy_meter: that.busy_meter,
				extra_required_participant_types: that.extra_required_participant_types,
				is_extra: that.extra_required_by_type[req.get('id')] > -1,
				event_opener: that.event_opener,
				participants: that.participants,
				event: that.event,
			}).render().el);
		});
		for (var type_id in that.available) {
			var participant_type = that.available[type_id];
			that.extra_required_by_type[participant_type.get('id')]--;
			that.$('#available_participant_list').append(new AssignableParticipantTypeView({
				model: participant_type,
				busy_meter: that.busy_meter,
				extra_required_participant_types: that.extra_required_participant_types,
				is_extra: that.extra_required_by_type[participant_type.get('id')] > -1,
				event_opener: that.event_opener,
				participants: that.participants,
				event: that.event,
			}).render().el);
		}
		return this;
	},

});
