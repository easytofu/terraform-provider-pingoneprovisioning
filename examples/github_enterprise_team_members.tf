variable "github_enterprise_slug" {
  type = string
}

variable "github_team_slug" {
  type = string
}

variable "github_team_member_usernames" {
  type = list(string)
}

resource "pingoneprovisioning_enterprise_team_members" "example" {
  enterprise = var.github_enterprise_slug
  team_slug  = var.github_team_slug
  usernames  = var.github_team_member_usernames
}
