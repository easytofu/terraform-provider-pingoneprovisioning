resource "pingoneprovisioning_propagation_store" "example" {
  environment_id = "00000000-0000-0000-0000-000000000000"
  name           = "Example SCIM Store"
  type           = "SCIM"

  configuration_scim {
    authentication_method  = "OAuth 2.0"
    authorization_type     = "Bearer"
    scim_url               = "https://example.com/scim/v2"
    scim_version           = "2.0"
    unique_user_identifier = "userName"
    user_filter            = "active eq true"
    users_resource         = "Users"
    groups_resource        = "Groups"
    oauth_access_token     = "example-access-token"
  }
}
