---
title: pingoneprovisioning_enterprise_team_member
page_title: "Resource: pingoneprovisioning_enterprise_team_member"
description: "Manages a single GitHub enterprise team membership. Requires a GitHub token configured on the provider."
slug: provider_resource_pingoneprovisioning_enterprise_team_member
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_resources
privacy:
  view: public
position: 8
---
## Resource: pingoneprovisioning_enterprise_team_member

Manages a single GitHub enterprise team membership. Requires a GitHub token configured on the provider.

## Example Usage

```terraform
resource "pingoneprovisioning_enterprise_team_member" "example" {
  enterprise = var.github_enterprise_slug
  team_slug  = var.github_team_slug
  username   = "octocat"
}
```

## Schema

### Required

- `enterprise` (String) The enterprise slug.
- `team_slug` (String) The enterprise team slug or ID.
- `username` (String) The GitHub username.

### Read-Only

- `id` (String) Composite ID of the team membership.

## Import

Import is supported using the following syntax:

```shell
terraform import pingoneprovisioning_enterprise_team_member.example <enterprise>/<team_slug>/<username>
```
