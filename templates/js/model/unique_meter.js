window.UniqueMeter = Backbone.Model.extend({
  url: function() {
	  return '/event_types/' + this.event.get('event_type') + '/unique';
	},
	initialize: function(data, options) {
	  this.event = options.event;
		this.event_types = options.event_types;
		this.set('ids', []);
	},
	isUnique: function() {
	  if (this.get('ids') != null) {
			for (var i = 0; i < this.get('ids').length; i++) {
				if (this.get('ids')[i] != this.event.get('id') && this.get('ids')[i] != this.event.get('recurrence_master')) {
					return false;
				}
			}
		}
		return true;
	},
	refresh: function() {
	  var current_type = this.event_types.get(this.event.get('event_type'));
		if (current_type != null && current_type.get('unique')) {
		  this.save({
			  start: this.event.get('start'),
				end: this.event.get('end'),
			});
		} else {
		  this.set('ids', []);
		}
	},
});
