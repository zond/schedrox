window.ContactsView = Backbone.View.extend({

	template: _.template($('#contacts_underscore').html()),

	events: {
		"change #new_contact": "addContact",
		"click .pagination a": "changePage",
		"keyup #search_contacts": "searchContacts",
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'refetch');
		this.show_contact = options.show_contact;
		this.collection = new Contacts();
		this.refetch();
		app.on('domainchange', this.refetch);
	},

	searchContacts: function(ev) {
		var val = $(ev.target).val();
		var values = val.split(" ");
		var doit = false;
		for (var i = 0; i < values.length; i++) {
			if (values[i].trim().length > 1) {
				doit = true;
			}
		}
		if (doit) {
			this.collection.search($(ev.target).val());
		} else if (this.collection.query != null && val.length < 2) {
			this.collection.page(0);
		}
	},

	refetch: function() {
		if (app.getDomain() != null) {
			this.collection.fetch({ reset: true });
		}
	},

	addContact: function(ev) {
		if (app.getDomain() != null) {
			var that = this;
			var newContact = new Contact({ name: $(ev.target).val() });
			newContact.save(null, {
				success: function() {
					that.collection.add(newContact);
				}
			});
		}
	},

	cleanup: function() {
		app.off('domainchange', this.refetch);
	},

	changePage: function(event) {
		event.stopPropagation();
		event.preventDefault();
		if (!$(event.target).parent().hasClass("disabled")) {
			this.collection.offset = (parseInt($(event.target).attr("data-page")) - 1) * this.collection.limit;
			this.refetch();
		}
	},

	render: function() {
		this.$el.html(this.template({ }));
		new ContactsResultsView({ 
			el: this.$("#contacts_results"),
			collection: this.collection,
		}).render();
		if (this.show_contact != null) {
			new ContactDetailsView({ model: this.show_contact }).modal(function() {
				app.navigate('/contacts');
			});
			this.show_contact = null;
		}
		return this;
	}

});
