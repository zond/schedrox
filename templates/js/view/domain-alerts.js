window.DomainAlertsView = Backbone.View.extend({

  template: _.template($('#domain_alerts_underscore').html()),

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.alerts = options.alerts;
    this.domains = options.domains;
  },

  render: function() {
    this.$el.html(this.template({ }));
    for (var dom in this.alerts) {
      var events = this.alerts[dom];
      var domain_name = '';
      for (var i = 0; i < this.domains.length; i++) {
        if (this.domains[i].id == dom) {
	  domain_name = this.domains[i].name;
	}
      }
      for (var i = 0; i < events.length; i++) {
	var event = events[i];
	this.$el.append(new DomainAlertView({ 
	  domain_name: domain_name,
	  event: event,
	}).render().el);
      }
    }
    return this;
  },

});
