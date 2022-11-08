window.UserPropertyForUserView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#user_property_for_user_underscore').html()),

	events: {
		"click .close": "removeProperty",
		"change .user-property-valid-until": "changeValidUntil",
		"change .user-property-assigned-at": "changeAssignedAt",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	changeAssignedAt: function(ev) {
		ev.preventDefault();
		if (anyDateConverter.format(this.model.get('assigned_at')) != $(ev.target).val()) {
			if (app.hasAuth({
				auth_type: 'Users',
				write: true,
			})) {
				this.model.modified = true;
				this.model.attributes['assigned_at'] = anyDateConverter.parse($(ev.target).val());
			}
		}
	},

	changeValidUntil: function(ev) {
		ev.preventDefault();
		if (anyDateConverter.format(this.model.get('valid_until')) != $(ev.target).val()) {
			if (app.hasAuth({
				auth_type: 'Users',
				write: true,
			})) {
				this.model.modified = true;
				this.model.attributes['valid_until'] = anyDateConverter.parse($(ev.target).val());
			}
		}
	},

	removeProperty: function(event) {
		if (app.hasAuth({
			auth_type: 'Users',
			write: true,
		})) {
			this.collection.remove(this.model);
		}
	},

	render: function() {
		this.$el.html(this.template({ 
			model: this.model,
			write_auth: app.hasAuth({
				auth_type: 'Users',
				write: true,
			}),
		}));
		if (this.model.get('valid_until') != null && this.model.get('valid_until').getTime() < new Date().getTime()) {
			this.$el.addClass('warning');
			this.$el.tooltip({
				placement: 'bottom',
				title: '{{.I "This property is no longer valid." }}',
			});
		}
		return this;
	},

});
