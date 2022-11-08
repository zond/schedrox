window.WeeklyRecurrenceView = Backbone.View.extend({

  template: _.template($('#weekly_recurrence_underscore').html()),

  pattern: /^DOW:(([0-6])(,[0-6])*)?\/([0-9]+)$/,

  format: 'DOW:{0}/{1}',

  def: 'DOW:/1',
  
  name: '{{.I "Weekly"}}',
  
  events: {
    "change .recurring-n-weeks": "changeNWeeks",
    "click .recurring-on": "toggleRecurringOn",
  },

  getRecurringOn: function() {
    var rval = [false, false, false, false, false, false, false];
    var dayBit = this.pattern.exec(this.model.get('recurrence'))[1];
    if (dayBit) {
      var days = dayBit.split(/,/);
      for (var i = 0; i < days.length; i++) {
	rval[parseInt(days[i])] = true;
      }
    }
    return rval;
  },

  toggleRecurringOn: function(ev) {
    if (app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: this.model.get('location'),
      event_kind: this.model.get('event_kind'),
      event_type: this.model.get('event_type'),
    })) {
      var old_n_weeks = this.pattern.exec(this.model.get('recurrence'))[4];
      var to_toggle = parseInt($(ev.target).attr('data-recurring-on'));
      var recurring_on = this.getRecurringOn();
      recurring_on[to_toggle] = !recurring_on[to_toggle];
      var new_list = [];
      for (var i = 0; i < recurring_on.length; i++) {
        if (recurring_on[i]) {
	  new_list.push(i);
	}
      }
      this.model.set('recurrence', this.format.format(new_list.join(','), old_n_weeks), { silent: true });
    }
  },

  changeNWeeks: function(ev) {
    if (app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: this.model.get('location'),
      event_kind: this.model.get('event_kind'),
      event_type: this.model.get('event_type'),
    })) {
      var new_n_weeks = parseInt($(event.target).val());
      if (new_n_weeks > 0) {
        var old_weekdays = this.pattern.exec(this.model.get('recurrence'))[1];
	this.model.set('recurrence', this.format.format(old_weekdays, '' + new_n_weeks), { silent: true });
      } else {
        $(event.target).val("1");
      }
    }
  },

  render: function() {
    var match = this.pattern.exec(this.model.get('recurrence'));
    var n_weeks = parseInt(match[4]);
    var write_auth = app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: this.model.get('location'),
      event_kind: this.model.get('event_kind'),
      event_type: this.model.get('event_type'),
    });
    this.$el.html(this.template({
      n_weeks: n_weeks,
      recurring_on: this.getRecurringOn(),
      write_auth: write_auth,
      model: this.model,
    })); 
    return this;
  },

});
