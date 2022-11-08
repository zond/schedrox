window.EventView = Backbone.View.extend({

	template: _.template($('#event_underscore').html()),

	className: 'event-details',

	events: {
		"click .available-event-type": "set_event_type",
		"click .available-event-kind": "set_event_kind",
		"click .available-location": "set_location",
		"click #event_recurring": "switchRecurring",
		"click #event_allDay": "switchAllDay",
		"click #event_start": "clickStart",
		"click #event_end": "clickEnd",
		"click #event_recurrence_end": "clickRecurrenceEnd",
		"change #event_title": "set_title",
		"change #event_start": "set_start",
		"change #event_end": "set_end",
		"change #event_information": "set_information",
	},

	initialize: function(options) {
		var that = this;
		_.bindAll(that, 'render');
		that.model.bind("change", that.render);

		that.event_kinds = options.event_kinds;
		that.event_types = options.event_types;
		that.available_participant_types = options.participant_types;
		that.locations = options.locations;
		that.busy_meter = new BusyMeter({
		  start: that.model.get('start'),
			end: that.model.get('end'),
			ignore_event: that.model.get('id'),
		});
		that.busy_meter.save();
		that.unique_meter = new UniqueMeter({}, {
			event: that.model, 
			event_types: that.event_types,
		});
		that.unique_meter.refresh();
		var that = that;
		_.each([that.event_kinds, that.event_types, that.available_participant_types, that.locations], function(coll) {
			coll.bind("change", that.render);
			coll.bind("reset", that.render);
			coll.bind("add", that.render);
			coll.bind("remove", that.render);
		});

		that.changes = new Changes(null, { event: that.model });
		that.participants = new Participants(null, { event: that.model });
		if (!that.model.isNew()) {
			if (app.hasAuth({
				auth_type: 'Events', 
				location: that.model.get('location'),
				event_kind: that.model.get('event_kind'), 
				event_type: that.model.get('event_type'), 
				write: true,
			})) {
				that.changes.fetch({ reset: true });
			}
			that.participants.fetch({ reset: true });
		}

		that.required_participant_types = new EventRequiredParticipantTypes(null, { event: that.model });
		that.extra_required_participant_types = new ExtraRequiredParticipantTypes(null, { event: that.model });
		if (that.model.get('event_type') != null) {
			that.required_participant_types.fetch({ reset: true });
			that.extra_required_participant_types.fetch({ reset: true });
		}
		var touch = function() {
			that.model.set('_touched', new Date());
		};
		_.each([that.extra_required_participant_types, that.participants], function(coll) {
			coll.bind('change', touch);
			coll.bind('add', touch);
			coll.bind('remove', touch);
		});
	},
	
	clickStart: function(ev) {
	  putBelow($(ev.target), $('#AnyTime--event_start'));
	},

	clickEnd: function(ev) {
	  putBelow($(ev.target), $('#AnyTime--event_end'));
	},

	clickRecurrenceEnd: function(ev) {
	  putBelow($(ev.target), $('#AnyTime--event_recurrence_end'));
	},

	modal: function(cb) {
		var that = this;
		var was_recurring = this.model.get('recurring');
		var original_data = _.clone(this.model.attributes);
		var event_opener = null;
		var event_kind = this.event_kinds.get(this.model.get('event_kind'));
		var series_editable = (event_kind == null ? true : event_kind.get('series_editable'));

		var unique_save_callback = function() {
			if (that.model.valid({
				event_types: that.event_types,
			})) {
				if (was_recurring) {
					that.participants.setTimes();
					var original_event = new Event(original_data);
					var save_all = function() {
						that.model.set_to_recurrence_master_times();
						that.save(cb);
					};
					var split = function() {
						that.model.split_recurrence(function(newEvent) {
							that.model = newEvent;
							that.anonymizeCollections();
							that.save(cb);
						});
					};
					var disconnect = function() {
						that.model.create_exception(function(newEvent) {
							that.model = newEvent;
							that.anonymizeCollections();
							that.save(cb);
						});
					};
					if (app.hasAnyAuth({
						auth_type: 'Participants', 
						location: that.model.get('location'),
						event_kind: that.model.get('event_kind'), 
						event_type: that.model.get('event_type'), 
						write: true,
					}) || app.hasAuth({
						auth_type: 'Events', 
						location: that.model.get('location'),
						event_kind: that.model.get('event_kind'), 
						event_type: that.model.get('event_type'), 
						write: true,
					})) {
						if (original_event.equalExceptRecurrence(that.model) && that.participants.oldNews() && that.extra_required_participant_types.oldNews()) {
							save_all();
						} else if (!series_editable || that.participants.anySwitchedDefaulted()) {
						  disconnect();
						} else {
							mymodal("<div class=\"alert-content\"><p>{{.I "Do you want to modify all events in this series, disconnect and save this specific occurence, or split the series and modify all future events?"}}</p></div>", {
								"{{.I "All" }}": save_all,
								"{{.I "Disconnect" }}": disconnect,
								"{{.I "All future" }}": split,
								"onCancel": cb,
							}, { 
								min_height: '5%',
							});
						}
					} else { // Only attend privs, don't even ask the question. save_all will only save participants anyway.
						save_all();
					}
				} else {
					that.save(cb);
				}
			} else {
				myalert('{{.I "This event is not valid because {0}." }}'.format({{.I "invalid_event_reasons" }}[that.model.why_invalid({ event_types: that.event_types })]), event_opener);
			}
		};
		var save_callback;
		var remove_callback = function() {
			if (was_recurring) {
				var original_event = new Event(original_data);
				var disconnect = function() {
					original_event.add_exception(original_event.get('start'));
					original_event.set_to_recurrence_master_times();
					original_event.save(null, {
						success: cb,
					});
				};
				if (series_editable) {
					mymodal("<div class=\"alert-content\"><p>{{.I "Do you want to remove all events in this series, disconnect and remove this specific occurence, or remove this and all future events?"}}</p></div>", {
						"{{.I "All" }}": function() {
							that.destroy(cb);
						},
						"{{.I "All future" }}": function() {
							var start = new Date(that.model.get('start').getTime());
							original_event.set_to_recurrence_master_times();
							original_event.set('recurrence_end', new Date(start.getTime() - (1000 * 60 * 60 * 12)));
							original_event.save(null, { success: cb });
						},
						"{{.I "Disconnect" }}": disconnect,
						"onCancel": cb,
					}, { 
						min_height: '5%',
					});
				} else {
					myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(that.model.describe(that.event_types).htmlEscape()), function() {
						disconnect();
					});
				}
			} else {
				myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(that.model.describe(that.event_types).htmlEscape()), function() {
					that.destroy(cb);
				});
			}
		};
		event_opener = function() {
			$.modal.close();
			if (!that.model.isNew()) {
				app.navigate('/calendar/' + that.model.get('id'));
			}
			if (app.hasAuth({ 
				auth_type: 'Events', 
				location: that.model.get('location'),
				event_kind: that.model.get('event_kind'), 
				event_type: that.model.get('event_type'), 
				write: true,
			})) {
				var opts = {
					"{{.I "Save"}}": save_callback,
					"onCancel": function() {
						if (JSON.stringify(that.model.attributes) != JSON.stringify(original_data)) {
							mymodal("<div class=\"alert-content\"><p>{{.I "This event has unsaved changes. Are you sure you want to close it without saving?" }}</p></div>", {
								"{{.I "No" }}": event_opener,
								"{{.I "Yes" }}": cb,
								"onCancel": event_opener,
							}, {
								min_height: '5%',
							});
						} else {
							cb();
						}
					},
				};
				if (!that.model.isNew()) {
					opts["{{.I "Remove"}}"] = remove_callback;
				}
				mymodal(that.render().el, opts, { min_height: '85%' });
			} else if (app.hasAnyAuth({
				auth_type: 'Participants',
				location: that.model.get('location'),
				event_kind: that.model.get('event_kind'),
				event_type: that.model.get('event_type'),
				write: true,
			})) {
				mymodal(that.render().el, {
					"{{.I "Save"}}": save_callback,
					"onCancel": cb,
				}, { min_height: '85%' });
			} else if (app.hasAnyAuth({
				auth_type: 'Attend',
				location: that.model.get('location'),
				event_kind: that.model.get('event_kind'),
				event_type: that.model.get('event_type'),
			})) {
				mymodal(that.render().el, {
					"{{.I "Save"}}": save_callback,
					"onCancel": cb,
				}, { min_height: '85%' });
			} else {
				mymodal(that.render().el, { "onCancel": cb }, { min_height: '85%' });
			}
			that.delegateEvents();
		};
		that.event_opener = event_opener;
		recommended_save_callback = function() {
		  if (that.model.recommended()) {
				unique_save_callback();
			} else {
				mymodal('{{.I "Are you sure you want to save this event, {0}?" }}'.format({{.I "invalid_event_reasons" }}[that.model.why_not_recommended()]), {
					'{{.I "Save anyway" }}': unique_save_callback,
					'{{.I "Do not save yet" }}': event_opener,
					'onCancel': event_opener,
				}, {
					min_width: '30%',
					min_height: '10%',
				});
			}
		};
		save_callback = function() {
		  if (!that.unique_meter.isUnique()) {
			  mymodal('{{.I "This event is supposed to be the only one if its kind at any time, if you save now it will not be."}}', {
				  '{{.I "Save anyway" }}': recommended_save_callback,
					'{{.I "Do not save yet" }}': event_opener,
					"onCancel": event_opener,
				}, {
          min_width: '30%',
					min_height: '10%',
				});
			} else {
			  recommended_save_callback();
			}
		};
		event_opener();
	},

	destroy: function(cb) {
		this.model.destroy({
			success: cb,
		});
	},

	switchAllDay: function(ev) {
		this.model.set('allDay', !this.model.get('allDay'));
	},

	switchRecurring: function(ev) {
		this.model.set('recurring', !this.model.get('recurring'));
	},

	anonymizeCollections: function() {
		var newExtraRequiredParticipantTypes = new ExtraRequiredParticipantTypes(null, { event: this.model });
		this.extra_required_participant_types.each(function(req) {    
			var attributes = _.clone(req.attributes);
			delete(attributes['id']);
			newExtraRequiredParticipantTypes.add(new RequiredParticipantType(attributes), { silent: true });
		});
		this.extra_required_participant_types = newExtraRequiredParticipantTypes;

		var newParticipants = new Participants(null, { event: this.model });
		this.participants.each(function(part) {    
			var attributes = _.clone(part.attributes);
			delete(attributes['id']);
			newParticipants.add(new Participant(attributes), { silent: true });
		});
		this.participants = newParticipants;
	},

	save: function(cb) {
		var that = this;
		var saveParticipants = function() {
			var after = new cbCounter(2, cb);
			if (app.hasAnyAuth({
				auth_type: 'Participants',
				location: that.model.get('location'),
				event_kind: that.model.get('event_kind'),
				event_type: that.model.get('event_type'),
				write: true,
			})) {
				that.participants.save(after.call);
				that.extra_required_participant_types.save(after.call);
			} else if (app.hasAnyAuth({
				auth_type: 'Attend',
				location: that.model.get('location'),
				event_kind: that.model.get('event_kind'),
				event_type: that.model.get('event_type'),
			})) {
				that.participants.save(after.call);
				that.extra_required_participant_types.save(after.call);
			} else {
				cb();
			}
		};
		if (app.hasAuth({
			auth_type: 'Events', 
			location: that.model.get('location'),
			event_kind: that.model.get('event_kind'), 
			event_type: that.model.get('event_type'), 
			write: true,
		})) {
			that.model.save(null, {
				success: function() {
					saveParticipants();
				},
			});
		} else {
			saveParticipants();
		}
	},

	set_default_minutes: function() {
		var event_type = this.event_types.get(this.model.get('event_type'));
		if (event_type != null) {
			if (event_type.get('default_minutes') != 0 && (this.model.get('end').getTime() - this.model.get('start').getTime()) / (1000 * 60) != event_type.get('default_minutes')) {
			  this.model.set('end', new Date(this.model.get('start').getTime() + (event_type.get('default_minutes') * 1000 * 60)));
			}
		}
	},

	check_default_minutes: function() {
		var event_type = this.event_types.get(this.model.get('event_type'));
		if (event_type != null) {
			if (event_type.get('default_minutes') != 0 && (this.model.get('end').getTime() - this.model.get('start').getTime()) / (1000 * 60) != event_type.get('default_minutes')) {
				this.$('#event_default_minutes_warning').removeClass('hidden');
			} else {
				this.$('#event_default_minutes_warning').addClass('hidden');
			}
		}
	},

	set_start: function(event) {
		var old_start = this.model.get('start');
		var new_start = anyTimeConverter.parse($(event.target).val());
		this.model.set('start', new_start, { silent: true });
		this.$('#event_end').val(anyTimeConverter.format(this.model.get('end')));
		this.model.set('end', new Date(this.model.get('end').getTime() + new_start.getTime() - old_start.getTime()), { silent: true });
		this.check_default_minutes();
		this.update_weekdays();
		this.busy_meter.save({
		  start: this.model.get('start'),
			end: this.model.get('end'),
		});
		this.unique_meter.refresh();
	},

	update_weekdays: function() {
		this.$('#start_day').text({{.I "day_names" }}[this.model.get('start').getDay()]);
		this.$('#end_day').text({{.I "day_names" }}[this.model.get('end').getDay()]);
	},

	set_end: function(event) {
		this.model.set('end', anyTimeConverter.parse($(event.target).val()), { silent: true });
		this.check_default_minutes();
		this.update_weekdays();
		this.busy_meter.save({
		  start: this.model.get('start'),
			end: this.model.get('end'),
		});
		this.unique_meter.refresh();
	},

	set_title: function(event) {
		this.model.set('title', $(event.target).val());
	},

	set_information: function(event) {
		this.model.set('information', $(event.target).val());
	},

	set_location: function(event) {
		event.preventDefault();
		var location_id = $(event.target).attr('data-location-id');
		if (location_id != '') {
			this.model.set('location', location_id, { silent: true });
		} else {
			this.model.set('location', null, { silent: true });
		}
		if (!this.model.valid({
			event_types: this.event_types,
		})) {
			this.model.set('event_kind', null, { silent: true });
			this.model.set('event_type', null, { silent: true });
		}
		this.model.trigger('change');
	},

	set_event_kind: function(event) {
		event.preventDefault();
		var kind_id = $(event.target).attr('data-event-kind-id');
		if (kind_id != '') {
			this.model.set('event_kind', kind_id, { silent: true });
		} else {
			this.model.set('event_kind', null, { silent: true });
		}
		this.model.set('event_type', null, { silent: true });
		this.model.trigger('change');
	},

	set_event_type: function(event) {
		event.preventDefault();
		var type_id = $(event.target).attr('data-event-type-id');
		if (type_id != '') {
			this.model.set('event_type', type_id);
			this.required_participant_types.fetch({ reset: true });
			this.set_default_minutes();
		} else {
			this.model.set('event_type', null);
		}
		this.unique_meter.refresh();
	},
	
	availableLocations: function() {
		return this.locations.filter(function(location) {
			return app.hasAnyAuth({
				auth_type: 'Events',
				location: location.get('id'),
				write: true,
			});
		});
	},

	availableKinds: function() {
	  var that = this;
		return this.event_kinds.filter(function(event_kind) {
			return that.model.valid_event_kind(event_kind, {
				event_types: that.event_types,
			});
		});
	},

	availableTypes: function() {
	  var that = this;
		return this.event_types.filter(function(event_type) {
			return that.model.valid_event_type(event_type);
		});
	},

	render: function() {
		var theLocs = this.availableLocations();
		var zLocation = new Location({ name: '{{.I "None" }}' });
		if (theLocs.length == 1) {
		  zLocation = theLocs[0];
		} else if (theLocs.length > 1) {
		  for (var i = 0; i < theLocs.length; i++) {
			  if (theLocs[i].get('id') == app.user.get('default_location')) {
				  zLocation = theLocs[i];
				}
			}
		}
		if (this.model.get('location') != null) {
			zLocation = new Location({ name: this.model.get('location_name'), id: this.model.get('location') });
		} else {
		  this.model.set('location', zLocation.get('id'), { silent: true });
		}

		var theKinds = this.availableKinds();
		var zKind = new EventKind({ name: '{{.I "None" }}' });
		if (theKinds.length == 1) {
		  zKind = theKinds[0];
		} else if (theKinds.length > 1) {
		  for (var i = 0; i < theKinds.length; i++) {
			  if (theKinds[i].get('id') == app.user.get('default_event_kind')) {
				  zKind = theKinds[i];
				}
			}
		}
		if (this.model.get('event_kind') != null) {
			zKind = new EventKind({ name: this.model.get('event_kind_name'), id: this.model.get('event_kind') });
		} else {
		  this.model.set('event_kind', zKind.get('id'), { silent: true });
		}

		var theTypes = this.availableTypes();
		var zType = new EventType({ name: '{{.I "None" }}' });
		if (theTypes.length == 1) {
		  zType = theTypes[0];
		} else if (theTypes.length > 1) {
		  for (var i = 0; i < theTypes.length; i++) {
			  if (theTypes[i].get('id') == app.user.get('default_event_type')) {
				  zType = theTypes[i];
				}
			}
		}
		if (this.model.get('event_type') != null) {
			zType = new EventType({ name: this.model.get('event_type_name'), id: this.model.get('event_type'), default_minutes: 0 });
		} else {
		  this.model.set('event_type', zType.get('id'), { silent: true });
			this.set_default_minutes();
		}

		var selected_location = this.locations.get(this.model.get('location')) || zLocation;
		var selected_event_kind = this.event_kinds.get(this.model.get('event_kind')) || zKind;
		var selected_event_type = this.event_types.get(this.model.get('event_type')) || zType;

		var write_auth = app.hasAuth({
			auth_type: 'Events',
			write: true,
			location: this.model.get('location'),
			event_kind: this.model.get('event_kind'),
			event_type: this.model.get('event_type'),
		});

		// render basic view
		this.$el.html(this.template({
			model: this.model,
			location: selected_location,
			event_type: selected_event_type,
			event_kind: selected_event_kind,
			write_auth: write_auth,
			default_minutes: selected_event_type.get('default_minutes'),
		})); 
		this.update_weekdays();
		this.check_default_minutes();
		var that = this;

		// render location alternatives
		_.each(theLocs, function(location) {
			that.$("#available_locations").append(new AvailableLocationView({ model: location }).render().el);
		});
		if (theLocs.length == 0) {
			that.$("#available_locations_dropdown").addClass('disabled');
		}

		// render kind alternatives
		_.each(theKinds, function(event_kind) {
			that.$("#available_event_kinds").append(new AvailableEventKindView({ model: event_kind }).render().el);
		});
		if (theKinds.length == 0) {
			that.$("#available_kinds_dropdown").addClass('disabled');
		}

		// render type alternatives
		_.each(theTypes, function(event_type) {
			that.$("#available_event_types").append(new AvailableEventTypeView({ model: event_type }).render().el);
		}); 
		if (theTypes.length == 0) {
			that.$("#available_types_dropdown").addClass('disabled');
		}

		// recurrence parts
		if (this.model.get('recurring')) {
			var subview = new RecurrenceOptionsView({
				model: this.model,
			}).render();
			this.$('#recurrence_options').append(subview.el);
		}

		// render participant parts
		if (app.hasAnyAuth({
			auth_type: 'Participants',
			location: this.model.get('location'),
			event_kind: this.model.get('event_kind'),
			event_type: this.model.get('event_type'),
		}) || app.hasAnyAuth({
			auth_type: 'Attend',
			location: this.model.get('location'),
			event_kind: this.model.get('event_kind'),
			event_type: this.model.get('event_type'),
		})) {
			new ParticipantsView({
				el: this.$("#participants"),
				event_opener: this.event_opener,
				busy_meter: this.busy_meter,
				extra_required_participant_types: this.extra_required_participant_types,
				participants: this.participants,
				event_types: this.event_types,
				available_participant_types: this.available_participant_types,
				required_participant_types: this.required_participant_types,
				event: this.model,
			}).render();
		}

		// render change log
		if (app.hasAuth({
			auth_type: 'Events',
			location: this.model.get('location'),
			event_kind: this.model.get('event_kind'),
			event_type: this.model.get('event_type'),
			write: true,
		})) {
			new EventChangesView({
				el: this.$('#changes'),
				collection: this.changes,
			}).render();
		}

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
				format: '{{.I "any_time_format_no_seconds" }}',
			};
			that.$('#event_start').AnyTime_noPicker().AnyTime_picker(options);
			that.$('#event_end').AnyTime_noPicker().AnyTime_picker(options);
			that.$('#event_start').removeAttr("readonly");
			that.$('#event_end').removeAttr("readonly");
		}, 500);
		if (this.model.get('information') != null) {
			var wantedInfoRows = this.model.get('information').split(/\n/).length;
			if (wantedInfoRows > 3) {
				that.$('#event_information').attr('rows', wantedInfoRows);
			}
		}
		return this;
	},

});
