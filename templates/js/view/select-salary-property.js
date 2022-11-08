window.SelectSalaryPropertyView = Backbone.View.extend({

  template: _.template($('#select_salary_property_underscore').html()),

	events: {
	  "click .select-property": "changeOption",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.error = null; 
		this.getter = options.getter;
		this.setter = options.setter;
	},

  changeOption: function(ev) {
		var props = this.getter() || {};
		props[this.model.name] = $(ev.target).attr('data-select-option');
		this.setter(props);
		this.render();
	},

  render: function() {
		this.$el.html(this.template({
			model: this.model,
			getter: this.getter,
			write_auth: this.setter != null,
		}));
    return this;
  },

});
