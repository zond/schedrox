window.ReportHoursDayView = Backbone.View.extend({

	template: _.template($('#report_hours_day_underscore').html()),

	tagName: 'tr',

	className: 'report-hours-day',

	events: {
	  "click .report-add": "addHours",
		"change .report-location": "showEventTypes",
		"change .report-event-type": "showParticipantTypes",
		"change .report-participant-type": "validate",
		"change .report-minutes": "validate",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.date = options.date;
		this.locations = options.locations;
		this.event_types = options.event_types;
		this.participant_types = options.participant_types;
		this.minutes = options.minutes;
		this.reported_hours = options.reported_hours;
		this.tabindexer = options.tabindexer;
	},

	addHours: function(ev) {
	  var loc = this.$('.report-location').select2('val');
		var type = this.$('.report-event-type').select2('val');
		var kind = this.event_types.get(type).get('event_kind');
		var participant = this.$('.report-participant-type').select2('val');
		var start = this.date;
		var end = new Date(start.getTime() + (1000 * 60 * parseInt(this.$('.report-minutes').select2('val'))));
		var information = this.$('.report-information').val();
		var that = this;
		this.reported_hours.create({
		  salary_attested_participant_type: participant,
		  start: start,
			end: end,
			location: loc,
			event_kind: kind,
			event_type: type,
			information: information,
		}, {
		  success: function(ev) {
			  that.$el.parent().find('tr').eq(that.el.rowIndex).after(new ReportedHoursView({ model: ev }).render().el);
				that.$('.report-location').select2('focus');
			},
		});
	},

	validate: function() {
	  if (this.$('.report-location').select2('val').trim() != '' &&
		    this.$('.report-event-type').select2('val').trim() != '' &&
				this.$('.report-participant-type').select2('val').trim() != '' &&
				this.$('.report-minutes').select2('val').trim() != '') {
			this.$('.report-add').removeAttr('disabled');
		} else {
			this.$('.report-add').attr('disabled', 'disabled');
		}
	},

	showEventTypes: function() {
		var that = this;
	  var loc = this.$('.report-location').select2('val');
		var previous = this.$('.report-event-type').select2('val');
		if ((typeof previous != 'string' || previous == '') && app.user.get('default_event_type') != null) {
		  previous = app.user.get('default_event_type');
		}
		var previous_allowed = false;
		var allowed_types = [];
		this.$('select.report-event-type').empty();
		that.$('select.report-event-type').append('<option></option>');
		this.event_types.filter(function(item) {
			if (app.hasAnyAuth({
				auth_type: 'Report hours',
				location: loc,
				event_kind: item.get('event_kind'),
				event_type: item.get('id'),
			})) {
			  if (item.get('id') == previous) {
				  previous_allowed = true;
				}
				allowed_types.push(item.get('id'));
				that.$('select.report-event-type').append('<option value="' + item.get('id') + '">' + item.get('name') + '</option>');
			}
		});
		that.$('select.report-event-type').select2({ placeholder: '{{.I "Event type" }}' });
		if (allowed_types.length == 1) {
		  that.$('select.report-event-type').select2('val', allowed_types[0]);
		} else if (previous_allowed) {
		  that.$('select.report-event-type').select2('val', previous);
		}
		this.showParticipantTypes();
	},

	showParticipantTypes: function() {
		var that = this;
	  var loc = this.$('.report-location').select2('val');
    var type = this.$('.report-event-type').select2('val');
		var previous = this.$('.report-participant-type').select2('val');
		if ((typeof previous != 'string' || previous == '') && app.user.get('default_participant_type') != null) {
		  previous = app.user.get('default_participant_type');
		}
		var previous_allowed = false;
		var kind = null;
		if (type != '' && type != null) {
			kind = this.event_types.get(type).get('event_kind');
		}
		var allowed_types = [];
		this.$('select.report-participant-type').empty();
		that.$('select.report-participant-type').append('<option></option>');
		this.participant_types.filter(function(item) {
			if (app.hasAnyAuth({
				auth_type: 'Report hours',
				location: loc,
				event_kind: kind,
				event_type: type,
				participant_type: item.get('id'),
			})) {
			  if (item.get('id') == previous) {
				  previous_allowed = true;
				}
				allowed_types.push(item.get('id'));
				that.$('select.report-participant-type').append('<option value="' + item.get('id') + '">' + item.get('name') + '</option>');
			}
		});
		that.$('select.report-participant-type').select2({ placeholder: '{{.I "Participant type" }}' });
		if (allowed_types.length == 1) {
		  that.$('select.report-participant-type').select2('val', allowed_types[0]);
		} else if (previous_allowed) {
		  that.$('select.report-participant-type').select2('val', previous);
		}
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({ 
		  finished: that.reported_hours.finished,
		  date: that.date,
			tabindexer: that.tabindexer,
		}));
		var previous = app.user.get('default_location');
		var previous_allowed = false;
		var allowed_locations = [];
		that.$('select.report-location').empty();
		that.$('select.report-location').append('<option></option>');
		that.locations.each(function(item) {
			if (app.hasAnyAuth({
				auth_type: 'Report hours',
				location: item.get('id'),
			})) {
			  if (item.get('id') == previous) {
				  previous_allowed = true;
				}
			  allowed_locations.push(item.get('id'));
			  that.$('select.report-location').append('<option value="' + item.get('id') + '">' + item.get('name') + '</option>');
			}
		});
		that.$('select.report-location').select2({ placeholder: '{{.I "Location" }}' });
		if (allowed_locations.length == 1) {
		  that.$('select.report-location').select2('val', allowed_locations[0]);
		} else if (previous_allowed) {
		  that.$('select.report-location').select2('val', previous);
		}
		that.showEventTypes();
		that.$('.report-minutes').select2({
		  data: that.minutes,
		});
		return that;
	},

});
