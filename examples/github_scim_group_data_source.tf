# Fetch a GitHub SCIM group by display name.

data "pingoneprovisioning_github_scim_group" "example" {
  enterprise   = "example-enterprise"
  display_name = "GitHub Example Repository Reader"
}

output "github_scim_group_id" {
  value = data.pingoneprovisioning_github_scim_group.example.id
}
