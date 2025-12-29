package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &propagationRuleResource{}
	_ resource.ResourceWithConfigure   = &propagationRuleResource{}
	_ resource.ResourceWithImportState = &propagationRuleResource{}
)

type propagationRuleResource struct {
	client *providerdata.Client
}

func NewPropagationRuleResource() resource.Resource {
	return &propagationRuleResource{}
}

func (r *propagationRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_propagation_rule"
}

func (r *propagationRuleResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PingOne provisioning propagation rule and its mappings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the propagation rule.",
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
			"plan_id": schema.StringAttribute{
				Description: "The ID of the propagation plan.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the propagation rule.",
				Required:    true,
			},
			"source_store_id": schema.StringAttribute{
				Description: "The source store ID for the propagation rule.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_store_id": schema.StringAttribute{
				Description: "The target store ID for the propagation rule.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the propagation rule is active.",
				Optional:    true,
			},
			"filter": schema.StringAttribute{
				Description: "Expression used by PingOne to select users to synchronize (maps to the API field `populationExpression`).",
				Optional:    true,
			},
			"deprovision": schema.BoolAttribute{
				Description: "Whether to deprovision users in the target store when they are removed from the source.",
				Optional:    true,
			},
			"population_ids": schema.ListAttribute{
				Description: "Optional list of population IDs in scope for this rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"group_ids": schema.ListAttribute{
				Description: "Optional list of group IDs to scope group provisioning for this rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"configuration": schema.MapAttribute{
				Description: "Optional rule configuration map (for example, `MFA_USER_DEVICE_MANAGEMENT`).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"mappings": schema.ListNestedAttribute{
				Description: "Optional list of attribute mappings for this rule.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The mapping ID.",
							Computed:    true,
						},
						"source_attribute": schema.StringAttribute{
							Description: "Source attribute expression.",
							Optional:    true,
						},
						"target_attribute": schema.StringAttribute{
							Description: "Target attribute expression.",
							Optional:    true,
						},
						"expression": schema.StringAttribute{
							Description: "Optional expression used to compute the target attribute value.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *propagationRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *propagationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.PropagationRuleModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mappingsAttr types.List
	diags = req.Plan.GetAttribute(ctx, path.Root("mappings"), &mappingsAttr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	manageMappings := !mappingsAttr.IsNull() && !mappingsAttr.IsUnknown()
	if !manageMappings {
		plan.Mappings = nil
	}

	validationDiags := validatePropagationRuleMappings(plan.Mappings)
	resp.Diagnostics.Append(validationDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	requestClient := apiClient

	payload, payloadDiags := propagationRulePayloadFromModel(ctx, &plan)
	resp.Diagnostics.Append(payloadDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := plan.EnvironmentId.ValueString()

	desiredActive := !plan.Active.IsNull() && !plan.Active.IsUnknown() && plan.Active.ValueBool()
	manageConfiguration := !plan.Configuration.IsNull() && !plan.Configuration.IsUnknown()

	payloadForCreate := cloneInterfaceMap(payload)
	// The plan-scoped create endpoint requires mappings to exist before enabling a rule.
	// Create inactive first, then apply mappings, then enable via update if desired.
	payloadForCreate["active"] = false
	delete(payloadForCreate, "configuration")

	ruleID, httpResp, err := createPropagationRuleViaPlan(ctx, requestClient, environmentID, plan.PlanId.ValueString(), payloadForCreate)
	if err != nil && shouldTryAlternateHostname(err, httpResp) {
		for _, hostname := range pingOneFallbackBaseHostnames(apiClient) {
			altClient, altErr := cloneManagementClientWithBaseHostname(apiClient, hostname)
			if altErr != nil || altClient == nil {
				continue
			}

			altRuleID, altResp, altReqErr := createPropagationRuleViaPlan(ctx, altClient, environmentID, plan.PlanId.ValueString(), payloadForCreate)
			httpResp = altResp
			err = altReqErr
			ruleID = altRuleID

			if err == nil {
				requestClient = altClient
				break
			}
			if !shouldTryAlternateHostname(err, httpResp) {
				break
			}
		}
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Propagation Rule",
			fmt.Sprintf("Could not create propagation rule: %s", err),
		)
		return
	}

	if manageMappings && len(plan.Mappings) > 0 {
		if err := ensurePropagationRuleMappings(ctx, requestClient, plan.EnvironmentId.ValueString(), ruleID, nil, plan.Mappings); err != nil {
			resp.Diagnostics.AddError(
				"Error Creating Propagation Rule Mappings",
				fmt.Sprintf("Could not reconcile mappings: %s", err),
			)
			return
		}
	}

	if desiredActive || manageConfiguration {
		updatePayload := cloneInterfaceMap(payload)
		if desiredActive {
			updatePayload["active"] = true
			updatePayload["populationExpression"] = populationExpressionForModel(ctx, &plan)
		}

		updateResp, updateErr := requestClient.PropagationRulesApi.
			EnvironmentsEnvironmentIDPropagationRulesStoreIDPut(ctx, environmentID, ruleID).
			Body(updatePayload).
			Execute()
		if updateErr != nil {
			action := "update"
			if desiredActive {
				action = "enable"
			}
			resp.Diagnostics.AddError(
				"Error Updating Propagation Rule",
				fmt.Sprintf("Could not %s propagation rule: %s", action, utils.HandleSDKError(updateErr, updateResp)),
			)
			return
		}
	}

	state := plan
	state.Id = types.StringValue(ruleID)

	if manageMappings {
		resolvedMappings, err := resolvePropagationRuleMappings(ctx, requestClient, state.EnvironmentId.ValueString(), ruleID, state.Mappings)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Propagation Rule Mappings",
				fmt.Sprintf("Could not read mappings: %s", err),
			)
			return
		}
		state.Mappings = resolvedMappings
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, revErr := createPropagationRevisionWithFallback(ctx, requestClient, environmentID); revErr != nil {
		resp.Diagnostics.AddWarning(
			"Propagation Revision Not Created",
			fmt.Sprintf("Rule was created, but the provider could not create a propagation revision. The PingOne UI may not reflect the latest propagation configuration until a revision is created. Error: %s", revErr),
		)
	}
}

func (r *propagationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.PropagationRuleModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	environmentID := state.EnvironmentId.ValueString()
	ruleID := state.Id.ValueString()

	ruleObj, httpResp, err := readPropagationRule(ctx, apiClient, environmentID, ruleID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Propagation Rule",
			fmt.Sprintf("Could not read propagation rule: %s", err),
		)
		return
	}

	apiDiags := applyRuleAPIToState(ctx, ruleObj, &state)
	resp.Diagnostics.Append(apiDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Mappings != nil {
		resolvedMappings, err := resolvePropagationRuleMappings(ctx, apiClient, environmentID, ruleID, state.Mappings)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Propagation Rule Mappings",
				fmt.Sprintf("Could not read mappings: %s", err),
			)
			return
		}
		state.Mappings = resolvedMappings
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *propagationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.PropagationRuleModel
	var state customtypes.PropagationRuleModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mappingsAttr types.List
	diags = req.Plan.GetAttribute(ctx, path.Root("mappings"), &mappingsAttr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	manageMappings := !mappingsAttr.IsNull() && !mappingsAttr.IsUnknown()
	if !manageMappings {
		plan.Mappings = nil
	}

	validationDiags := validatePropagationRuleMappings(plan.Mappings)
	resp.Diagnostics.Append(validationDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	environmentID := state.EnvironmentId.ValueString()
	ruleID := state.Id.ValueString()

	desiredActive := !plan.Active.IsNull() && !plan.Active.IsUnknown() && plan.Active.ValueBool()

	if manageMappings {
		if err := ensurePropagationRuleMappings(ctx, apiClient, environmentID, ruleID, state.Mappings, plan.Mappings); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Propagation Rule Mappings",
				fmt.Sprintf("Could not reconcile mappings: %s", err),
			)
			return
		}
	}

	payload, payloadDiags := propagationRulePayloadFromModel(ctx, &plan)
	resp.Diagnostics.Append(payloadDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if desiredActive {
		payload["populationExpression"] = populationExpressionForModel(ctx, &plan)
	}

	httpResp, err := apiClient.PropagationRulesApi.
		EnvironmentsEnvironmentIDPropagationRulesStoreIDPut(ctx, environmentID, ruleID).
		Body(payload).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Propagation Rule",
			fmt.Sprintf("Could not update propagation rule: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	newState := plan
	newState.Id = types.StringValue(ruleID)

	if manageMappings {
		resolvedMappings, err := resolvePropagationRuleMappings(ctx, apiClient, environmentID, ruleID, newState.Mappings)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Propagation Rule Mappings",
				fmt.Sprintf("Could not read mappings: %s", err),
			)
			return
		}
		newState.Mappings = resolvedMappings
	} else {
		newState.Mappings = nil
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, revErr := createPropagationRevisionWithFallback(ctx, apiClient, environmentID); revErr != nil {
		resp.Diagnostics.AddWarning(
			"Propagation Revision Not Created",
			fmt.Sprintf("Rule was updated, but the provider could not create a propagation revision. The PingOne UI may not reflect the latest propagation configuration until a revision is created. Error: %s", revErr),
		)
	}
}

func (r *propagationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customtypes.PropagationRuleModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client.API
	environmentID := state.EnvironmentId.ValueString()
	ruleID := state.Id.ValueString()

	_ = deleteAllMappings(ctx, apiClient, environmentID, ruleID)

	httpResp, err := apiClient.PropagationRulesApi.
		EnvironmentsEnvironmentIDPropagationRulesRuleIDDelete(ctx, environmentID, ruleID).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Propagation Rule",
			fmt.Sprintf("Could not delete propagation rule: %s", utils.HandleSDKError(err, httpResp)),
		)
		return
	}

	if _, revErr := createPropagationRevisionWithFallback(ctx, apiClient, environmentID); revErr != nil {
		resp.Diagnostics.AddWarning(
			"Propagation Revision Not Created",
			fmt.Sprintf("Rule was deleted, but the provider could not create a propagation revision. The PingOne UI may not reflect the latest propagation configuration until a revision is created. Error: %s", revErr),
		)
	}
}

func (r *propagationRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := utils.SplitImportID(req.ID, 2)
	if idParts == nil {
		resp.Diagnostics.AddError(
			"Error Importing Propagation Rule",
			fmt.Sprintf("Unexpected import identifier format: %s. Expected '<environment_id>/<rule_id>'.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func propagationRulePayloadFromModel(ctx context.Context, model *customtypes.PropagationRuleModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := map[string]interface{}{
		"plan": map[string]interface{}{
			"id": model.PlanId.ValueString(),
		},
		"environment": map[string]interface{}{
			"id": model.EnvironmentId.ValueString(),
		},
		"sourceStore": map[string]interface{}{
			"id": model.SourceStoreId.ValueString(),
		},
		"targetStore": map[string]interface{}{
			"id": model.TargetStoreId.ValueString(),
		},
		"name": model.Name.ValueString(),
	}

	if !model.Active.IsNull() && !model.Active.IsUnknown() {
		payload["active"] = model.Active.ValueBool()
	}
	if expr, ok := populationExpressionFromModel(ctx, model); ok {
		payload["populationExpression"] = expr
	}
	if !model.Deprovision.IsNull() && !model.Deprovision.IsUnknown() {
		payload["deprovision"] = model.Deprovision.ValueBool()
	}
	if !model.GroupIds.IsNull() && !model.GroupIds.IsUnknown() {
		var groupIDs []string
		listDiags := model.GroupIds.ElementsAs(ctx, &groupIDs, false)
		diags.Append(listDiags...)
		if !listDiags.HasError() {
			payload["groups"] = groupRefsFromIDs(groupIDs)
		}
	}
	if !model.Configuration.IsNull() && !model.Configuration.IsUnknown() {
		var cfg map[string]string
		mapDiags := model.Configuration.ElementsAs(ctx, &cfg, false)
		diags.Append(mapDiags...)
		if !mapDiags.HasError() {
			payload["configuration"] = cfg
		}
	}

	return payload, diags
}

func validatePropagationRuleMappings(mappings []customtypes.PropagationRuleMappingModel) diag.Diagnostics {
	var diags diag.Diagnostics

	for i, m := range mappings {
		source := ""
		if !m.SourceAttribute.IsNull() && !m.SourceAttribute.IsUnknown() {
			source = strings.TrimSpace(m.SourceAttribute.ValueString())
		}
		target := ""
		if !m.TargetAttribute.IsNull() && !m.TargetAttribute.IsUnknown() {
			target = strings.TrimSpace(m.TargetAttribute.ValueString())
		}
		expression := ""
		if !m.Expression.IsNull() && !m.Expression.IsUnknown() {
			expression = strings.TrimSpace(m.Expression.ValueString())
		}

		if target == "" {
			diags.AddAttributeError(
				path.Root("mappings").AtListIndex(i).AtName("target_attribute"),
				"Missing Required Argument",
				"`target_attribute` must be set for each mapping.",
			)
		}

		if source == "" && expression == "" {
			diags.AddAttributeError(
				path.Root("mappings").AtListIndex(i).AtName("source_attribute"),
				"Missing Required Argument",
				"Either `source_attribute` or `expression` must be set for each mapping.",
			)
			diags.AddAttributeError(
				path.Root("mappings").AtListIndex(i).AtName("expression"),
				"Missing Required Argument",
				"Either `source_attribute` or `expression` must be set for each mapping.",
			)
		}

		if source != "" && expression != "" {
			diags.AddAttributeError(
				path.Root("mappings").AtListIndex(i).AtName("expression"),
				"Conflicting Arguments",
				"`expression` cannot be used when `source_attribute` is set; choose one.",
			)
		}
	}

	return diags
}

func isBadAuthorizationHeaderError(err error, httpResp *http.Response) bool {
	if err == nil || httpResp == nil {
		return false
	}
	if httpResp.StatusCode != http.StatusForbidden {
		return false
	}

	body, readErr := utils.ReadAndRestoreResponseBody(httpResp)
	if readErr != nil {
		return false
	}

	msg := strings.ToLower(string(body))
	return strings.Contains(msg, "invalid key=value pair") && strings.Contains(msg, "authorization header")
}

func shouldTryAlternateHostname(err error, httpResp *http.Response) bool {
	if err == nil || httpResp == nil {
		return false
	}

	if isBadAuthorizationHeaderError(err, httpResp) {
		return true
	}

	// Region/hostname mismatches often manifest as NOT_FOUND for plan/store IDs.
	return httpResp.StatusCode == http.StatusNotFound
}

func createPropagationRuleViaPlan(ctx context.Context, apiClient *management.APIClient, environmentID string, planID string, payload map[string]interface{}) (string, *http.Response, error) {
	if apiClient == nil {
		return "", nil, fmt.Errorf("nil api client")
	}

	httpResp, err := createPropagationRuleForPlan(ctx, apiClient, environmentID, planID, payload)
	if err != nil {
		if httpResp != nil && (httpResp.StatusCode == http.StatusNotFound || httpResp.StatusCode == http.StatusMethodNotAllowed) {
			httpResp, err = apiClient.PropagationRulesApi.
				EnvironmentsEnvironmentIDPropagationRulesPost(ctx, environmentID).
				Body(payload).
				Execute()
			if err != nil {
				return "", httpResp, fmt.Errorf("%s", utils.HandleSDKError(err, httpResp))
			}
		} else {
			return "", httpResp, err
		}
	}

	name, _ := payload["name"].(string)
	sourceStoreID, _ := utils.NestedString(payload, "sourceStore", "id")
	targetStoreID, _ := utils.NestedString(payload, "targetStore", "id")

	ruleID, err := propagationRuleIDFromCreateResponse(ctx, apiClient, environmentID, planID, name, sourceStoreID, targetStoreID, httpResp)
	if err != nil {
		return "", httpResp, err
	}

	return strings.TrimSpace(ruleID), httpResp, nil
}

func createPropagationRuleForPlan(ctx context.Context, apiClient *management.APIClient, environmentID string, planID string, payload map[string]interface{}) (*http.Response, error) {
	cfg := apiClient.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("api client has nil config")
	}
	if cfg.HTTPClient == nil {
		return nil, fmt.Errorf("api client has nil http client")
	}

	basePath, err := cfg.ServerURLWithContext(ctx, "PropagationRulesApiService.EnvironmentsEnvironmentIDPropagationRulesPost")
	if err != nil {
		return nil, err
	}
	basePath = strings.TrimRight(strings.TrimSpace(basePath), "/")

	endpoint := fmt.Sprintf(
		"%s/environments/%s/propagation/plans/%s/rules",
		basePath,
		url.PathEscape(environmentID),
		url.PathEscape(planID),
	)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
		if readErr != nil {
			return resp, readErr
		}
	}

	if resp.StatusCode >= 300 {
		return resp, fmt.Errorf("%s", utils.HandleSDKError(fmt.Errorf("%s", resp.Status), resp))
	}

	return resp, nil
}

func createPropagationRevision(ctx context.Context, apiClient *management.APIClient, environmentID string) (*http.Response, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("nil api client")
	}

	httpResp, err := apiClient.PropagationRevisionsApi.
		EnvironmentsEnvironmentIDPropagationRevisionsPost(ctx, environmentID).
		Execute()
	if err != nil {
		return httpResp, fmt.Errorf("%s", utils.HandleSDKError(err, httpResp))
	}

	return httpResp, nil
}

func createPropagationRevisionWithFallback(ctx context.Context, apiClient *management.APIClient, environmentID string) (*http.Response, error) {
	httpResp, err := createPropagationRevision(ctx, apiClient, environmentID)
	if err == nil {
		return httpResp, nil
	}

	if shouldTryAlternateHostname(err, httpResp) {
		for _, hostname := range pingOneFallbackBaseHostnames(apiClient) {
			altClient, altErr := cloneManagementClientWithBaseHostname(apiClient, hostname)
			if altErr != nil || altClient == nil {
				continue
			}

			httpResp, err = createPropagationRevision(ctx, altClient, environmentID)
			if err == nil {
				return httpResp, nil
			}
			if !shouldTryAlternateHostname(err, httpResp) {
				break
			}
		}
	}

	return httpResp, err
}

func deletePropagationMapping(ctx context.Context, apiClient *management.APIClient, environmentID string, mappingID string) (*http.Response, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("nil api client")
	}

	cfg := apiClient.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("api client has nil config")
	}
	if cfg.HTTPClient == nil {
		return nil, fmt.Errorf("api client has nil http client")
	}

	// Use a known-good server definition to derive the base path (the SDK's DELETE endpoint is currently incorrect).
	basePath, err := cfg.ServerURLWithContext(ctx, "PropagationMappingsApiService.EnvironmentsEnvironmentIDPropagationMappingsMappingIDGet")
	if err != nil {
		return nil, err
	}
	basePath = normalizePropagationMappingBasePath(basePath)

	endpoint, err := url.JoinPath(
		basePath,
		"environments",
		environmentID,
		"propagation",
		"mappings",
		mappingID,
	)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.Body != nil {
		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
		if readErr != nil {
			return resp, readErr
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		return resp, nil
	}

	if resp.StatusCode >= 300 {
		return resp, fmt.Errorf("%s", utils.HandleSDKError(fmt.Errorf("%s", resp.Status), resp))
	}

	return resp, nil
}

func normalizePropagationMappingBasePath(basePath string) string {
	basePath = strings.TrimRight(strings.TrimSpace(basePath), "/")

	// Trim any accidental propagation mapping suffix to keep the base path stable.
	for _, suffix := range []string{"/propagation/mapping", "/propagation/mappings"} {
		if strings.HasSuffix(basePath, suffix) {
			return strings.TrimSuffix(basePath, suffix)
		}
	}

	return basePath
}

func deletePropagationMappingWithFallback(ctx context.Context, apiClient *management.APIClient, environmentID string, mappingID string) (*http.Response, error) {
	httpResp, err := deletePropagationMapping(ctx, apiClient, environmentID, mappingID)
	if err == nil {
		return httpResp, nil
	}

	if shouldTryAlternateHostname(err, httpResp) {
		for _, hostname := range pingOneFallbackBaseHostnames(apiClient) {
			altClient, altErr := cloneManagementClientWithBaseHostname(apiClient, hostname)
			if altErr != nil || altClient == nil {
				continue
			}

			httpResp, err = deletePropagationMapping(ctx, altClient, environmentID, mappingID)
			if err == nil {
				return httpResp, nil
			}
			if !shouldTryAlternateHostname(err, httpResp) {
				break
			}
		}
	}

	return httpResp, err
}

func cloneInterfaceMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func populationExpressionFromModel(ctx context.Context, model *customtypes.PropagationRuleModel) (string, bool) {
	if model == nil {
		return "", false
	}

	filter := ""
	if !model.Filter.IsNull() && !model.Filter.IsUnknown() {
		filter = strings.TrimSpace(model.Filter.ValueString())
	}

	var popIDs []string
	if !model.PopulationIds.IsNull() && !model.PopulationIds.IsUnknown() {
		if diags := model.PopulationIds.ElementsAs(ctx, &popIDs, false); diags.HasError() {
			popIDs = nil
		}
	}

	var popExprParts []string
	seen := make(map[string]bool)
	for _, id := range popIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		popExprParts = append(popExprParts, fmt.Sprintf("population.id eq %q", id))
	}
	sort.Strings(popExprParts)

	popExpr := ""
	if len(popExprParts) > 0 {
		popExpr = strings.Join(popExprParts, " or ")
	}

	switch {
	case popExpr != "" && filter != "":
		return fmt.Sprintf("(%s) and (%s)", popExpr, filter), true
	case popExpr != "":
		return popExpr, true
	case filter != "":
		return filter, true
	default:
		return "", false
	}
}

func populationExpressionForModel(ctx context.Context, model *customtypes.PropagationRuleModel) string {
	if expr, ok := populationExpressionFromModel(ctx, model); ok {
		return expr
	}
	return "population.id pr"
}

func cloneManagementClientWithBaseHostname(apiClient *management.APIClient, baseHostname string) (*management.APIClient, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("nil api client")
	}

	baseHostname = strings.TrimSpace(baseHostname)
	if baseHostname == "" {
		return nil, fmt.Errorf("empty base hostname")
	}

	origCfg := apiClient.GetConfig()
	if origCfg == nil {
		return nil, fmt.Errorf("api client has nil config")
	}

	cfg := management.NewConfiguration()
	cfg.HTTPClient = origCfg.HTTPClient
	cfg.UserAgent = origCfg.UserAgent
	cfg.Debug = origCfg.Debug
	for k, v := range origCfg.DefaultHeader {
		cfg.DefaultHeader[k] = v
	}

	cfg.SetDefaultServerIndex(1)
	if err := cfg.SetDefaultServerVariableDefaultValue("baseHostname", baseHostname); err != nil {
		return nil, err
	}
	if err := cfg.SetDefaultServerVariableDefaultValue("protocol", "https"); err != nil {
		return nil, err
	}

	return management.NewAPIClient(cfg), nil
}

func pingOneFallbackBaseHostnames(apiClient *management.APIClient) []string {
	current := currentPingOneHostname(apiClient)

	var candidates []string
	switch {
	case strings.HasSuffix(current, ".com.au"):
		candidates = []string{"api.pingone.asia", "api.pingone.com", "api.pingone.eu", "api.pingone.ca"}
	case strings.HasSuffix(current, ".asia"):
		candidates = []string{"api.pingone.com.au", "api.pingone.com", "api.pingone.eu", "api.pingone.ca"}
	case strings.HasSuffix(current, ".eu"):
		candidates = []string{"api.pingone.com", "api.pingone.ca", "api.pingone.asia", "api.pingone.com.au"}
	case strings.HasSuffix(current, ".ca"):
		candidates = []string{"api.pingone.com", "api.pingone.eu", "api.pingone.asia", "api.pingone.com.au"}
	case strings.HasSuffix(current, ".com"):
		candidates = []string{"api.pingone.eu", "api.pingone.ca", "api.pingone.asia", "api.pingone.com.au"}
	default:
		candidates = []string{"api.pingone.com", "api.pingone.eu", "api.pingone.ca", "api.pingone.asia", "api.pingone.com.au"}
	}

	var out []string
	seen := make(map[string]bool)
	for _, host := range candidates {
		host = strings.TrimSpace(host)
		if host == "" || host == current {
			continue
		}
		if seen[host] {
			continue
		}
		seen[host] = true
		out = append(out, host)
	}

	return out
}

func currentPingOneHostname(apiClient *management.APIClient) string {
	if apiClient == nil {
		return ""
	}
	cfg := apiClient.GetConfig()
	if cfg == nil {
		return ""
	}

	switch cfg.DefaultServerIndex {
	case 0:
		if len(cfg.Servers) < 1 {
			return ""
		}
		baseDomain := strings.TrimSpace(cfg.Servers[0].Variables["baseDomain"].DefaultValue)
		suffix := strings.TrimSpace(cfg.Servers[0].Variables["suffix"].DefaultValue)
		baseDomain = strings.TrimSuffix(baseDomain, ".")
		suffix = strings.TrimPrefix(suffix, ".")
		if baseDomain == "" || suffix == "" {
			return ""
		}
		return baseDomain + "." + suffix
	case 1:
		if len(cfg.Servers) < 2 {
			return ""
		}
		return strings.TrimSpace(cfg.Servers[1].Variables["baseHostname"].DefaultValue)
	default:
		return ""
	}
}

func propagationRuleIDFromCreateResponse(ctx context.Context, apiClient *management.APIClient, environmentID string, planID string, name string, sourceStoreID string, targetStoreID string, httpResp *http.Response) (string, error) {
	decoded, err := utils.DecodeResponseJSON(httpResp)
	if err != nil {
		return "", err
	}

	if m, ok := decoded.(map[string]interface{}); ok {
		if id, ok := utils.NestedString(m, "id"); ok && id != "" {
			return id, nil
		}
	}

	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		rules, err := listPropagationRulesForPlan(ctx, apiClient, environmentID, planID)
		if err != nil {
			lastErr = err
		} else {
			var matches []string
			for _, rule := range rules {
				ruleName, _ := utils.NestedString(rule, "name")
				if ruleName != name {
					continue
				}

				if sourceStoreID != "" {
					srcID, _ := utils.NestedString(rule, "sourceStore", "id")
					if srcID != sourceStoreID {
						continue
					}
				}
				if targetStoreID != "" {
					tgtID, _ := utils.NestedString(rule, "targetStore", "id")
					if tgtID != targetStoreID {
						continue
					}
				}

				id, _ := utils.NestedString(rule, "id")
				if id != "" {
					matches = append(matches, id)
				}
			}

			if len(matches) == 1 {
				return matches[0], nil
			}
			if len(matches) > 1 {
				return "", fmt.Errorf("found %d rules matching name=%q source=%q target=%q after create; unable to determine ID", len(matches), name, sourceStoreID, targetStoreID)
			}

			lastErr = fmt.Errorf("rule not found yet")
		}

		time.Sleep(time.Duration(attempt+1) * 300 * time.Millisecond)
	}

	return "", fmt.Errorf("could not locate created rule (name=%q source=%q target=%q): %v", name, sourceStoreID, targetStoreID, lastErr)
}

func readPropagationRule(ctx context.Context, apiClient *management.APIClient, environmentID string, ruleID string) (map[string]interface{}, *http.Response, error) {
	httpResp, err := apiClient.PropagationRulesApi.
		EnvironmentsEnvironmentIDPropagationRulesRuleIDGet(ctx, environmentID, ruleID).
		Execute()
	if err != nil {
		return nil, httpResp, fmt.Errorf("%s", utils.HandleSDKError(err, httpResp))
	}

	decoded, err := utils.DecodeResponseJSON(httpResp)
	if err != nil {
		return nil, httpResp, err
	}

	ruleObj, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, httpResp, fmt.Errorf("unexpected rule response shape")
	}

	return ruleObj, httpResp, nil
}

func listPropagationRulesForPlan(ctx context.Context, apiClient *management.APIClient, environmentID string, planID string) ([]map[string]interface{}, error) {
	httpResp, err := apiClient.PropagationRulesApi.
		EnvironmentsEnvironmentIDPropagationPlansPlanIDRulesGet(ctx, environmentID, planID).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("%s", utils.HandleSDKError(err, httpResp))
	}

	decoded, err := utils.DecodeResponseJSON(httpResp)
	if err != nil {
		return nil, err
	}

	list, err := utils.ExtractEmbeddedArray(decoded, "rules", "items")
	if err != nil {
		return nil, err
	}

	var rules []map[string]interface{}
	for _, v := range list {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		rules = append(rules, m)
	}

	return rules, nil
}

func applyRuleAPIToState(ctx context.Context, apiObj map[string]interface{}, state *customtypes.PropagationRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if id, ok := utils.NestedString(apiObj, "id"); ok && id != "" {
		state.Id = types.StringValue(id)
	}
	if name, ok := utils.NestedString(apiObj, "name"); ok && name != "" {
		state.Name = types.StringValue(name)
	}
	if planID, ok := utils.NestedString(apiObj, "plan", "id"); ok && planID != "" {
		state.PlanId = types.StringValue(planID)
	}
	if srcID, ok := utils.NestedString(apiObj, "sourceStore", "id"); ok && srcID != "" {
		state.SourceStoreId = types.StringValue(srcID)
	}
	if tgtID, ok := utils.NestedString(apiObj, "targetStore", "id"); ok && tgtID != "" {
		state.TargetStoreId = types.StringValue(tgtID)
	}

	if !state.Filter.IsNull() && !state.Filter.IsUnknown() {
		if v, ok := apiObj["populationExpression"]; ok {
			if s, ok := v.(string); ok {
				state.Filter = types.StringValue(s)
			}
		}
	}
	if !state.Active.IsNull() && !state.Active.IsUnknown() {
		if v, ok := apiObj["active"]; ok {
			if b, ok := v.(bool); ok {
				state.Active = types.BoolValue(b)
			}
		}
	}
	if !state.Deprovision.IsNull() && !state.Deprovision.IsUnknown() {
		if v, ok := apiObj["deprovision"]; ok {
			if b, ok := v.(bool); ok {
				state.Deprovision = types.BoolValue(b)
			}
		}
	}

	if !state.Configuration.IsNull() && !state.Configuration.IsUnknown() {
		if v, ok := apiObj["configuration"]; ok && v != nil {
			if rawMap, ok := v.(map[string]interface{}); ok {
				cfg := make(map[string]string, len(rawMap))
				for k, rawVal := range rawMap {
					if s, ok := rawVal.(string); ok {
						cfg[k] = s
					}
				}

				mapVal, mapDiags := types.MapValueFrom(ctx, types.StringType, cfg)
				diags.Append(mapDiags...)
				state.Configuration = mapVal
			}
		}
	}

	if !state.PopulationIds.IsNull() && !state.PopulationIds.IsUnknown() {
		if v, ok := apiObj["populations"]; ok && v != nil {
			if rawList, ok := v.([]interface{}); ok {
				var ids []string
				for _, item := range rawList {
					if m, ok := item.(map[string]interface{}); ok {
						if id, ok := utils.NestedString(m, "id"); ok && id != "" {
							ids = append(ids, id)
						}
					}
				}

				listVal, listDiags := types.ListValueFrom(ctx, types.StringType, ids)
				diags.Append(listDiags...)
				state.PopulationIds = listVal
			}
		}
	}
	if !state.GroupIds.IsNull() && !state.GroupIds.IsUnknown() {
		if v, ok := apiObj["groups"]; ok && v != nil {
			if rawList, ok := v.([]interface{}); ok {
				var ids []string
				for _, item := range rawList {
					if m, ok := item.(map[string]interface{}); ok {
						if id, ok := utils.NestedString(m, "id"); ok && id != "" {
							ids = append(ids, id)
						}
					}
				}
				sort.Strings(ids)

				listVal, listDiags := types.ListValueFrom(ctx, types.StringType, ids)
				diags.Append(listDiags...)
				state.GroupIds = listVal
			}
		}
	}

	return diags
}

func groupRefsFromIDs(ids []string) []map[string]interface{} {
	if len(ids) == 0 {
		return []map[string]interface{}{}
	}

	seen := make(map[string]bool, len(ids))
	var filtered []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		filtered = append(filtered, id)
	}
	sort.Strings(filtered)

	out := make([]map[string]interface{}, 0, len(filtered))
	for _, id := range filtered {
		out = append(out, map[string]interface{}{"id": id})
	}

	return out
}

func ensurePropagationRuleMappings(ctx context.Context, apiClient *management.APIClient, environmentID string, ruleID string, prior []customtypes.PropagationRuleMappingModel, desired []customtypes.PropagationRuleMappingModel) error {
	existing, err := listPropagationRuleMappings(ctx, apiClient, environmentID, ruleID)
	if err != nil {
		return err
	}

	existingByKey := make(map[string]map[string]interface{})
	for _, m := range existing {
		key := mappingKeyFromAPI(m)
		if key == "" {
			continue
		}
		existingByKey[key] = m
	}

	desiredKeys := make(map[string]customtypes.PropagationRuleMappingModel)
	for _, m := range desired {
		key := mappingKey(m.SourceAttribute.ValueString(), m.TargetAttribute.ValueString(), m.Expression.ValueString())
		if key == "" {
			continue
		}
		desiredKeys[key] = m
	}

	// Delete mappings not desired.
	for key, m := range existingByKey {
		if _, ok := desiredKeys[key]; ok {
			continue
		}
		id, _ := utils.NestedString(m, "id")
		if id == "" {
			continue
		}
		delResp, delErr := deletePropagationMappingWithFallback(ctx, apiClient, environmentID, id)
		if delErr != nil {
			return fmt.Errorf("delete mapping %s: %s", id, utils.HandleSDKError(delErr, delResp))
		}
	}

	// Create mappings that are missing.
	requestClient := apiClient
	for key, m := range desiredKeys {
		if _, ok := existingByKey[key]; ok {
			continue
		}

		source := strings.TrimSpace(m.SourceAttribute.ValueString())
		target := strings.TrimSpace(m.TargetAttribute.ValueString())
		expression := strings.TrimSpace(m.Expression.ValueString())

		payload := map[string]interface{}{
			"targetAttribute": target,
		}
		if expression != "" {
			payload["expression"] = expression
		} else {
			payload["sourceAttribute"] = source
		}

		httpResp, createErr := requestClient.PropagationMappingsApi.
			EnvironmentsEnvironmentIDPropagationRulesRuleIDMappingsPost(ctx, environmentID, ruleID).
			Body(payload).
			Execute()
		if createErr != nil && shouldTryAlternateHostname(createErr, httpResp) {
			for _, hostname := range pingOneFallbackBaseHostnames(requestClient) {
				altClient, altErr := cloneManagementClientWithBaseHostname(requestClient, hostname)
				if altErr != nil || altClient == nil {
					continue
				}

				altResp, altReqErr := altClient.PropagationMappingsApi.
					EnvironmentsEnvironmentIDPropagationRulesRuleIDMappingsPost(ctx, environmentID, ruleID).
					Body(payload).
					Execute()

				httpResp = altResp
				createErr = altReqErr

				if createErr == nil {
					requestClient = altClient
					break
				}
				if !shouldTryAlternateHostname(createErr, httpResp) {
					break
				}
			}
		}
		if createErr != nil {
			return fmt.Errorf("create mapping %s: %s", key, utils.HandleSDKError(createErr, httpResp))
		}
	}

	return nil
}

func resolvePropagationRuleMappings(ctx context.Context, apiClient *management.APIClient, environmentID string, ruleID string, preferredOrder []customtypes.PropagationRuleMappingModel) ([]customtypes.PropagationRuleMappingModel, error) {
	existing, err := listPropagationRuleMappings(ctx, apiClient, environmentID, ruleID)
	if err != nil {
		return nil, err
	}

	existingByKey := make(map[string]customtypes.PropagationRuleMappingModel)
	for _, m := range existing {
		id, _ := utils.NestedString(m, "id")
		source, _ := utils.NestedString(m, "sourceAttribute")
		target, _ := utils.NestedString(m, "targetAttribute")
		expression, _ := utils.NestedString(m, "expression")

		key := mappingKey(source, target, expression)
		if key == "" {
			continue
		}

		mapping := customtypes.PropagationRuleMappingModel{
			Id:              types.StringValue(id),
			TargetAttribute: types.StringValue(target),
		}
		if strings.TrimSpace(source) != "" {
			mapping.SourceAttribute = types.StringValue(source)
		} else {
			mapping.SourceAttribute = types.StringNull()
		}
		if strings.TrimSpace(expression) != "" {
			mapping.Expression = types.StringValue(expression)
		} else {
			mapping.Expression = types.StringNull()
		}

		existingByKey[key] = mapping
	}

	var resolved []customtypes.PropagationRuleMappingModel
	seen := make(map[string]bool)

	for _, preferred := range preferredOrder {
		source := preferred.SourceAttribute.ValueString()
		target := preferred.TargetAttribute.ValueString()
		expression := preferred.Expression.ValueString()
		key := mappingKey(source, target, expression)
		if key == "" {
			continue
		}
		if v, ok := existingByKey[key]; ok {
			resolved = append(resolved, v)
		} else {
			// Preserve configured mapping even if API doesn't return it yet.
			resolved = append(resolved, preferred)
		}
		seen[key] = true
	}

	// Add any remaining mappings in a stable order.
	var remainingKeys []string
	for key := range existingByKey {
		if seen[key] {
			continue
		}
		remainingKeys = append(remainingKeys, key)
	}
	sort.Strings(remainingKeys)
	for _, key := range remainingKeys {
		resolved = append(resolved, existingByKey[key])
	}

	return resolved, nil
}

func deleteAllMappings(ctx context.Context, apiClient *management.APIClient, environmentID string, ruleID string) error {
	mappings, err := listPropagationRuleMappings(ctx, apiClient, environmentID, ruleID)
	if err != nil {
		return err
	}

	for _, m := range mappings {
		id, _ := utils.NestedString(m, "id")
		if id == "" {
			continue
		}
		_, _ = deletePropagationMappingWithFallback(ctx, apiClient, environmentID, id)
	}
	return nil
}

func listPropagationRuleMappings(ctx context.Context, apiClient *management.APIClient, environmentID string, ruleID string) ([]map[string]interface{}, error) {
	httpResp, err := apiClient.PropagationMappingsApi.
		EnvironmentsEnvironmentIDPropagationRulesRuleIDMappingsGet(ctx, environmentID, ruleID).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("%s", utils.HandleSDKError(err, httpResp))
	}

	decoded, err := utils.DecodeResponseJSON(httpResp)
	if err != nil {
		return nil, err
	}

	list, err := utils.ExtractEmbeddedArray(decoded, "mappings", "items")
	if err != nil {
		// Some endpoints return a raw array.
		if arr, ok := decoded.([]interface{}); ok {
			list = arr
		} else {
			return nil, err
		}
	}

	var mappings []map[string]interface{}
	for _, v := range list {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		// Ensure the map uses consistent key casing for downstream logic.
		if _, has := m["sourceAttribute"]; !has {
			if v, ok := m["source_attribute"]; ok {
				m["sourceAttribute"] = v
			}
		}
		if _, has := m["targetAttribute"]; !has {
			if v, ok := m["target_attribute"]; ok {
				m["targetAttribute"] = v
			}
		}

		mappings = append(mappings, m)
	}

	return mappings, nil
}

func mappingKey(source string, target string, expression string) string {
	source = strings.TrimSpace(source)
	target = strings.TrimSpace(target)
	expression = strings.TrimSpace(expression)

	if target == "" {
		return ""
	}

	switch {
	case expression != "" && source == "":
		return "expr:" + target + "->" + expression
	case source != "" && expression == "":
		return "src:" + source + "->" + target
	case source != "" && expression != "":
		return "src_expr:" + source + "->" + target + "->" + expression
	default:
		return ""
	}
}

func mappingKeyFromAPI(m map[string]interface{}) string {
	source, _ := utils.NestedString(m, "sourceAttribute")
	target, _ := utils.NestedString(m, "targetAttribute")
	expression, _ := utils.NestedString(m, "expression")
	return mappingKey(source, target, expression)
}
