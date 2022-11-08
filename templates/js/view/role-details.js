window.RoleDetailsView = Backbone.View.extend({

  template: _.template($('#role_details_underscore').html()),

  className: 'role-details',

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.auths = new Auths([], { url: '/roles/' + this.model.id + '/auths' });
    this.auths.fetch({ reset: true });

    this.model.bind("change", this.render);
    
    this.locations = new Locations();
    this.locations.fetch({ reset: true });
    this.event_types = new EventTypes();
    this.event_types.fetch({ reset: true });
    this.event_kinds = new EventKinds();
    this.event_kinds.fetch({ reset: true });
    this.participant_types = new ParticipantTypes();
    this.participant_types.fetch({ reset: true });
		this.roles = new Roles([], { url: '/roles' });
		this.roles.fetch({ reset: true });
  },

  modal: function(cb) {
    var that = this;
    if (app.hasAuth({
      auth_type: 'Roles',
      write: true,
    })) {
      $.modal.close();
      app.navigate('/settings/roles/' + that.model.get('id'));
      mymodal(that.render().el, {
	'{{.I "Save" }}': function() {
	  that.auths.save(null, {
	     success: cb,
	  });
	},
	'onCancel': cb,
      });
    } else {
      mymodal(that.render().el, { 'onClose': cb });
    }
  },

  render: function() {
    this.$el.html(this.template({ model: this.model }));
    new AuthsView({
      el: this.$("#auths"),
      collection: this.auths,
      locations: this.locations,
      event_types: this.event_types,
      event_kinds: this.event_kinds,
      participant_types: this.participant_types,
			roles: this.roles,
    }).render();
    return this;
  },

});
