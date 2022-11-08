window.ReportedEvents = Backbone.Collection.extend({
  url: function() {
	  return '/reported?from=' + (this.from.getISOTime() / 1000) + '&to=' + (this.to.getISOTime() / 1000);
	},
  model: ReportedEvent,
	initialize: function(models, options) {
	  this.attested = false;
		this.from = options.from;
		this.to = options.to;
		this.finished = false;
	},
	attest: function(from, to) {
	  var that = this;
		$.ajax(/^(.*)\?.*/.exec(this.url)[1], {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify({
			  from: from,
				to: to,
				events: that.models,
			}),
			success: function(data) {
			  that.attested = data.attested;
				that.models = _.map(data.events, function(ev) {
				  return new AttestableEvent(ev);
				});
				that.trigger('reset');
			},
		});
	},
	revert: function() {
	  var that = this;
		$.ajax(this.url, {
		  type: 'DELETE',
			dataType: 'json',
			success: function(data) {
			  that.attested = data.attested;
				that.models = _.map(data.events, function(ev) {
				  return new AttestableEvent(ev);
				});
				that.trigger('reset');
			},
		});
	},
	parse: function(data) {
	  this.finished = data.finished;
		return data.events;
	},
});
