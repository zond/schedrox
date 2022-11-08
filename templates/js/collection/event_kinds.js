window.EventKinds = Backbone.Collection.extend({
  model: EventKind,
  url: "/event_kinds",
});
