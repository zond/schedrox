window.RequiredParticipantTypeView = Backbone.View.extend({

	tagName: 'tr',

	template: _.template($('#required_participant_type_underscore').html()),

	events: {
		"click .close": "remove_type",
		"click .available-per-participant-type": "set_per_type",
		"change .participant-type-min": "set_min",
		"change .participant-type-offset": "set_offset",
		"change .participant-type-max": "set_max",
		"change .participant-type-per-num": "set_per_num",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.deps = options.deps;
		this.model.bind("change", this.render);
		this.available_participant_types = options.available_participant_types;
		this.collection.bind("change", this.render);
		this.collection.bind("reset", this.render);
		this.collection.bind("add", this.render);
		this.collection.bind("remove", this.render);
	},

	remove_type: function(event) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			this.collection.remove(this.model);
		}
	},

	set_per_type: function(event) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var type_id = $(event.target).attr('data-participant-type-id');
			if (type_id == '') {
				this.model.set('per_type', null);
			} else {
				this.model.set('per_type', type_id);
			}
			this.model.modified = true;
		}
	},

	set_per_num: function(event) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var new_per_num = parseInt($(event.target).val());
			if (new_per_num > 0) {
				this.model.set('per_num', new_per_num, { silent: true });
				this.model.modified = true;
			} else {
				$(event.target).val("1");
			}
		}
	},

	set_max: function(event) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var new_max = parseInt($(event.target).val());
			if (new_max > -1) {
				this.model.set('max', new_max, { silent: true });
				if (this.model.get('min') > new_max) {
					this.model.set('min', new_max, { silent: true });
					this.$('.participant-type-min').val(new_max);
				}
				this.model.modified = true;
			} else {
				$(event.target).val("0");
			}
		}
	},

	set_offset: function(event) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var new_min = parseInt($(event.target).val());
			this.model.set('min', new_min, { silent: true });
			this.model.modified = true;
		}
	},

set_min: function(event) {
		if (app.hasAuth({
			auth_type: 'Event types',
			write: true,
		})) {
			var new_min = parseInt($(event.target).val());
			if (new_min > -1) {
				this.model.set('min', new_min, { silent: true });
				if (this.model.get('max') < new_min) {
					this.model.set('max', new_min, { silent: true });
					this.$('.participant-type-max').val(new_min);
				}
				this.model.modified = true;
			} else {
				$(event.target).val("0");
			}
		}
	},

	render: function() {
		var that = this;
		this.$el.html(this.template({ 
			model: this.model,
			per_type: this.available_participant_types.get(this.model.get('per_type')),
			participant_type: this.available_participant_types.get(this.model.get('participant_type')),
		}));
		this.$('.per-tooltip').tooltip({
		  title: '{{.I "The required number will be calculated as Max(0, [nr present of dependent type] - [offset]) / [scaling factor]" }}',
			html: true,
		});
		this.$(".available-event-participant-types").append(new AvailableParticipantTypeView({ 
			model: new ParticipantType({ id: null, name: '{{.I "Event"}}' }),
			klass: 'available-per-participant-type',
		}).render().el);
		this.collection.forEach(function(type) {
			if (type.get('participant_type') != that.model.get('participant_type') && that.deps[type.get('participant_type')] != that.model.get('participant_type')) {
				var real_type = that.available_participant_types.get(type.get('participant_type'));
				if (real_type != null) {
					that.$(".available-event-participant-types").append(new AvailableParticipantTypeView({ 
						model: real_type, 
						klass: 'available-per-participant-type',
					}).render().el);
				}
			}
		});
		return this;
	},

});
