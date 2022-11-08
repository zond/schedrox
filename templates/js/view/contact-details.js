window.ContactDetailsView = Backbone.View.extend({

	template: _.template($('#contact_details_underscore').html()),

	className: 'contact-details',

	events: {
		"change #contact_contact_given_name": "changeContactGivenName",
		"change #contact_contact_family_name": "changeContactFamilyName",
		"change #contact_name": "changeName",
		"change #contact_email": "changeEmail",
		"change #contact_mobile_phone": "changeMobilePhone",
		"change #contact_address_line_1": "changeAddressLine1",
		"change #contact_address_line_2": "changeAddressLine2",
		"change #contact_address_line_3": "changeAddressLine3",
		"change #contact_billing_address_line_1": "changeBillingAddressLine1",
		"change #contact_billing_address_line_2": "changeBillingAddressLine2",
		"change #contact_billing_address_line_3": "changeBillingAddressLine3",
		"change #contact_information": "changeInformation",
		"change #contact_reference": "changeReference",
		"change #contact_organization_number": "changeOrgNumber",
		"click #clear_email_bounce": "clearEmailBounce",
		"click #contact_events": "openContactEvents",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	openContactEvents: function(ev) {
	  ev.preventDefault();
		var that = this;
		$.modal.close();
    new ContactEventsView({ 
			model: this.model,
		}).modal(function() {
		  that.modal(that.onClose);
    });
	},

	modal: function(cb) {
		var that = this;
		$.modal.close();
		app.navigate('/contacts/' + that.model.get('id'));
		if (app.hasAuth({
			auth_type: 'Contacts',
			write: true,
		})) {
      this.render();
			this.delegateEvents();
			mymodal(this.el, {
				"{{.I "Save"}}": function() {
					that.model.save(null, {
						success: cb,
					});
				},
				'onCancel': cb,
			});
		} else {
			mymodal(this.render().el, { 'onClose': cb });
		}
	},

	clearEmailBounce: function(event) {
		event.preventDefault();
		this.model.set('email_bounce', '');
	},

	changeContactGivenName: function(event) {
		this.model.set('contact_given_name', $(event.target).val(), { silent: true });
	},

	changeName: function(event) {
		this.model.set('name', $(event.target).val(), { silent: true });
	},

	changeContactFamilyName: function(event) {
		this.model.set('contact_family_name', $(event.target).val(), { silent: true });
	},

	changeEmail: function(event) {
		if ($(event.target).val() == '' || $(event.target).val().isEmail()) {
			this.model.set('email', $(event.target).val(), { silent: true });
			if (this.model.errors && this.model.errors.email) {
				this.model.errors.email = null;
				this.render();
			}
		} else {
			this.model.errors = {};
			this.model.errors.email = '{{.I "{0} is not a valid email address." }}'.format($(event.target).val());
			this.render();
			this.delegateEvents();
		}
	},

	changeReference: function(event) {
		this.model.set('reference', $(event.target).val(), { silent: true });
	},

	changeOrgNumber: function(event) {
		this.model.set('organization_number', $(event.target).val(), { silent: true });
	},

	changeInformation: function(event) {
		this.model.set('information', $(event.target).val(), { silent: true });
	},

	changeMobilePhone: function(event) {
		this.model.set('mobile_phone', $(event.target).val(), { silent: true });
	},

	changeBillingAddressLine1: function(event) {
		this.model.set('billing_address_line_1', $(event.target).val(), { silent: true });
	},

	changeBillingAddressLine2: function(event) {
		this.model.set('billing_address_line_2', $(event.target).val(), { silent: true });
	},

	changeBillingAddressLine3: function(event) {
		this.model.set('billing_address_line_3', $(event.target).val(), { silent: true });
	},

	changeAddressLine1: function(event) {
		this.model.set('address_line_1', $(event.target).val(), { silent: true });
	},

	changeAddressLine2: function(event) {
		this.model.set('address_line_2', $(event.target).val(), { silent: true });
	},

	changeAddressLine3: function(event) {
		this.model.set('address_line_3', $(event.target).val(), { silent: true });
	},

	render: function() {
		this.$el.html(this.template({ 
		  write_auth: app.hasAuth({ auth_type: 'Contacts', write: true }),
			model: this.model,
		}));
		return this;
	},

});
