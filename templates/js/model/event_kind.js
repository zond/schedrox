window.EventKind = Backbone.Model.extend({
  urlRoot: '/event_kinds',
  // options.event_types: available event types
  // options.location: current location
  valid: function(options) {
    var that = this;
    return (options.event_types.any(function(type) {
      return type.valid({
        location: options.location,
	event_kind: that.get('id'),
      });
    }) || app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: that.get('location'),
    }));
  },
});
