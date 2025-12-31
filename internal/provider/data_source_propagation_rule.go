package provider

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	customtypes "github.com/easytofu/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
)

var (
	_ datasource.DataSource              = &propagationRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &propagationRuleDataSource{}
)

type propagationRuleDataSource struct {
	client *client.Client
}

func NewPropagationRuleDataSource() datasource.DataSource {
	return &propagationRuleDataSource{}
}

func (d *propagationRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_propagation_rule"
}

func (d *propagationRuleDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a PingOne provisioning propagation rule and its mappings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the propagation rule.",
				Optional:    true,
				Computed:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment.",
				Required:    true,
			},
			"plan_id": schema.StringAttribute{
				Description: "Optional plan ID to scope name lookups.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Optional name to lookup the propagation rule.",
				Optional:    true,
				Computed:    true,
			},
			"source_store_id": schema.StringAttribute{
				Description: "The source store ID for the propagation rule.",
				Computed:    true,
			},
			"target_store_id": schema.StringAttribute{
				Description: "The target store ID for the propagation rule.",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the propagation rule is active.",
				Computed:    true,
			},
			"filter": schema.StringAttribute{
				Description: "SCIM filter expression for selecting users to synchronize.",
				Computed:    true,
			},
			"deprovision": schema.BoolAttribute{
				Description: "Whether to deprovision users in the target store when they are removed from the source.",
				Computed:    true,
			},
			"population_ids": schema.ListAttribute{
				Description: "List of population IDs in scope for this rule.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"group_ids": schema.ListAttribute{
				Description: "List of group IDs in scope for group provisioning.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"configuration": schema.MapAttribute{
				Description: "Rule configuration map (for example, `MFA_USER_DEVICE_MANAGEMENT`).",
				Computed:    true,
				ElementType: types.StringType,
			},
			"mappings": schema.ListNestedAttribute{
				Description: "List of attribute mappings for this rule.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The mapping ID.",
							Computed:    true,
						},
						"source_attribute": schema.StringAttribute{
							Description: "Source attribute expression.",
							Computed:    true,
						},
						"target_attribute": schema.StringAttribute{
							Description: "Target attribute expression.",
							Computed:    true,
						},
						"expression": schema.StringAttribute{
							Description: "Expression used to compute the target attribute value.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *propagationRuleDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		propagationRuleLookupValidator{},
	}
}

type propagationRuleLookupValidator struct{}

func (v propagationRuleLookupValidator) Description(_ context.Context) string {
	return "Validates lookup arguments for a propagation rule."
}

func (v propagationRuleLookupValidator) MarkdownDescription(_ context.Context) string {
	return "Validates lookup arguments for a propagation rule."
}

func (v propagationRuleLookupValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config customtypes.PropagationRuleModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, validationDiags := propagationRuleLookupModeFromValues(config.Id, config.PlanId, config.Name)
	resp.Diagnostics.Append(validationDiags...)
}

type propagationRuleLookupMode int

const (
	propagationRuleLookupModeInvalid propagationRuleLookupMode = iota
	propagationRuleLookupModeName
	propagationRuleLookupModeId
)

func propagationRuleLookupModeFromValues(id, planID, name types.String) (propagationRuleLookupMode, diag.Diagnostics) {
	var diags diag.Diagnostics

	nameSet := !name.IsNull() && !name.IsUnknown() && name.ValueString() != ""
	planSet := !planID.IsNull() && !planID.IsUnknown() && planID.ValueString() != ""
	idSet := !id.IsNull() && !id.IsUnknown() && id.ValueString() != ""

	if nameSet {
		if planSet {
			return propagationRuleLookupModeName, diags
		}
		return propagationRuleLookupModeName, diags
	}

	if planSet {
		diags.AddAttributeError(
			path.Root("name"),
			"Missing Required Argument",
			"When configuring a lookup by `plan_id`, `name` must be set.",
		)
		return propagationRuleLookupModeInvalid, diags
	}

	if idSet {
		return propagationRuleLookupModeId, diags
	}

	diags.AddError(
		"Missing Required Arguments",
		"Configure either `id` or `name` (optionally with `plan_id`) to lookup a propagation rule.",
	)
	return propagationRuleLookupModeInvalid, diags
}

func (d *propagationRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clientData, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = clientData
}

func (d *propagationRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customtypes.PropagationRuleModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := state.EnvironmentId.ValueString()
	apiClient := d.client.API

	lookupMode, validationDiags := propagationRuleLookupModeFromValues(state.Id, state.PlanId, state.Name)
	resp.Diagnostics.Append(validationDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ruleID string

	switch lookupMode {
	case propagationRuleLookupModeId:
		ruleID = state.Id.ValueString()
	case propagationRuleLookupModeName:
		targetName := state.Name.ValueString()
		targetPlanID := ""
		if !state.PlanId.IsNull() && !state.PlanId.IsUnknown() {
			targetPlanID = state.PlanId.ValueString()
		}

		tflog.Info(ctx, "Looking up propagation rule by name", map[string]interface{}{
			"environment_id": environmentID,
			"plan_id":        targetPlanID,
			"name":           targetName,
		})

		rules, err := listPropagationRules(ctx, apiClient, environmentID, targetPlanID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Propagation Rules",
				fmt.Sprintf("Could not list propagation rules: %s", err),
			)
			return
		}

		var matches []string
		for _, rule := range rules {
			name, _ := utils.NestedString(rule, "name")
			if name != targetName {
				continue
			}
			id, _ := utils.NestedString(rule, "id")
			if id != "" {
				matches = append(matches, id)
			}
		}

		if len(matches) == 0 {
			resp.Diagnostics.AddError(
				"Propagation Rule Not Found",
				fmt.Sprintf("No propagation rule found with name %q in environment %q.", targetName, environmentID),
			)
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Propagation Rules Found",
				fmt.Sprintf("Found %d propagation rules with name %q in environment %q; refine your lookup.", len(matches), targetName, environmentID),
			)
			return
		}
		ruleID = matches[0]
	default:
		resp.Diagnostics.AddError("Invalid Lookup Configuration", "Unable to determine lookup configuration for propagation rule.")
		return
	}

	ruleObj, httpResp, err := readPropagationRuleDataSource(ctx, apiClient, environmentID, ruleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Propagation Rule",
			fmt.Sprintf("Could not read propagation rule: %s", err),
		)
		return
	}
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Propagation Rule Not Found",
			fmt.Sprintf("Propagation rule %q was not found in environment %q.", ruleID, environmentID),
		)
		return
	}

	state.EnvironmentId = types.StringValue(environmentID)
	state.Id = types.StringValue(ruleID)

	applyRuleAPIToStateDataSource(ruleObj, &state)

	mappings, err := readPropagationRuleMappings(ctx, apiClient, environmentID, ruleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Propagation Rule Mappings",
			fmt.Sprintf("Could not read mappings: %s", err),
		)
		return
	}
	state.Mappings = mappings

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func listPropagationRules(ctx context.Context, apiClient *management.APIClient, environmentID string, planID string) ([]map[string]interface{}, error) {
	var httpResp *http.Response
	var err error

	if planID != "" {
		httpResp, err = apiClient.PropagationRulesApi.
			EnvironmentsEnvironmentIDPropagationPlansPlanIDRulesGet(ctx, environmentID, planID).
			Execute()
	} else {
		httpResp, err = apiClient.PropagationRulesApi.
			EnvironmentsEnvironmentIDPropagationRulesGet(ctx, environmentID).
			Execute()
	}

	if err != nil {
		return nil, fmt.Errorf("%s", utils.HandleSDKError(err, httpResp))
	}

	decoded, err := utils.DecodeResponseJSON(httpResp)
	if err != nil {
		return nil, err
	}

	list, err := utils.ExtractEmbeddedArray(decoded, "rules", "items")
	if err != nil {
		if arr, ok := decoded.([]interface{}); ok {
			list = arr
		} else {
			return nil, err
		}
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

func readPropagationRuleDataSource(ctx context.Context, apiClient *management.APIClient, environmentID string, ruleID string) (map[string]interface{}, *http.Response, error) {
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

func applyRuleAPIToStateDataSource(apiObj map[string]interface{}, state *customtypes.PropagationRuleModel) {
	if planID, ok := utils.NestedString(apiObj, "plan", "id"); ok && planID != "" {
		state.PlanId = types.StringValue(planID)
	}
	if name, ok := utils.NestedString(apiObj, "name"); ok && name != "" {
		state.Name = types.StringValue(name)
	}
	if srcID, ok := utils.NestedString(apiObj, "sourceStore", "id"); ok && srcID != "" {
		state.SourceStoreId = types.StringValue(srcID)
	}
	if tgtID, ok := utils.NestedString(apiObj, "targetStore", "id"); ok && tgtID != "" {
		state.TargetStoreId = types.StringValue(tgtID)
	}

	if v, ok := apiObj["populationExpression"]; ok {
		if s, ok := v.(string); ok {
			state.Filter = types.StringValue(s)
		}
	}
	if v, ok := apiObj["active"]; ok {
		if b, ok := v.(bool); ok {
			state.Active = types.BoolValue(b)
		}
	}
	if v, ok := apiObj["deprovision"]; ok {
		if b, ok := v.(bool); ok {
			state.Deprovision = types.BoolValue(b)
		}
	}

	var cfg map[string]string
	if v, ok := apiObj["configuration"]; ok && v != nil {
		if rawMap, ok := v.(map[string]interface{}); ok {
			cfg = make(map[string]string, len(rawMap))
			for k, rawVal := range rawMap {
				if s, ok := rawVal.(string); ok {
					cfg[k] = s
				}
			}
		}
	}
	if cfg != nil {
		mapVal, _ := types.MapValueFrom(context.Background(), types.StringType, cfg)
		state.Configuration = mapVal
	} else {
		state.Configuration = types.MapNull(types.StringType)
	}

	var populationIDs []string
	if v, ok := apiObj["populations"]; ok && v != nil {
		if rawList, ok := v.([]interface{}); ok {
			for _, item := range rawList {
				if m, ok := item.(map[string]interface{}); ok {
					if id, ok := utils.NestedString(m, "id"); ok && id != "" {
						populationIDs = append(populationIDs, id)
					}
				}
			}
		}
	}
	sort.Strings(populationIDs)
	if len(populationIDs) > 0 {
		listVal, _ := types.ListValueFrom(context.Background(), types.StringType, populationIDs)
		state.PopulationIds = listVal
	} else {
		state.PopulationIds = types.ListNull(types.StringType)
	}

	var groupIDs []string
	if v, ok := apiObj["groups"]; ok && v != nil {
		if rawList, ok := v.([]interface{}); ok {
			for _, item := range rawList {
				if m, ok := item.(map[string]interface{}); ok {
					if id, ok := utils.NestedString(m, "id"); ok && id != "" {
						groupIDs = append(groupIDs, id)
					}
				}
			}
		}
	}
	sort.Strings(groupIDs)
	if len(groupIDs) > 0 {
		listVal, _ := types.ListValueFrom(context.Background(), types.StringType, groupIDs)
		state.GroupIds = listVal
	} else {
		state.GroupIds = types.ListNull(types.StringType)
	}
}

func readPropagationRuleMappings(ctx context.Context, apiClient *management.APIClient, environmentID string, ruleID string) ([]customtypes.PropagationRuleMappingModel, error) {
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
		if arr, ok := decoded.([]interface{}); ok {
			list = arr
		} else {
			return nil, err
		}
	}

	type rawMapping struct {
		id         string
		source     string
		target     string
		expression string
		key        string
	}

	var raw []rawMapping
	for _, v := range list {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := utils.NestedString(m, "id")
		source, _ := utils.NestedString(m, "sourceAttribute")
		if source == "" {
			if v, ok := m["source_attribute"].(string); ok {
				source = v
			}
		}
		target, _ := utils.NestedString(m, "targetAttribute")
		if target == "" {
			if v, ok := m["target_attribute"].(string); ok {
				target = v
			}
		}
		expression, _ := utils.NestedString(m, "expression")

		source = strings.TrimSpace(source)
		target = strings.TrimSpace(target)
		expression = strings.TrimSpace(expression)

		if target == "" {
			continue
		}

		key := ""
		switch {
		case expression != "" && source == "":
			key = "expr:" + target + "->" + expression
		case source != "" && expression == "":
			key = "src:" + source + "->" + target
		case source != "" && expression != "":
			key = "src_expr:" + source + "->" + target + "->" + expression
		default:
			continue
		}

		raw = append(raw, rawMapping{
			id:         id,
			source:     source,
			target:     target,
			expression: expression,
			key:        key,
		})
	}

	sort.Slice(raw, func(i, j int) bool {
		return raw[i].key < raw[j].key
	})

	var mappings []customtypes.PropagationRuleMappingModel
	for _, m := range raw {
		model := customtypes.PropagationRuleMappingModel{
			Id:              types.StringValue(m.id),
			TargetAttribute: types.StringValue(m.target),
		}
		if m.source != "" {
			model.SourceAttribute = types.StringValue(m.source)
		} else {
			model.SourceAttribute = types.StringNull()
		}
		if m.expression != "" {
			model.Expression = types.StringValue(m.expression)
		} else {
			model.Expression = types.StringNull()
		}

		mappings = append(mappings, model)
	}

	return mappings, nil
}
