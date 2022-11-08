window.SalaryPropertiesView = Backbone.View.extend({

  template: _.template($('#salary_properties_underscore').html()),

  initialize: function(options) {
    _.bindAll(this, 'render');
		this.set_name = options.set_name;
		this.getter = options.getter;
		this.setter = options.setter;
  },

  render: function() {
		var that = this;
    that.$el.html(that.template({
		  model: that.model,
		}));
		_.each(app.getDomain().get('salary_config')[that.set_name] || [], function(prop) {
      if (prop.type == 'free_text') {
			  that.$el.append(new FreeTextSalaryPropertyView({
				  model: prop,
					getter: that.getter,
					setter: that.setter,
				}).render().el);
			} else if (prop.type == 'integer') {
			  that.$el.append(new IntegerSalaryPropertyView({
				  model: prop,
					getter: that.getter,
					setter: that.setter,
				}).render().el);
			} else if (prop.type == 'float') {
			  that.$el.append(new FloatSalaryPropertyView({
				  model: prop,
					getter: that.getter,
					setter: that.setter,
				}).render().el);
			} else if (prop.type == 'select') {
			  that.$el.append(new SelectSalaryPropertyView({
				  model: prop,
					getter: that.getter,
					setter: that.setter,
				}).render().el);
			}
		});
    return that;
  },

});
