window.AvailableParticipantTypeView = Backbone.View.extend({

  tagName: 'li',

  template: _.template($('#available_participant_type_underscore').html()),

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.model.bind("change", this.render);
    this.klass = options.klass || 'available-participant-type';
  },

  render: function() {
    this.$el.html(this.template({ 
      model: this.model, 
      klass: this.klass, 
    }));
    return this;
  },

});
