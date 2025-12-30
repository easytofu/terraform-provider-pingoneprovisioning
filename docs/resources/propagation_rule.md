---
title: pingoneprovisioning_propagation_rule
page_title: "Resource: pingoneprovisioning_propagation_rule"
description: "Manages a PingOne provisioning propagation rule and its mappings."
slug: provider_resource_pingoneprovisioning_propagation_rule
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_resources
privacy:
  view: public
position: 3
---
## Resource: pingoneprovisioning_propagation_rule

Manages a PingOne provisioning propagation rule and its mappings.

## Example Usage

```terraform
resource "pingoneprovisioning_propagation_rule" "example" {
  environment_id = var.environment_id
  plan_id        = var.plan_id
  name           = "Users to SCIM"

  source_store_id = var.source_store_id
  target_store_id = var.target_store_id

  active      = true
  deprovision = true
  filter      = "active eq true"

  mappings = [
    {
      source_attribute = "userName"
      target_attribute = "userName"
    },
    {
      source_attribute = "emails[primary eq true].value"
      target_attribute = "emails[0].value"
    }
  ]
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.
- `plan_id` (String) The ID of the propagation plan.
- `name` (String) The name of the propagation rule.
- `source_store_id` (String) The source store ID for the propagation rule.
- `target_store_id` (String) The target store ID for the propagation rule.

### Optional

- `active` (Boolean) Whether the propagation rule is active.
- `configuration` (Map of String) Optional rule configuration map (for example, `MFA_USER_DEVICE_MANAGEMENT`).
- `deprovision` (Boolean) Whether to deprovision users in the target store when they are removed from the source.
- `filter` (String) Expression used by PingOne to select users to synchronize (maps to the API field `populationExpression`).
- `group_ids` (List of String) Optional list of group IDs to scope group provisioning for this rule.
- `population_ids` (List of String) Optional list of population IDs in scope for this rule.
- `mappings` (List of Object) Optional list of attribute mappings for this rule. (see [below for nested schema](#nestedblock--mappings))

### Read-Only

- `id` (String) The unique ID of the propagation rule.

<a id="nestedblock--mappings"></a>
### Nested Schema for `mappings`

Optional:

- `expression` (String) Optional expression used to compute the target attribute value.
- `source_attribute` (String) Source attribute expression.
- `target_attribute` (String) Target attribute expression.

Read-Only:

- `id` (String) The mapping ID.

## Import

Import is supported using the following syntax:

```shell
terraform import pingoneprovisioning_propagation_rule.example <environment_id>/<rule_id>
```
