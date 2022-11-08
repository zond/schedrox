window.AttestableEventView = Backbone.View.extend({

	template: _.template($('#attestable_event_underscore').html()),

	tagName: 'tr',

	events: {
	  "click .close": "runDeleter",
	},

  attestable: function() {
		if (this.isAttestable == null) {
    	this.isAttestable = app.hasAuth({
    		auth_type: 'Attest',
    		location: this.model.get('location'), 
    		event_kind: this.model.get('event_kind'),
    		event_type: this.model.get('event_type'),
    		participant_type: this.model.get('salary_attested_participant_type'),
    	});
		}
		return this.isAttestable;
	},

	className: function() {
		var suffix = 'unattestable';
		if (this.attestable()) {
      suffix = (this.model.get('salary_time_reported') ? 'by-user' : 'by-calendar');
		}
	  return 'reported-hours-' + suffix;
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.deleter = options.deleter;
		if (!this.attestable()) {
			this.deleter = 'no';
		}
		this.model.bind("change", this.render);
		this.show_attester = options.show_attester;
	},

	runDeleter: function(ev) {
	  ev.preventDefault();
		if (this.deleter != null) {
		  this.deleter();
		}
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({ 
		  model: that.model,
			deleter: that.deleter,
			duration: hoursMinutesForDates(that.model.get('start'), that.model.get('end')),
			show_attester: that.show_attester,
		}));
		return that;
	},

});
