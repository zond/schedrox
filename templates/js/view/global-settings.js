window.GlobalSettingsView = Backbone.View.extend({

	template: _.template($('#global_settings_underscore').html()),

	events: {
		"change #new_domain": "addDomain",
		"click #clean_search": "cleanSearch",
		"click #clean_user_properties": "cleanProperties",
		"click #remove_encoded_keys": "removeEncodedKeys",
		"click #convert_recurrence_exceptions": "convertRecurrenceExceptions",
		"click #clean_detached_event_weeks": "cleanDetachedEventWeeks",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.domains = new Domains();
		this.domains.bind("change", this.render);
		this.domains.bind("reset", this.render);
		this.domains.bind("add", this.render);
		this.domains.bind("remove", this.render);
		this.domains.fetch({ reset: true });
	},

	cleanDetachedEventWeeks: function(ev) {
	  ev.preventDefault();
		$.ajax("/maintenance/clean_detached_event_weeks", {
		  dataType: 'json',
			success: function(data) {
			  console.log(data);
			},
		});
	},

	convertRecurrenceExceptions: function(ev) {
	  ev.preventDefault();
		$.ajax("/maintenance/convert_recurrence_exceptions", {
		  dataType: 'json',
			success: function(data) {
			  alert('{{.I "Converted {0} recurrence exceptions"}}'.format(data.deleted));
			},
		});
	},

	removeEncodedKeys: function(ev) {
		ev.preventDefault();
		$.ajax("/maintenance/remove_encoded_keys", {
		  dataType: 'json',
			success: function(data) {
			  alert('{{.I "Removed {0} encoded keys"}}'.format(data.deleted));
			},
		});
	},

  cleanSearch: function(ev) {
		ev.preventDefault();
		$.ajax("/maintenance/cleansearch", {
		  dataType: 'json',
		  success: function(data) {
			  alert('{{.I "Deleted {0} bad search terms"}}'.format(data.deleted));
			},
		});
	},

  cleanProperties: function(ev) {
		ev.preventDefault();
		$.ajax("/maintenance/cleanproperties", {
		  dataType: 'json',
		  success: function(data) {
			  alert('{{.I "Deleted {0} bad properties"}}'.format(data.deleted));
			},
		});
	},

	render: function() {
		this.$el.html(this.template({}));
		this.domains.forEach(function(domain) {
			this.$("#domain_list").append(new DomainView({ model: domain }).render().el);
		});
		return this;
	},

	addDomain: function(event) {
		var that = this;
		var newDomain = new Domain({ 
			name: $("#new_domain").val(),
		});
		newDomain.save(null, {
			success: function() {
				that.domains.add(newDomain);
				if (app.getDomain() == null) {
					app.setDomain(newDomain);
				}
				app.user.addDomain(newDomain.attributes);
				app.menu.addDomain(newDomain.attributes);
			}
		});
	},

});
