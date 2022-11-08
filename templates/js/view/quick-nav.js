window.QuickNavView = Backbone.View.extend({

	template: _.template($('#quick_nav_underscore').html()),

	tagName: 'span',

	events: {
	  "change #goto_date": "gotoDate",
		"change #goto_week": "gotoWeek",
		"click .back-to-today": "backToToday",
	},

	initialize: function(options) {
		_.bindAll(this, 'render');
		this.parent = options.parent;
	},

	backToToday: function(ev) {
	  ev.preventDefault();
		this.parent.date = new Date();
		this.parent.firstDay = null;
		this.parent.render();
	},

	gotoDate: function(ev) {
	  var toDate = new Date(parseInt(ev.val));
		this.parent.date = toDate;
		this.parent.firstDay = toDate.getDay();
		this.parent.render();
	},

  gotoWeek: function(ev) {
	  var toDate = new Date();
		for (var i = 0; i < 54; i++) {
			if (toDate.getWeek() == parseInt(ev.val)) {
			  break;
			}
			toDate = new Date(toDate.getFullYear(), toDate.getMonth(), toDate.getDate() + 7);
		}
		this.parent.date = toDate;
		this.parent.firstDay = {{.I "firstDOW" }};
		this.parent.render();
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({ 
		}));
		that.$('#goto_week').select2({
		  placeholder: '{{.I "Week" }}',
		  query: function(q) {
			  var data = [];
				var wn = new Date().getWeek();
				for (var i = wn; i < 54; i++) {
				  if (('' + i).indexOf(q.term) == 0) {
						data.push({id: '' + i, text: '' + i});
					}
				}
				for (var i = 1; i < wn; i++) {
				  if (('' + i).indexOf(q.term) == 0) {
						data.push({id: '' + i, text: '' + i});
					}
				}
				q.callback({results: data});
			},
		});
		that.$('#goto_date').select2({
		  placeholder: '{{.I "Date" }}',
			query: function(q) {
			  var data = [];
				var today = new Date();
				for (var i = 0; i < 400; i++) {
				  var d = new Date(today.getFullYear(), today.getMonth(), today.getDate() + i);
				  var dateString = anyDateConverter.format(d);
          if (dateString.indexOf(q.term) != -1) {
					  data.push({id: d.getTime(), text: dateString});
					}
				}
				for (var i = 1; i < 400; i++) {
				  var d = new Date(today.getFullYear(), today.getMonth(), today.getDate() - i);
				  var dateString = anyDateConverter.format(d);
          if (dateString.indexOf(q.term) != -1) {
					  data.push({id: d.getTime(), text: dateString});
					}
				}
				q.callback({results: data});
			},
		});
		return that;
	},

});
