package types

import "github.com/hashicorp/terraform-plugin-framework/types"

// PropagationRuleMappingModel describes a single propagation mapping for a rule.
type PropagationRuleMappingModel struct {
	Id              types.String `tfsdk:"id"`
	SourceAttribute types.String `tfsdk:"source_attribute"`
	TargetAttribute types.String `tfsdk:"target_attribute"`
	Expression      types.String `tfsdk:"expression"`
}

// PropagationRuleModel describes the Terraform model for a PingOne propagation rule.
type PropagationRuleModel struct {
	Id            types.String                  `tfsdk:"id"`
	EnvironmentId types.String                  `tfsdk:"environment_id"`
	PlanId        types.String                  `tfsdk:"plan_id"`
	Name          types.String                  `tfsdk:"name"`
	SourceStoreId types.String                  `tfsdk:"source_store_id"`
	TargetStoreId types.String                  `tfsdk:"target_store_id"`
	Active        types.Bool                    `tfsdk:"active"`
	Filter        types.String                  `tfsdk:"filter"`
	Deprovision   types.Bool                    `tfsdk:"deprovision"`
	PopulationIds types.List                    `tfsdk:"population_ids"`
	GroupIds      types.List                    `tfsdk:"group_ids"`
	Configuration types.Map                     `tfsdk:"configuration"`
	Mappings      []PropagationRuleMappingModel `tfsdk:"mappings"`
}
