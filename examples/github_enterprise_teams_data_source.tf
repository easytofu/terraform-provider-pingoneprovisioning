variable "github_enterprise_slug" {
  type = string
}

data "pingoneprovisioning_enterprise_teams" "all" {
  enterprise = var.github_enterprise_slug
}

output "github_enterprise_team_slugs" {
  value = [for team in data.pingoneprovisioning_enterprise_teams.all.teams : team.slug]
}
