---
title: pingoneprovisioning_enterprise_team_organizations
page_title: "Resource: pingoneprovisioning_enterprise_team_organizations"
description: "Manages organization assignments for a GitHub enterprise team. Requires a GitHub token configured on the provider."
slug: provider_resource_pingoneprovisioning_enterprise_team_organizations
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_resources
privacy:
  view: public
position: 7
---
## Resource: pingoneprovisioning_enterprise_team_organizations

Manages organization assignments for a GitHub enterprise team. Requires a GitHub token configured on the provider.

## Example Usage

```terraform
resource "pingoneprovisioning_enterprise_team_organizations" "example" {
  enterprise         = var.github_enterprise_slug
  team_slug          = var.github_team_slug
  organization_slugs = ["example-org", "another-org"]
}
```

## Schema

### Required

- `enterprise` (String) The enterprise slug.
- `team_slug` (String) The enterprise team slug or ID.
- `organization_slugs` (Set of String) Organization slugs to assign to the enterprise team.

### Read-Only

- `id` (String) Composite ID of the organization assignments.

## Import

Import is supported using the following syntax:

```shell
terraform import pingoneprovisioning_enterprise_team_organizations.example <enterprise>/<team_slug>
```
