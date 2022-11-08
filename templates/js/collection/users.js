window.Users = Backbone.Collection.extend({
	model: User,
	url: function() {
		if (this.query == '' && this.filters.length == 0) {
			return "/users";
		} else {
			var filterpart = [];
			if (this.query != '') {
				filterpart.push('q=' + encodeURIComponent(this.query));
			}
			for (var i = 0; i < this.filters.length; i++) {
				if (this.filters[i].value != '') {
					filterpart.push('filter=' + encodeURIComponent(this.filters[i].type) + ':' + encodeURIComponent(this.filters[i].value));
				}
			}
			return '/users/search?' + filterpart.join('&');
		}
	},
	initialize: function(models, options) {
		this.query = '';
		this.filters = [];
	},
	search: function(query, filters) {
		this.query = query;
		this.filters = filters;
		this.fetch({ reset: true });
	},
});
