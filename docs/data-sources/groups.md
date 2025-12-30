---
title: pingoneprovisioning_groups
page_title: "Data Source: pingoneprovisioning_groups"
description: "Fetches PingOne groups for an environment."
slug: provider_datasource_pingoneprovisioning_groups
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_datasources
privacy:
  view: public
position: 5
---
## Data Source: pingoneprovisioning_groups

Fetches PingOne groups for an environment.

## Example Usage

```terraform
data "pingoneprovisioning_groups" "all" {
  environment_id = var.pingone_environment_id
}
```

## Schema

### Required

- `environment_id` (String) The ID of the environment.

### Optional

- `filter` (String) Optional SCIM filter to apply when listing groups.

### Read-Only

- `ids` (List of String) List of group IDs found.
- `groups` (List of Object) List of groups returned by the query. (see [below for nested schema](#nestedatt--groups))

<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Read-Only:

- `id` (String)
- `environment_id` (String)
- `name` (String)
- `display_name` (String)
- `description` (String)
- `population_id` (String)
- `user_filter` (String)
- `external_id` (String)
- `source_id` (String)
- `source_type` (String)
