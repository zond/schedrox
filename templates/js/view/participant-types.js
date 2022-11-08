window.ParticipantTypesView = Backbone.View.extend({

	template: _.template($('#participant_types_underscore').html()),

	events: {
		"change #new_type": "newType",
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'refetch');
		this.show_type = options.show_type;
		this.participant_types = new ParticipantTypes();
		this.participant_types.bind("change", this.render);
		this.participant_types.bind("reset", this.render);
		this.participant_types.bind("add", this.render);
		this.participant_types.bind("remove", this.render);
		this.refetch();
		app.on('domainchange', this.refetch);
	},

	cleanup: function() {
		app.off('domainchange', this.refetch);
	},

	newType: function(ev) {
		if (app.getDomain() != null) {
			var that = this;
			var newParticipantType = new ParticipantType({ name: $(ev.target).val() });
			newParticipantType.save(null, {
				success: function() {
					that.participant_types.add(newParticipantType);
				}
			});
		}
	},

	refetch: function() {
		if (app.getDomain() != null) {
			this.participant_types.fetch({ reset: true });
		}
	},

	render: function() {
		this.$el.html(this.template({}));
		this.participant_types.forEach(function(participant_type) {
			this.$('#participant_type_list').append(new ParticipantTypeView({ 
				model: participant_type,
			}).render().el);
		});
		if (this.show_type != null) {
			new ParticipantTypeDetailsView({ model: this.show_type }).modal(function() {
				app.navigate('/events/participants');
			});
			this.show_type = null;
		}
		return this;
	}

});
