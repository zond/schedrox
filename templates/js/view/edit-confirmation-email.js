window.EditConfirmationEmailView = Backbone.View.extend({

  template: _.template($('#edit_confirmation_email_underscore').html()),

  events: {
    "change #event_type_confirmation_email_body_template": "updateBodyTemplate",
    "change #event_type_confirmation_email_subject_template": "updateSubjectTemplate",
  },

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.model.bind("change", this.render);
  },

  updateBodyTemplate: function(ev) {
    this.model.set('confirmation_email_body_template', $(ev.target).val(), { silent: true })
  },

  updateSubjectTemplate: function(ev) {
    this.model.set('confirmation_email_subject_template', $(ev.target).val(), { silent: true })
  },

  render: function() {
    this.$el.html(this.template({ 
      model: this.model,
      write_auth: app.hasAuth({
        auth_type: 'Event types',
	write: true,
      }),
    }));
    return this;
  },

});
