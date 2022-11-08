window.SalaryReport = Backbone.Model.extend({
  url: function() {
	  return '/salary/report?from=' + parseInt(this.get('from').getISOTime() / 1000) + '&to=' + parseInt(this.get('to').getISOTime() / 1000);
	},
	parse: function(data) {
	  data.from = new Date(Date.parse(data.from));
	  data.to = new Date(Date.parse(data.to));
		var newEvents = _.map(data.events, function(ev) {
      ev.start = new Date(Date.parse(ev.start));
      ev.end = new Date(Date.parse(ev.end));
			return ev;
		});
		data.events = newEvents;
		return data;
	},
});
