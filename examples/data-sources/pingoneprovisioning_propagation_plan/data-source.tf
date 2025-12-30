data "pingoneprovisioning_propagation_plan" "example" {
  environment_id = "00000000-0000-0000-0000-000000000000"
  name           = "Default Plan"
}

output "propagation_plan_id" {
  value = data.pingoneprovisioning_propagation_plan.example.id
}
