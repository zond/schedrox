window.FloatSalaryPropertyView = Backbone.View.extend({

  template: _.template($('#float_salary_property_underscore').html()),

	events: {
	  "change .float-property": "changeFloat",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
		this.user = options.user;
    this.error = null; 
		this.getter = options.getter;
		this.setter = options.setter;
	},

  changeFloat: function(ev) {
	  if (!isNaN($(ev.target).val()) && 
		    (this.model.minimum == null || parseFloat($(ev.target).val()) >= this.model.minimum) && 
				(this.model.maximum == null || parseFloat($(ev.target).val()) <= this.model.maximum)) {
			var props = this.getter() || {};
			props[this.model.name] = parseFloat($(ev.target).val());
			this.setter(props);
			if (this.error != null) {
				this.error = null;
				this.render();
			}
		} else {
		  this.error = '{{.I "Invalid value: {0}"}}'.format($(ev.target).val());
			this.render();
		}
	},

  render: function() {
    this.$el.html(this.template({
		  model: this.model,
			error: this.error,
			getter: this.getter,
			write_auth: this.setter != null,
		}));
    return this;
  },

});
