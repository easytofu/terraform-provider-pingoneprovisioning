---
title: pingoneprovisioning_enterprise_teams
page_title: "Data Source: pingoneprovisioning_enterprise_teams"
description: "Lists GitHub enterprise teams. Requires a GitHub token configured on the provider."
slug: provider_datasource_pingoneprovisioning_enterprise_teams
category:
  uri: PingOne Provisioning Terraform Provider
parent:
  uri: provider_datasources
privacy:
  view: public
position: 6
---
## Data Source: pingoneprovisioning_enterprise_teams

Lists GitHub enterprise teams. Requires a GitHub token configured on the provider.

## Example Usage

```terraform
data "pingoneprovisioning_enterprise_teams" "all" {
  enterprise = var.github_enterprise_slug
}

output "github_enterprise_team_slugs" {
  value = [for team in data.pingoneprovisioning_enterprise_teams.all.teams : team.slug]
}
```

## Schema

### Required

- `enterprise` (String) The enterprise slug.

### Read-Only

- `teams` (List of Object) Enterprise teams. (see [below for nested schema](#nestedatt--teams))

<a id="nestedatt--teams"></a>
### Nested Schema for `teams`

Read-Only:

- `id` (String) Team ID.
- `name` (String) Team name.
- `description` (String) Team description.
- `slug` (String) Team slug.
- `group_id` (String) IdP group ID.
- `url` (String) API URL for the team.
- `html_url` (String) Web URL for the team.
- `members_url` (String) API URL for team members.
- `created_at` (String) Team creation timestamp.
- `updated_at` (String) Team update timestamp.
