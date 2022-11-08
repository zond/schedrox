window.ReportedHoursView = Backbone.View.extend({

	template: _.template($('#reported_hours_underscore').html()),

	tagName: 'tr',

  events: {
	  "click .report-remove": "removeHours",
	},

	className: function() {
	  return 'reported-hours-' + (this.model.get('salary_time_reported') ? 'by-user' : 'by-calendar');
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.finished = options.finished;
	},

	removeHours: function(ev) {
	  ev.preventDefault();
		var that = this;
		this.model.destroy({
		  success: function() {
			  that.$el.remove();
			},
		});
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({ 
		  finished: that.finished,
		  model: that.model,
		}));
		return that;
	},

});
