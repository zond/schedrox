window.EventChangeView = Backbone.View.extend({

	template: _.template($('#event_change_underscore').html()),

	translations: {{.I "ChangeActions"}},

	events: {
		"click .change-heading": "toggleExpand",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
		this.user = new User({ id: this.model.get('user'), email: this.model.get('user_email') });
		this.expanded = false;
	},

	toggleExpand: function(ev) {
		var that = this;
		ev.preventDefault();
		this.user.fetch({
			success: function() {
				that.expanded = !that.expanded;
				that.render();
			},
			error: function() {
				that.expanded = !that.expanded;
				that.render();
			},
		});
	},

	render: function() {
		this.$el.html(this.template({ 
			user_auth: app.hasAuth({
				auth_type: 'Users',
			}),
			translations: this.translations, 
			model: this.model,
			user: this.user,
			expanded: this.expanded,
		})); 
		return this;
	},

});
