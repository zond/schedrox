window.Participants = Backbone.Collection.extend({

  model: Participant,
  
  url: function() {
		if (this.event.get('recurring') && this.event.get('recurrence_master')) {
			return '/events/' + this.event.get('recurrence_master') + '/participants';
		} else {
			return '/events/' + this.event.get('id') + '/participants';
		}
  },
  
  initialize: function(models, options) {
    this.removed = [];
    this.event = options.event
    _.bindAll(this, 'removal');
    this.bind("remove", this.removal);
  },
  
  removal: function(part) {
    if (this.event.get('id') != null) {
      part.url = this.url() + '/' + part.id;
      this.removed.push(part);
    }
  },

  allowedWrite: function(participant) {
    var that = this;
    return (app.hasAuth({
      auth_type: 'Participants',
      location: that.event.get('location'),
      event_kind: that.event.get('event_kind'),
      event_type: that.event.get('event_type'),
      participant_type: participant.get('participant_type'),
      write: true,
    }) || app.hasAuth({
      auth_type: 'Attend',
      location: that.event.get('location'),
      event_kind: that.event.get('event_kind'),
      event_type: that.event.get('event_type'),
      participant_type: participant.get('participant_type'),
    }));
  },

	anySwitchedDefaulted: function() {
    for (var i = 0; i < this.length; i++) {
      if (this.at(i).switched_defaulted) {
        return true;
      }
    }
    return false;
	},

  oldNews: function() {
    if (this.removed.length > 0) {
      return false;
    }
    for (var i = 0; i < this.length; i++) {
      if (this.at(i).modified || this.at(i).switched_defaulted || this.at(i).isNew()) {
        return false;
      }
    }
    return true;
  },

  setTimes: function() {
    var that = this;
    this.forEach(function(part) {
      part.set('event_start', that.event.get('start').toISOString());
      part.set('event_end', that.event.get('end').toISOString());
    });
  },

  save: function(cb) {
    var that = this;
		var saver = function() {
			if (that.length > 0) {
				var after = new cbCounter(that.length, function() {
					cb();
				});
				that.forEach(function(part) {
					if (part.isNew()) {
						if (that.allowedWrite(part)) {
							part.save(null, {
								success: after.call,
							});
						} else {
							after.call();
						}
					} else if (part.modified || part.switched_defaulted) {
						if (that.allowedWrite(part)) {
							part.save(null, {
								success: after.call,
							});
						} else {
							after.call();
						}
					} else {
						after.call();
					}
				});
			} else {
				cb();
			}
		};
		if (that.removed.length > 0) {
			var after = new cbCounter(that.removed.length, function() {
				saver();
			});
			_.each(that.removed, function(part) {
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
		} else {
			saver();
		}
  },
});

