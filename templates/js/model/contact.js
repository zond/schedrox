window.Contact = Backbone.Model.extend({
  urlRoot: '/contacts',

  gravatarImageURL: function(options) {
    return gravatarImage(this.get('gravatar_hash'), options);
  },

  contactName: function() {
    return "{{.I "name_order"}}".format(this.get('contact_given_name'), this.get('contact_family_name'));
  },
});

