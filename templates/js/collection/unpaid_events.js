window.UnpaidEvents = Backbone.Collection.extend({
  model: UnpaidEvent,
  url: function() {
   	return '/events/reports/unpaid?from=' + (this.from.getISOTime() / 1000) + '&to=' + (this.to.getISOTime() / 1000);
  },
  comparator: function(item) {
		var contact = item.get('contact') || {
			'organization_number': '',
      'name': '',
		};
    return [item.get('location_name'), contact.organization_number, contact.name];
  },
  initialize: function(coll, opts) {
   	this.from = opts.from;
   	this.to = opts.to;
  },
});
