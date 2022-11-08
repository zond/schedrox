window.MyEventsView = Backbone.View.extend({

  template: _.template($('#my_events_underscore').html()),

  initialize: function(options) {
		_.bindAll(this, 'render');
		this.collection = new Events([], { url: '/events/mine' });
		this.collection.bind("reset", this.render);
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
				isOpen: false,
			}).render().el);
		});
    return this;
  },

});
