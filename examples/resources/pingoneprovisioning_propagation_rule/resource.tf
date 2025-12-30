resource "pingoneprovisioning_propagation_rule" "example" {
  environment_id  = "00000000-0000-0000-0000-000000000000"
  plan_id         = "22222222-2222-2222-2222-222222222222"
  name            = "Users to SCIM"
  source_store_id = "33333333-3333-3333-3333-333333333333"
  target_store_id = "44444444-4444-4444-4444-444444444444"

  active      = true
  deprovision = true
  filter      = "active eq true"

  mappings = [
    {
      source_attribute = "userName"
      target_attribute = "userName"
    },
    {
      source_attribute = "emails[primary eq true].value"
      target_attribute = "emails[0].value"
    }
  ]
}
