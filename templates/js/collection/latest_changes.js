window.LatestChanges = Backbone.Collection.extend({

  model: Change,
  
  url: function() {
    return '/changes/latest';
  },
  
});

