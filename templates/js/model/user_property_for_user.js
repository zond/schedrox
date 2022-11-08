window.UserPropertyForUser = Backbone.Model.extend({
  initialize: function(data) {
    this.modified = false;
    this.set(this.parse(data));
  },
  parse: function(data) {
    data.assigned_at = this.convertDate(data.assigned_at);
		if (data.valid_until != null) {
		  if (data.valid_until == '0001-01-01T00:00:00Z') {
				delete(data['valid_until']);
			} else {
				data.valid_until = this.convertDate(data.valid_until);
			}
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
