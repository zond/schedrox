window.Location = Backbone.Model.extend({
  urlRoot: '/locations',
  // options.event_kinds: available event kinds
  // options.event_types: available event types
  valid: function(options) {
    var that = this;
    return (options.event_kinds.any(function(kind) {
      return kind.valid({
        location: that.get('id'),
	event_types: options.event_types,
      });
    }) || app.hasAuth({
      auth_type: 'Events',
      write: true,
    }));
  },
});
