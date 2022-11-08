window.LocationView = Backbone.View.extend({

  tagName: 'tr',

  template: _.template($('#location_underscore').html()),

  events: {
    "click .close": "removeLocation",
		"click .open-location": "openLocation",
  },

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.model.bind("change", this.render);
  },

	openLocation: function(event) {
		event.preventDefault();
		new LocationDetailsView({ model: this.model }).modal(function() {
			app.navigate('/settings/domain');
		});
	},

  removeLocation: function(event) {
    if (app.hasAuth({
      auth_type: 'Domain',
      write: true,
    })) {
      var that = this;
      myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(this.model.get("name").htmlEscape()), function() {
				that.model.destroy();
      });
    }
  },

  render: function() {
		var hasProps = (window.app != null && app.getDomain() != null && app.getDomain().get('salary_mod') && (app.getDomain().get('salary_config').salary_location_properties || []).length > 0);
    this.$el.html(this.template({ 
			model: this.model,
		  hasProps: hasProps,
		}));
    return this;
  },

});
