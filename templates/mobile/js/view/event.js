window.EventView = Backbone.View.extend({

  template: _.template($('#event_underscore').html()),

  events: {
		"click .collapse-toggle": "collapseToggle",
    "click .attend-event": "attendEvent",
	},

  initialize: function(options) {
		_.bindAll(this, 'render');
		this.participants = new Participants(null, { event: this.model });
		this.isOpen = options.isOpen;
		this.model.bind("change", this.render);
		this.participants.bind('reset', this.render);
		this.participants.bind('add', this.render);
		window.session.participant_types.bind('reset', this.render);
		this.loadedParticipants = false;
		this.collapsed = true;
  },

  collapseToggle: function(ev) {
	  ev.preventDefault();
		this.collapsed = !this.collapsed;
		if (!this.loadedParticipants) {
			this.loadedParticipants = true;
			this.participants.fetch({ reset: true });
		}
		this.render();
	},

	attendEvent: function(ev) {
		var that = this;
    if (window.confirm('{{.I "Are you sure you wish to attend to this event? You will not be able to revoke this decision." }}')) {
			that.participants.create({
				user: window.session.user.get('id'),
				multiple: 1,
				event_start: that.model.get('start').toISOString(),
				event_end: that.model.get('end').toISOString(),
				participant_type: $(ev.target).attr('data-participant-type'),
				participant_type_name: $(ev.target).attr('data-participant-type-name'),
				given_name: window.session.user.get('given_name'),
				family_name: window.session.user.get('family_name'),
				gravatar_hash: window.session.user.get('gravatar_hash'),
				always_create_exception: true,
				email: window.session.user.get('email'),
				mobile_phone: window.session.user.get('mobile_phone'),
			}, {
				success: function() {
					that.collection.fetch({ reset: true });
				},
			});
		}
	},

  render: function() {
    var participantTypeById = {};
		window.session.participant_types.each(function(typ) {
			participantTypeById[typ.get('id')] = typ;
		});
		var partsByType = {};
		var users = [];
		var contacts = [];
		this.participants.each(function(part) {
			var current = partsByType[part.get('participant_type')];
			if (current == null) {
				current = 0;
			}
			current++;
			partsByType[part.get('participant_type')] = current;
			if (participantTypeById[part.get('participant_type')].get('is_contact')) {
				contacts.push(part);
			} else {
				users.push(part);
			}
		});
		var allowedTypes = [];
		_.each(this.model.get('currently_required_participant_types'), function(typ) {
      var current = partsByType[typ.participant_type];
			if (current == null) {
				current = 0;
			}
			if (current < typ.max) {
				var fullType = participantTypeById[typ.participant_type];
				if (fullType != null && !fullType.get('is_contact')) {
					allowedTypes.push(fullType);
				}
			}
		});
    this.$el.html(this.template({
		  event: this.model,
			isOpen: this.isOpen,
		  users: users,
			contacts: contacts,
			allowedTypes: allowedTypes,
		  collapsed: this.collapsed,
		})); 
    return this;
  },

});
