window.ContactEventView = Backbone.View.extend({

	template: _.template($('#contact_event_underscore').html()),

	tagName: 'tr',

	events: {
	  "click .open-event": "openEvent",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	openEvent: function(ev) {
	  ev.preventDefault();
		$.modal.close();
		app.showEvent(this.model.get('id'));
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({ 
		  model: that.model,
			duration: hoursMinutesForDates(that.model.get('start'), that.model.get('end')),
		}));
		return that;
	},

});
