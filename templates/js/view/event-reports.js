window.EventReportsView = Backbone.View.extend({

	template: _.template($('#event_reports_underscore').html()),

  events: {
		"click .unpaid-events-link": "showReport",
		"click .changed-events-link": "showReport",
	  	"click .export-events-link": "showReport",
	  	"click .contact-events-link": "showReport",
		"click .user-events-link": "showReport",
	},

	render: function() {
		this.$el.html(this.template({ 
		}));
		return this;
	},

  	showReport: function(ev) {
		ev.preventDefault();
    		app.navigate($(ev.currentTarget).attr("href"), { trigger: true });
  	},

});
