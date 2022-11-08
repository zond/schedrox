window.SalaryConfigureView = Backbone.View.extend({

  template: _.template($('#salary_configure_underscore').html()),

	events: {
		"click .available-period-type": "changePeriod",
		"change #domain_salary_breakpoint": "changeBreakpoint",
		"click .available-salary-breakpoint": "changeWeekdayBreakpoint",
		"click #upload_code": "uploadCode",
		"change #new_salary_participant_type_property": "addSalaryParticipantTypeProperty",
		"change #new_salary_user_property": "addSalaryUserProperty",
		"change #new_salary_event_type_property": "addSalaryEventTypeProperty",
		"change #new_salary_event_kind_property": "addSalaryEventKindProperty",
		"change #new_salary_location_property": "addSalaryLocationProperty",
		"change #domain_salary_report_hours_min_minutes": "changeMinMinutes",
	},

  initialize: function(options) {
    _.bindAll(this, 'render', 'refetch', 'saveModel');
    app.on('domainchange', this.refetch);
		this.model = new SalaryConfig({});
		this.model.bind("change", this.render);
		this.refetch();
  },

  cleanup: function() {
    app.off('domainchange', this.refetch);
  },

	refetch: function() {
		if (app.getDomain() != null) {
			this.model.fetch();
		}
  },

	uploadCode: function(ev) {
	  ev.preventDefault();
		this.$('.upload-salary').submit();
	},

  saveModel: function(cb) {
	  var that = this;
		this.model.save({}, {
			success: function() {
			  if (cb == null) {
					that.render();
				} else {
					cb();
				}
				app.getDomain().set('salary_config', that.model.attributes);
			}
		});
	},

	addSalaryLocationProperty: function(ev) {
	  if (!_.any(this.model.get('salary_location_properties'), function(prop) {
		  return prop.name == $(ev.target).val();
		})) {
			this.model.get('salary_location_properties').push({
				type: salary_property_types[0],
				name: $(ev.target).val(),
			});
			this.saveModel();
		}
	},

	addSalaryEventKindProperty: function(ev) {
	  if (!_.any(this.model.get('salary_event_kind_properties'), function(prop) {
		  return prop.name == $(ev.target).val();
		})) {
			this.model.get('salary_event_kind_properties').push({
				type: salary_property_types[0],
				name: $(ev.target).val(),
			});
			this.saveModel();
		}
	},

	addSalaryEventTypeProperty: function(ev) {
	  if (!_.any(this.model.get('salary_event_type_properties'), function(prop) {
		  return prop.name == $(ev.target).val();
		})) {
			this.model.get('salary_event_type_properties').push({
				type: salary_property_types[0],
				name: $(ev.target).val(),
			});
			this.saveModel();
		}
	},

	addSalaryParticipantTypeProperty: function(ev) {
	  if (!_.any(this.model.get('salary_participant_type_properties'), function(prop) {
		  return prop.name == $(ev.target).val();
		})) {
			this.model.get('salary_participant_type_properties').push({
				type: salary_property_types[0],
				name: $(ev.target).val(),
			});
			this.saveModel();
		}
	},

	addSalaryUserProperty: function(ev) {
	  if (!_.any(this.model.get('salary_user_properties'), function(prop) {
		  return prop.name == $(ev.target).val();
		})) {
			this.model.get('salary_user_properties').push({
				type: salary_property_types[0],
				name: $(ev.target).val(),
			});
			this.saveModel();
		}
	},

	changePeriod: function(ev) {
	  ev.preventDefault();
		this.model.set('salary_period', $(ev.target).attr('data-period-type'));
		this.saveModel();
	},

	changeMinMinutes: function(ev) {
		ev.preventDefault();
		this.model.set('salary_report_hours_min_minutes', parseInt($(ev.target).val()));
		this.saveModel();
	},

	changeWeekdayBreakpoint: function(ev) {
	  ev.preventDefault();
		this.model.set('salary_breakpoint', parseInt($(ev.target).attr('data-salary-breakpoint')), { silent: true });
		this.saveModel();
	},

	changeBreakpoint: function(ev) {
    var maxBreakpoint = 31;
		var neu = parseInt($(ev.target).val());
		if (neu < 1) {
		  $(ev.target).val('1');
		} else if (neu > maxBreakpoint) {
		  $(ev.target).val('' + maxBreakpoint);
		} else {
			this.model.set('salary_breakpoint', neu, { silent: true });
			this.model.save();
		}
	},

	render: function() {
	  var that = this;
		var write_auth = app.hasAuth({
			auth_type: 'Salary configuration',
			write: true,
		});
		that.$el.html(that.template({
			write_auth: write_auth,
		  model: that.model,
		}));
		_.each(that.model.get('salary_event_type_properties'), function(prop) {
		  that.$('#salary_event_type_properties_list').append(new SalaryPropertyForDomainView({ 
				model: prop,
				parent: that.model,
				saveModel: that.saveModel,
				set_name: 'salary_event_type_properties',
			}).render().el);
		});
		_.each(that.model.get('salary_participant_type_properties'), function(prop) {
		  that.$('#salary_participant_type_properties_list').append(new SalaryPropertyForDomainView({ 
				model: prop,
				parent: that.model,
				saveModel: that.saveModel,
				set_name: 'salary_participant_type_properties',
			}).render().el);
		});
		_.each(that.model.get('salary_user_properties'), function(prop) {
		  that.$('#salary_user_properties_list').append(new SalaryPropertyForDomainView({ 
				model: prop,
				parent: that.model,
				saveModel: that.saveModel,
				set_name: 'salary_user_properties',
			}).render().el);
		});
		_.each(that.model.get('salary_event_kind_properties'), function(prop) {
		  that.$('#salary_event_kind_properties_list').append(new SalaryPropertyForDomainView({ 
				model: prop,
				parent: that.model,
				saveModel: that.saveModel,
				set_name: 'salary_event_kind_properties',
			}).render().el);
		});
		_.each(that.model.get('salary_location_properties'), function(prop) {
		  that.$('#salary_location_properties_list').append(new SalaryPropertyForDomainView({ 
				model: prop,
				parent: that.model,
				saveModel: that.saveModel,
				set_name: 'salary_location_properties',
			}).render().el);
		});
		return this;
  }

});
