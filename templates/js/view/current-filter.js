window.CurrentFilterView = Backbone.View.extend({

	template: _.template($('#current_filter_underscore').html()),

	events: {
		"change #locations_filter": "changeLocations",
		"change #kinds_filter": "changeKinds",
		"change #types_filter": "changeTypes",
    "change #users_filter": "changeUsers",
		"click #save_calendar_filter": "saveFilters",
    "mouseover": "expand",
    "mouseout": "contract",
	},

  contract: function(ev) {
		this.expanded = false;
		this.$('.current-filter').height('36');
	},

  expand: function(ev) {
		this.expanded = true;
		this.$('.current-filter').removeAttr('style');
	},

	initialize: function(options) {
		_.bindAll(this, 'render', 'formatUser', 'textForUser', 'findUsers', 'initUser');
		this.event_kinds = options.event_kinds;
		this.event_types = options.event_types;
		this.locations = options.locations;
		this.refetch = options.refetch;
		this.model.bind('change', this.render);
		this.model.bind('change', this.refetch);
		this.usercache = options.usercache;
		this.expanded = false;
		var that = this;
		_.each([this.event_kinds, this.event_types, this.locations], function(coll) {
			coll.bind("change", that.render);
			coll.bind("reset", that.render);
			coll.bind("add", that.render);
			coll.bind("remove", that.render);
		});
	},

	changeUsers: function(ev) {
		this.model.set('users', $(ev.target).val().split(","));
		this.model.storeInLocalStorage();
	},

	changeLocations: function(ev) {
		this.model.set('locations', $(ev.target).val());
		this.model.storeInLocalStorage();
	},

	changeKinds: function(ev) {
		this.model.set('kinds', $(ev.target).val());
		this.model.storeInLocalStorage();
	},

	changeTypes: function(ev) {
		this.model.set('types', $(ev.target).val());
		this.model.storeInLocalStorage();
	},

	saveFilters: function(ev) {
		var that = this;
		var newModel = new CustomFilter({
			locations: that.model.get('locations'),
			kinds: that.model.get('kinds'),
			types: that.model.get('types'),
			users: that.model.get('users'),
		});
		mymodal(new NameCustomFilterView({
			model: newModel,
		}).render().el, {
			"{{.I "Save"}}": function() {
				newModel.save(null, {
					success: function() {
						that.collection.add(newModel);
					},
				});
			},
		}, { min_height: '5%', min_width: '5%' });
	},

	textForUser: function(part) {
		var rval = []
		if (part.given_name != '' && part.family_name != '') {
			rval.push("{{.I "name_order" }}".format(part.given_name, part.family_name))
		} else if (part.given_name != '') {
			rval.push(part.given_name);
		} else if (part.family_name != '') {
			rval.push(part.family_name);
		}
		return rval.join(', ');
	},

	formatUser: function(part) {
		var that = this;
		var text = that.textForUser(part);
		if (that.current_term != null && that.current_term != '') {
			var values = that.current_term.split(" ")
			for (var i = 0; i < values.length; i++) {
				var value = values[i];
				if (value.length > 1) {
					var matchStart = text.toLowerCase().indexOf(value.toLowerCase());
					if (matchStart != -1) {
						var matchEnd = matchStart + value.length;
						text = text.substr(0, matchStart) + '<strong>' + text.substr(matchStart, value.length) + '</strong>' + text.substr(matchEnd);
					}
				}
			}
		}
		return '<img class="gravatar-small" src="' + gravatarImage(part.gravatar_hash, {s: 20}) + '">' + text;
	},

	findUsers: function(options) {
		var that = this;
		that.current_term = options.term;
		$.ajax('/users/search?q=' + encodeURIComponent(that.current_term), {
			type: 'GET',
			dataType: 'json',
			success: function(data) {
				_.each(data, function(user) {
					that.usercache[user.id] = user;
				});
				options.callback({
					results: data,
				});
			},
		});
	},

	initUser: function(el, cb) {
		var that = this;
		if (el.val() == '') {
			cb([]);
			return;
		}
		var found = 0;
		var ids = el.val().split(",");
		var done = function() {
			var result = [];
			_.each(ids, function(id) {
				result.push(that.usercache[id]);
			});
			cb(result);
		};
		_.each(ids, function(id) {
			var user = that.usercache[id];
			if (user == null) {
				$.ajax('/users/' + id, {
					type: 'GET',
					dataType: 'json',
					success: function(data) {
						that.usercache[id] = data;
						found++;
						if (found == ids.length) {
							done();
						}
					},
				});
			} else {
				found++;
				if (found == ids.length) {
					done();
				}
			}
		});
	},

	render: function() {
	  var that = this;
		that.$el.html(that.template({ model: that.model }));
		that.locations.forEach(function(location) {
			that.$('#locations_filter').append('<option value="' + location.get('id') + '">' + location.get('name') + '</option>');
		});
		that.$('#locations_filter').select2({
		  placeholder: '{{.I "Location" }}',		
		})
		if (that.model.get('locations') != null && that.model.get('locations').length > 0) {
			that.$('#locations_filter').select2('val', that.model.get('locations'));
		}

		that.event_kinds.forEach(function(kind) {
			that.$('#kinds_filter').append('<option value="' + kind.get('id') + '">' + kind.get('name') + '</option>');
		});
		that.$('#kinds_filter').select2({
			placeholder: '{{.I "Kind" }}',
		});
	  if (that.model.get('kinds') != null && that.model.get('kinds').length > 0) {
      that.$('#kinds_filter').select2('val', that.model.get('kinds'));
		}

		that.event_types.forEach(function(type) {
			that.$('#types_filter').append('<option value="' + type.get('id') + '">' + type.get('name') + '</option>');
		});
		that.$('#types_filter').select2({
		  placeholder: '{{.I "Type" }}',	
		});
		if (that.model.get('types') != null && that.model.get('types').length > 0) {
			that.$('#types_filter').select2('val', that.model.get('types'));
		}

		that.$('#users_filter').select2({
			query: that.findUsers,
			placeholder: '{{.I "User" }}',
			initSelection: that.initUser,
			minimumInputLength: 2,
			formatResult: that.formatUser,
			formatSelection: that.formatUser,
			multiple: true,
		});
		if (that.model.get('users') != null && that.model.get('users').length > 0) {
			that.$('#users_filter').select2('val', that.model.get('users'));
		}

		that.delegateEvents();
		if (!that.expanded) {
			that.$('.current-filter').height(36);
		}
		return that;
	},

});
