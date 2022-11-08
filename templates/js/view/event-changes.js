window.EventChangesView = Backbone.View.extend({

  template: _.template($('#event_changes_underscore').html()),

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.collection.bind("change", this.render);
    this.collection.bind("reset", this.render);
    this.collection.bind("add", this.render);
    this.collection.bind("remove", this.render);
  },

  render: function() {
    this.$el.html(this.template({ })); 
    this.collection.forEach(function(change) {
      this.$('#changes_list').append(new EventChangeView({ model: change }).render().el);
    });
    return this;
  },

});
