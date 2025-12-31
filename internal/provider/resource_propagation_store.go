// pingone/providers/pingone-propagation/internal/resources/propagation_store.go
package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/mappers"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/schemas"
	customtypes "github.com/easytofu/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &propagationStoreResource{}
	_ resource.ResourceWithConfigure   = &propagationStoreResource{}
	_ resource.ResourceWithImportState = &propagationStoreResource{}
)

// propagationStoreResource is the resource implementation.
type propagationStoreResource struct {
	client *client.Client
}

// NewPropagationStoreResource is a helper function to simplify the provider implementation.
func NewPropagationStoreResource() resource.Resource {
	return &propagationStoreResource{}
}

func (r *propagationStoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_propagation_store"
}

func (r *propagationStoreResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PingOne provisioning propagation store.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the propagation store.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "A name for the identity store.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the identity store.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the identity store. Options are `Aquera`, `AzureADSAMLV2`, `GithubEMU`, `GitHubEMU`, `GoogleApps`, `LDAPGateway`, `PingOne`, `Salesforce`, `SalesforceContacts`, `SCIM` (alias: `scim`), `ServiceNow`, `Slack`, `Workday`, `Zoom`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Aquera", "AzureADSAMLV2", "GithubEMU", "GitHubEMU", "GoogleApps", "LDAPGateway",
						"PingOne", "Salesforce", "SalesforceContacts", "SCIM", "ServiceNow",
						"scim", "Slack", "Workday", "Zoom",
					),
				},
			},
			"image_id": schema.StringAttribute{
				Description: "The image ID for the identity store resource.",
				Optional:    true,
			},
			"image_href": schema.StringAttribute{
				Description: "The URL for the identity store resource image file.",
				Computed:    true,
			},
			"managed": schema.BoolAttribute{
				Description: "Indicates whether or not to enable deprovisioning of users from the target store.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"status": schema.StringAttribute{
				Description: "The status of the propagation store.",
				Optional:    true,
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
			"configuration_aquera":              schemas.AqueraConfigSchema(false),
			"configuration_azure_ad_saml_v2":    schemas.AzureAdSamlV2ConfigSchema(false),
			"configuration_github_emu":          schemas.GithubEmuConfigSchema(false),
			"configuration_google_apps":         schemas.GoogleAppsConfigSchema(false),
			"configuration_ldap_gateway":        schemas.LdapGatewayConfigSchema(false),
			"configuration_ping_one":            schemas.PingOneConfigSchema(false),
			"configuration_salesforce":          schemas.SalesforceConfigSchema(false),
			"configuration_salesforce_contacts": schemas.SalesforceContactsConfigSchema(false),
			"configuration_scim":                schemas.ScimConfigSchema(false),
			"scim_configuration":                schemas.ScimConfigSchema(false),
			"configuration_service_now":         schemas.ServiceNowConfigSchema(false),
			"configuration_slack":               schemas.SlackConfigSchema(false),
			"configuration_workday":             schemas.WorkdayConfigSchema(false),
			"configuration_zoom":                schemas.ZoomConfigSchema(false),
		},
	}
}

func (r *propagationStoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *propagationStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.PropagationStoreModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configMap, err := mappers.ModelToConfigurationMap(&plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Propagation Store",
			fmt.Sprintf("Could not build configuration map: %s", err),
		)
		return
	}

	apiClient := r.client.API
	payload := buildPropagationStorePayload(&plan, configMap)

	result, httpResp, err := apiClient.PropagationStoresApi.
		CreatePropagationStore(ctx, plan.EnvironmentId.ValueString()).
		PropagationStore(*payload).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Propagation Store",
			fmt.Sprintf("Could not create propagation store: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	state, mapErr := r.apiToModel(result, httpResp, plan.EnvironmentId.ValueString(), &plan)
	if mapErr != nil {
		resp.Diagnostics.AddError(
			"Error Creating Propagation Store",
			fmt.Sprintf("Could not map API response to state: %s", mapErr),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *propagationStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.PropagationStoreModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	result, httpResp, err := apiClient.PropagationStoresApi.
		ReadOnePropagationStore(ctx, state.EnvironmentId.ValueString(), state.Id.ValueString()).
		Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Propagation Store",
			fmt.Sprintf("Could not read propagation store: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	newState, mapErr := r.apiToModel(result, httpResp, state.EnvironmentId.ValueString(), &state)
	if mapErr != nil {
		resp.Diagnostics.AddError(
			"Error Reading Propagation Store",
			fmt.Sprintf("Could not map API response to state: %s", mapErr),
		)
		return
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *propagationStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.PropagationStoreModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configMap, err := mappers.ModelToConfigurationMap(&plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Propagation Store",
			fmt.Sprintf("Could not build configuration map: %s", err),
		)
		return
	}

	apiClient := r.client.API
	payload := buildPropagationStorePayload(&plan, configMap)

	result, httpResp, err := apiClient.PropagationStoresApi.
		UpdatePropagationStore(ctx, plan.EnvironmentId.ValueString(), plan.Id.ValueString()).
		PropagationStore(*payload).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Propagation Store",
			fmt.Sprintf("Could not update propagation store: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	state, mapErr := r.apiToModel(result, httpResp, plan.EnvironmentId.ValueString(), &plan)
	if mapErr != nil {
		resp.Diagnostics.AddError(
			"Error Updating Propagation Store",
			fmt.Sprintf("Could not map API response to state: %s", mapErr),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *propagationStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customtypes.PropagationStoreModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	httpResp, err := apiClient.PropagationStoresApi.
		DeletePropagationStore(ctx, state.EnvironmentId.ValueString(), state.Id.ValueString()).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Propagation Store",
			fmt.Sprintf("Could not delete propagation store: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}
}

func (r *propagationStoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := utils.SplitImportID(req.ID, 2)
	if idParts == nil {
		resp.Diagnostics.AddError(
			"Error Importing Propagation Store",
			fmt.Sprintf("Unexpected import identifier format: %s. Expected '<environment_id>/<store_id>'.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func (r *propagationStoreResource) apiToModel(apiObj *management.PropagationStore, httpResp *http.Response, environmentId string, plan *customtypes.PropagationStoreModel) (customtypes.PropagationStoreModel, error) {
	preferredType := ""
	if plan != nil && !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		preferredType = plan.Type.ValueString()
	}

	rawType, rawStatus, err := utils.ExtractPropagationStoreTypeStatus(httpResp)
	if err != nil {
		return customtypes.PropagationStoreModel{}, err
	}

	apiType := rawType
	if apiType == "" || apiType == "UNKNOWN" {
		apiType = string(apiObj.GetType())
	}
	if apiType == "" || apiType == "UNKNOWN" {
		apiType = preferredType
	}
	tfType := utils.NormalizePropagationStoreTypeForTerraform(apiType, preferredType)

	model := customtypes.PropagationStoreModel{
		Id:            types.StringValue(apiObj.GetId()),
		EnvironmentId: types.StringValue(environmentId),
		Name:          types.StringValue(apiObj.GetName()),
		Type:          types.StringValue(tfType),
	}

	if v, ok := apiObj.GetDescriptionOk(); ok {
		model.Description = types.StringValue(*v)
	} else {
		model.Description = types.StringNull()
	}

	if image, ok := apiObj.GetImageOk(); ok && image != nil {
		if v, ok := image.GetIdOk(); ok && v != nil {
			model.ImageId = types.StringValue(*v)
		} else {
			model.ImageId = types.StringNull()
		}

		if v, ok := image.GetHrefOk(); ok && v != nil {
			model.ImageHref = types.StringValue(*v)
		} else {
			model.ImageHref = types.StringNull()
		}
	} else {
		model.ImageId = types.StringNull()
		model.ImageHref = types.StringNull()
	}

	if v, ok := apiObj.GetManagedOk(); ok {
		model.Managed = types.BoolValue(*v)
	} else {
		model.Managed = types.BoolNull()
	}

	status := rawStatus
	if status == "" {
		if v, ok := apiObj.GetStatusOk(); ok && v != nil && string(*v) != "" && string(*v) != "UNKNOWN" {
			status = string(*v)
		}
	}
	if status == "" && plan != nil && !plan.Status.IsNull() && !plan.Status.IsUnknown() && plan.Status.ValueString() != "" {
		status = plan.Status.ValueString()
	}
	if status != "" {
		model.Status = types.StringValue(status)
	} else {
		model.Status = types.StringNull()
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

		model.SyncStatus = types.ObjectValueMust(map[string]attr.Type{
			"last_sync_time": types.StringType,
			"next_sync_time": types.StringType,
			"status":         types.StringType,
			"details":        types.StringType,
		}, attrs)
	} else {
		model.SyncStatus = types.ObjectNull(map[string]attr.Type{
			"last_sync_time": types.StringType,
			"next_sync_time": types.StringType,
			"status":         types.StringType,
			"details":        types.StringType,
		})
	}

	config := apiObj.GetConfiguration()
	mappers.ApplyPropagationStoreConfigurationFromMap(&model, tfType, config, plan)

	return model, nil
}

func buildPropagationStorePayload(plan *customtypes.PropagationStoreModel, config map[string]interface{}) *management.PropagationStore {
	storeType := ""
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		storeType = plan.Type.ValueString()
	}

	payload := management.NewPropagationStore(
		config,
		plan.Name.ValueString(),
		management.EnumPropagationStoreType(utils.NormalizePropagationStoreTypeForAPI(storeType)),
	)

	if !plan.Description.IsNull() {
		payload.SetDescription(plan.Description.ValueString())
	}

	if !plan.Managed.IsNull() {
		payload.SetManaged(plan.Managed.ValueBool())
	}

	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		statusValue := plan.Status.ValueString()
		if statusValue != "" {
			payload.SetStatus(management.EnumPropagationStoreStatus(statusValue))
		}
	}

	if !plan.ImageId.IsNull() && !plan.ImageId.IsUnknown() && plan.ImageId.ValueString() != "" {
		image := management.NewPropagationStoreImage()
		image.SetId(plan.ImageId.ValueString())
		payload.SetImage(*image)
	}

	return payload
}
