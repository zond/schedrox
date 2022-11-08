window.AvailableAuthTypeView = Backbone.View.extend({

  tagName: 'li',

  template: _.template($('#available_auth_type_underscore').html()),

	initialize: function(options) {
	  this.auth_type = options.auth_type;
	},

	render: function() {
		this.$el.html(this.template({ auth_type: this.auth_type }));
		return this;
	},

});
