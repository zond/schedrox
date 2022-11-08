window.SalaryReportOverviewView = Backbone.View.extend({

	template: _.template($('#salary_report_overview_underscore').html()),

  initialize: function(options) {
    _.bindAll(this, 'render');
	},

	render: function() {
		var that = this;
	  var nUsers = 0;
		for (var uid in that.model.get('users')) {
		  nUsers++;
		}
		var time = 0;
		_.each(that.model.get('events'), function(ev) {
		  time += (ev.end.getTime() - ev.start.getTime());
		});
		$.getScript("/salary/{{.Version}}/code.js", function() {
			that.$el.html(that.template({ 
				model: that.model,
				report_mime_type: getReportMimeType(that.model.attributes),
				report_content: getReportContent(that.model.attributes), 
				nUsers: nUsers,
				hoursMinutes: hoursMinutesForMinutes(parseInt(time / (1000 * 60))),
			}));
		});
		return that;
	}

});
