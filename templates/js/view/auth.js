window.AuthView = Backbone.View.extend({

  tagName: 'tr',

  template: _.template($('#auth_underscore').html()),

  events: {
    "click .close": "removeAuth",
    "click .write-switch": "switchWrite",
    "click .available-event-kind": "setEventKind",
    "click .available-location": "setLocation",
    "click .available-event-type": "setEventType",
    "click .available-participant-type": "setParticipantType",
    "click .available-role": "setRole",
  },

  initialize: function(options) {
    this.locations = options.locations;
    this.event_kinds = options.event_kinds;
    this.event_types = options.event_types;
    this.participant_types = options.participant_types;
		this.roles = options.roles;
    _.bindAll(this, 'render');
    this.model.bind("change", this.render);
  },

  removeAuth: function(event) {
    if (app.hasAuth({
      auth_type: 'Roles',
      write: true,
    })) {
      this.collection.remove(this.model);
    }
  },

  setRole: function(event) {
    if (app.hasAuth({
      auth_type: 'Roles',
      write: true,
    })) {
      var role_name = $(event.target).attr('data-role-name');
			this.model.set('role', role_name);
      this.model.modified = true;
    }
  },

  setLocation: function(event) {
    if (app.hasAuth({
      auth_type: 'Roles',
      write: true,
    })) {
      var location_id = $(event.target).attr('data-location-id');
      if (location_id == "") {
	this.model.set('location', null);
      } else {
	this.model.set('location', location_id);
      }
      this.model.modified = true;
    }
  },

  setEventKind: function(event) {
    if (app.hasAuth({
      auth_type: 'Roles',
      write: true,
    })) {
      var kind_id = $(event.target).attr('data-event-kind-id');
      if (kind_id == "") {
	this.model.set('event_kind', null);
      } else {
	this.model.set('event_kind', kind_id);
      }
      this.model.modified = true;
    }
  },

  setEventType: function(event) {
    if (app.hasAuth({
      auth_type: 'Roles',
      write: true,
    })) {
      var type_id = $(event.target).attr('data-event-type-id');
      if (type_id == "") {
	this.model.set('event_type', null);
      } else {
	this.model.set('event_type', type_id);
      }
      this.model.modified = true;
    }
  },

  setParticipantType: function(event) {
    if (app.hasAuth({
      auth_type: 'Roles',
      write: true,
    })) {
      var type_id = $(event.target).attr('data-participant-type-id');
      if (type_id == "") {
	this.model.set('participant_type', null);
      } else {
	this.model.set('participant_type', type_id);
      }
      this.model.modified = true;
    }
  },

  switchWrite: function(event) {
    if (app.hasAuth({
      auth_type: 'Roles',
      write: true,
    })) {
      this.model.set('write', !this.model.get('write'));
      this.model.modified = true;
    }
  },

  render: function() {
    var role_name = '{{.I "All" }}';
		if (this.model.get('role') != null && this.model.get('role') != '') {
			role_name = this.model.get('role');
		}
    var location_name = '{{.I "All" }}';
    var location = this.locations.get(this.model.get('location'));
    if (location != null) location_name = location.get('name');
    var event_kind_name = '{{.I "All" }}';
    var event_kind = this.event_kinds.get(this.model.get('event_kind'));
    if (event_kind != null) event_kind_name = event_kind.get('name');
    var event_type_name = '{{.I "All" }}';
    var event_type = this.event_types.get(this.model.get('event_type'));
    if (event_type != null) event_type_name = event_type.get('name');
    var participant_type_name = '{{.I "All" }}';
    var participant_type = this.participant_types.get(this.model.get('participant_type'));
    if (participant_type != null) participant_type_name = participant_type.get('name');
    this.$el.html(this.template({ 
      model: this.model,
      type: app.authTypes[this.model.get('auth_type')],
      location_name: location_name,
      event_kind_name: event_kind_name,
      event_type_name: event_type_name,
      participant_type_name: participant_type_name,
			role_name: role_name,
    }));
    var that = this;
    that.$("#available_locations").append(new AvailableLocationView({ model: new Location({ id: null, name: '{{.I "All"}}' }) }).render().el);
    this.locations.forEach(function(location) {
      that.$("#available_locations").append(new AvailableLocationView({ model: location }).render().el);
    });
    that.$("#available_event_kinds").append(new AvailableEventKindView({ model: new EventKind({ id: null, name: '{{.I "All" }}' }) }).render().el);
    this.event_kinds.forEach(function(event_kind) {
      that.$("#available_event_kinds").append(new AvailableEventKindView({ model: event_kind }).render().el);
    });
    that.$("#available_event_types").append(new AvailableEventTypeView({ model: new EventType({ id: null, name: '{{.I "All" }}' }) }).render().el);
    this.event_types.forEach(function(event_type) {
      if (that.model.get('event_kind') == null || event_type.get('event_kind') == that.model.get('event_kind')) {
	that.$("#available_event_types").append(new AvailableEventTypeView({ model: event_type }).render().el);
      }
    });
    that.$("#available_participant_types").append(new AvailableParticipantTypeView({ model: new ParticipantType({ id: null, name: '{{.I "All" }}' }) }).render().el);
    this.participant_types.forEach(function(participant_type) {
      that.$("#available_participant_types").append(new AvailableParticipantTypeView({ model: participant_type }).render().el);
    });
		that.$('#available_roles').append(new AvailableRoleView({ model: new Role({ id: null, name: '{{.I "All" }}' }) }).render().el);
    this.roles.forEach(function(role) {
      that.$("#available_roles").append(new AvailableRoleView({ model: role }).render().el);
    });
    return this;
  },

});
