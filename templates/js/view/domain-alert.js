window.DomainAlertView = Backbone.View.extend({

  template: _.template($('#domain_alert_underscore').html()),

  className: 'domain-alert',

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.domain_name = options.domain_name;
    this.event = options.event;
  },

  render: function() {
    this.$el.html(this.template({ 
      domain_name: this.domain_name,
      event: this.event,
    }));
    return this;
  },

});
