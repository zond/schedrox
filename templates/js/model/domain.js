window.Domain = Backbone.Model.extend({
  urlRoot: '/domains',
  initialize: function(data) {
    this.set(this.parse(data));
  },
	parse: function(data) {
	  if (data.salary_config == null) {
		  data.salary_config = {};
		}
	  if (data.earliest_event != null) {
			data.earliest_event = this.convertDate(data.earliest_event);
		}
    if (data.latest_event != null) {
			data.latest_event = this.convertDate(data.latest_event);
		}
		if (data.salary_properties == null) {
		  data.salary_properties = {};
		}
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
