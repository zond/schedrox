window.TopNavView = Backbone.View.extend({

  template: _.template($('#top_nav_underscore').html()),

  events: {
	 	"click .domain_link": "changeDomain",
    "click .filter_link": "changeFilter",
	},

  initialize: function(options) {
		_.bindAll(this, 'render');
		this.model.bind("change", this.render);
		this.model.bind("reset", this.render);
		window.session.custom_filters.bind('reset', this.render);
		window.session.user.bind('change', this.render);
  },

  changeFilter: function(ev) {
		ev.preventDefault();
		var id = $(ev.target).attr('data-filter-id');
		if (id == "_all_") {
			window.session.menu.set('active_filter', new CustomFilter({
				name: '{{.I "Filter" }}',
			}));
		} else {
			window.session.menu.set('active_filter', window.session.custom_filters.get(id));
		}
	},

  changeDomain: function(ev) {
		ev.preventDefault();
		window.session.menu.set('domain', new Domain(_.find(window.session.user.get('domains'), function(dom) {
			return dom.id == $(ev.target).attr('data-domain-id');
		})));
		window.session.app.reloader();
	},

  render: function() {
		var title = {
			'events/mine': '{{.I "My events" }}',
			'events/open': '{{.I "Open events" }}',
		}[this.model.get('active')];
    this.$el.html(this.template({
			domains: (window.session.user.get('domains') || []),
		  domain: window.session.menu.get('domain'),
		  active: this.model.get('active'),
			title: title,
  		filters: window.session.custom_filters,
  		active_filter: window.session.menu.get('active_filter'),
		})); 
    return this;
  },

});

