---
title: pingoneprovisioning_propagation_store
page_title: "Data Source: pingoneprovisioning_propagation_store"
description: "Fetches a PingOne provisioning propagation store by ID or by name and type."
slug: provider_datasource_pingoneprovisioning_propagation_store
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_datasources
privacy:
  view: public
position: 1
---
## Data Source: pingoneprovisioning_propagation_store

Fetches a PingOne provisioning propagation store by ID or by name and type.

## Example Usage

```terraform
data "pingoneprovisioning_propagation_store" "by_name" {
  environment_id = var.environment_id
  name           = "Example SCIM Store"
  type           = "SCIM"
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.

### Optional

- `id` (String) The unique ID of the propagation store.
- `name` (String) The name of the identity store.
- `type` (String) The type of the identity store.

### Read-Only

- `description` (String) A description of the identity store.
- `image_id` (String) The image ID for the identity store resource.
- `image_href` (String) The URL for the identity store resource image file.
- `managed` (Boolean) Indicates whether or not to enable deprovisioning of users from the target store.
- `status` (String) The status of the propagation store.
- `sync_status` (Object) Sync status for the propagation store. (see [below for nested schema](#nestedatt--sync_status))
- `configuration_aquera` (Block) Aquera configuration. (see [below for nested schema](#nestedatt--configuration_aquera))
- `configuration_azure_ad_saml_v2` (Block) Azure AD SAML v2 configuration. (see [below for nested schema](#nestedatt--configuration_azure_ad_saml_v2))
- `configuration_google_apps` (Block) Google Apps configuration. (see [below for nested schema](#nestedatt--configuration_google_apps))
- `configuration_ldap_gateway` (Block) LDAP Gateway configuration. (see [below for nested schema](#nestedatt--configuration_ldap_gateway))
- `configuration_ping_one` (Block) PingOne configuration. (see [below for nested schema](#nestedatt--configuration_ping_one))
- `configuration_salesforce` (Block) Salesforce configuration. (see [below for nested schema](#nestedatt--configuration_salesforce))
- `configuration_salesforce_contacts` (Block) Salesforce Contacts configuration. (see [below for nested schema](#nestedatt--configuration_salesforce_contacts))
- `configuration_scim` (Block) SCIM configuration. (see [below for nested schema](#nestedatt--configuration_scim))
- `scim_configuration` (Block) SCIM configuration alias. (see [below for nested schema](#nestedatt--scim_configuration))
- `configuration_service_now` (Block) ServiceNow configuration. (see [below for nested schema](#nestedatt--configuration_service_now))
- `configuration_slack` (Block) Slack configuration. (see [below for nested schema](#nestedatt--configuration_slack))
- `configuration_workday` (Block) Workday configuration. (see [below for nested schema](#nestedatt--configuration_workday))
- `configuration_zoom` (Block) Zoom configuration. (see [below for nested schema](#nestedatt--configuration_zoom))

<a id="nestedatt--sync_status"></a>
### Nested Schema for `sync_status`

Read-Only:

- `details` (String)
- `last_sync_time` (String)
- `next_sync_time` (String)
- `status` (String)

<a id="nestedatt--configuration_aquera"></a>
### Nested Schema for `configuration_aquera`

Read-Only:

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

<a id="nestedatt--configuration_azure_ad_saml_v2"></a>
### Nested Schema for `configuration_azure_ad_saml_v2`

Read-Only:

- `base_url` (String)
- `bearer_token` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `remove_action` (String)
- `scim_url` (String)
- `update_users` (Boolean)

<a id="nestedatt--configuration_google_apps"></a>
### Nested Schema for `configuration_google_apps`

Read-Only:

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

<a id="nestedatt--configuration_ldap_gateway"></a>
### Nested Schema for `configuration_ldap_gateway`

Read-Only:

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

<a id="nestedatt--configuration_ping_one"></a>
### Nested Schema for `configuration_ping_one`

Read-Only:

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

<a id="nestedatt--configuration_salesforce"></a>
### Nested Schema for `configuration_salesforce`

Read-Only:

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

<a id="nestedatt--configuration_salesforce_contacts"></a>
### Nested Schema for `configuration_salesforce_contacts`

Read-Only:

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

<a id="nestedatt--configuration_scim"></a>
### Nested Schema for `configuration_scim`

Read-Only:

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

<a id="nestedatt--scim_configuration"></a>
### Nested Schema for `scim_configuration`

Read-Only:

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

<a id="nestedatt--configuration_service_now"></a>
### Nested Schema for `configuration_service_now`

Read-Only:

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

<a id="nestedatt--configuration_slack"></a>
### Nested Schema for `configuration_slack`

Read-Only:

- `base_url` (String)
- `bearer_token` (String)
- `create_users` (Boolean)
- `deprovision_users` (Boolean)
- `disable_users` (Boolean)
- `group_name_source` (String)
- `remove_action` (String)
- `scim_url` (String)
- `update_users` (Boolean)

<a id="nestedatt--configuration_workday"></a>
### Nested Schema for `configuration_workday`

Read-Only:

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

<a id="nestedatt--configuration_zoom"></a>
### Nested Schema for `configuration_zoom`

Read-Only:

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
