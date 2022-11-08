window.UserSalariesView = Backbone.View.extend({

  template: _.template($('#user_salaries_underscore').html()),

	events: {
	  "click .earlier": "moveBack",
		"click .later": "moveForward",
		"click .attest": "openAttest",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
		this.from = lastSalaryBreakpointBefore(today());
		this.to = firstSalaryBreakpointAfter(this.from);
		this.userOpener = options.userOpener;
  },

	openAttest: function(event) {
		event.preventDefault();
		var that = this;
		new UserAttestView({ 
			model: that.model, 
			from: that.from,
			to: that.to,
		}).modal(function() {
      that.userOpener();
		});
	},

	moveBack: function(ev) {
	  this.from = lastSalaryBreakpointBefore(this.from);
		this.to = firstSalaryBreakpointAfter(this.from);
		this.render();
	},

  moveForward: function(ev) {
	  this.to = firstSalaryBreakpointAfter(this.to);
		this.from = lastSalaryBreakpointBefore(this.to);
		this.render();
	},

  render: function() {
    this.$el.html(this.template({
		  from: this.from,
			to: this.to,
			hasForward: this.to.getTime() <= new Date().getTime(),
		}));
		this.delegateEvents();
    return this;
  },

});
