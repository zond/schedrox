window.ContactEvent = Backbone.Model.extend({
	initialize: function(data) {
		this.set(this.parse(data));
	},
	parse: function(data) {
		data.start = this.convertDate(data.start);
		if (data.allDay == true && data.end == null) {
			data.end = data.start;
		}
		data.end = this.convertDate(data.end);
		return data;
	},
	convertDate: function(d) {
		if (typeof(d) == 'object') {
			return d;
		} else if (typeof(d) == 'string') {
			return new Date(Date.parse(d));
		} else if (d == null) {
			return null;
		} else {
			throw "Unknown date type for " + d
		}
	},
});
