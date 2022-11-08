window.AvailableEventTypeView = Backbone.View.extend({

  tagName: 'li',

  template: _.template($('#available_event_type_underscore').html()),

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	render: function() {
		this.$el.html(this.template({ model: this.model }));
		return this;
	},

});
