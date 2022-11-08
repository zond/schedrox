window.RequiredParticipantTypesView = Backbone.View.extend({

  template: _.template($('#required_participant_types_underscore').html()),

  events: {
    "click .available-participant-type": "addParticipantType",
  },

  initialize: function(options) {
    _.bindAll(this, 'render');
    this.collection.bind("change", this.render);
    this.collection.bind("reset", this.render);
    this.collection.bind("add", this.render);
    this.collection.bind("remove", this.render);
    this.available_participant_types = options.available_participant_types;
    this.available_participant_types.bind("change", this.render);
    this.available_participant_types.bind("reset", this.render);
    this.available_participant_types.bind("add", this.render);
    this.available_participant_types.bind("remove", this.render);
  },

  addParticipantType: function(event) {
    event.preventDefault();
    var newParticipantType = new RequiredParticipantType({ 
      participant_type: $(event.target).attr('data-participant-type-id'),
      per_num: 0,
      min: 0,
      max: 0,
    });
    this.collection.add(newParticipantType);
  },

  render: function() {
    var that = this;
    this.$el.html(this.template({}));
    var deps = {};
    this.collection.forEach(function(participant_type) {
      if (participant_type.get('per_type') != null) {      
				deps[participant_type.get('id')] = participant_type.get('per_type');
			}
		});
		this.collection.forEach(function(participant_type) {
			this.$('#participant_type_list').append(new RequiredParticipantTypeView({ 
				model: participant_type,
				collection: that.collection,
				available_participant_types: that.available_participant_types,
				deps: deps,
			}).render().el);
		});
    this.available_participant_types.forEach(function(participant_type) {
      that.$('#available_participant_types').append(new AvailableParticipantTypeView({ model: participant_type }).render().el);
    });
    dbg = this.collection;
    return this;
  }

});
