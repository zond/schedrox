window.UnpaidEventsView = Backbone.View.extend({

	template: _.template($('#unpaid_events_underscore').html()),

  events: {
		"click .fetch-button": "fetchEvents",
	},

  initialize: function() {
		this.unpaid_events = new UnpaidEvents([], { from: new Date(), to: new Date(), });
		_.bindAll(this, 'render');
		this.unpaid_events.bind("reset", this.render);
		this.from = new Date();
		this.to = new Date();
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({ 
			from: that.from,
			to: that.to,
		}));
		that.unpaid_events.each(function(ev) {
			that.$('.unpaid-events').append(new UnpaidEventView({
				model: ev,
			}).render().el);
		});
		// make AnyTime fix pickers a bit later
		setTimeout(function() {
			var options = {
		  	askSecond: false,
			  dayAbbreviations: {{.I "day_names_short"}},
		    dayNames: {{.I "day_names"}},
		    firstDOW: {{.I "firstDOW"}},
		    labelDayOfMonth: '{{.I "labelDayOfMonth"}}',
		    labelHour: '{{.I "labelHour"}}',
		    labelMinute: '{{.I "labelMinute"}}',
 		    labelMonth: '{{.I "labelMonth"}}',
 		    labelTitle: '{{.I "labelTitle"}}',
 		    labelYear: '{{.I "labelYear"}}',
 		    monthAbbreviations: {{.I "month_names_short"}},
 		    monthNames: {{.I "month_names"}},
 		    format: '{{.I "any_date_format" }}',
			};
			that.$('#report_from').AnyTime_noPicker().AnyTime_picker(options);
			that.$('#report_to').AnyTime_noPicker().AnyTime_picker(options);
		}, 500);
		return that;
	},

	fetchEvents: function(ev) {
		ev.preventDefault();
		this.to = anyDateConverter.parse(this.$('#report_to').val());
  	this.unpaid_events.to = this.to;
		var from = anyDateConverter.parse(this.$('#report_from').val());
		if (this.to.getTime() < from.getTime()) {
			from = this.to;
		} else if (this.to.getTime() - from.getTime() > (1000 * 60 * 60 * 24 * 62)) {
			from = new Date(this.to.getTime() - (1000 * 60 * 60 * 24 * 62));
		}
		this.from = from;
  	this.unpaid_events.from = this.from;
  	this.unpaid_events.fetch({reset: true});
 },

});
