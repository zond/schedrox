window.EventTypesView = Backbone.View.extend({

	template: _.template($('#event_types_underscore').html()),

	events: {
		"change #new_type": "newType",
		"change #new_kind": "newKind",
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'refetch');
		this.show_type = options.show_type;
		this.show_kind = options.show_kind;
		this.event_types = new EventTypes();
		this.event_types.bind("change", this.render);
		this.event_types.bind("reset", this.render);
		this.event_types.bind("add", this.render);
		this.event_types.bind("remove", this.render);
		this.event_kinds = new EventKinds();
		this.event_kinds.bind("change", this.render);
		this.event_kinds.bind("reset", this.render);
		this.event_kinds.bind("add", this.render);
		this.event_kinds.bind("remove", this.render);
		this.refetch();
		app.on('domainchange', this.refetch);
	},

	cleanup: function() {
		app.off('domainchange', this.refetch);
		$('.colorpicker').each(function(x, el) {
			$(el).remove();
		});
	},

	newType: function(ev) {
		if (app.getDomain() != null) {
			var that = this;
			var newEventType = new EventType({ name: $(ev.target).val() });
			newEventType.save(null, {
				success: function() {
					that.event_types.add(newEventType);
				}
			});
		}
	},

	newKind: function(ev) {
		if (app.getDomain() != null) {
			var that = this;
			var newEventKind = new EventKind({ name: $(ev.target).val() });
			newEventKind.save(null, {
				success: function() {
					that.event_kinds.add(newEventKind);
				}
			});
		}
	},

	refetch: function() {
	  var that = this;
		if (app.getDomain() != null) {
			that.event_types.fetch({ reset: true });
			that.event_kinds.fetch({ reset: true });
		}
	},

	render: function() {
		this.$el.html(this.template({ 
			event_types: this.event_types, 
			event_kinds: this.event_kinds,
		}));
		this.event_kinds.forEach(function(event_kind) {
			this.$('#event_kind_list').append(new EventKindView({ model: event_kind }).render().el);
		});
		var that = this;
		this.event_types.forEach(function(event_type) {
			this.$('#event_type_list').append(new EventTypeView({ 
				model: event_type, 
				kinds: that.event_kinds, 
			}).render().el);
		});
		if (this.show_type != null) {
			new EventTypeDetailsView({ model: this.show_type }).modal(function() {
				app.navigate('/events/types');
			});
			this.show_type = null;
		} else if (this.show_kind != null) {
			new EventKindDetailsView({ model: this.show_kind }).modal(function() {
				app.navigate('/events/types');
			});
			this.show_kind = null;
		}
		return this;
	}

});
