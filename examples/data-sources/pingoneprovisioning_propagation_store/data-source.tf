data "pingoneprovisioning_propagation_store" "example" {
  environment_id = "00000000-0000-0000-0000-000000000000"
  name           = "Example SCIM Store"
  type           = "SCIM"
}

output "propagation_store_id" {
  value = data.pingoneprovisioning_propagation_store.example.id
}
