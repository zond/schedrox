window.ExportEventsView = Backbone.View.extend({

	template: _.template($('#export_events_underscore').html()),

	events: {
		"click .available-event-type": "set_event_type",
		"change #report_from": "set_from",
		"change #report_to": "set_to",
	},

	initialize: function() {
		_.bindAll(this, 'render');
		this.from = new Date();
		this.to = new Date();
		this.event_types = new EventTypes();
		this.event_types.fetch({ reset: true });
		this.event_types.bind("reset", this.render);
		this.event_type = new Event({
			name: '{{.I "All" }}',
			id: -1,
		});
	},

	set_from: function(event) {
		this.from = anyDateConverter.parse($(event.target).val());
		this.$('#unix_start').val(this.from.getISOTime() / 1000);
	},

	set_to: function(event) {
		this.to = anyDateConverter.parse($(event.target).val());
		this.$('#unix_to').val(this.to.getISOTime() / 1000);
	},

	set_event_type: function(event) {
		event.preventDefault();
		var type_id = $(event.target).attr('data-event-type-id');
		this.event_type = new Event({
			name: $(event.target).attr('data-event-type-name'),
			id: $(event.target).attr('data-event-type-id'),
		});
		this.$('#event_type_id').val(this.event_type.get('id'));
		this.render();
	},
	
	render: function() {
		var that = this;
		that.$el.html(that.template({ 
			from: that.from,
			to: that.to,
			event_types: that.event_types,
			event_type: that.event_type,
		}));
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
		that.$("#available_event_types").append(new AvailableEventTypeView({ model: new EventType({name: '{{.I "All" }}', id: -1 }) }).render().el);
		that.event_types.each(function(event_type) {
			that.$("#available_event_types").append(new AvailableEventTypeView({ model: event_type }).render().el);
		}); 
		return that;
	},

});
