
window.clobberEventIds = true;

window.Event = Backbone.Model.extend({
	urlRoot: "/events",
	initialize: function(data) {
		this.set(this.parse(data));
	},
	parse: function(data) {
		data.start = this.convertDate(data.start);
		if (data.allDay == true && data.end == null) {
			data.end = data.start;
		}
		data.end = this.convertDate(data.end);
		this.original_start = data.start;
		this.original_end = data.end;
		if (data.recurrence_end != null) {
			data.recurrence_end = this.convertDate(data.recurrence_end);
		}
		if (data.recurrence_master_start != null) {
			data.recurrence_master_start = this.convertDate(data.recurrence_master_start);
		}
		if (data.recurrence_master_end != null) {
			data.recurrence_master_end = this.convertDate(data.recurrence_master_end);
		}
		if (window.clobberEventIds && data['recurrence_master'] != null) {
			data['id'] = data['recurrence_master'];
		}
		return data;
	},
	add_exception: function(except) {
		var current = (this.get('recurrence_exceptions') || '').split(/,/);
		var year = '' + except.getFullYear();
		while (year.length < 4) {
			year = '0' + year;
		}
		var month = '' + (except.getMonth() + 1);
		while (month.length < 2) {
			month = '0' + month;
		}
		var day = '' + except.getDate();
		while (day.length < 2) {
			day = '0' + day;
		}
		current.push(year + month + day);
		this.set('recurrence_exceptions', current.join(','));
	},
	// Don't even ask how this works, pure trial and error. Timezones, daylight savings, date conversions
	set_to_recurrence_master_times: function() {
		this.attributes.start.setTime(this.attributes.recurrence_master_start.getTime() + this.get('start').getTime() - this.original_start.getTime());
		this.attributes.end.setTime(this.attributes.recurrence_master_end.getTime() + this.get('end').getTime() - this.original_end.getTime());
	},
	convertDate: function(d) {
		if (typeof(d) == 'object') {
			return d;
		} else if (typeof(d) == 'string') {
			return new Date(Date.parse(d));
		} else if (d == null) {
			return null;
		} else {
			throw "Unknown date type for " + d
		}
	},
	split_recurrence: function(cb) {
		var that = this;
		var my_attributes = _.clone(that.attributes);
		$.ajax(that.url() + '/splits', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify({
				at: this.get('start'),
			}),
			success: function(data) {
				my_attributes.recurring = data.recurring;
				my_attributes.recurrence = data.recurrence
				my_attributes.id = data.id;
				delete(my_attributes.recurrence_master);
				cb(new Event(my_attributes));
			},
		});
	},
	create_exception: function(cb) {
		var that = this;
		var my_attributes = _.clone(that.attributes);
		$.ajax(that.url() + '/exceptions', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify({
				start: this.get('start'),
				end: this.get('end'),
			}),
			success: function(data) {
				my_attributes.recurring = data.recurring;
				my_attributes.recurrence = data.recurrence
				my_attributes.id = data.id;
				delete(my_attributes.recurrence_master);
				cb(new Event(my_attributes));
			},
		});
	},
	describe: function(event_types) {
		var rval = [];
		if (this.get('title') != null && this.get('title') != '') {
			rval.push(this.get('title'));
		}
		var type = event_types.get(this.get('event_type'));
		if (type != null) {
			rval.push(type.get('name'));
		}
		return rval.join('/');
	},
	validTimes: function() {
		if (this.get('allDay')) {
			return true;
		} else {
			if (app.getDomain().get('latest_event').getHours() > 0) {
				return betweenDayTimes(app.getDomain().get('earliest_event'), app.getDomain().get('latest_event'), this.get('start')) && betweenDayTimes(app.getDomain().get('earliest_event'), app.getDomain().get('latest_event'), this.get('end'));
			} else {
				return !beforeDayTime(this.get('start'), app.getDomain().get('earliest_event')) && !beforeDayTime(this.get('end'), app.getDomain().get('earliest_event'));
			}
		}
	},
	recommended: function(options) {
		return this.validTimes();
	},
	valid: function(options) {
		return (
			this.get('location') != null && 
			this.get('event_kind') != null && 
			this.get('event_type') != null && 
			options.event_types.get(this.get('event_type')).get('event_kind') == this.get('event_kind')
		);
	},
	why_not_recommended: function(options) {
		if (this.recommended(options)) {
			return "... strange, there should be no problems";
		} else if (!this.validTimes()) {
			return "it is outside the earliest/latest allowed event times";
		}
	},
	why_invalid: function(options) {
		if (this.valid(options)) {
			return "... strange, it should be";
		} else if (this.get('location') == null) {
			return "it lacks location";
		} else if (this.get('event_kind') == null) {
			return "it lacks kind";
		} else if (this.get('event_type') == null) {
			return "it lacks type";
		} else if (options.event_types.get(this.get('event_type')).get('event_kind') != this.get('event_kind')) {
			return "the type is of a different kind than the event";
		} else if (!app.hasAuth({
			auth_type: 'Events',
			write: true,
			location: this.get('location'),
			event_kind: this.get('event_kind'),
			event_type: this.get('event_type'),
		})) {
			return "you lack authorization to create events like this";
		}
	},
	equalExceptRecurrence: function(other) {
		var myAttributes = _.clone(this.attributes);
		var otherAttributes = _.clone(other.attributes);
		otherAttributes['start'] = otherAttributes['start'].toISOString()
		otherAttributes['end'] = otherAttributes['end'].toISOString()
		myAttributes['start'] = myAttributes['start'].toISOString()
		myAttributes['end'] = myAttributes['end'].toISOString()
		delete(otherAttributes['recurring']);
		delete(otherAttributes['recurrence']);
		delete(otherAttributes['recurrence_end']);
		delete(otherAttributes['recurrence_exceptions']);
		delete(otherAttributes['recurrence_master']);
		delete(otherAttributes['recurrence_master_start']);
		delete(otherAttributes['recurrence_master_end']);
		delete(myAttributes['recurring']);
		delete(myAttributes['recurrence']);
		delete(myAttributes['recurrence_end']);
		delete(myAttributes['recurrence_exceptions']);
		delete(myAttributes['recurrence_master']);
		delete(myAttributes['recurrence_master_start']);
		delete(myAttributes['recurrence_master_end']);
		return JSON.stringify(myAttributes) == JSON.stringify(otherAttributes);
	},
	// options.event_types: available event types
	valid_event_kind: function(kind, options) {
		return kind.valid({
			location: this.get('location'),
			event_types: options.event_types,
		});
	},
	valid_event_type: function(type, options) {
		return type.valid({
			location: this.get('location'),
			event_kind: this.get('event_kind'),
		});
	},
});
