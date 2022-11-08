window.ContactsResultsView = Backbone.View.extend({

  template: _.template($('#contacts_results_underscore').html()),

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.collection.bind("change", this.render);
    this.collection.bind("reset", this.render);
    this.collection.bind("add", this.render);
    this.collection.bind("remove", this.render);
  },

  render: function() {
    var that = this;
    this.$el.html(this.template({
      pages: Math.ceil(this.collection.total / this.collection.limit),
      page: Math.ceil(this.collection.offset / this.collection.limit) + 1,
      isSearch: this.collection.query != null,
    }));
    this.collection.forEach(function(contact) {
      that.$("#contact_list").append(new ContactView({ model: contact }).render().el);
    });
    return this;
  }

});
