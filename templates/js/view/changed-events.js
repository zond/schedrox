window.ChangedEventsView = Backbone.View.extend({

	template: _.template($('#changed_events_underscore').html()),

  initialize: function() {
		var that = this;
		_.bindAll(that, 'render');
		that.collection = new LatestChanges();
		that.collection.bind("change", that.render);
		that.collection.bind("reset", that.render);
		that.collection.bind("add", that.render);
		that.collection.bind("remove", that.render);
		that.collection.fetch({ reset: true });
	},

	render: function() {
		this.$el.html(this.template({ }));
    this.collection.forEach(function(change) {
      this.$('.changed-events').append(new EventChangeView({ model: change }).render().el);
    });
		return this;
	},

});
