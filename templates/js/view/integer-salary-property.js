window.IntegerSalaryPropertyView = Backbone.View.extend({

  template: _.template($('#integer_salary_property_underscore').html()),

	events: {
	  "blur .integer-property": "changeInteger",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.error = null; 
		this.getter = options.getter;
		this.setter = options.setter;
	},

  changeInteger: function(ev) {
	  if (('' + parseInt($(ev.target).val())) == $(ev.target).val() && 
		    (this.model.minimum == null || parseInt($(ev.target).val()) >= this.model.minimum) && 
				(this.model.maximum == null || parseInt($(ev.target).val()) <= this.model.maximum)) {
			var props = this.getter() || {};
			props[this.model.name] = parseInt($(ev.target).val());
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
