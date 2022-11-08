window.FreeTextSalaryPropertyView = Backbone.View.extend({

  template: _.template($('#free_text_salary_property_underscore').html()),

	events: {
	  "change .text-property": "changeText",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.error = null; 
		this.getter = options.getter;
		this.setter = options.setter;
	},

  changeText: function(ev) {
	  if (this.model.regexp == null || this.model.regexp == '' || $(ev.target).val().match(new RegExp(this.model.regexp)) != null) {
			var props = this.getter() || {};
			props[this.model.name] = $(ev.target).val();
			this.setter(props);
			this.error = null;
		} else {
		  this.error = '{{.I "Invalid value: {0}"}}'.format($(ev.target).val());
		}
		this.render();
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
