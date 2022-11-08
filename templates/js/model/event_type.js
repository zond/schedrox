window.EventType = Backbone.Model.extend({
  urlRoot: '/event_types',
  // options.event_kind: current event kind
  // options.location: current location
  valid: function(options) {
    return options.event_kind == this.get('event_kind') && app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: options.location,
      event_kind: this.get('event_kind'),
      event_type: this.get('id'),
    });
  },
});
