window.PreviewConfirmationEmailView = Backbone.View.extend({

  template: _.template($('#preview_confirmation_email_underscore').html()),

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.model.bind("change", this.render);
    this.subject = "";
    this.body = "";
    var that = this;
    $.ajax('/example_confirmation', {
      type: 'POST',
      dataType: 'json',
      data: JSON.stringify({
        subject_template: this.model.get('confirmation_email_subject_template'),
	body_template: this.model.get('confirmation_email_body_template'),
      }),
      success: function(data) {
        that.subject = data.subject,
	that.body = data.body;
        that.render();
      },
    });
  },

  render: function() {
    this.$el.html(this.template({
      subject: this.subject,
      body: this.body,
    }));
    return this;
  },

});
