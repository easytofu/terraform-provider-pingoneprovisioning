---
title: pingoneprovisioning_propagation_plan
page_title: "Data Source: pingoneprovisioning_propagation_plan"
description: "Fetches a PingOne provisioning propagation plan by ID or name."
slug: provider_datasource_pingoneprovisioning_propagation_plan
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_datasources
privacy:
  view: public
position: 3
---
## Data Source: pingoneprovisioning_propagation_plan

Fetches a PingOne provisioning propagation plan by ID or name.

## Example Usage

```terraform
data "pingoneprovisioning_propagation_plan" "example" {
  environment_id = var.environment_id
  name           = "Default Plan"
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.

### Optional

- `id` (String) The unique ID of the propagation plan.
- `name` (String) Unique name of the propagation plan.

### Read-Only

- `status` (String) Status of the propagation plan.
