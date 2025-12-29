package mappers

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
)

// GroupToModel maps a PingOne Group to the Terraform data model.
func GroupToModel(apiObj *management.Group, environmentId string) customtypes.GroupModel {
	model := customtypes.GroupModel{
		Id:            types.StringNull(),
		EnvironmentId: types.StringValue(environmentId),
		Name:          types.StringNull(),
		DisplayName:   types.StringNull(),
		Description:   types.StringNull(),
		PopulationId:  types.StringNull(),
		UserFilter:    types.StringNull(),
		ExternalId:    types.StringNull(),
		SourceId:      types.StringNull(),
		SourceType:    types.StringNull(),
	}

	if apiObj == nil {
		return model
	}

	if id := strings.TrimSpace(apiObj.GetId()); id != "" {
		model.Id = types.StringValue(id)
	}
	if name := strings.TrimSpace(apiObj.GetName()); name != "" {
		model.Name = types.StringValue(name)
	}
	if displayName, ok := apiObj.GetDisplayNameOk(); ok && displayName != nil {
		if v := strings.TrimSpace(*displayName); v != "" {
			model.DisplayName = types.StringValue(v)
		}
	}
	if description, ok := apiObj.GetDescriptionOk(); ok && description != nil {
		if v := strings.TrimSpace(*description); v != "" {
			model.Description = types.StringValue(v)
		}
	}
	if population, ok := apiObj.GetPopulationOk(); ok && population != nil {
		if id := strings.TrimSpace(population.GetId()); id != "" {
			model.PopulationId = types.StringValue(id)
		}
	}
	if filter, ok := apiObj.GetUserFilterOk(); ok && filter != nil {
		if v := strings.TrimSpace(*filter); v != "" {
			model.UserFilter = types.StringValue(v)
		}
	}
	if externalId, ok := apiObj.GetExternalIdOk(); ok && externalId != nil {
		if v := strings.TrimSpace(*externalId); v != "" {
			model.ExternalId = types.StringValue(v)
		}
	}
	if sourceId, ok := apiObj.GetSourceIdOk(); ok && sourceId != nil {
		if v := strings.TrimSpace(*sourceId); v != "" {
			model.SourceId = types.StringValue(v)
		}
	}
	if sourceType, ok := apiObj.GetSourceTypeOk(); ok && sourceType != nil {
		if v := strings.TrimSpace(string(*sourceType)); v != "" && v != "UNKNOWN" {
			model.SourceType = types.StringValue(v)
		}
	}

	return model
}
