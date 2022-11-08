window.AttestableEvents = Backbone.Collection.extend({
  model: AttestableEvent,
	initialize: function(models, options) {
		this.finished = true;
	},
	attest: function(from, to, cb) {
	  var that = this;
		var match = /\/users\/(.*)\/attestable_events(\?from=(.*)&to=(.*))/.exec(_.result(that, 'url'));
		$.ajax('/users/' + match[1] + '/attested_events' + match[2], {
			type: 'PUT',
			data: JSON.stringify(that.models),
			success: function() {
			  that.set([]);
				that.trigger('reset');
				cb();
			},
		});
	},
  parse: function(data) {
		this.finished = data.finished;
    return data.events;
  },
});
