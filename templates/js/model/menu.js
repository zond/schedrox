window.Menu = Backbone.Model.extend({
  deleteDomain: function(id) {
    var domains = this.get('domains');
    domains = _.reject(domains, function(d) {
      return d.id == id;
    });
    this.set('domains', domains);
    this.trigger('change');
  },

  addDomain: function(d) {
    this.get('domains').push(d);
    this.trigger('change');
  },
});
