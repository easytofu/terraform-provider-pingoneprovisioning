---
title: pingoneprovisioning_propagation_rule
page_title: "Data Source: pingoneprovisioning_propagation_rule"
description: "Fetches a PingOne provisioning propagation rule and its mappings by ID or by name (optionally scoped by plan ID)."
slug: provider_datasource_pingoneprovisioning_propagation_rule
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_datasources
privacy:
  view: public
position: 4
---
## Data Source: pingoneprovisioning_propagation_rule

Fetches a PingOne provisioning propagation rule and its mappings by ID or by name (optionally scoped by plan ID).

## Example Usage

```terraform
data "pingoneprovisioning_propagation_rule" "by_name" {
  environment_id = var.environment_id
  plan_id        = var.plan_id
  name           = "Users to SCIM"
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.

### Optional

- `id` (String) The unique ID of the propagation rule.
- `name` (String) Optional name to lookup the propagation rule.
- `plan_id` (String) Optional plan ID to scope name lookups.

### Read-Only

- `active` (Boolean) Whether the propagation rule is active.
- `configuration` (Map of String) Rule configuration map (for example, `MFA_USER_DEVICE_MANAGEMENT`).
- `deprovision` (Boolean) Whether to deprovision users in the target store when they are removed from the source.
- `filter` (String) SCIM filter expression for selecting users to synchronize.
- `group_ids` (List of String) List of group IDs in scope for group provisioning.
- `population_ids` (List of String) List of population IDs in scope for this rule.
- `source_store_id` (String) The source store ID for the propagation rule.
- `target_store_id` (String) The target store ID for the propagation rule.
- `mappings` (List of Object) List of attribute mappings for this rule. (see [below for nested schema](#nestedatt--mappings))

<a id="nestedatt--mappings"></a>
### Nested Schema for `mappings`

Read-Only:

- `expression` (String) Expression used to compute the target attribute value.
- `id` (String) The mapping ID.
- `source_attribute` (String) Source attribute expression.
- `target_attribute` (String) Target attribute expression.
