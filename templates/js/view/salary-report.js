window.SalaryReportView = Backbone.View.extend({

	template: _.template($('#salary_report_underscore').html()),

	events: {
	  "click .earlier": "moveBack",
		"click .later": "moveForward",
	  "click .fetch": "fetchData",
	},

  initialize: function(options) {
    _.bindAll(this, 'render', 'reRender');
		this.from = lastSalaryBreakpointBefore(today());
		this.to = firstSalaryBreakpointAfter(this.from);
		this.report = null;
		app.on('domainchange', this.reRender);
	},

	cleanup: function() {
		app.off('domainchange', this.reRender);
	},

	moveBack: function(ev) {
	  this.from = lastSalaryBreakpointBefore(this.from);
		this.to = firstSalaryBreakpointAfter(this.from);
		this.report = null;
		this.render();
	},

  moveForward: function(ev) {
	  this.to = firstSalaryBreakpointAfter(this.to);
		this.from = lastSalaryBreakpointBefore(this.to);
		this.report = null;
		this.render();
	},

	fetchData: function(ev) {
	  var that = this;
		that.report = new SalaryReport({
		  from: that.from,
			to: that.to,
		});
	  that.report.fetch({
		  success: that.render,
		});
	},

	reRender: function() {
	  this.report = null;
		this.render();
	},

	render: function() {
		this.$el.html(this.template({ 
			from: this.from,
			to: this.to,
			hasForward: this.to.getTime() <= new Date().getTime(),
		}));
		if (this.report != null) {
      new SalaryReportOverviewView({
			  el: this.$('#report_overview'),
			  model: this.report,
			}).render();
		}
		return this;
	}

});
