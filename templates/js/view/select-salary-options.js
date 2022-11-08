window.SelectSalaryOptionsView = Backbone.View.extend({

  template: _.template($('#select_salary_options_underscore').html()),

	events: {
	  "change .salary-option-options": "changeOptions",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
		this.saveModel = options.saveModel;
		this.parent = options.parent;
		this.set_name = options.set_name;
  },

	changeOptions: function(ev) {
	  var that = this;
	  that.model.options = $(ev.target).val().split(",");
		that.saveModel(function() {
		  that.model = _.find(that.parent.get(that.set_name), function(prop) {
				return prop.name == that.model.name;
			});
		});
	},

	render: function() {
    this.$el.html(this.template({ 
      model: this.model,
    }));
		this.$('.salary-option-options').select2({
			tags:[],
			tokenSeparators: [",", " "]
		});
    return this;
  },

});
