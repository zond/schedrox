window.UserPropertiesForDomain = Backbone.Collection.extend({
  model: UserPropertyForDomain,
  url: "/user_properties",
});
