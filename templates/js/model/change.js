window.Change = Backbone.Model.extend({
  initialize: function(data) {
    this.set(this.parse(data));
  },
  parse: function(data) {
    data.at = new Date(new Date().getTime() - (data.ago * 1000));
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
