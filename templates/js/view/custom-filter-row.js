window.CustomFilterRowView = Backbone.View.extend({

  template: _.template($('#custom_filter_row_underscore').html()),

  tagName: 'tr',

  events: {
    "click .close": "deleteFilter",
  },

  deleteFilter: function(ev) {
    var that = this;
    myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(that.model.get("name").htmlEscape()), function() {
      that.model.destroy();
    });
  },

  render: function() {
    this.$el.html(this.template({ model: this.model }));
    return this;
  },

});
