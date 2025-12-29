package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/providerdata"
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/utils"
)

var (
	_ resource.Resource                = &propagationPlanResource{}
	_ resource.ResourceWithConfigure   = &propagationPlanResource{}
	_ resource.ResourceWithImportState = &propagationPlanResource{}
)

type propagationPlanResource struct {
	client *providerdata.Client
}

func NewPropagationPlanResource() resource.Resource {
	return &propagationPlanResource{}
}

func (r *propagationPlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_propagation_plan"
}

func (r *propagationPlanResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PingOne provisioning propagation plan.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the propagation plan.",
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
				Description: "Unique name of the propagation plan.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the propagation plan.",
				Computed:    true,
			},
		},
	}
}

func (r *propagationPlanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clientData, ok := req.ProviderData.(*providerdata.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerdata.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = clientData
}

func (r *propagationPlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.PropagationPlanModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API

	payload := management.NewIdentityPropagationPlan(plan.Name.ValueString())
	result, httpResp, err := apiClient.IdentityPropagationPlansApi.
		CreatePlan(ctx, plan.EnvironmentId.ValueString()).
		IdentityPropagationPlan(*payload).
		Execute()
	if err != nil {
		if isPropagationPlanEnvironmentAlreadyHasPlanError(httpResp) {
			detail := "A propagation plan already exists in this environment. Import the existing plan into state or use the propagation plan data source."

			if existingPlan, listErr := readSingletonPropagationPlan(ctx, apiClient, plan.EnvironmentId.ValueString()); listErr == nil {
				detail = fmt.Sprintf(
					"Propagation plan %q (%s) already exists in this environment. Import it into state or use the propagation plan data source.",
					existingPlan.GetName(),
					existingPlan.GetId(),
				)
			}

			resp.Diagnostics.AddError(
				"Propagation Plan Already Exists",
				detail,
			)
			return
		}

		resp.Diagnostics.AddError(
			"Error Creating Propagation Plan",
			fmt.Sprintf("Could not create propagation plan: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	state := propagationPlanFromAPI(result, plan.EnvironmentId.ValueString())

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func isPropagationPlanEnvironmentAlreadyHasPlanError(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusConflict {
		return false
	}

	decoded, err := utils.DecodeResponseJSON(resp)
	if err != nil || decoded == nil {
		return false
	}

	root, ok := decoded.(map[string]interface{})
	if !ok {
		return false
	}

	// Some PingOne errors include the message at the root; others include a details array.
	if msg, ok := root["message"].(string); ok {
		if strings.Contains(strings.ToLower(msg), "existing plan") {
			return true
		}
	}

	detailsRaw, ok := root["details"]
	if !ok || detailsRaw == nil {
		return false
	}

	details, ok := detailsRaw.([]interface{})
	if !ok {
		return false
	}

	for _, detailRaw := range details {
		detail, ok := detailRaw.(map[string]interface{})
		if !ok {
			continue
		}

		msg, _ := detail["message"].(string)
		target, _ := detail["target"].(string)

		if target == "environment" && strings.Contains(strings.ToLower(msg), "existing plan") {
			return true
		}
	}

	return false
}

func readSingletonPropagationPlan(ctx context.Context, apiClient *management.APIClient, environmentID string) (*management.IdentityPropagationPlan, error) {
	iterator := apiClient.IdentityPropagationPlansApi.ReadAllPlans(ctx, environmentID).Execute()

	var plans []management.IdentityPropagationPlan

	for cursor, iterErr := range iterator {
		if iterErr != nil {
			return nil, iterErr
		}
		if cursor.EntityArray == nil {
			continue
		}

		embedded, ok := cursor.EntityArray.GetEmbeddedOk()
		if !ok || embedded == nil {
			continue
		}

		embeddedPlans, ok := embedded.GetPlansOk()
		if !ok || embeddedPlans == nil {
			continue
		}

		plans = append(plans, embeddedPlans...)
	}

	if len(plans) == 0 {
		return nil, fmt.Errorf("no propagation plan found in environment %q", environmentID)
	}
	if len(plans) > 1 {
		return nil, fmt.Errorf("found %d propagation plans in environment %q; expected exactly 1", len(plans), environmentID)
	}

	plan := plans[0]
	return &plan, nil
}

func (r *propagationPlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.PropagationPlanModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	planID := state.Id.ValueString()
	environmentID := state.EnvironmentId.ValueString()

	result, httpResp, err := apiClient.IdentityPropagationPlansApi.
		ReadOnePlan(ctx, environmentID, planID).
		Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Propagation Plan",
			fmt.Sprintf("Could not read propagation plan: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	newState := propagationPlanFromAPI(result, environmentID)

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r *propagationPlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.PropagationPlanModel
	var state customtypes.PropagationPlanModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	environmentID := state.EnvironmentId.ValueString()
	planID := state.Id.ValueString()

	payload := management.NewIdentityPropagationPlan(plan.Name.ValueString())
	result, httpResp, err := apiClient.IdentityPropagationPlansApi.
		UpdatePlan(ctx, environmentID, planID).
		IdentityPropagationPlan(*payload).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Propagation Plan",
			fmt.Sprintf("Could not update propagation plan: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	newState := propagationPlanFromAPI(result, environmentID)

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r *propagationPlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customtypes.PropagationPlanModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	httpResp, err := apiClient.IdentityPropagationPlansApi.
		DeletePlan(ctx, state.EnvironmentId.ValueString(), state.Id.ValueString()).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Propagation Plan",
			fmt.Sprintf("Could not delete propagation plan: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}
}

func (r *propagationPlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := utils.SplitImportID(req.ID, 2)
	if idParts == nil {
		resp.Diagnostics.AddError(
			"Error Importing Propagation Plan",
			fmt.Sprintf("Unexpected import identifier format: %s. Expected '<environment_id>/<plan_id>'.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func propagationPlanFromAPI(apiObj *management.IdentityPropagationPlan, environmentID string) customtypes.PropagationPlanModel {
	model := customtypes.PropagationPlanModel{
		Id:            types.StringValue(apiObj.GetId()),
		EnvironmentId: types.StringValue(environmentID),
		Name:          types.StringValue(apiObj.GetName()),
		Status:        types.StringNull(),
	}

	if v, ok := apiObj.GetStatusOk(); ok && v != nil {
		model.Status = types.StringValue(string(*v))
	}

	return model
}
