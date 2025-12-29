// pingone/providers/pingone-propagation/internal/datasources/propagation_store.go
package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/mappers"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/providerdata"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/schemas"
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &propagationStoreDataSource{}
	_ datasource.DataSourceWithConfigure = &propagationStoreDataSource{}
)

// propagationStoreDataSource is the data source implementation.
type propagationStoreDataSource struct {
	client *providerdata.Client
}

// NewPropagationStoreDataSource is a helper function to simplify the provider implementation.
func NewPropagationStoreDataSource() datasource.DataSource {
	return &propagationStoreDataSource{}
}

func (d *propagationStoreDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_propagation_store"
}

func (d *propagationStoreDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a PingOne provisioning propagation store.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the propagation store.",
				Optional:    true,
				Computed:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the identity store.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the identity store.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the identity store.",
				Computed:    true,
			},
			"image_id": schema.StringAttribute{
				Description: "The image ID for the identity store resource.",
				Computed:    true,
			},
			"image_href": schema.StringAttribute{
				Description: "The URL for the identity store resource image file.",
				Computed:    true,
			},
			"managed": schema.BoolAttribute{
				Description: "Indicates whether or not to enable deprovisioning of users from the target store.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the propagation store.",
				Computed:    true,
			},
			"sync_status": schema.ObjectAttribute{
				Description: "Sync status for the propagation store.",
				Computed:    true,
				AttributeTypes: map[string]attr.Type{
					"last_sync_time": types.StringType,
					"next_sync_time": types.StringType,
					"status":         types.StringType,
					"details":        types.StringType,
				},
			},
		},
		Blocks: map[string]schema.Block{
			"configuration_aquera":              schemas.AqueraConfigSchema(true),
			"configuration_azure_ad_saml_v2":    schemas.AzureAdSamlV2ConfigSchema(true),
			"configuration_github_emu":          schemas.GithubEmuConfigSchema(true),
			"configuration_google_apps":         schemas.GoogleAppsConfigSchema(true),
			"configuration_ldap_gateway":        schemas.LdapGatewayConfigSchema(true),
			"configuration_ping_one":            schemas.PingOneConfigSchema(true),
			"configuration_salesforce":          schemas.SalesforceConfigSchema(true),
			"configuration_salesforce_contacts": schemas.SalesforceContactsConfigSchema(true),
			"configuration_scim":                schemas.ScimConfigSchema(true),
			"scim_configuration":                schemas.ScimConfigSchema(true),
			"configuration_service_now":         schemas.ServiceNowConfigSchema(true),
			"configuration_slack":               schemas.SlackConfigSchema(true),
			"configuration_workday":             schemas.WorkdayConfigSchema(true),
			"configuration_zoom":                schemas.ZoomConfigSchema(true),
		},
	}
}

func (d *propagationStoreDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		propagationStoreLookupValidator{},
	}
}

type propagationStoreLookupValidator struct{}

func (v propagationStoreLookupValidator) Description(_ context.Context) string {
	return "Validates lookup arguments for a propagation store."
}

func (v propagationStoreLookupValidator) MarkdownDescription(_ context.Context) string {
	return "Validates lookup arguments for a propagation store."
}

func (v propagationStoreLookupValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config customtypes.PropagationStoreModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, validationDiags := propagationStoreLookupModeFromValues(config.Id, config.Name, config.Type)
	resp.Diagnostics.Append(validationDiags...)
}

func (d *propagationStoreDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*providerdata.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *providerdata.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *propagationStoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customtypes.PropagationStoreModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := state.EnvironmentId.ValueString()
	apiClient := d.client.API

	var apiResult *management.PropagationStore
	var apiTypeRaw string
	var apiStatusRaw string

	lookupMode, validationDiags := propagationStoreLookupModeFromValues(state.Id, state.Name, state.Type)
	resp.Diagnostics.Append(validationDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch lookupMode {
	case propagationStoreLookupModeNameType:
		targetName := state.Name.ValueString()
		targetType := state.Type.ValueString()
		targetTypeAPI := utils.NormalizePropagationStoreTypeForAPI(targetType)

		tflog.Info(ctx, "Reading propagation store by Name and Type", map[string]interface{}{
			"environment_id": environmentID,
			"name":           targetName,
			"type":           targetType,
		})

		iterator := apiClient.PropagationStoresApi.ReadAllStores(ctx, environmentID).Execute()

		type foundStore struct {
			obj       *management.PropagationStore
			rawType   string
			rawStatus string
		}
		var foundStores []foundStore

		for cursor, iterErr := range iterator {
			if iterErr != nil {
				resp.Diagnostics.AddError(
					"Error Reading Propagation Stores",
					fmt.Sprintf("Could not iterate propagation stores: %s", iterErr),
				)
				return
			}

			if cursor.HTTPResponse == nil {
				continue
			}

			bodyBytes, readErr := io.ReadAll(cursor.HTTPResponse.Body)
			_ = cursor.HTTPResponse.Body.Close()
			if readErr != nil {
				resp.Diagnostics.AddError(
					"Error Reading Propagation Stores",
					fmt.Sprintf("Could not read propagation stores response: %s", readErr),
				)
				return
			}

			var raw map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &raw); err != nil {
				resp.Diagnostics.AddError(
					"Error Parsing Propagation Stores",
					fmt.Sprintf("Could not parse propagation stores response: %s", err),
				)
				return
			}

			embedded, ok := raw["_embedded"].(map[string]interface{})
			if !ok {
				continue
			}

			stores, ok := embedded["stores"].([]interface{})
			if !ok {
				continue
			}

			for _, s := range stores {
				storeMap, ok := s.(map[string]interface{})
				if !ok {
					continue
				}

				storeName, _ := storeMap["name"].(string)
				storeTypeRaw, _ := storeMap["type"].(string)
				storeStatusRaw, _ := storeMap["status"].(string)

				if storeName != targetName || !strings.EqualFold(storeTypeRaw, targetTypeAPI) {
					continue
				}

				storeJSON, err := json.Marshal(storeMap)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error Parsing Propagation Store",
						fmt.Sprintf("Could not marshal propagation store response: %s", err),
					)
					return
				}

				var storeObj management.PropagationStore
				if err := json.Unmarshal(storeJSON, &storeObj); err != nil {
					resp.Diagnostics.AddError(
						"Error Parsing Propagation Store",
						fmt.Sprintf("Could not unmarshal propagation store response: %s", err),
					)
					return
				}

				foundStores = append(foundStores, foundStore{
					obj:       &storeObj,
					rawType:   storeTypeRaw,
					rawStatus: storeStatusRaw,
				})
			}
		}

		if len(foundStores) == 0 {
			resp.Diagnostics.AddError(
				"Propagation Store Not Found",
				fmt.Sprintf("No propagation store found with name '%s' and type '%s' in environment '%s'.", targetName, targetType, environmentID),
			)
			return
		} else if len(foundStores) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Propagation Stores Found",
				fmt.Sprintf("Found %d stores with name '%s' and type '%s' in environment '%s'. Use the 'id' argument to select a specific store.", len(foundStores), targetName, targetType, environmentID),
			)
			return
		}

		apiResult = foundStores[0].obj
		apiTypeRaw = foundStores[0].rawType
		apiStatusRaw = foundStores[0].rawStatus
	case propagationStoreLookupModeId:
		// =========================================================================================
		// Scenario B: Lookup by ID if provided
		// =========================================================================================
		storeID := state.Id.ValueString()

		tflog.Info(ctx, "Reading propagation store by ID", map[string]interface{}{
			"environment_id": environmentID,
			"id":             storeID,
		})

		result, httpResp, err := apiClient.PropagationStoresApi.
			ReadOnePropagationStore(ctx, environmentID, storeID).
			Execute()
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddError(
					"Propagation Store Not Found",
					fmt.Sprintf("No propagation store found with ID '%s' in environment '%s'.", storeID, environmentID),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Error Reading Propagation Store",
				fmt.Sprintf("Could not read propagation store: %s", utils.HandleSDKError(err, httpResp)),
			)
			return
		}

		apiResult = result
		rawType, rawStatus, parseErr := utils.ExtractPropagationStoreTypeStatus(httpResp)
		if parseErr != nil {
			resp.Diagnostics.AddError(
				"Error Reading Propagation Store",
				fmt.Sprintf("Could not parse propagation store response: %s", parseErr),
			)
			return
		}
		apiTypeRaw = rawType
		apiStatusRaw = rawStatus
	default:
		resp.Diagnostics.AddError("Invalid Lookup Configuration", "Unable to determine lookup configuration for propagation store.")
		return
	}

	// Map API response to state
	d.apiToModel(apiResult, apiTypeRaw, apiStatusRaw, environmentID, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

type propagationStoreLookupMode int

const (
	propagationStoreLookupModeInvalid propagationStoreLookupMode = iota
	propagationStoreLookupModeNameType
	propagationStoreLookupModeId
)

func propagationStoreLookupModeFromValues(id, name, storeType types.String) (propagationStoreLookupMode, diag.Diagnostics) {
	var diags diag.Diagnostics

	idSet := !id.IsNull() && !id.IsUnknown() && id.ValueString() != ""
	nameSet := !name.IsNull() && !name.IsUnknown() && name.ValueString() != ""
	typeSet := !storeType.IsNull() && !storeType.IsUnknown() && storeType.ValueString() != ""

	// If either name or type is configured, require both. We intentionally do
	// not treat `id` as conflicting since Optional+Computed attributes can carry
	// forward prior state values even when users do not configure them.
	if nameSet || typeSet {
		if !nameSet {
			diags.AddAttributeError(
				path.Root("name"),
				"Missing Required Argument",
				"When configuring a lookup by `name` and `type`, `name` must be set.",
			)
		}
		if !typeSet {
			diags.AddAttributeError(
				path.Root("type"),
				"Missing Required Argument",
				"When configuring a lookup by `name` and `type`, `type` must be set.",
			)
		}
		if diags.HasError() {
			return propagationStoreLookupModeInvalid, diags
		}
		return propagationStoreLookupModeNameType, diags
	}

	if idSet {
		return propagationStoreLookupModeId, diags
	}

	diags.AddError(
		"Missing Required Arguments",
		"Configure either `id` or both `name` and `type` to lookup a propagation store.",
	)
	return propagationStoreLookupModeInvalid, diags
}

func (d *propagationStoreDataSource) apiToModel(apiObj *management.PropagationStore, rawType string, rawStatus string, environmentId string, state *customtypes.PropagationStoreModel) {
	state.Id = types.StringValue(apiObj.GetId())
	state.EnvironmentId = types.StringValue(environmentId)
	state.Name = types.StringValue(apiObj.GetName())

	preferredType := ""
	if !state.Type.IsNull() && !state.Type.IsUnknown() {
		preferredType = state.Type.ValueString()
	}
	apiType := rawType
	if apiType == "" || apiType == "UNKNOWN" {
		apiType = string(apiObj.GetType())
	}
	if apiType == "" || apiType == "UNKNOWN" {
		apiType = preferredType
	}
	tfType := utils.NormalizePropagationStoreTypeForTerraform(apiType, preferredType)
	state.Type = types.StringValue(tfType)

	if v, ok := apiObj.GetDescriptionOk(); ok {
		state.Description = types.StringValue(*v)
	} else {
		state.Description = types.StringNull()
	}

	if image, ok := apiObj.GetImageOk(); ok && image != nil {
		if v, ok := image.GetIdOk(); ok && v != nil {
			state.ImageId = types.StringValue(*v)
		} else {
			state.ImageId = types.StringNull()
		}

		if v, ok := image.GetHrefOk(); ok && v != nil {
			state.ImageHref = types.StringValue(*v)
		} else {
			state.ImageHref = types.StringNull()
		}
	} else {
		state.ImageId = types.StringNull()
		state.ImageHref = types.StringNull()
	}

	if v, ok := apiObj.GetManagedOk(); ok {
		state.Managed = types.BoolValue(*v)
	} else {
		state.Managed = types.BoolNull()
	}

	status := rawStatus
	if status == "" {
		if v, ok := apiObj.GetStatusOk(); ok && v != nil && string(*v) != "" && string(*v) != "UNKNOWN" {
			status = string(*v)
		}
	}
	if status != "" {
		state.Status = types.StringValue(status)
	} else {
		state.Status = types.StringNull()
	}

	if v, ok := apiObj.GetSyncStatusOk(); ok && v != nil {
		attrs := map[string]attr.Value{
			"last_sync_time": types.StringNull(),
			"next_sync_time": types.StringNull(),
			"status":         types.StringNull(),
			"details":        types.StringNull(),
		}

		if vv, ok := v.GetLastSyncAtOk(); ok && vv != nil {
			attrs["last_sync_time"] = types.StringValue(vv.Format(time.RFC3339))
		}
		if vv, ok := v.GetSyncStateOk(); ok && vv != nil {
			attrs["status"] = types.StringValue(string(*vv))
		}
		if vv, ok := v.GetDetailsOk(); ok && vv != nil {
			attrs["details"] = types.StringValue(*vv)
		}

		state.SyncStatus = types.ObjectValueMust(map[string]attr.Type{
			"last_sync_time": types.StringType,
			"next_sync_time": types.StringType,
			"status":         types.StringType,
			"details":        types.StringType,
		}, attrs)
	} else {
		state.SyncStatus = types.ObjectNull(map[string]attr.Type{
			"last_sync_time": types.StringType,
			"next_sync_time": types.StringType,
			"status":         types.StringType,
			"details":        types.StringType,
		})
	}

	config := apiObj.GetConfiguration()
	mappers.ApplyPropagationStoreConfigurationFromMap(state, tfType, config, nil)
}
