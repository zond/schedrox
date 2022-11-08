window.ParticipantTypeView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#participant_type_underscore').html()),

	events: {
		"click .close": "removeType",
		"click .open-type": "openType",
		"click .is-contact": "setIsContact",
	}, 

	openType: function(event) {
		event.preventDefault();
		new ParticipantTypeDetailsView({ model: this.model }).modal(function() {
			app.navigate('/events/participants');
		});
	},

	removeType: function(event) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var that = this;
			myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(this.model.get("name").htmlEscape()), function() {
				that.model.destroy();
			});
		}
	},

	setIsContact: function(event) {
		event.preventDefault();
		if ($(event.target).attr('data-is-contact') == 'true') {
			this.model.set('is_contact', true);
		} else {
			this.model.set('is_contact', false);
		}
		this.model.save();
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	render: function() {
		this.$el.html(this.template({ 
			model: this.model,
		}));
		return this;
	},

});
