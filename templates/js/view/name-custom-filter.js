window.NameCustomFilterView = Backbone.View.extend({

  template: _.template($('#name_custom_filter_underscore').html()),

  events: {
    "change #custom_filter_name": "updateName",
  },

  updateName: function(ev) {
    this.model.set('name', $(ev.target).val());
  },

  render: function() {
    this.$el.html(this.template({ }));
    return this;
  },

});
