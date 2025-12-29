package types

import "github.com/hashicorp/terraform-plugin-framework/types"

// UserCustomAttributesModel describes the Terraform model for PingOne user custom attributes.
type UserCustomAttributesModel struct {
	Id            types.String  `tfsdk:"id"`
	EnvironmentId types.String  `tfsdk:"environment_id"`
	UserId        types.String  `tfsdk:"user_id"`
	Attributes    types.Dynamic `tfsdk:"attributes"`
}
