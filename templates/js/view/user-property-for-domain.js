window.UserPropertyForDomainView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#user_property_for_domain_underscore').html()),

	events: {
		"click .close": "removeProperty",
		"change .user-property-days-valid": "changeDaysValid",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
	},

	changeDaysValid: function(ev) {
		var neu = parseInt($(ev.target).val());
		if (neu > -1) {
			this.model.set('days_valid', neu, { silent: true });
		} else {
			$(ev.target).val('1');
		}
		this.model.save();
	},

	removeProperty: function(event) {
		if (app.hasAuth({
			auth_type: 'Domain',
			write: true,
		})) {
			var that = this;
			myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(this.model.get("name").htmlEscape()), function() {
				that.model.destroy();
			});
		}
	},

	render: function() {
		this.$el.html(this.template({ 
			model: this.model,
			write_auth: app.hasAuth({
				auth_type: 'Domain',
				write: true,
			}),
		}));
		return this;
	},

});
