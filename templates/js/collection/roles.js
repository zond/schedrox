window.Roles = Backbone.Collection.extend({

  model: Role,

  initialize: function(models, options) {
    this.removed = [];
    _.bindAll(this, 'removal');
    this.bind("remove", this.removal);
  },

  removal: function(role) {
    role.url = this.url + '/' + role.id;
    this.removed.push(role);
  },

  allowedWrite: function(r) {
    return app.hasAuth({
      auth_type: 'Roles',
      write: true,
		  role: r.get('name'),
    });
  },

  save: function(cb) {
    var that = this;
		if (this.removed.length + this.length == 0) {
			cb();
		} else {
			var after = new cbCounter(this.removed.length + this.length, cb);
			_.each(this.removed, function(part) {
				if (part.isNew()) {
					after.call();
				} else {
					if (that.allowedWrite(part)) {
						part.destroy({
							success: after.call,
						});
					} else {
						after.call();
					}
				}
			});
			this.forEach(function(part) {
				if (part.isNew()) {
					if (that.allowedWrite(part)) {
						part.save(null, {
							success: after.call,
						});
					} else {
						after.call();
					}
				} else if (part.modified) {
					if (that.allowedWrite(part)) {
						part.destroy({
							success: function() {
								part.set('id', null);
								part.save(null, {
								  success: after.call,
								});
							},
						});
					} else {
						after.call();
					}
				} else {
					after.call();
				}
      });
    }
  },

});
