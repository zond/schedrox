window.ParticipantTypes = Backbone.Collection.extend({
  model: ParticipantType,
  url: "/participant_types",
});
