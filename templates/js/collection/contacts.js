window.Contacts = Backbone.Collection.extend({
  model: Contact,
  url: function() {
    if (this.query == null) {
      return "/contacts?offset=" + this.offset + "&limit=" + this.limit;
    } else {
      return '/contacts/search?q=' + this.query;
    }
  },
  initialize: function(models, options) {
    this.total = 0;
    this.offset = 0;
    this.limit = 20;
    this.query = null;
  },
  page: function(n) {
    this.query = null;
    this.offset = 0;
    this.fetch();
  },
  search: function(query, options) {
    this.query = query;
		if (options == null) {
			options = {};
		}
		options.reset = true;
    this.fetch(options);
  },
  parse: function(data) {
    if (this.query == null) {
      this.total = data.total;
      return data.results;
    } else {
      return data;
    }
  },
});
