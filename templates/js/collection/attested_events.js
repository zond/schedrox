window.AttestedEvents = Backbone.Collection.extend({
  model: AttestedEvent,
	revert: function(cb) {
	  var that = this;
		$.ajax(this.url, {
		  type: 'DELETE',
			dataType: 'json',
			success: function(data) {
				that.set([]);
				that.trigger('reset');
				cb();
			},
		});
	},
});
