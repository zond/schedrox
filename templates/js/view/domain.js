window.DomainView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#domain_underscore').html()),

	events: {
		"click .close": "removeDomain",
		"click .domain-salary-mod": "toggleSalaryMod",
		"change .domain-closed-and-redirected-to": "changeClosed",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	changeClosed: function(ev) {
	  ev.preventDefault();
		this.model.set('closed_and_redirected_to', $(ev.target).val());
		this.model.save();
	},

	toggleSalaryMod: function(ev) {
	  this.model.set('salary_mod', !this.model.get('salary_mod'));
    if (app.getDomain().get('id') == this.model.get('id')) {
		  app._domain.set('salary_mod', this.model.get('salary_mod'));			
		}
		for (var i = 0; i < app.user.get('domains').length; i++) {
		  if (app.user.get('domains')[i].id == this.model.get('id')) {
			  app.user.get('domains')[i].salary_mod = this.model.get('salary_mod');
			}
		}
		this.model.save();
	},

	removeDomain: function(event) {
		var that = this;
		myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(this.model.get("name").htmlEscape()), function() {
			if (app.getDomain() != null && app.getDomain().get('id') == that.model.get('id')) {
				app.setDomain(null);
			}
			window.app.menu.deleteDomain(that.model.get('id'));
			window.app.user.deleteDomain(that.model.get('id'));
			that.model.destroy();
		});
	},

	render: function() {
		this.$el.html(this.template({ model: this.model }));
		return this;
	},

});
