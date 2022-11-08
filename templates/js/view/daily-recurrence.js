window.DailyRecurrenceView = Backbone.View.extend({

  template: _.template($('#daily_recurrence_underscore').html()),

  pattern: /^DAY:([0-9]+)$/,

  format: 'DAY:{0}',
  
  def: 'DAY:1',
  
  name: '{{.I "Daily" }}',

  events: {
    "change .recurring-n-days": "changeNDays",
  },
  
  changeNDays: function(ev) {
    if (app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: this.model.get('location'),
      event_kind: this.model.get('event_kind'),
      event_type: this.model.get('event_type'),
    })) {
      var new_n_days = parseInt($(event.target).val());
      if (new_n_days > 0) {
	this.model.set('recurrence', this.format.format('' + new_n_days), { silent: true });
      } else {
        $(event.target).val("1");
      }
    }
  },

  render: function() {
    var match = this.pattern.exec(this.model.get('recurrence'));
    var n_days = parseInt(match[1]);
    var write_auth = app.hasAuth({
      auth_type: 'Events',
      write: true,
      location: this.model.get('location'),
      event_kind: this.model.get('event_kind'),
      event_type: this.model.get('event_type'),
    });
    this.$el.html(this.template({
      write_auth: write_auth,
      model: this.model,
      n_days: n_days
    })); 
    return this;
  },

});
