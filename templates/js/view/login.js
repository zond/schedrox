window.LoginView = Backbone.View.extend({

  template: _.template($('#login_underscore').html()),

  initialize: function(options) {
  },

  render: function() {
    this.$el.html(this.template({})); 
    return this;
  },

});
