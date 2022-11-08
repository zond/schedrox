window.RecurrenceOptionsView = Backbone.View.extend({

  template: _.template($('#recurrence_options_underscore').html()),

  events: {
    "click .available-recurrence-type": "changeRecurrenceType",
    "change #event_recurrence_end": "setRecurrenceEnd",
  },

  recurrenceTypes: {
    'DAY': function() {
      return DailyRecurrenceView;
    },
    'DOM': function() {
      return MonthlyRecurrenceView;
    },
    'DOW': function() {
      return WeeklyRecurrenceView;
    },
  },

  setRecurrenceEnd: function(event) {
    this.model.set('recurrence_end', anyTimeConverter.parse($(event.target).val()), { silent: true });
  },

  changeRecurrenceType: function(ev) {
    var name = $(ev.target).attr('data-recurrence-type');
    for (var type_name in this.recurrenceTypes) {
      if (name == type_name) {
        var view = new (this.recurrenceTypes[name]())({ model: this.model });
        this.model.set('recurrence', view.def);
	return;
      }
    }
    throw "Unknown recurrence type " + name;
  },

  getRecurrenceView: function() {
    for (var name in this.recurrenceTypes) {
      var view = new (this.recurrenceTypes[name]())({ model: this.model });
      if (view.pattern.exec(this.model.get('recurrence')) != null) {
	return view
      }
    }
    var rval = new DailyRecurrenceView({ model: this.model });
    this.model.set('recurrence', rval.def);
    return rval;
  },

  render: function() {
    var that = this;
    var write_auth = app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: this.model.get('location'),
      event_kind: this.model.get('event_kind'),
      event_type: this.model.get('event_type'),
    });
    var recurrence_view = this.getRecurrenceView();
    this.$el.html(this.template({
      recurrence_view: recurrence_view,
      write_auth: write_auth,
      model: this.model,
    })); 
    this.$('#recurrence_type_options').append(recurrence_view.render().el);
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
		    format: '{{.I "any_time_format" }}',
      };
      that.$('#event_recurrence_end').AnyTime_noPicker().AnyTime_picker(options);
			that.$('#event_recurrence_end').removeAttr("readonly");
    }, 500);
    return this;
  },

});
