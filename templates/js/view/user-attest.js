window.UserAttestView = Backbone.View.extend({

	template: _.template($('#user_attest_underscore').html()),

	events: {
	  "click .attest": "attest",
	  "click .revert": "revert",
		"click .toggle-finished": "toggleFinished",
		"click .create_new_report": "addReportedHours",
		"change .report-location": "validate",
		"change .report-event-type": "showLocations",
		"change .report-participant-type": "showEventTypes",
		"change .create_new_report_minutes": "validate",
	},

	initialize: function(options) {
		var that = this;
		_.bindAll(that, 'render');
		that.model.bind("change", that.render);
		that.from = options.from;
		that.to = options.to;
		that.attested_events = new AttestedEvents(null, { url: "/users/" + that.model.get('id') + '/attested_events?from=' + parseInt(that.from.getISOTime() / 1000) + '&to=' + parseInt(that.to.getISOTime() / 1000) });
		that.attestable_events = new AttestableEvents(null, { url: "/users/" + that.model.get('id') + '/attestable_events?from=' + parseInt(that.from.getISOTime() / 1000) + '&to=' + parseInt(that.to.getISOTime() / 1000) });
		_.each([that.attested_events, that.attestable_events], function(coll) {
			coll.bind("change", that.render);
			coll.bind("reset", that.render);
			coll.bind("add", that.render);
			coll.bind("remove", that.render);
			coll.fetch({ reset: true });
		});
		that.event_types = new EventTypes();
		that.participant_types = new ParticipantTypes();
		that.locations = new Locations();
		_.each([that.locations, that.event_types, that.participant_types], function(coll) {
			coll.bind("reset", that.render);
			coll.fetch({ reset: true });
		});
	},

	validate: function() {
	  if (this.$('.create_new_report_location').select2('val').trim() != '' &&
		    this.$('.create_new_report_event_type').select2('val').trim() != '' &&
				this.$('.create_new_report_participant_type').select2('val').trim() != '' &&
				this.$('.create_new_report_minutes').select2('val').trim() != '') {
			this.$('.create_new_report').removeAttr('disabled');
		} else {
			this.$('.create_new_report').attr('disabled', 'disabled');
		}
	},

	addReportedHours: function(ev) {
		var part= this.$('.create_new_report_participant_type').select2('val');
	  var loc = this.$('.create_new_report_location').select2('val');
		var type = this.$('.create_new_report_event_type').select2('val');
		var kind = this.event_types.get(type).get('event_kind');
		var start = anyDateConverter.parse(this.$('.create_new_report_date').val());
		var end = new Date(start.getTime() + (1000 * 60 * parseInt(this.$('.create_new_report_minutes').select2('val'))));
		var information = this.$('.create_new_report_information').val();
		var that = this;
	  ev.preventDefault();
		that.attestable_events.create({
		  salary_attested_participant_type: part,
		  salary_attested_participant_type_name: that.participant_types.get(part).get('name'),
		  start: start,
			end: end,
			location: loc,
			location_name: that.locations.get(loc).get('name'),
			event_kind: kind,
			event_type: type,
      salary_time_reported: true,
			event_type_name: that.event_types.get(type).get('name'),
			information: information,
		});
	},

	revert: function(ev) {
	  var that = this;
		that.attested_events.revert(function() {
		  that.attestable_events.fetch({ reset: true });
		});
	},

	attest: function(ev) {
	  var that = this;
		this.attestable_events.attest(that.from, that.to, function() {
		  that.attested_events.fetch({ reset: true });
		});
	},

	toggleFinished: function(ev) {
	  ev.preventDefault();
		var that = this;
		if (this.attestable_events.finished) {
			$.ajax('/users/' + this.model.get('id') + '/reported/finished?from=' + (that.from.getISOTime() / 1000) + '&to=' + (that.to.getISOTime() / 1000), {
				type: 'DELETE',
				headers: {
					'Accept': 'application/json',
				},
				success: function() {
					that.attestable_events.finished = false;
					that.render();
				},
			});
		} else {
			$.ajax('/users/' + this.model.get('id') + '/reported/finished?from=' + (that.from.getISOTime() / 1000) + '&to=' + (that.to.getISOTime() / 1000), {
				type: 'POST',
				dataType: 'json',
				data: JSON.stringify(''),
				success: function() {
					that.attestable_events.finished = true;
					that.render();
				},
			});
		}
	},

	modal: function(cb) {
		var that = this;
		$.modal.close();
		app.navigate('/users/' + that.model.get('id') + '/attest/' + that.from.getTime() + '-' + that.to.getTime());
		mymodal(that.render().el, {
			'onClose': function() {
				if (cb != null) {
					cb();
				}
			},
		});
	},

	showEventTypes: function() {
		var that = this;
	  var participant_type = that.$('.create_new_report_participant_type').select2('val');
		var previous = that.$('.create_new_report_event_type').select2('val');
		if ((typeof previous != 'string' || previous == '') && app.user.get('default_event_type') != null) {
		  previous = app.user.get('default_event_type');
		}
		var previous_allowed = false;
		var allowed_types = [];
		that.$('select.create_new_report_event_type').empty();
		that.$('select.create_new_report_event_type').append('<option></option>');
		that.event_types.each(function(item) {
			if (app.hasAnyAuth({
				auth_type: 'Attest',
				participant_type: participant_type,
				event_type: item.get('id'),
				event_kind: item.get('event_kind'),
			})) {
			  if (item.get('id') == previous) {
				  previous_allowed = true;
				}
				allowed_types.push(item.get('id'));
				that.$('select.create_new_report_event_type').append('<option value="' + item.get('id') + '">' + item.get('name') + '</option>');
			}
			that.$('select.create_new_report_event_type').select2({ placeholder: '{{.I "Event type" }}' });
		});
		if (allowed_types.length == 1) {
		  that.$('select.create_new_report_event_type').select2('val', allowed_types[0]);
		} else if (previous_allowed) {
		  that.$('select.create_new_report_event_type').select2('val', previous);
		}
		this.showLocations();
	},

	showLocations: function() {
		var that = this;
	  var participant_type = that.$('.create_new_report_participant_type').select2('val');
    var event_type = this.$('.create_new_report_event_type').select2('val');
		var previous = this.$('.create_new_report_location').select2('val');
		if ((typeof previous != 'string' || previous == '') && app.user.get('default_participant_type') != null) {
		  previous = app.user.get('default_location');
		}
		var previous_allowed = false;
		var kind = null;
		if (event_type != '' && event_type != null) {
		  var t = that.event_types.get(event_type);
			if (t != null) {
				kind = t.get('event_kind');
			}
		}
		var allowed_types = [];
		this.$('select.report-participant-type').empty();
		that.$('select.report-participant-type').append('<option></option>');
		that.locations.each(function(item) {
			if (app.hasAnyAuth({
				auth_type: 'Attest',
				participant_type: participant_type,
				event_kind: kind,
				event_type: event_type,
				location: item.get('id'),
			})) {
				if (item.get('id') == previous) {
					previous_allowed = true;
				}
				allowed_types.push(item.get('id'));
				that.$('select.create_new_report_location').append('<option value="' + item.get('id') + '">' + item.get('name') + '</option>');
			}
			that.$('select.create_new_report_location').select2({ placeholder: '{{.I "Location" }}' });
		});
		if (allowed_types.length == 1) {
		  that.$('select.create_new_report_location').select2('val', allowed_types[0]);
		} else if (previous_allowed) {
		  that.$('select.create_new_report_location').select2('val', previous);
		}
	},

	render: function() {
		var that = this;
		var deletable = !this.attestable_events.attested && !this.attestable_events.finished;
		that.$el.html(that.template({ 
		  attested: this.attestable_events.attested,
			finished: this.attestable_events.finished,
			deletable: deletable,
			from: that.from,
		}));
		var lastColumn = '';
		if (deletable) {
		  lastColumn = '<td></td>';
		}
		var allMinutes = 0;
		var lastDate = null;
		var lastMinutes = 0;
		var thisDate = null;
		that.attestable_events.each(function(ev) {
			thisDate = anyDateConverter.format(ev.get('start'));
			if (lastDate != null && thisDate != lastDate) {
			  that.$('#unattested_events').append('<tr><td colspan=5></td><td class="align-right"><strong>{{.I "Sum"}} ' + hoursMinutesForMinutes(lastMinutes) + '</strong></td>' + lastColumn + '</tr>');
				lastMinutes = 0;
			}
			var options = { 
				model: ev,
			};
      if (deletable) {
			  if (ev.get('salary_time_reported')) {
					options.deleter = function() {
						ev.destroy();
					};
				} else {
				  options.deleter = 'no';
				}
			}
			that.$('#unattested_events').append(new AttestableEventView(options).render().el);
			lastDate = thisDate;
			var theseMinutes = (ev.get('end').getTime() - ev.get('start').getTime()) / (1000 * 60);
			lastMinutes += theseMinutes;
			allMinutes += theseMinutes;
		});
		that.$('#unattested_events').append('<tr><td colspan=5></td><td class="align-right"><strong>{{.I "Sum"}} ' + hoursMinutesForMinutes(lastMinutes) + '</strong></td>' + lastColumn + '</tr>');
		that.$('#unattested_events').append('<tr><td colspan=5></td><td class="align-right"><strong>{{.I "Total sum"}} ' + hoursMinutesForMinutes(allMinutes) + '</strong></td>' + lastColumn + '</tr>');
		allMinutes = 0;
		if (that.attestable_events.finished) {
			that.$(".attest").removeAttr('disabled');
			if (that.attestable_events.length == 0) {
				that.$('.attest-row').addClass('hidden');
			} else {
				that.$('.attest-row').removeClass('hidden');
			}
		} else {
		  that.$(".attest").attr('disabled', 'disabled');
		}
		that.attested_events.each(function(ev) {
			thisDate = anyDateConverter.format(ev.get('start'));
			if (lastDate != null && thisDate != lastDate) {
			  that.$('#attested_events').append('<tr><td colspan=5></td><td class="align-right"><strong>{{.I "Sum"}} ' + hoursMinutesForMinutes(lastMinutes) + '</strong></td>' + lastColumn + '</tr>');
				lastMinutes = 0;
			}
			var options = { 
				model: ev,
				show_attester: true,
			};
      if (deletable) {
			  if (ev.get('salary_time_reported')) {
					options.deleter = function() {
						ev.destroy();
					};
				} else {
				  options.deleter = 'no';
				}
			}
			that.$('#attested_events').append(new AttestableEventView(options).render().el);
			lastDate = thisDate;
			var theseMinutes = (ev.get('end').getTime() - ev.get('start').getTime()) / (1000 * 60);
			lastMinutes += theseMinutes;
			allMinutes += theseMinutes;
		});
		that.$('#attested_events').append('<tr><td colspan=5></td><td class="align-right"><strong>{{.I "Sum"}} ' + hoursMinutesForMinutes(lastMinutes) + '</strong></td>' + lastColumn + '</tr>');
		that.$('#attested_events').append('<tr><td colspan=5></td><td class="align-right"><strong>{{.I "Total sum"}} ' + hoursMinutesForMinutes(allMinutes) + '</strong></td>' + lastColumn + '</tr>');
		if (that.attested_events.length == 0) {
		  that.$('.revert-row').addClass('hidden');
		} else {
		  that.$('.revert-row').removeClass('hidden');
		}
		if (deletable) {
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
				that.$('#create_new_report_date').AnyTime_noPicker().AnyTime_picker(options);
			}, 500);
			var allowed_participant_types = [];
			var previous = app.user.get('default_participant_type');
			var previous_allowed = false;
			that.participant_types.each(function(item) {
				if (app.hasAnyAuth({
					auth_type: 'Attest',
					participant_type: item.get('id'),
				})) {
					if (item.get('id') == previous) {
						previous_allowed = true;
					}
					allowed_participant_types.push(item.get('id'));
					that.$('select.create_new_report_participant_type').append('<option value="' + item.get('id') + '">' + item.get('name') + '</option>');
				}
				that.$('select.create_new_report_participant_type').select2({ placeholder: '{{.I "Participant type" }}' });
			});
			if (allowed_participant_types.length == 1) {
				that.$('select.create_new_report_participant_type').select2('val', allowed_participant_types[0]);
			} else if (previous_allowed) {
				that.$('select.create_new_report_participant_type').select2('val', previous);
			}
			that.showEventTypes();
		}
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
		that.$('.create_new_report_minutes').select2({
		  data: minutes,
		});
		return that;
	},

});
