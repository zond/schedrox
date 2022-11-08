window.MonthlyRecurrenceView = Backbone.View.extend({

  template: _.template($('#monthly_recurrence_underscore').html()),

  pattern: /^DOM:(([0-9]{1,2})(,[0-9]{1,2})*)?\/([0-9]+)$/,

  format: 'DOM:{0}/{1}',
  
  def: 'DOM:1/1',
  
  name: '{{.I "Monthly" }}',

  events: {
    "change .recurring-n-months": "changeNMonths",
    "change #recurring_on": "changeRecurringOn",
  },

  changeRecurringOn: function(ev) {
    var old_n_months = this.pattern.exec(this.model.get('recurrence'))[4];
    this.model.set('recurrence', this.format.format(ev.val.join(','), old_n_months), { silent: true });
  },

  changeNMonths: function(ev) {
    if (app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: this.model.get('location'),
      event_kind: this.model.get('event_kind'),
      event_type: this.model.get('event_type'),
    })) {
      var new_n_months = parseInt($(event.target).val());
      if (new_n_months > 0) {
        var old_dates = this.pattern.exec(this.model.get('recurrence'))[1];
	this.model.set('recurrence', this.format.format(old_dates, '' + new_n_months), { silent: true });
      } else {
        $(event.target).val("1");
      }
    }
  },

  render: function() {
    var match = this.pattern.exec(this.model.get('recurrence'));
    var n_months = parseInt(match[4]);
    var write_auth = app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: this.model.get('location'),
      event_kind: this.model.get('event_kind'),
      event_type: this.model.get('event_type'),
    });
    this.$el.html(this.template({
      n_months: n_months,
      write_auth: write_auth,
      model: this.model,
    })); 
    var that = this;
    this.$('#recurring_on').select2();
    this.$('#recurring_on').select2('val', that.pattern.exec(that.model.get('recurrence'))[1].split(/,/));
    return this;
  },

});
