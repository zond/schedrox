window.RoleView = Backbone.View.extend({

  tagName: 'tr',

  template: _.template($('#role_underscore').html()),

  events: {
    "click .close": "removeRole",
    "click .open": "openRole",
  },

  initialize: function(options) {
    this.removal = options.removal;
    this.hideDetails = options.hideDetails;
    _.bindAll(this, 'render');
    this.model.bind("change", this.render);
  },

  openRole: function(event) {
    event.preventDefault();
    if (!self.hideDetails) {
      new RoleDetailsView({ model: this.model }).modal(function() {
        app.navigate('/settings/roles');
      });
    }
  },

  removeRole: function(event) {
    if (app.hasAuth({
      auth_type: 'Roles',
		  role: this.model.get('name'),
      write: true,
    })) {
      this.removal(this.model);
    }
  },

  render: function() {
    this.$el.html(this.template({ 
      model: this.model,
      hideDetails: this.hideDetails,
    }));
    return this;
  },

});
