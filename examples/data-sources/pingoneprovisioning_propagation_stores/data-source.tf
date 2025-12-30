data "pingoneprovisioning_propagation_stores" "all" {
  environment_id = "00000000-0000-0000-0000-000000000000"
}

output "propagation_store_ids" {
  value = data.pingoneprovisioning_propagation_stores.all.ids
}
