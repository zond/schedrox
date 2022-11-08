window.EventTypeView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#event_type_underscore').html()),

	events: {
		"click .close": "removeType",
		"click .open-type": "openType",
		"click .available-event-kind": "setKind",
		"change .colorpicker input": "changeColor",
	}, 

	initialize: function(options) {
		this.kinds = options.kinds;
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	changeColor: function(ev) {
		var val = $(ev.target).val()
		if (val == '' || /^#[0-9A-Fa-f]{6,8}$/.exec(val) != null) {
			this.model.set('color', val);
		} else {
			$(ev.target).val(this.model.get('color'));
		}
		this.model.save();
	},

	openType: function(event) {
		event.preventDefault();
		new EventTypeDetailsView({ model: this.model }).modal(function() {
			app.navigate('/events/types');
		});
	},

	setKind: function(event) {
		event.preventDefault();
		var newKindId = $(event.target).attr('data-event-kind-id');
		if (newKindId == '') {
			this.model.set('event_kind', null);
		} else {
			this.model.set('event_kind', newKindId);
		}
		this.model.save();
	},

	removeType: function(event) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var that = this;
			myconfirm("{{.I "Are you sure you want to remove {0}?" }}".format(this.model.get("name").htmlEscape()), function() {
				that.model.destroy();
			});
		}
	},

	render: function() {
		var kind_name = '{{.I "None" }}';
		var that = this;
		this.kinds.forEach(function(kind) {
			if (kind.get('id') == that.model.get('event_kind')) {
				kind_name = kind.get('name');
			}
		});
		var col = this.model.get('color');
		if (col == '') {
			col = '#ffffff';
		}
		this.$el.html(this.template({ 
			model: this.model,
			col: col,
			write_auth: app.hasAuth({ auth_type: 'Event types', write: true }),
			kind_name: kind_name,
		}));
		this.kinds.forEach(function(kind) {
			that.$("#available_kinds").append(new AvailableEventKindView({ 
				model: kind,
			}).render().el);
		});
		if (app.hasAuth({ auth_type: 'Event types', write: true })) { 
			this.$('.colorpicker').colorpicker().on('hide', function(ev) {
				that.model.set('color', ev.color.toHex());
				that.model.save();
			});
		}
		return this;
	},

});
