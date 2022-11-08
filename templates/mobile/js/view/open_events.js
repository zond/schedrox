window.OpenEventsView = Backbone.View.extend({

  template: _.template($('#open_events_underscore').html()),

  initialize: function(options) {
		_.bindAll(this, 'render', 'updateFilter');
		this.collection = new Events([]);
		this.collection.bind("reset", this.render);
		window.session.menu.bind('change', this.updateFilter);
		this.updateFilter();
  },

  cleanup: function() {
		window.session.menu.unbind('change', this.updateFilter);
	},

  updateFilter: function() {
		var filters = [];
		_.each(window.session.menu.get('active_filter').get('locations'), function(loc) {
			filters.push('locations=' + loc);
		});
		_.each(window.session.menu.get('active_filter').get('kinds'), function(loc) {
			filters.push('kinds=' + loc);
		});
		_.each(window.session.menu.get('active_filter').get('types'), function(loc) {
			filters.push('types=' + loc);
		});
		this.collection.url = '/events/open?' + filters.join("&");
		this.collection.fetch({ reset: true });
	},

  render: function() {
		var that = this;
    this.$el.html(this.template({})); 
		var at = null;
		this.collection.each(function(event) {
			var dateString = anyDateConverter.format(event.get('start'))
			if (at == null || dateString != at) {
				that.$('.events').append('<h4>' + {{.I "day_names" }}[event.get('start').getDay()] + " " + dateString + '</h4>');
			}
			at = anyDateConverter.format(event.get('start'));
			that.$('.events').append(new EventView({
				model: event,
				collection: that.collection,
				isOpen: true,
			}).render().el);
		});
    return this;
  },

});
