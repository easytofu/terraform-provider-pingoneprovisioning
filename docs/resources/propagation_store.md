---
title: pingoneprovisioning_propagation_store
page_title: "Resource: pingoneprovisioning_propagation_store"
description: "Manages a PingOne provisioning propagation store. Choose the configuration block that matches the `type` you set."
slug: provider_resource_pingoneprovisioning_propagation_store
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_resources
privacy:
  view: public
position: 1
---
## Resource: pingoneprovisioning_propagation_store

Manages a PingOne provisioning propagation store. Choose the configuration block that matches the `type` you set.

## Example Usage

```terraform
resource "pingoneprovisioning_propagation_store" "scim" {
  environment_id = var.environment_id
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
  }
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.
- `name` (String) A name for the identity store.
- `type` (String) The type of the identity store. Options are `Aquera`, `AzureADSAMLV2`, `GoogleApps`, `LDAPGateway`, `PingOne`, `Salesforce`, `SalesforceContacts`, `SCIM` (alias: `scim`), `ServiceNow`, `Slack`, `Workday`, `Zoom`.

### Optional

- `description` (String) A description of the identity store.
- `image_id` (String) The image ID for the identity store resource.
- `managed` (Boolean) Indicates whether or not to enable deprovisioning of users from the target store.
- `status` (String) The status of the propagation store.
- `configuration_aquera` (Block) Aquera configuration. (see [below for nested schema](#nestedblock--configuration_aquera))
- `configuration_azure_ad_saml_v2` (Block) Azure AD SAML v2 configuration. (see [below for nested schema](#nestedblock--configuration_azure_ad_saml_v2))
- `configuration_google_apps` (Block) Google Apps configuration. (see [below for nested schema](#nestedblock--configuration_google_apps))
- `configuration_ldap_gateway` (Block) LDAP Gateway configuration. (see [below for nested schema](#nestedblock--configuration_ldap_gateway))
- `configuration_ping_one` (Block) PingOne configuration. (see [below for nested schema](#nestedblock--configuration_ping_one))
- `configuration_salesforce` (Block) Salesforce configuration. (see [below for nested schema](#nestedblock--configuration_salesforce))
- `configuration_salesforce_contacts` (Block) Salesforce Contacts configuration. (see [below for nested schema](#nestedblock--configuration_salesforce_contacts))
- `configuration_scim` (Block) SCIM configuration. (see [below for nested schema](#nestedblock--configuration_scim))
- `scim_configuration` (Block) SCIM configuration alias. (see [below for nested schema](#nestedblock--scim_configuration))
- `configuration_service_now` (Block) ServiceNow configuration. (see [below for nested schema](#nestedblock--configuration_service_now))
- `configuration_slack` (Block) Slack configuration. (see [below for nested schema](#nestedblock--configuration_slack))
- `configuration_workday` (Block) Workday configuration. (see [below for nested schema](#nestedblock--configuration_workday))
- `configuration_zoom` (Block) Zoom configuration. (see [below for nested schema](#nestedblock--configuration_zoom))

### Read-Only

- `id` (String) The unique ID of the propagation store.
- `image_href` (String) The URL for the identity store resource image file.
- `sync_status` (Object) Sync status for the propagation store. (see [below for nested schema](#nestedblock--sync_status))

<a id="nestedblock--sync_status"></a>
### Nested Schema for `sync_status`

Read-Only:

- `details` (String)
- `last_sync_time` (String)
- `next_sync_time` (String)
- `status` (String)

<a id="nestedblock--configuration_aquera"></a>
### Nested Schema for `configuration_aquera`

Optional:

- `api_key` (String)
- `api_secret` (String)
- `authentication_method` (String)
- `base_url` (String)
- `bearer_token` (String)
- `create_users` (Boolean)
- `disable_users` (Boolean)
- `domain` (String)
- `group_name_source` (String)
- `password` (String)
- `remove_action` (String)
- `update_users` (Boolean)
- `username` (String)

<a id="nestedblock--configuration_azure_ad_saml_v2"></a>
### Nested Schema for `configuration_azure_ad_saml_v2`

Optional:

- `base_url` (String)
- `bearer_token` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `remove_action` (String)
- `scim_url` (String)
- `update_users` (Boolean)

<a id="nestedblock--configuration_google_apps"></a>
### Nested Schema for `configuration_google_apps`

Optional:

- `authentication_method` (String)
- `base_url` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `domain` (String)
- `group_name_source` (String)
- `oauth_client_id` (String)
- `oauth_client_secret` (String)
- `oauth_refresh_token` (String)
- `oauth_token_url` (String)
- `remove_action` (String)
- `update_users` (Boolean)

<a id="nestedblock--configuration_ldap_gateway"></a>
### Nested Schema for `configuration_ldap_gateway`

Optional:

- `api_key` (String)
- `api_secret` (String)
- `authentication_method` (String)
- `base_url` (String)
- `bearer_token` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `ldap_gateway_id` (String)
- `ldap_gateway_region` (String)
- `password` (String)
- `remove_action` (String)
- `update_users` (Boolean)
- `username` (String)

<a id="nestedblock--configuration_ping_one"></a>
### Nested Schema for `configuration_ping_one`

Optional:

- `authentication_method` (String)
- `base_url` (String)
- `bearer_token` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `domain` (String)
- `group_name_source` (String)
- `oauth_client_id` (String)
- `oauth_client_secret` (String)
- `oauth_refresh_token` (String)
- `oauth_token_url` (String)
- `password` (String)
- `remove_action` (String)
- `scim_url` (String)
- `update_users` (Boolean)
- `username` (String)

<a id="nestedblock--configuration_salesforce"></a>
### Nested Schema for `configuration_salesforce`

Optional:

- `authentication_method` (String)
- `base_url` (String)
- `bearer_token` (String)
- `consumer_key` (String)
- `consumer_secret` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `password` (String)
- `record_type` (String)
- `remove_action` (String)
- `scim_url` (String)
- `security_token` (String)
- `update_users` (Boolean)
- `username` (String)

<a id="nestedblock--configuration_salesforce_contacts"></a>
### Nested Schema for `configuration_salesforce_contacts`

Optional:

- `authentication_method` (String)
- `base_url` (String)
- `bearer_token` (String)
- `consumer_key` (String)
- `consumer_secret` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `password` (String)
- `record_type` (String)
- `remove_action` (String)
- `scim_url` (String)
- `security_token` (String)
- `update_users` (Boolean)
- `username` (String)

<a id="nestedblock--configuration_scim"></a>
### Nested Schema for `configuration_scim`

Optional:

- `authentication_method` (String)
- `authorization_type` (String)
- `basic_auth_password` (String)
- `basic_auth_user` (String)
- `create_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `groups_resource` (String)
- `oauth_access_token` (String)
- `oauth_client_id` (String)
- `oauth_client_secret` (String)
- `oauth_token_request` (String)
- `remove_action` (String)
- `scim_url` (String)
- `scim_version` (String)
- `unique_user_identifier` (String)
- `update_users` (Boolean)
- `user_filter` (String)
- `users_resource` (String)

<a id="nestedblock--scim_configuration"></a>
### Nested Schema for `scim_configuration`

Optional:

- `authentication_method` (String)
- `authorization_type` (String)
- `basic_auth_password` (String)
- `basic_auth_user` (String)
- `create_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `groups_resource` (String)
- `oauth_access_token` (String)
- `oauth_client_id` (String)
- `oauth_client_secret` (String)
- `oauth_token_request` (String)
- `remove_action` (String)
- `scim_url` (String)
- `scim_version` (String)
- `unique_user_identifier` (String)
- `update_users` (Boolean)
- `user_filter` (String)
- `users_resource` (String)

<a id="nestedblock--configuration_service_now"></a>
### Nested Schema for `configuration_service_now`

Optional:

- `base_url` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `password` (String)
- `remove_action` (String)
- `scim_url` (String)
- `update_users` (Boolean)
- `username` (String)

<a id="nestedblock--configuration_slack"></a>
### Nested Schema for `configuration_slack`

Optional:

- `base_url` (String)
- `bearer_token` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `remove_action` (String)
- `scim_url` (String)
- `update_users` (Boolean)

<a id="nestedblock--configuration_workday"></a>
### Nested Schema for `configuration_workday`

Optional:

- `authentication_method` (String)
- `base_url` (String)
- `client_id` (String)
- `client_secret` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `password` (String)
- `remove_action` (String)
- `scim_url` (String)
- `token_url` (String)
- `update_users` (Boolean)
- `username` (String)

<a id="nestedblock--configuration_zoom"></a>
### Nested Schema for `configuration_zoom`

Optional:

- `api_key` (String)
- `api_secret` (String)
- `authentication_method` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `oauth_account_id` (String)
- `oauth_client_id` (String)
- `oauth_client_secret` (String)
- `oauth_token_url` (String)
- `remove_action` (String)
- `scim_url` (String)
- `update_users` (Boolean)

## Import

Import is supported using the following syntax:

```shell
terraform import pingoneprovisioning_propagation_store.example <environment_id>/<store_id>
```
