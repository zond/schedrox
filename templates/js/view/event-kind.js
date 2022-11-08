window.EventKindView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#event_kind_underscore').html()),

	events: {
		"click .close": "removeKind",
		"click .open-kind": "openKind",
		"change .colorpicker input": "changeColor",
	}, 

	openKind: function(event) {
		event.preventDefault();
		new EventKindDetailsView({ model: this.model }).modal(function() {
			app.navigate('/events/types');
		});
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

	removeKind: function(event) {
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

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
	},

	render: function() {
		var that = this;
		var col = this.model.get('color');
		if (col == '') {
			col = '#ffffff';
		}
		this.$el.html(this.template({ 
			model: this.model,
			col: col,
		}));
		if (app.hasAuth({ auth_type: 'Event types', write: true })) { 
			this.$('.colorpicker').colorpicker().on('hide', function(ev) {
				that.model.set('color', ev.color.toHex());
				that.model.save();
			});
		}
		return this;
	},

});
