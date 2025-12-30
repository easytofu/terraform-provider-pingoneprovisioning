data "pingoneprovisioning_propagation_rule" "example" {
  environment_id = "00000000-0000-0000-0000-000000000000"
  plan_id        = "22222222-2222-2222-2222-222222222222"
  name           = "Users to SCIM"
}

output "propagation_rule_id" {
  value = data.pingoneprovisioning_propagation_rule.example.id
}
