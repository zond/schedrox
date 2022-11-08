window.SalaryPropertyForDomainView = Backbone.View.extend({

  tagName: 'tr',

  template: _.template($('#salary_property_for_domain_underscore').html()),

  events: {
    "click .available-salary-property-type": "changeType",
		"click .close": "removeProperty",
		"change .salary-property-name": "changeName",
	},

  initialize: function(options) {
    _.bindAll(this, 'render');
		this.saveModel = options.saveModel;
		this.parent = options.parent;
		this.set_name = options.set_name;
  },

	removeProperty: function(ev) {
	  ev.preventDefault();
		var oldProps = this.parent.get(this.set_name);
		var newProps = [];
		_.each(oldProps, function(prop) {
		  if (prop.name != $(ev.target).attr('data-salary-property-name')) {
			  newProps.push(prop);
			}
		});
		this.parent.set(this.set_name, newProps);
		this.saveModel();
	},

	changeName: function(ev) {
	  ev.preventDefault();
		this.model.name = $(ev.target).val();
		this.saveModel();
	},

	changeType: function(ev) {
	  ev.preventDefault();
		this.model.type = $(ev.target).attr('data-salary-property-type');
		this.saveModel();
	},

  render: function() {
		var write_auth = app.hasAuth({
			auth_type: 'Salary configuration',
			write: true,
		});
    this.$el.html(this.template({ 
      model: this.model,
			write_auth: write_auth,
    }));
		if (this.model.type == 'free_text') {
		  new FreeTextSalaryOptionsView({ 
				model: this.model,
				saveModel: this.saveModel,
				el: this.$('.salary-property-options'),
			}).render();
		} else if (this.model.type == 'integer') {
		  new IntegerSalaryOptionsView({ 
				model: this.model,
				saveModel: this.saveModel,
				el: this.$('.salary-property-options'),
			}).render();
		} else if (this.model.type == 'float') {
		  new IntegerSalaryOptionsView({ 
				model: this.model,
				saveModel: this.saveModel,
				el: this.$('.salary-property-options'),
			}).render();
		} else if (this.model.type == 'select') {
		  new SelectSalaryOptionsView({ 
				model: this.model,
				parent: this.parent,
				set_name: this.set_name,
				saveModel: this.saveModel,
				el: this.$('.salary-property-options'),
			}).render();
		}
    return this;
  },

});
