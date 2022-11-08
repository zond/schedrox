window.EventRequiredParticipantTypes = Backbone.Collection.extend({

  model: RequiredParticipantType,
  
  url: function() {
    if (this.ev.get('location') != null) {
      return '/event_types/' + this.ev.get('event_type') + '/required_participant_types/' + this.ev.get('location');
    } else {
      return '/event_types/' + this.ev.get('event_type') + '/required_participant_types';
    }
  },
  
  initialize: function(models, options) {
    _.bindAll(this, 'removal', 'refetch');
    this.removed = [];
    this.ev = options.event;
    this.ev.bind("change", this.refetch);
    this.bind("remove", this.removal);
    this.last_event_type = this.ev.get('event_type');
  },

  refetch: function() {
    if (this.ev.get('event_type') != null && this.ev.get('event_type') != this.last_event_type) {
      this.last_event_type = this.ev.get('event_type');
      this.fetch();
    }
  },

  removal: function(type) {
    type.url = this.url() + '/' + type.id;
    this.removed.push(type);
  },

  save: function() {
    var that = this;
    _.each(this.removed, function(type) {
      type.destroy();
    });
    this.forEach(function(type) {
      if (type.isNew()) {
	type.save();
      } else if (type.modified) {
	type.destroy({
	  success: function() {
	    type.set('id', null);
	    type.save();
	  },
	});
      }
    });
  },
});

