window.EventUsersView = Backbone.View.extend({

	template: _.template($('#event_users_underscore').html()),

	events: {
		"click .available-user": "set_user",
		"change #report_from": "set_from",
		"change #report_to": "set_to",
	},

	initialize: function() {
		_.bindAll(this, 'render');
		this.from = new Date();
		this.to = new Date();
		this.users = new Users();
		this.users.fetch({ reset: true });
		this.users.bind("reset", this.render);
		this.user = new User({
			given_name: '',
			family_name: '{{.I "All" }}',
			id: -1,
		});
	},

	set_from: function(event) {
		this.from = anyDateConverter.parse($(event.target).val());
		this.$('#unix_start').val(this.from.getISOTime() / 1000);
	},

	set_to: function(event) {
		this.to = anyDateConverter.parse($(event.target).val());
		this.$('#unix_to').val(this.to.getISOTime() / 1000);
	},

	set_user: function(event) {
		event.preventDefault();
		var user_id = $(event.target).attr('data-user-id');
		this.user = new User({
			given_name: $(event.target).attr('data-given-name'),
			family_name: $(event.target).attr('data-family-name'),
			id: $(event.target).attr('data-id'),
		});
		this.$('#user_id').val(this.user.get('id'));
		this.render();
	},
	
	render: function() {
		var that = this;
		that.$el.html(that.template({ 
			from: that.from,
			to: that.to,
			users: that.users,
			user: that.user,
		}));
		// make AnyTime fix pickers a bit later
		setTimeout(function() {
			var options = {
				askSecond: false,
				dayAbbreviations: {{.I "day_names_short"}},
				dayNames: {{.I "day_names"}},
				firstDOW: {{.I "firstDOW"}},
				labelDayOfMonth: '{{.I "labelDayOfMonth"}}',
				labelHour: '{{.I "labelHour"}}',
				labelMinute: '{{.I "labelMinute"}}',
				labelMonth: '{{.I "labelMonth"}}',
				labelTitle: '{{.I "labelTitle"}}',
				labelYear: '{{.I "labelYear"}}',
				monthAbbreviations: {{.I "month_names_short"}},
				monthNames: {{.I "month_names"}},
				format: '{{.I "any_date_format" }}',
			};
			that.$('#report_from').AnyTime_noPicker().AnyTime_picker(options);
			that.$('#report_to').AnyTime_noPicker().AnyTime_picker(options);
		}, 500);
		that.$("#available_users").append(new AvailableUserView({ model: new User({given_name: '', family_name: '{{.I "All" }}', id: -1 }) }).render().el);
		that.users.each(function(user) {
			that.$("#available_users").append(new AvailableUserView({ model: user }).render().el);
		}); 
		return that;
	},

});
