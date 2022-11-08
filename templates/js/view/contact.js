window.ContactView = Backbone.View.extend({

  tagName: 'tr',

  template: _.template($('#contact_underscore').html()),

  events: {
    "click .close": "removeContact",
    "click .open": "openContact",
  },

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.model.bind("change", this.render);
  },

  openContact: function(event) {
    event.preventDefault();
    new ContactDetailsView({ model: this.model }).modal(function() {
      app.navigate('/contacts');
    });
  },

  removeContact: function(event) {
    if (app.hasAuth({
      auth_type: 'Contacts',
      write: true,
    })) {
      var that = this;
      myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(this.model.get("name").htmlEscape()), function() {
	that.model.destroy();
      });
    }
  },

  render: function() {
    this.$el.html(this.template({ model: this.model }));
    var messages = [];
    if (this.model.get('email_bounce') != null && this.model.get('email_bounce') != '') {
      messages.push('{{.I "{0} has a non operational email address." }}'.format(this.model.get('name')));
    }
    if (messages.length > 0) {
      this.$el.addClass('warning');
      this.$el.tooltip({
	placement: 'bottom',
	title: messages.join('<br/>'),
	html: true,
      });
    }
    return this;
  },

});
