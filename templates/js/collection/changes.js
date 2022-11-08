window.Changes = Backbone.Collection.extend({

  model: Change,
  
  url: function() {
    return '/events/' + this.event.get('id') + '/changes';
  },
  
  initialize: function(models, options) {
    this.event = options.event
  },
  
});

