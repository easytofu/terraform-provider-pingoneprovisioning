package types

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GroupModel describes a PingOne group for data sources.
type GroupModel struct {
	Id            types.String `tfsdk:"id"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	DisplayName   types.String `tfsdk:"display_name"`
	Description   types.String `tfsdk:"description"`
	PopulationId  types.String `tfsdk:"population_id"`
	UserFilter    types.String `tfsdk:"user_filter"`
	ExternalId    types.String `tfsdk:"external_id"`
	SourceId      types.String `tfsdk:"source_id"`
	SourceType    types.String `tfsdk:"source_type"`
}

var GroupModelAttrTypes = map[string]attr.Type{
	"id":             types.StringType,
	"environment_id": types.StringType,
	"name":           types.StringType,
	"display_name":   types.StringType,
	"description":    types.StringType,
	"population_id":  types.StringType,
	"user_filter":    types.StringType,
	"external_id":    types.StringType,
	"source_id":      types.StringType,
	"source_type":    types.StringType,
}

// GroupModelType returns the object type for group list elements.
func GroupModelType() attr.Type {
	return types.ObjectType{AttrTypes: GroupModelAttrTypes}
}
