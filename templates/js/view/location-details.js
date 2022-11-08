window.LocationDetailsView = Backbone.View.extend({

	template: _.template($('#location_details_underscore').html()),

	className: 'location-details',

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	modal: function(cb) {
		var that = this;
		if (app.hasAuth({
			auth_type: 'Domain',
			write: true,
		})) {
			var locationOpener = function() {
				$.modal.close();
				app.navigate('/locations/' + that.model.get('id'));
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
			locationOpener();
		} else {
			mymodal(that.render().el, { 'onClose': cb });
		}
	},

	render: function() {
		var that = this;
		var write_auth = app.hasAuth({
			auth_type: 'Domain',
			write: true,
		});
		that.$el.html(that.template({ 
			model: that.model,
			write_auth: write_auth,
		}));
		if (window.app != null && app.getDomain() != null && app.getDomain().get('salary_mod') && (app.getDomain().get('salary_config').salary_location_properties || []).length > 0) {
			new SalaryPropertiesView({
				el: that.$('#salary_properties'),
				set_name: 'salary_location_properties',
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
