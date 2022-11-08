window.ExtraRequiredParticipantTypes = Backbone.Collection.extend({

  model: RequiredParticipantType,
  
  url: function() {
    return '/events/' + this.event.get('id') + '/required_participant_types/';
  },
  
  initialize: function(models, options) {
    _.bindAll(this, 'removal');
    this.removed = [];
    this.event = options.event;
    this.bind("remove", this.removal);
  },

  removal: function(type) {
    type.url = this.url() + '/' + type.id;
    this.removed.push(type);
  },

  allowedWrite: function(req) {
    var that = this;
    return app.hasAuth({
      auth_type: 'Participants',
      location: that.event.get('location'),
      event_kind: that.event.get('event_kind'),
      event_type: that.event.get('event_type'),
      participant_type: req.get('participant_type'),
      write: true,
    });
  },

  oldNews: function() {
    if (this.removed.length > 0) {
      return false;
    }
    for (var i = 0; i < this.length; i++) {
      if (this.at(i).modified || this.at(i).isNew()) {
        return false;
      }
    }
    return true;
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
    }
  },
});

