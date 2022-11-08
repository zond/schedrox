window.ContactEventsView = Backbone.View.extend({

	template: _.template($('#contact_events_underscore').html()),

	className: 'contact-events',

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.collection = new ContactEvents([], { url: '/contacts/' + this.model.get('id') + '/events' });
    this.collection.bind("change", this.render);
    this.collection.bind("reset", this.render);
    this.collection.bind("add", this.render);
    this.collection.bind("remove", this.render);
    this.collection.fetch();
	},

	modal: function(cb) {
		$.modal.close();
		app.navigate('/contacts/' + this.model.get('id'));
		mymodal(this.render().el, { 'onClose': cb });
	},

	render: function() {
		this.$el.html(this.template({ 
			model: this.model,
		}));
		var that = this;
		this.collection.each(function(ev) {
			thisDate = anyDateConverter.format(ev.get('start'));
			that.$('#event_list').append(new ContactEventView({ 
				model: ev,
			}).render().el);
		});

		return this;
	},

});
