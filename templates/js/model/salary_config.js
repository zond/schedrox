window.SalaryConfig = Backbone.Model.extend({
  url: '/salary/config',
  initialize: function(data) {
    this.set(this.parse(data));
  },
	parse: function(data) {
		if (data.salary_event_kind_properties == null) {
		  data.salary_event_kind_properties = [];
		}
		if (data.salary_event_type_properties == null) {
		  data.salary_event_type_properties = [];
		}
		if (data.salary_location_properties == null) {
		  data.salary_location_properties = [];
		}
		if (data.salary_participant_type_properties == null) {
		  data.salary_participant_type_properties = [];
		}
		if (data.salary_user_properties == null) {
		  data.salary_user_properties = [];
		}
    return data;
  },
});
