variable "pingone_environment_id" {
  type = string
}

data "pingoneprovisioning_groups" "all" {
  environment_id = var.pingone_environment_id
}

output "pingone_group_ids" {
  value = data.pingoneprovisioning_groups.all.ids
}
