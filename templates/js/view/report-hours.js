window.ReportHoursView = Backbone.View.extend({

	template: _.template($('#report_hours_underscore').html()),

	events: {
	  "click .earlier": "moveBack",
		"click .later": "moveForward",
		"click #period_finished": "toggleFinished",
	},

  initialize: function(options) {
    _.bindAll(this, 'render', 'updateSum');
		this.from = lastSalaryBreakpointBefore(today());
		this.to = firstSalaryBreakpointAfter(this.from);
		this.event_types = new EventTypes();
		this.participant_types = new ParticipantTypes();
		this.locations = new Locations();
		this.reported_hours = new ReportedEvents([], {
		  from: this.from,
			to: this.to,
		});
		var that = this;
		_.each([this.reported_hours, this.locations, this.event_types, this.participant_types], function(coll) {
			coll.bind("reset", that.render);
			coll.fetch({ reset: true });
		});
		this.reported_hours.bind('remove', this.updateSum);
		this.reported_hours.bind('add', this.updateSum);
	},

	toggleFinished: function(ev) {
	  ev.preventDefault();
		var that = this;
		if (!this.reported_hours.finished) {
		  myconfirm("{{.I "You will not be able to revoke this decision. Are you sure?" }}", function() {
				$.ajax('/reported/finished?from=' + (that.from.getISOTime() / 1000) + '&to=' + (that.to.getISOTime() / 1000), {
					type: 'POST',
					dataType: 'json',
					data: JSON.stringify(''),
					success: function() {
					  that.reported_hours.finished = true;
					  that.render();
					},
				});
			});
		}
	},

	moveBack: function(ev) {
	  this.from = lastSalaryBreakpointBefore(this.from);
		this.to = firstSalaryBreakpointAfter(this.from);
		this.reported_hours.from = this.from;
		this.reported_hours.to = this.to;
    this.reported_hours.fetch({ reset: true });
	},

  moveForward: function(ev) {
	  this.to = firstSalaryBreakpointAfter(this.to);
		this.from = lastSalaryBreakpointBefore(this.to);
		this.reported_hours.from = this.from;
		this.reported_hours.to = this.to;
    this.reported_hours.fetch({ reset: true });
	},

	updateSum: function() {
		var minutes = 0;
		this.reported_hours.each(function(ev) {
			minutes += (ev.get('end').getTime() - ev.get('start').getTime()) / (1000 * 60);
		});
		var hours = parseInt(minutes / 60);
		var minutes = minutes - (hours * 60);
		var minStr = '' + minutes;
		if (minStr.length == 1) {
			minStr = '0' + minStr;
		}
		this.$('.total-sum').text('' + hours + ':' + minStr);
	},

	render: function() {
		this.$el.html(this.template({ 
		  finished: this.reported_hours.finished,
			from: this.from,
			to: this.to,
		}));
		this.updateSum();
		var minutes = [];
		var minMinutes = app.getDomain().get('salary_config').salary_report_hours_min_minutes;
		if (minMinutes == 0) {
		  minMinutes = 1;
		}
		for (var i = minMinutes; i < 16 * 60; i += minMinutes) {
		  minutes.push({
			  id: i,
				text: hoursMinutesForMinutes(i),
			});
		}
		var reportedByDay = {};
		this.reported_hours.each(function(ev) {
		  var d = anyDateConverter.format(ev.get('start'));
		  if (reportedByDay[d] == null) {
			  reportedByDay[d] = [];
			}
			reportedByDay[d].push(ev);
		});
		var t = this.from;
		var tabindex = 0;
		while (t.getTime() < this.to.getTime()) {
			this.$('#days_list').append(new ReportHoursDayView({
				date: t,
				reported_hours: this.reported_hours,
				locations: this.locations,
				event_types: this.event_types,
				participant_types: this.participant_types,
				minutes: minutes,
				tabindexer: function() {
				  tabindex++;
					return tabindex;
				},
			}).render().el);
			var that = this;
			var d = anyDateConverter.format(t);
			_.each(reportedByDay[d] || [], function(ev) {
				this.$('#days_list').append(new ReportedHoursView({
					finished: that.reported_hours.finished,
					model: ev,
				}).render().el);
			});
			t = new Date(t.getFullYear(), t.getMonth(), t.getDate() + 1);
		}
		$('.report-location').first().select2('focus');
		return this;
	}

});
