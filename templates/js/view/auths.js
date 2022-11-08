window.AuthsView = Backbone.View.extend({

	template: _.template($('#auths_underscore').html()),

	events: {
		"click .auth-type": "addAuth",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');

		this.collection.bind("change", this.render);
		this.collection.bind("reset", this.render);
		this.collection.bind("add", this.render);
		this.collection.bind("remove", this.render);

		this.locations = options.locations;
		this.locations.bind("change", this.render);
		this.locations.bind("reset", this.render);
		this.locations.bind("add", this.render);
		this.locations.bind("remove", this.render);

		this.event_types = options.event_types;
		this.event_types.bind("change", this.render);
		this.event_types.bind("reset", this.render);
		this.event_types.bind("add", this.render);
		this.event_types.bind("remove", this.render);

		this.event_kinds = options.event_kinds;
		this.event_kinds.bind("change", this.render);
		this.event_kinds.bind("reset", this.render);
		this.event_kinds.bind("add", this.render);
		this.event_kinds.bind("remove", this.render);

		this.participant_types = options.participant_types;
		this.participant_types.bind("change", this.render);
		this.participant_types.bind("reset", this.render);
		this.participant_types.bind("add", this.render);
		this.participant_types.bind("remove", this.render);
		
		this.roles = options.roles;
		this.roles.bind("change", this.render);
		this.roles.bind("reset", this.render);
		this.roles.bind("add", this.render);
		this.roles.bind("remove", this.render);
	},

	addAuth: function(event) {
		event.preventDefault();
		var authType = app.authTypes[$(event.target).attr('data-auth-type')];
		var newAuth = new Auth({ 
			_auth_type: authType,
			auth_type: authType.name,
			translation: authType.translation,
		});
		this.collection.add(newAuth);
	},

	render: function() {
		var that = this;
		this.$el.html(this.template({}));
		this.collection.forEach(function(auth) {
			that.$("#auth_list").append(new AuthView({ 
				model: auth,
				collection: that.collection,
				locations: that.locations,
				event_types: that.event_types,
				event_kinds: that.event_kinds,
				participant_types: that.participant_types,
				roles: that.roles,
			}).render().el);
		});
		for (var authTypeName in app.authTypes) {
		  if ((authTypeName != 'Attest' && authTypeName != 'Salary report') || app.getDomain().get('salary_mod')) {
				this.$("#available_auth_types").append(new AvailableAuthTypeView({ auth_type: app.authTypes[authTypeName] }).render().el);
			}
		}
		return this;
	},

});
