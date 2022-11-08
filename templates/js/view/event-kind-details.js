window.EventKindDetailsView = Backbone.View.extend({

	template: _.template($('#event_kind_details_underscore').html()),

	className: 'event-kind-details',

	events: {
		"click .toggle-alert": "toggleAlert",
		"click .toggle-block": "toggleBlock",
		"click .toggle-series-editable": "toggleSeriesEditable",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	toggleSeriesEditable: function(ev) {
		this.model.set('series_editable', !this.model.get('series_editable'));
	},

	toggleAlert: function(ev) {
		this.model.set('alert', !this.model.get('alert'));
	},

	toggleBlock: function(ev) {
		this.model.set('block', !this.model.get('block'));
	},

	modal: function(cb) {
		var that = this;
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var typeOpener = function() {
				$.modal.close();
				app.navigate('/events/kinds/' + that.model.get('id'));
				mymodal(that.render().el, {
					'{{.I "Save"}}': function() {
						that.model.save(null, {
							success: function() {
								cb();
							},
						});
					},
					'onCancel': cb,
				});
			};
			typeOpener();
		} else {
			mymodal(that.render().el, { 'onClose': cb });
		}
	},

	render: function() {
		var that = this;
		var write_auth = app.hasAuth({
			auth_type: 'Event types',
			write: true,
		});
		that.$el.html(that.template({ 
			model: that.model,
			write_auth: write_auth,
		}));
		if (window.app != null && app.getDomain() != null && app.getDomain().get('salary_mod') && (app.getDomain().get('salary_config').salary_event_kind_properties || []).length > 0) {
			new SalaryPropertiesView({
				el: that.$('#salary_properties'),
				set_name: 'salary_event_kind_properties',
				model: that.model,
				getter: function() {
					return that.model.get('salary_properties');			  
				},
				setter: (write_auth ? function(props) {
					that.model.set('salary_properties', props, { silent: true });
				} : null),
			}).render();
		}
		return that;
	},

});
