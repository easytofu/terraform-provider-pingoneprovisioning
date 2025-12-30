data "pingoneprovisioning_groups" "all" {
  environment_id = "00000000-0000-0000-0000-000000000000"
}

output "pingone_group_ids" {
  value = data.pingoneprovisioning_groups.all.ids
}
