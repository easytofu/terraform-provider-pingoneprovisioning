---
title: pingoneprovisioning_propagation_stores
page_title: "Data Source: pingoneprovisioning_propagation_stores"
description: "Fetches PingOne provisioning propagation stores for an environment."
slug: provider_datasource_pingoneprovisioning_propagation_stores
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_datasources
privacy:
  view: public
position: 2
---
## Data Source: pingoneprovisioning_propagation_stores

Fetches PingOne provisioning propagation stores for an environment.

## Example Usage

```terraform
data "pingoneprovisioning_propagation_stores" "all" {
  environment_id = var.environment_id
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.

### Optional

- `store_id` (String) Optional filter by a specific propagation store ID.
- `type` (String) Optional filter by propagation store type.

### Read-Only

- `ids` (List of String) List of propagation store IDs found.
- `stores` (List of Object) List of propagation stores. Each object matches the schema of the `pingoneprovisioning_propagation_store` data source, including configuration blocks.
