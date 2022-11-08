window.UsersResultsView = Backbone.View.extend({

	template: _.template($('#users_results_underscore').html()),

	events: {
		"change #new_user": "addUser",
		"click #toggle_select_user_for_modification": "toggleSelections",
		"click .select-user-for-modification": "userSelected",
		"click #available_modification_modes a": "changeModificationMode",
		"click #available_modifications a": "changeModification",
		"click #execute_modification": "executeModification",
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'addUser');
		this.collection.bind("change", this.render);
		this.collection.bind("reset", this.render);
		this.collection.bind("add", this.render);
		this.collection.bind("remove", this.render);
		this.user_properties_for_domain = options.user_properties_for_domain;
		this.filtersView = options.filtersView;
		this.available_properties = options.available_properties;
		this.available_roles = options.available_roles;
		this.checked = {};
		this.allChecked = false;
		this.anyChecked = false;
		this.modificationMode = 'add_role';
		this.modificationName = '';
	},

	executeModification: function(ev) {
		var that = this;
		var users = [];
		for (var id in that.checked) {
			users.push(id);
		}
		var data = {
			users: users,
		};
		if (this.modificationMode == 'add_role') {
			data.add = true;
			data.role = that.modificationName;
			data.property = "";
		} else if (this.modificationMode == 'remove_role') {
			data.add = false;
			data.role = that.modificationName;
			data.property = "";
		} else if (this.modificationMode == 'add_property') {
			data.add = true;
			data.role = "";
			data.property = that.modificationName;
		} else if (this.modificationMode == 'remove_property') {
			data.add = false;
			data.role = "";
			data.property = that.modificationName;
		}
		$.ajax('/users/bulk/update', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify(data),
			success: function(data) {
				alert("{{.I "Done" }}");
			},
		});
	},

	changeModification: function(ev) {
		if (this.modificationMode.indexOf('role') != -1) {
			this.modificationName = $(ev.target).attr('data-role-name');
		} else if (this.modificationMode.indexOf('property') != -1) {
			this.modificationName = $(ev.target).attr('data-user-property-name');
		}
		this.render();
	},

	changeModificationMode: function(ev) {
		this.modificationMode = $(ev.target).attr('data-modification-mode');
		this.modificationName = '';
		this.render();
	},

	userSelected: function(ev) {
		if (!!$(ev.target).attr('checked')) {
			this.checked[$(ev.target).attr('data-user-id')] = true;
		} else {
			delete(this.checked[$(ev.target).attr('data-user-id')]);
		}
		this.anyChecked = false;
		for (var id in this.checked) {
			this.anyChecked = true;
		}
		this.render();
	},

	toggleSelections: function(ev) {
		if (!!$(ev.target).attr('checked')) {
			this.checkAll();
		} else {
			this.uncheckAll();
		}
		this.render();
	},

	checkAll: function() {
		var that = this;
		that.checked = {};
		that.allChecked = true;
		that.anyChecked = true;
		this.$('.select-user-for-modification').each(function(index, el) {
    	$(el).attr('checked', true);
    	that.checked[$(el).attr('data-user-id')] = true;
		});
	},

	uncheckAll: function() {
		var that = this;
		this.$('.select-user-for-modification').each(function(index, el) {
			$(el).attr('checked', false);
		});
		that.allChecked = false;
		that.anyChecked = false;
		that.checked = {};
	},

	addUser: function(ev) {
		if (app.getDomain() != null) {
			var that = this;
			if ($(ev.target).val() == '' || $(ev.target).val().isEmail()) {
				var newUser = new User({ email: $(ev.target).val() });
				newUser.save(null, {
					success: function() {
						that.collection.add(newUser);
						that.render();
					}
				});
			} else {
				myalert('{{.I "{0} is not a valid email address." }}'.format($(ev.target).val()));
			}
		}
	},
	render: function() {
		if (this.modificationName == '') {
			if (this.modificationMode.indexOf('role') != -1 && this.available_roles.length > 0) {
				this.modificationName = this.available_roles.at(0).get('name');
			} else if (this.modificationMode.indexOf('property') != -1 && this.available_properties.length > 0) {
				this.modificationName = this.available_properties.at(0).get('name');
			}
		}
		var that = this;
		var attestFilter = that.filtersView.attestFilter();
		this.$el.html(this.template({
			emails: _.collect(this.collection.models, function(u) {
				return u.get('email');
			}),
			allChecked: that.allChecked,
			anyChecked: that.anyChecked,
			modificationMode: that.modificationMode,
			modificationName: that.modificationName,
			attestFilter: attestFilter,
		}));
		if (this.anyChecked) {
			if (this.modificationMode.indexOf('role') != -1) {
				this.available_roles.forEach(function(role) {
					that.$('#available_modifications').append(new AvailableRoleView({ model: role }).render().el);
				});
			} else {
				this.available_properties.forEach(function(property) {
					that.$('#available_modifications').append(new AvailableUserPropertyView({ model: property }).render().el);
				});
			}
		}
		this.collection.forEach(function(user) {
			that.$("#user_list").append(new UserView({ 
				model: user,
				available_roles: that.available_roles,
				user_properties_for_domain: that.user_properties_for_domain,
				checked: that.checked,
				attestFilter: attestFilter,
			}).render().el);
		});
		return this;
	}

});
