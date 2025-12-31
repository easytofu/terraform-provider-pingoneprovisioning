---
title: pingoneprovisioning_enterprise_team
page_title: "Resource: pingoneprovisioning_enterprise_team"
description: "Manages a GitHub enterprise team. Requires a GitHub token configured on the provider."
slug: provider_resource_pingoneprovisioning_enterprise_team
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_resources
privacy:
  view: public
position: 5
---
## Resource: pingoneprovisioning_enterprise_team

Manages a GitHub enterprise team. Requires a GitHub token configured on the provider.

## Example Usage

```terraform
resource "pingoneprovisioning_enterprise_team" "example" {
  enterprise = var.github_enterprise_slug
  name       = "GitHub Example Team"

  organization_selection_type = "selected"
  organization_slugs          = var.github_org_slugs

  group_id = var.github_idp_group_id
}
```

## Schema

### Required

- `enterprise` (String) The enterprise slug.
- `name` (String) The name of the team.

### Optional

- `description` (String) The team description.
- `group_id` (String) Optional IdP group ID used to manage membership for the enterprise team.
- `organization_selection_type` (String) Specifies which organizations in the enterprise should have access to this team. Allowed values: `disabled`, `selected`, `all`.
- `organization_slugs` (Set of String) Optional organization slugs to assign when `organization_selection_type` is `selected`.

### Read-Only

- `id` (String) The unique ID of the enterprise team.
- `slug` (String) The team slug.
- `url` (String) API URL for the enterprise team.
- `html_url` (String) Web URL for the enterprise team.
- `members_url` (String) API URL for team members.
- `created_at` (String) Team creation timestamp.
- `updated_at` (String) Team update timestamp.

## Import

Import is supported using the following syntax:

```shell
terraform import pingoneprovisioning_enterprise_team.example <enterprise>/<team_slug>
```
