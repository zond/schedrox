window.CustomFiltersView = Backbone.View.extend({

	template: _.template($('#custom_filters_underscore').html()),

	events: {
		"click .clear-current-filter": "clearCurrentFilter",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.collection.bind("change", this.render);
		this.collection.bind("reset", this.render);
		this.collection.bind("add", this.render);
		this.collection.bind("remove", this.render);
		this.model.bind('change', this.render);
	},

	clearCurrentFilter: function(event) {
		event.preventDefault();
		this.model.set('locations', [], { silent: true });
		this.model.set('kinds', [], { silent: true });
		this.model.set('types', [], { silent: true });
		this.model.set('users', [], { silent: true });
		this.model.storeInLocalStorage();
		this.model.trigger('change');
	},

	render: function() {
		var that = this;
		this.$el.html(this.template({ 
			show_all: that.collection.length > 0,
		}));
		if ((this.model.get('users') == null || this.model.get('users').length == 0) && (this.model.get('locations') == null || this.model.get('locations').length == 0) && (this.model.get('kinds') == null || this.model.get('kinds').length == 0) && (this.model.get('types') == null || this.model.get('types').length == 0)) {
			this.$('.clear-current-filter').addClass('btn-primary');
		}
		this.collection.forEach(function(filter) {
			that.$el.append(new CustomFilterView({ 
				model: filter,
				current: that.model,
			}).render().el);
		});
		this.delegateEvents();
		return this;
	},

});
