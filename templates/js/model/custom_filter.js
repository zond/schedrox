window.CustomFilter = Backbone.Model.extend({
  urlRoot: '/custom_filters',

  loadFromLocalStorage: function() {
		if (app.getDomain() != null) {
			var locations = JSON.parse(window.localStorage.getItem('calendar-location-filter-' + app.getDomain().get('id')));
			var kinds = JSON.parse(window.localStorage.getItem('calendar-kind-filter-' + app.getDomain().get('id')));
			var types = JSON.parse(window.localStorage.getItem('calendar-type-filter-' + app.getDomain().get('id')));
			var users = JSON.parse(window.localStorage.getItem('calendar-user-filter-' + app.getDomain().get('id')));
			if (locations != null && locations.length > 0) {
				this.set('locations', locations, { silent: true });
			} else {
				this.set('locations', [], { silent: true });
			}
			if (kinds != null && kinds.length > 0) {
				this.set('kinds', kinds, { silent: true });
			} else {
				this.set('kinds', [], { silent: true });
			}
			if (types != null && types.length > 0) {
				this.set('types', types, { silent: true });
			} else {
				this.set('types', [], { silent: true });
			}
			if (users != null && users.length > 0) {
				this.set('users', users, { silent: true});
			} else {
				this.set('users', [], { silent: true });
			}
		}
	},

  storeInLocalStorage: function() {
		if (app.getDomain() != null) {
			window.localStorage.setItem('calendar-location-filter-' + app.getDomain().get('id'), JSON.stringify(this.get('locations')));
			window.localStorage.setItem('calendar-kind-filter-' + app.getDomain().get('id'), JSON.stringify(this.get('kinds')));
			window.localStorage.setItem('calendar-type-filter-' + app.getDomain().get('id'), JSON.stringify(this.get('types')));
			window.localStorage.setItem('calendar-user-filter-' + app.getDomain().get('id'), JSON.stringify(this.get('users')));
		}
	},
});
