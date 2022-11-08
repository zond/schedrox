window.CustomFilterView = Backbone.View.extend({

  template: _.template($('#custom_filter_underscore').html()),

  events: {
    "click .custom-filter": "enableCustomFilter",
  },

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.current = options.current;
  },

  enableCustomFilter: function() {
    this.current.set('locations', this.model.get('locations'), { silent: true });
    this.current.set('kinds', this.model.get('kinds'), { silent: true });
    this.current.set('types', this.model.get('types'), { silent: true });
    this.current.set('users', this.model.get('users'), { silent: true });
		this.current.storeInLocalStorage();
    this.current.trigger('change');
  },

  isCurrent: function() {
    return _.isEqual(this.current.get('users'), this.model.get('users')) && _.isEqual(this.current.get('locations'), this.model.get('locations')) && _.isEqual(this.current.get('kinds'), this.model.get('kinds')) && _.isEqual(this.current.get('types'), this.model.get('types'));
  },

  render: function() {
    this.$el.html(this.template({ model: this.model }));
    if (this.isCurrent()) {
      this.$('.btn').addClass('btn-primary');
    }
    return this;
  },

});
