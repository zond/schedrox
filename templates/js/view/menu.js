window.MenuView = Backbone.View.extend({

  template: _.template($('#menu_underscore').html()),

  events: {
    "click .nav-link": "navigate",
    "click .domain-link": "changeDomain",
  },

  initialize: function(options) {
    this.app = options.app;
    _.bindAll(this, 'render');
    this.model.bind("change", this.render);
    if (this.app.user != null) {
      this.app.user.bind('change', this.render);
    }
  },

  navigate: function(event) {
    event.preventDefault();
    app.navigate($(event.currentTarget).attr("href"), { trigger: true });
  },

  changeDomain: function(event) {
    event.preventDefault();
		var newDomain = new Domain({
			id: $(event.target).attr('data-id'),
			name: $(event.target).attr('data-name'),
		});
		for (var i = 0; i < app.user.get('domains').length; i++) {
		  if (app.user.get('domains')[i].id == newDomain.get('id')) {
			  newDomain = new Domain(app.user.get('domains')[i]);
			}
		}
    app.setDomain(newDomain);
		this.render();
  },

  addClassFor: function(data, name) {
    if (this.model.get('active') == name) {
      data[name] = 'active';
    } else {
      data[name] = '';
    }
  },

  render: function() {
    var classes = {}
    this.addClassFor(classes, 'events');
    this.addClassFor(classes, 'settings');
    this.addClassFor(classes, 'calendar');
    this.addClassFor(classes, 'users');
    this.addClassFor(classes, 'contacts');
		this.addClassFor(classes, 'salaries');
    this.$el.html(this.template({
      classes: classes,
      model: this.model,
      app: this.app,
    }));
    if (window.app != null && window.app.getDomain() != null) {
		  var dom = window.app.getDomain();
      $("#selected_domain").text(dom.get('name'));
			if (dom.get('closed_and_redirected_to') != '') {
			  myalert('{{.I "{0} has moved to <a href=\"{1}\">{1}</a>." }}'.format(dom.get('name'), dom.get('closed_and_redirected_to')));
			}
    }
    this.delegateEvents();
    return this;
  }

});
