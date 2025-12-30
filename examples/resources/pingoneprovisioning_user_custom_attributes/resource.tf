resource "pingoneprovisioning_user_custom_attributes" "example" {
  environment_id = "00000000-0000-0000-0000-000000000000"
  user_id        = "11111111-1111-1111-1111-111111111111"

  attributes = {
    customRoles = ["example_role"]
  }
}
