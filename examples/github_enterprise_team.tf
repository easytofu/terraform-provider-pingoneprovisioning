variable "github_enterprise_slug" {
  type = string
}

variable "github_org_slugs" {
  type = list(string)
}

variable "github_idp_group_id" {
  type = string
}


resource "pingoneprovisioning_enterprise_team" "example" {
  enterprise = var.github_enterprise_slug
  name       = "GitHub Example Team"

  organization_selection_type = "selected"
  organization_slugs          = var.github_org_slugs

  group_id = var.github_idp_group_id
}
