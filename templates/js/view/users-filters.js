window.UsersFiltersView = Backbone.View.extend({

	template: _.template($('#users_filters_underscore').html()),

	events: {
		"keyup #search_users": "updateSearchTerm",
		"click #available_filter_types a": "addFilter",
		"click #filter_list .close": "removeFilter",
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'updateUsers');
		this.filters = options.filters;
		this.terms = [];
		this.users = options.users;
		this.user_properties_for_domain = options.user_properties_for_domain;
		this.available_roles = options.available_roles;
		this.results_view = null;
	},

	removeFilter: function(ev) {
		this.filters = _.reject(this.filters, function(filter) {
			return filter.type == $(ev.target).attr('data-filter-type') && ('' + filter.value) == ('' + $(ev.target).attr('data-filter-value'));
		});
		this.render();
		this.updateUsers();
	},

	attestFilter: function() {
	  return _.find(this.filters, function(filter) {
		  return filter.type == 'attested' || filter.type == 'unattested';
		});
	},

	addFilter: function(ev) {
		ev.preventDefault();
		this.filters.push({
			type: $(ev.target).attr('data-filter-type'),
			name: $(ev.target).attr('data-filter-name'),
			value: '',
		});
		this.render();
		this.updateUsers();
	},

	updateSearchTerm: function(ev) {
		var val = $(ev.target).val();
		this.terms = [];
		var values = val.split(" ");
		for (var i = 0; i < values.length; i++) {
			if (values[i].trim().length > 1) {
				this.terms.push(values[i].trim());
			}
		}
		this.updateUsers();
	},

	updateUsers: function() {
		this.results_view.uncheckAll();
		var properFilters = _.filter(this.filters, function(filter) {
			return filter.value != '';
		});
		if (this.terms.length > 0 || properFilters.length > 0) {
			this.users.search(this.terms.join(" "), properFilters);
		} else {
			this.users.query = '';
			this.users.filters = [];
			this.users.fetch({ reset: true });
		}
	},

	render: function() {
		var that = this;
		this.$el.html(that.template({ 
			query: that.terms.join(" "),
		}));
		_.each(this.filters, function(filter) {
			that.$('#filter_list').append(new UserFilterView({ 
				available_roles: that.available_roles,
				user_properties_for_domain: that.user_properties_for_domain,
				model: filter,
				update: that.updateUsers,
			}).render().el);
		});
		return this;
	}

});
