package types

import "github.com/hashicorp/terraform-plugin-framework/types"

// PropagationPlanModel describes the Terraform model for a PingOne propagation plan.
type PropagationPlanModel struct {
	Id            types.String `tfsdk:"id"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	Status        types.String `tfsdk:"status"`
}
