window.RolesSettingsView = Backbone.View.extend({

	template: _.template($('#roles_settings_underscore').html()),

	events: {
		"change #new_role": "addRole",
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'refetch');
		this.show_role = options.show_role;
		this.roles = new Roles([], { url: '/roles' });
		this.roles.bind("change", this.render);
		this.roles.bind("reset", this.render);
		this.roles.bind("add", this.render);
		this.roles.bind("remove", this.render);
		this.refetch();
		app.on('domainchange', this.refetch);
	},

	refetch: function() {
		if (app.getDomain() != null) {
			this.roles.fetch({ reset: true });
		}
	},

	cleanup: function() {
		app.off('domainchange', this.refetch);
	},

	render: function() {
		this.$el.html(this.template({}));
		this.roles.forEach(function(role) {
			this.$("#role_list").append(new RoleView({ 
				model: role,
				removal: function(role) {
					myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(role.get("name").htmlEscape()), function() {
						role.destroy();
					});
				},
			}).render().el);
		});
		if (this.show_role != null) {
			new RoleDetailsView({ model: this.show_role }).modal(function() {
				app.navigate('/settings/roles');
			});
		}
		return this;
	},

	addRole: function(event) {
		if (app.getDomain() != null) {
			var that = this;
			var newRole = new Role({ name: $("#new_role").val() });
			newRole.url = this.roles.url;
			newRole.save(null, {
				success: function() {
					newRole.url = that.roles.url + '/' + newRole.get('id');
					that.roles.add(newRole);
				},
			});
		}
	},

});
