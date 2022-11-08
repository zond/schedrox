window.UserFilterView = Backbone.View.extend({

  tagName: 'tr',

  template: _.template($('#user_filter_underscore').html()),

  events: {
    "click .role-filter a": "changeRole",
    "click .norole-filter a": "changeNoRole",
    "click .noprop-filter a": "changeNoProperty",
    "click .property-filter a": "changeProperty",
    "click .disabled-filter a": "changeDisabled",
		"click .unattested-filter a": "changeUnattested",
		"click .attested-filter a": "changeAttested",
  },

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.user_properties_for_domain = options.user_properties_for_domain;
    this.available_roles = options.available_roles;
    this.update = options.update;
  },

  changeNoProperty: function(ev) {
    ev.preventDefault();
    this.model.value = $(ev.target).attr('data-user-property-name');
    this.model.desc = $(ev.target).attr('data-user-property-name');
    this.render();
    this.update();
  },

  changeProperty: function(ev) {
    ev.preventDefault();
    this.model.value = $(ev.target).attr('data-user-property-name');
    this.model.desc = $(ev.target).attr('data-user-property-name');
    this.render();
    this.update();
  },

  changeDisabled: function(ev) {
    ev.preventDefault();
    this.model.value = $(ev.target).attr('data-disabled-state')
    this.model.desc = $(ev.target).attr('data-disabled-text');
    this.render();
    this.update();
  },

  changeNoRole: function(ev) {
    ev.preventDefault();
    this.model.value = $(ev.target).attr('data-role-name');
    this.model.desc = $(ev.target).attr('data-role-name');
    this.render();
    this.update();
  },

  changeRole: function(ev) {
    ev.preventDefault();
    this.model.value = $(ev.target).attr('data-role-name');
    this.model.desc = $(ev.target).attr('data-role-name');
    this.render();
    this.update();
  },

	changeUnattested: function(ev) {
	  ev.preventDefault();
		this.model.value = $(ev.target).attr('data-period');
		this.model.desc = $(ev.target).text();
    this.render();
		this.update();
	},

	changeAttested: function(ev) {
	  ev.preventDefault();
		this.model.value = $(ev.target).attr('data-period');
		this.model.desc = $(ev.target).text();
    this.render();
		this.update();
	},

  render: function() {
    var that = this;
    this.$el.html(this.template({ model: this.model }));
    this.user_properties_for_domain.forEach(function(prop) {
      that.$('.property-filter').append(new AvailableUserPropertyView({ model: prop }).render().el);
      that.$('.noprop-filter').append(new AvailableUserPropertyView({ model: prop }).render().el);
    })
    this.available_roles.forEach(function(role) {
      that.$('.role-filter').append(new AvailableRoleView({ model: role }).render().el);
      that.$('.norole-filter').append(new AvailableRoleView({ model: role }).render().el);
    })
		var t = firstSalaryBreakpointAfter(today());
		for (var i = 0; i < 5; i++) {
		  var back = lastSalaryBreakpointBefore(t);
			that.$('.unattested-filter').append('<li><a href="#" data-period="' + (back.getISOTime() / 1000) + '-' + (t.getISOTime() / 1000) + '">' + anyDateConverter.format(back) + ' - ' + anyDateConverter.format(t) + '</a></li>');
			t = back;
		}
		t = firstSalaryBreakpointAfter(today());
		for (var i = 0; i < 5; i++) {
		  var back = lastSalaryBreakpointBefore(t);
			that.$('.attested-filter').append('<li><a href="#" data-period="' + (back.getISOTime() / 1000) + '-' + (t.getISOTime() / 1000) + '">' + anyDateConverter.format(back) + ' - ' + anyDateConverter.format(t) + '</a></li>');
			t = back;
		}
    that.$('.disabled-filter').append('<li><a href="#" data-disabled-state="true" data-disabled-text="{{.I "Yes" }}">{{.I "Yes" }}</a></li>')
    that.$('.disabled-filter').append('<li><a href="#" data-disabled-state="false" data-disabled-text={{.I "No" }}>{{.I "No" }}</a></li>')
    return this;
  },

});
