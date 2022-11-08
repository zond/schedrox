window.UserRolesView = Backbone.View.extend({

  template: _.template($('#user_roles_underscore').html()),

  events: {
    "click .available-role": "addRole",
  },

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.collection.bind("change", this.render);
    this.collection.bind("reset", this.render);
    this.collection.bind("add", this.render);
    this.collection.bind("remove", this.render);
    this.available_roles = options.available_roles;
  },

  addRole: function(event) {
    event.preventDefault();
    var newRole = new Role({ 
      name: $(event.target).attr('data-role-name'),
    });
    this.collection.add(newRole);
  },

  render: function() {
    var that = this;
    this.$el.html(this.template({}));
    this.collection.forEach(function(role) {
      that.$("#role_list").append(new RoleView({ 
				model: role,
				hideDetails: true,
				removal: function(role) {
					that.collection.remove(role);
				},
      }).render().el);
    });
    this.available_roles.forEach(function(role) {
      that.$("#available_roles").append(new AvailableRoleView({ model: role }).render().el);
    });
    return this;
  },

});
