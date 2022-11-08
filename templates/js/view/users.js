window.UsersView = Backbone.View.extend({

	template: _.template($('#users_underscore').html()),

	initialize: function(options) {
		_.bindAll(this, 'render','refetch');
		this.show_user = options.show_user;
		this.attest_user = options.attest_user;
		this.attest_from = options.attest_from;
		this.attest_to = options.attest_to;
		this.collection = new Users();
		this.collection.filters.push({
			name: '{{.I "Disabled" }}',
      type: 'disabled',
      value: 'false',
			desc: '{{.I "No" }}',
		});
		this.user_properties_for_domain = new UserPropertiesForDomain([], { url: '/user_properties' });
		this.available_roles = new Roles([], { url: '/roles' });
		this.refetch();
		app.on('domainchange', this.refetch);
	},

	refetch: function() {
		if (app.getDomain() != null) {
			this.collection.fetch({ reset: true });
			if (app.hasAnyAuth({
				auth_type: 'Roles',
			})) {
				this.available_roles.fetch({ reset: true});
			}
			if (app.hasAuth({
				auth_type: 'Domain',
			})) {
				this.user_properties_for_domain.fetch({ reset: true });
			}
		}
	},

	cleanup: function() {
		app.off('domainchange', this.refetch);
	},

	render: function() {
		var that = this;
		this.$el.html(this.template({ 
		}));
		var filtersView = new UsersFiltersView({
			filters: this.collection.filters,
			el: this.$('#users_filters'),
			users: this.collection,
			available_roles: this.available_roles,
			user_properties_for_domain: this.user_properties_for_domain,
		}).render();
		var results_view = new UsersResultsView({ 
			el: this.$("#users_results"),
			user_properties_for_domain: this.user_properties_for_domain,
			available_roles: this.available_roles,
			available_properties: this.user_properties_for_domain,
			collection: this.collection,
			filtersView: filtersView,
		}).render();
		filtersView.results_view = results_view;
		if (that.show_user != null) {
			var that = this;
			var uid = that.show_user.get('id');
			new UserDetailsView({ 
				model: that.show_user,
				user_properties_for_domain: that.user_properties_for_domain,
				available_roles: that.available_roles,
				opener: function() {
				  app.showUser(uid);
				},
			}).modal(function() {
				app.navigate('/users');
			});
			this.show_user = null;
		} else if (that.attest_user != null) {
			var that = this;
			var uid = that.attest_user.get('id');
			new UserAttestView({ 
				model: that.attest_user, 
				from: that.attest_from,
				to: that.attest_to,
			}).modal(function() {
				app.navigate('/users');
			});
			this.attest_user = null;
		}
		return this;
	}

});
