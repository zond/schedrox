window.EventTypeRequiredParticipantTypes = Backbone.Collection.extend({

  model: RequiredParticipantType,

  comparator: function(a, b) {
		if (this.available_types.length == 0) {
			return 0;
		}
		var typeA = this.available_types.get(a.get('participant_type'));
		var typeB = this.available_types.get(b.get('participant_type'));
		if (typeA.get('is_contact') != typeB.get('is_contact')) {
			if (typeA.get('is_contact')) {
				return -1;
			} else {
				return 1;
			}
		}
		if (typeA.get('name') < typeB.get('name')) {
			return -1;
		} else if (typeA.get('name') > typeB.get('name')) {
			return 1;
		}
		return 0;
	},

  initialize: function(models, options) {
    this.removed = [];
		this.available_types = options.available_types;
    _.bindAll(this, 'removal');
    this.bind("remove", this.removal);
  },

  removal: function(type) {
    type.url = this.url + '/' + type.id;
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

