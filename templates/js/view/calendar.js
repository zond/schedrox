window.CalendarView = Backbone.View.extend({

	template: _.template($('#calendar_underscore').html()),

	initialize: function(opts) {
		_.bindAll(this, 'cleanup', 'open_event', 'selected', 'fetchEvents', 'refetchAll', 'fetchMeta', 'render_event', 'rerender', 'set_date', 'deliverEvents', 'toggleMenu');
		this.start = new Date();
		this.end = new Date();
		this.fetching = {};
		this.minTime = "00:00"
		this.maxTime = "24:00"
		this.timeAdd = 0;
		this.firstDay = null;
		this.currentView = 'agendaWeek';
		this.date = new Date();
		var loadedUnixtime = JSON.parse(window.localStorage.getItem('calendar-unixtime'));
		if (loadedUnixtime != null) {
			this.date = new Date(loadedUnixtime);
		}
		this.setDefaultLimits();
		this.events = [];
		this.show_event = opts.show_event;
		this.event_types = new EventTypes();
		this.event_kinds = new EventKinds();
		this.participant_types = new ParticipantTypes();
		this.locations = new Locations();
		this.custom_filters = new CustomFilters();
		this.usercache = {};
		this.current_filter_view = new CurrentFilterView({
			usercache: this.usercache,
			event_types: this.event_types,
			event_kinds: this.event_kinds,
			locations: this.locations,
			model: new CustomFilter({
				locations: [],
				kinds: [],
				types: [],
			  users: [],
			}),
			refetch: this.fetchEvents,
			collection: this.custom_filters,
		});
		this.current_filter_view.model.loadFromLocalStorage();
		this.custom_filters_view = new CustomFiltersView({
			collection: this.custom_filters,
			model: this.current_filter_view.model,
		});
		if (app.getDomain() != null) {
			var that = this;
			_.each([this.locations, this.event_kinds, this.event_types], function(coll) {
				coll.bind("change", that.rerender);
				coll.bind("reset", that.rerender);
				coll.bind("add", that.rerender);
				coll.bind("remove", that.rerender);
			});
		}
		app.on('domainchange', this.refetchAll);
		this.showMenu = true;
		window.addEventListener('shake', this.toggleMenu, false);
		this.popups = {};
		this.fetchMeta();
	},

	toggleMenu: function() {
		this.showMenu = !this.showMenu;
		if (this.showMenu) {
			$('#menu').show();
			$('.current-filter').show();
			$('.fc-header-title').show();
			$('.fc-header-right').show();
			$('.back-to-today').parent().show();
			$('.fc-content').css('margin-top', '90px');
      $('#content').css('margin-top', '0px');
		} else {
			$('#menu').hide();
			$('.current-filter').hide();
			$('.fc-header-title').hide();
			$('.fc-header-right').hide();
			$('.back-to-today').parent().hide();
			$('.fc-content').css('margin-top', '0px');
      $('#content').css('margin-top', '3.5em');
		}
	},

	cleanup: function() {
		app.off('domainchange', this.refetchAll);
		window.removeEventListener('shake', this.toggleMenu, false);
	},

	cleanPopups: function() {
		_.each(this.popups, function(value, key) {
			value.remove();
		});
		this.popups = {};
		this.$('.tooltip.top.fade.in').remove();
		this.$('.popup-container').remove();
	},

	refetchAll: function() {
		this.current_filter_view.model.loadFromLocalStorage();
	  this.fetchMeta();
		this.fetchEvents();
	},

	rerender: function() {
		this.$('#calendar').fullCalendar('rerenderEvents');
	},

	fetchMeta: function() {
		if (app.getDomain() != null) {
			_.each([this.locations, this.event_kinds, this.event_types, this.participant_types, this.custom_filters], function(coll) {
				coll.fetch({ reset: true });
			});
		}
	},

	render: function() {
	  var that = this;
	  var width = '100';
		var height = 100;
		var daysBack = 0;
		if (app.user != null) {
		  if (app.user.get('calendar_width') != null && app.user.get('calendar_width') != 0) {
				width = app.user.get('calendar_width');
			}
			if (app.user.get('calendar_days_back') != null && app.user.get('calendar_days_back') != 0) {
			  daysBack = app.user.get('calendar_days_back');
			}
			if (app.user.get('calendar_height') != null && app.user.get('calendar_height') != 0) {
				height = app.user.get('calendar_height');
			}
		}
		that.$el.html(that.template({
		  width: width,
		}));
		var options = {
		  year: this.date.getFullYear(),
			month: this.date.getMonth(),
			date: this.date.getDate(),
			header: {
				left: 'prev,next',
				center: 'title',
				right: 'agendaDay,agendaWeek,month',
			},
			columnFormat: {
				month: 'ddd',
				week: '{{.I "week_column_format" }}',
				day: '{{.I "day_column_format" }}',
			},
			timeFormat: {
				agenda: {{.I "agendaTimeFormat"}}, 
				'': {{.I "timeFormat"}},          
			},
			buttonText: {
				prev:     '&nbsp;&#9668;&nbsp;',  // left triangle
				next:     '&nbsp;&#9658;&nbsp;',  // right triangle
				prevYear: '&nbsp;&lt;&lt;&nbsp;', // <<
				nextYear: '&nbsp;&gt;&gt;&nbsp;', // >>
				today:    '{{.I "today"}}',
				month:    '{{.I "month"}}',
				week:     '{{.I "week"}}',
				day:      '{{.I "day"}}'
			},
			minTime: this.minTime,
			slotEventOverlap: false,
			maxTime: this.maxTime,
			allDayText: '{{.I "all-day (in calendar)"}}',
			axisFormat: {{.I "axisFormat"}},
			monthNames: {{.I "month_names" }},
			snapMinutes: 15,
			monthNamesShort: {{.I "month_names_short" }},
			dayNames: {{.I "day_names" }},
			dayNamesShort: {{.I "day_names_short" }},
			eventClick: that.open_event,
			firstDay: this.firstDay || ((new Date().getDay() + (7 - daysBack)) % 7),
			weekNumbers: true,
			weekNumberTitle: '{{.I "week_number_title" }}',
			selectHelper: true,
			selectable: app.hasAnyAuth({
				auth_type: "Events",
				write: true,
			}),
			events: that.deliverEvents,
			unselectAuto: false,
			defaultView: that.currentView,
			ignoreTimezone: true,
			aspectRatio: 1.5 / (height / 100),
			select: that.selected,
			eventAfterRender: that.render_event,
			eventAfterAllRender: that.set_date,
		};
		that.$('#calendar').fullCalendar(options);
		that.current_filter_view.render();
		that.$el.append(that.current_filter_view.el);
		that.custom_filters_view.render();
		that.$('.fc-header-center').append(that.custom_filters_view.el);
		that.$('.fc-header-left').append(new QuickNavView({
			parent: that,
		}).render().el);
		if (that.show_event != null) {
			that.open_event(that.show_event.attributes);
			that.show_event = null;
		}
		for (var i = 0; i < 48; i++) {
		  var slot = that.$('.fc-slot' + i);
			var t = new Date(((i / 2.0) * 3600000) + this.timeAdd);
			slot.find('td').attr('title', $.fullCalendar.formatDate(t, {{.I "fullTimeFormat" }}));
			slot.bind('mouseover', function(ev) {
			  $(ev.target).closest('tr').find('.fc-agenda-axis').addClass('highlight-axis');
			});
			slot.bind('mouseout', function(ev) {
			  $(ev.target).closest('tr').find('.fc-agenda-axis').removeClass('highlight-axis');
			});
		}
		return that;
	},

	set_date: function() {
	  this.date = this.$('#calendar').fullCalendar('getDate');
		window.localStorage.setItem('calendar-unixtime', JSON.stringify(this.date.getTime()));
	},

	setDefaultLimits: function() {
		if (app.getDomain() != null) {
			this.minTime = $.fullCalendar.formatDate(app.getDomain().get('earliest_event'), 'HH:mm');
			this.timeAdd = app.getDomain().get('earliest_event').getTime();
			if (app.getDomain().get('latest_event').getHours() > 0) {
				this.maxTime = $.fullCalendar.formatDate(app.getDomain().get('latest_event'), 'HH:mm');
			}
		}
	},

	deliverEvents: function(start, end, cb) {
		this.set_date();
	  this.start = start;
		this.end = end;
		this.cleanPopups();
	  cb(this.events);
		this.fetchEvents()
	},

	fetchEvents: function() {
		var that = this;
		var start = that.start;
		var end = that.end;
		var fetchKey = ('' + start + '-' + end);
		if (that.fetching[fetchKey] == null) {
			that.fetching[fetchKey] = true;
			var url = '/events?start=' + (start.getTime() / 1000) + '&end=' + (end.getTime() / 1000) + '&';
			var filters = [];
			_.each(this.current_filter_view.model.get('locations'), function(loc) {
				filters.push('locations=' + loc);
			});
			_.each(this.current_filter_view.model.get('kinds'), function(loc) {
				filters.push('kinds=' + loc);
			});
			_.each(this.current_filter_view.model.get('types'), function(loc) {
				filters.push('types=' + loc);
			});
			_.each(this.current_filter_view.model.get('users'), function(usr) {
				if (usr != '') {
					filters.push('users=' + usr);
				}
			});
			var oldMinTime = that.minTime;
			var oldMaxTime = that.maxTime;
			this.currentView = this.$('#calendar').fullCalendar('getView').name;
			that.setDefaultLimits();
			$.ajax(url + filters.join('&'), {
				dataType: 'json',
				success: function(data) {
					if (that.start == start && that.end == end) {
						that.events = data;
						_.each(data, function(ev) {
							if (!ev.allDay) {
								var startTime = $.fullCalendar.formatDate($.fullCalendar.parseDate(ev.start), 'HH:mm');
								var endTime = $.fullCalendar.formatDate($.fullCalendar.parseDate(ev.end), 'HH:mm');
								if (startTime < that.minTime) {
									that.minTime = startTime;
								}
								if (startTime > that.maxTime) {
									that.maxTime = startTime;
								}
								if (endTime < that.minTime) {
									that.minTime = endTime;
								}
								if (endTime > that.maxTime) {
									that.maxTime = endTime;
								}
							}
						});
						if (that.minTime != oldMinTime || that.maxTime != oldMaxTime) {
							that.render();
						} else {
							that.$('#calendar').fullCalendar('refetchEvents');
						}
					}
					delete(that.fetching[fetchKey]);
				},
			});
		}
	},

	render_event: function(ev, el, view) {
		if (ev.end == null && ev.start != null) {
			ev.end = ev.start;
		}
		var minHeight = 15;
    if (parseInt($(el).css('height')) < minHeight) {
			$(el).css('height', minHeight + 'px');
		}
		if (ev._start != null && ev._end == null) {
			$(el).css('height', minHeight + 'px');
		}
		var that = this;
		var viewName = this.$('#calendar').fullCalendar('getView').name;
		var elName = '';
		var tagName = '';
		var text = [];
		if (viewName == 'month') {
			tagName = 'span';
			if (ev.allDay) {
				elName = '.fc-event-inner';
			} else {
				elName = '.fc-event-time';
			}
			if (ev.title == null || ev.title == '') {
				$(el).find('.fc-event-title').remove();
			}
		} else {
			tagName = 'div';
			elName = '.fc-event-inner';
			if (ev.title == null || ev.title == '') {
				$(el).find('.fc-event-title').remove();
			}
		}
		var type = this.event_types.get(ev.event_type);
		var kind = this.event_kinds.get(ev.event_kind);
		var color_chosen = parseColor('#3a87ad');
		if (type != null && /^#[0-9a-fA-F]{6,8}$/.exec(type.get('color')) != null) {
		  color_chosen = parseColor(type.get('color'));
			$(el).css('background-color', type.get('color'));
		} else {
			if (kind != null && /^#[0-9a-fA-F]{6,8}$/.exec(kind.get('color')) != null) {
				color_chosen = parseColor(kind.get('color'));
				$(el).css('background-color', kind.get('color'));
			}
		}
		if (luminosity(color_chosen) > 60) {
			$(el).find('.fc-event-inner').css('color', 'black');
		}
		if (ev.event_type_name != null && (type == null || !type.get('name_hidden_in_calendar'))) {
		  text.push(ev.event_type_name);
		}
		if (this.locations.length > 1 && ev.location_name != null) {
			text.push(ev.location_name);
		}
		if (ev.participants != null && type != null && type.get('display_users_in_calendar')) {
		  _.each(ev.participants, function(participant) {
			  if (participant.user != null) {
					text.push('{{.I "name_order" }}'.format(participant.given_name, participant.family_name));
				}
			});
		}
		if (
			(ev.user_participants != null && ev.wanted_user_participants != null && ev.contact_participants != null && ev.allowed_contact_participants != null && ev.required_contact_participants != null) && 
			(ev.user_participants > 0 || ev.wanted_user_participants > 0 || ev.contact_participants > 0 || ev.allowed_contact_participants > 0 || ev.required_contact_participants > 0)) {
			var form = eventParticipantsDefaultFormat;
			if (type != null && type.get('participants_format') != null && type.get('participants_format') != '') {
			  form = type.get('participants_format');
			}
			text.push(form.format(ev.contact_participants, ev.required_contact_participants, ev.allowed_contact_participants, ev.user_participants, ev.wanted_user_participants, ev.required_user_participants, ev.allowed_user_participants));
		}

		var contacts_state = '';
		if (ev.contact_participants != null && ev.required_contact_participants != null && ev.contact_participants < ev.required_contact_participants) {
			contacts_state = 'fc';
		} else if (ev.contact_participants != null && ev.allowed_contact_participants != null && ev.contact_participants < ev.allowed_contact_participants) {
			contacts_state = 'mc';
		}
		var users_state = '';
		if (ev.user_participants != null && ev.required_user_participants != null && ev.user_participants < ev.required_user_participants) {
			users_state = 'fu';
		} else if (ev.user_participants != null && ev.allowed_user_participants != null && ev.user_participants < ev.allowed_user_participants) {
			users_state = 'mu';
		}
		var event_type = this.event_types.get(ev.event_type);
		if (event_type != null) {
			if (!event_type.get('signal_colors_when_0_contacts') && ev.contact_participants == 0) {
				users_state = '';
				contacts_state = '';
			}
			if (!event_type.get('signal_colors_when_more_possible_contacts') && contacts_state == 'mc') {
				contacts_state = '';
			}
			if (!event_type.get('signal_colors_when_more_possible_users') && users_state == 'mu') {
				users_state = '';
			}
		}
		if (users_state != '' || contacts_state != '') {
			$(el).find('.fc-event-time').addClass('event-attention-' + users_state + contacts_state);
		}

		var messages = [];
		if (ev.participants != null) {
			for (var i = 0; i < ev.participants.length; i++) {
				if (ev.participants[i].user != null) {
					var part = new Participant(ev.participants[i]);
					messages.push(part.get('participant_type_name') + ': ' + part.name());
				}
			}
		}
		if (users_state == 'fu') {
			messages.push('<span style="color: red;">{{.I "Not enough users." }}</span>');
		} else if (users_state == 'mu') {
			messages.push('<span style="color: blue;">{{.I "More users allowed." }}</span>');
		}
		if (contacts_state == 'fc') {
			messages.push('<span style="color: orange;">{{.I "Not enough contacts." }}</span>');
		} else if (contacts_state == 'mc') {
			messages.push('<span style="color: green;">{{.I "More contacts allowed." }}</span>');
		}

		if (messages.length > 0) {
			var time_container = $(el).find(elName);
			var popup_container = $('<div class="popup-container"></div>');
			popup_container.css('position', 'absolute');
			popup_container.css('z-index', '1001');
			var offset = time_container.offset();
			popup_container.css('left', offset.left + 'px');
			popup_container.css('top', offset.top + 'px');
			popup_container.css('width', time_container.width() + 'px');
			popup_container.css('height', time_container.height() + 'px');
			that.$el.append(popup_container);
			popup_container.tooltip({
				placement: 'top',
				title: messages.join('<br/>'),
				html: true,
			});
			popup_container.on('click', function(e) {
				that.open_event(ev);
			});
			that.popups[ev.id] = popup_container;
		}
		if (event_type != null && event_type.get('title_size') != null && event_type.get('title_size') != 0) {
		  $(el).find('.fc-event-title').css('font-size', '' + event_type.get('title_size') + '%');
		  $(el).find('.participant-name').css('font-size', '' + event_type.get('title_size') + '%');
		}
		if (text.length > 0) {
			if (!ev.allDay && (ev.end - ev.start)/(1000*60) < 60) {
				var timeEl = $(el).find(elName + ' .fc-event-time');
				// right, fullcalendar stops displaying the end time for short events. Don't know why.
				var fStart = $.fullCalendar.formatDate(ev.start, {{.I "fullTimeFormat" }})
				var fEnd = $.fullCalendar.formatDate(ev.end, {{.I "fullTimeFormat" }})
				timeEl.html(fStart + ' - ' + fEnd + ", " + text.join(", "));
			} else {
				$(el).find(elName).append('<' + tagName + '>' + text.join(", ") + '</' + tagName + '>');
			}
		}
		if (ev.information != '') {
		  var curr = $(el).find('.fc-event-time').html();
			$(el).find('.fc-event-time').html(curr + ' <strong>(!)</strong>');
		}
	},

	open_event: function(ev) {
		this.$('.tooltip.top.fade.in').remove();
		var that = this;
		new EventView({
			participant_types: that.participant_types,
			event_types: that.event_types,
			event_kinds: that.event_kinds,
			locations: that.locations,
			model: new Event(ev),
		}).modal(function() {
			app.navigate('/calendar');
			that.fetchEvents();
		});
	},

	selected: function(startDate, endDate, allDay, jsEvent, view) {
		this.$('.tooltip.top.fade.in').remove();
		var that = this;
		if (app.hasAnyAuth({
			auth_type: "Events",
			write: true,
		}) && app.getDomain() != null) {
			var blocker = null;
			_.each(that.events, function(ev) {
				var kind = that.event_kinds.get(ev.event_kind);
				if (kind != null && kind.get('block')) {
					if (overlaps(startDate, endDate, ev.start, ev.end)) {
						blocker = {};
						if (ev.title != null && ev.title != '') {
							blocker.title = ev.title;
						} else if (ev.event_type_name != null && ev.event_type_name != '') {
							blocker.title = ev.event_type_name;
						} else if (ev.event_kind_name != null && ev.event_kind_name != '') {
							blocker.title = ev.event_kind_name;
						} else {
							blocker.title = anyTimeConverter.format(ev.start) + '-' + anyTimeConverter.format(ev.end);
						}
						if (ev.location_name != null && ev.location_name != '') {
							blocker.location = ev.location_name;
						} else {
							blocker.location = '{{.I "unknown location" }}';
						}
					}
				}
			});
			var openView = function() {
				var ev = new Event({
					allDay: allDay,
					title: '',
					start: startDate,
					end: endDate,
					recurrence_end: endDate,
				});
				new EventView({
					participant_types: that.participant_types,
					event_types: that.event_types,
					event_kinds: that.event_kinds,
					locations: that.locations,
					model: ev,
				}).modal(function() {
					that.fetchEvents();
					that.$("#calendar").fullCalendar('unselect');
				});
			};
			if (blocker != null) {
				myconfirm("{{.I "This interval is blocked by {0} at {1}. Still create an event?" }}".format(blocker.title, blocker.location), openView);
			} else {
				openView();
			}
		}
	},

});
