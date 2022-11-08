window.IntegerSalaryOptionsView = Backbone.View.extend({

  template: _.template($('#integer_salary_options_underscore').html()),

	events: {
    "blur .salary-option-minimum": "changeMinimum",
    "blur .salary-option-maximum": "changeMaximum",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
		this.saveModel = options.saveModel;
  },

	changeMaximum: function(ev) {
	  this.model.maximum = parseInt($(ev.target).val());
		if (this.model.minimum > this.model.maximum) {
		  this.model.minimum = this.model.maximum;
		}
		this.saveModel();
	},

	changeMinimum: function(ev) {
	  this.model.minimum = parseInt($(ev.target).val());
		if (this.model.minimum > this.model.maximum) {
		  this.model.maximum = this.model.minimum;
		}
		this.saveModel();
	},

	render: function() {
    this.$el.html(this.template({ 
      model: this.model,
    }));
    return this;
  },

});
