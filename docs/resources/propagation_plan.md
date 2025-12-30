---
title: pingoneprovisioning_propagation_plan
page_title: "Resource: pingoneprovisioning_propagation_plan"
description: "Manages a PingOne provisioning propagation plan."
slug: provider_resource_pingoneprovisioning_propagation_plan
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_resources
privacy:
  view: public
position: 2
---
## Resource: pingoneprovisioning_propagation_plan

Manages a PingOne provisioning propagation plan.

## Example Usage

```terraform
resource "pingoneprovisioning_propagation_plan" "example" {
  environment_id = var.environment_id
  name           = "Default Plan"
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.
- `name` (String) Unique name of the propagation plan.

### Read-Only

- `id` (String) The unique ID of the propagation plan.
- `status` (String) Status of the propagation plan.

## Import

Import is supported using the following syntax:

```shell
terraform import pingoneprovisioning_propagation_plan.example <environment_id>/<plan_id>
```
