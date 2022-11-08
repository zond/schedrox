window.FreeTextSalaryOptionsView = Backbone.View.extend({

  template: _.template($('#free_text_salary_options_underscore').html()),

	events: {
	  "change .salary-option-regexp": "changeRegexp",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
		this.saveModel = options.saveModel;
  },

  changeRegexp: function(ev) {
	  this.model.regexp = $(ev.target).val();
		this.saveModel();
	},
  
	render: function() {
    this.$el.html(this.template({ 
      model: this.model,
    }));
    return this;
  },

});
