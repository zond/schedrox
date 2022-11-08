window.Participant = Backbone.Model.extend({
  initialize: function(attributes) {
    this.modified = false;
  },
  gravatarImageURL: function(options) {
    return gravatarImage(this.get('gravatar_hash'), options);
  },
  name: function() {
    if (this.get('user') != null) {
      if ((this.get('given_name') != null && this.get('given_name') != '') || (this.get('family_name') != null && this.get('family_name') != '')) {
	return '{{.I "name_order"}}'.format(this.get('given_name'), this.get('family_name'));
      } else {
	return this.get('email');
      }
    } else {
      return this.get('name');
    }
  },
});
