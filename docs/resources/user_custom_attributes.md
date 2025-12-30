---
title: pingoneprovisioning_user_custom_attributes
page_title: "Resource: pingoneprovisioning_user_custom_attributes"
description: "Manages custom schema attributes for an existing PingOne user. Removing the resource only removes it from Terraform state."
slug: provider_resource_pingoneprovisioning_user_custom_attributes
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_resources
privacy:
  view: public
position: 4
---
## Resource: pingoneprovisioning_user_custom_attributes

Manages custom schema attributes for an existing PingOne user. Removing the resource only removes it from Terraform state.

## Example Usage

```terraform
resource "pingoneprovisioning_user_custom_attributes" "example" {
  environment_id = "00000000-0000-0000-0000-000000000000"
  user_id        = "11111111-1111-1111-1111-111111111111"

  attributes = {
    customRoles = ["example_role"]
  }
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.
- `user_id` (String) The PingOne user ID to update.
- `attributes` (Dynamic) Map of custom user attribute values keyed by schema attribute name.

### Read-Only

- `id` (String) Internal identifier for this custom attribute mapping.

## Import

Import is supported using the following syntax:

```shell
terraform import pingoneprovisioning_user_custom_attributes.example <environment_id>/<user_id>
```
