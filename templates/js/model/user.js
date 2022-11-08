window.User = Backbone.Model.extend({
  urlRoot: '/users',

  gravatarImageURL: function(options) {
    return gravatarImage(this.get('gravatar_hash'), options);
  },

  name: function() {
    if ((this.get('given_name') != null && this.get('given_name') != '') || (this.get('family_name') != null && this.get('family_name') != '')) {
      return '{{.I "name_order"}}'.format(this.get('given_name'), this.get('family_name'));
    } else {
      return this.get('email');
    }
  },

  withGravatarProfileData: function(cb) {
    var hash = this.get('gravatar_hash');
    if (hash == null || hash == '') {
      cb({
				missing_gravatar_profile: true,
				has_gravatar_profile: false,
      });
    } else {
      app.displayLoader();
      $.ajax('https://www.gravatar.com/' + this.get('gravatar_hash') + '.json', {
				dataType: 'jsonp',
				timeout: 5000,
				success: function(data) {
					app.hideLoader();
					data.has_gravatar_profile = true;
					data.missing_gravatar_profile = false;
					cb(data);
				},
				error: function() {
					app.hideLoader();
					cb({
						missing_gravatar_profile: true,
						has_gravatar_profile: false,
					});
				},
			});
    }			
  },

  gravatarProfileURL: function() {
    var hash = this.get('gravatar_hash');
    if (hash == null || hash == "") {
      return null;
    }
    return 'https://www.gravatar.com/' + hash;
  },

  addDomain: function(attributes) {
    attributes.owner = true;
    this.get('domains').push(attributes);
    this.trigger('change');
  },
  
  deleteDomain: function(id) {
    var domains = this.get('domains');
    domains = _.reject(domains, function(d) {
      return d.id == id;
    });
    this.set('domains', domains);
  },

});
