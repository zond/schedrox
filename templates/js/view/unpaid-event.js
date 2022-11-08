window.UnpaidEventView = Backbone.View.extend({

	template: _.template($('#unpaid_event_underscore').html()),

	tagName: 'tr',

  events: {
		"click .participant-paid": "togglePaid",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

  togglePaid: function(ev) {
		var that = this;
	  ev.preventDefault();
		$.ajax("/events/" + that.model.get('reported_event') + "/participants/" + that.model.get('participants')[0].id + "/set_paid",
			{
        type: 'POST',
			  dataType: 'json',
			  data: {
					paid: $(ev.target).is(':checked'),
				},
			  success: function(data) {
					that.model.get('participants')[0] = data;
					that.model.trigger('change');
				},
			});
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({ 
		  model: that.model,
			contact: that.model.get('contact') || {
				'name': '',
        'organization_number': '',
			  'billing_address_line_1': '',
			  'billing_address_line_2': '',
			  'billing_address_line_3': '',
			  'contact_given_name': '',
			  'contact_family_name': '',
			  'reference': '',
			},
		}));
		return that;
	},

});
