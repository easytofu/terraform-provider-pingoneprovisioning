package provider

import (
	"context"
	"fmt"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	customtypes "github.com/easytofu/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &propagationPlanDataSource{}
	_ datasource.DataSourceWithConfigure = &propagationPlanDataSource{}
)

type propagationPlanDataSource struct {
	client *client.Client
}

func NewPropagationPlanDataSource() datasource.DataSource {
	return &propagationPlanDataSource{}
}

func (d *propagationPlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_propagation_plan"
}

func (d *propagationPlanDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a PingOne provisioning propagation plan.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the propagation plan.",
				Optional:    true,
				Computed:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Unique name of the propagation plan.",
				Optional:    true,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the propagation plan.",
				Computed:    true,
			},
		},
	}
}

func (d *propagationPlanDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		propagationPlanLookupValidator{},
	}
}

type propagationPlanLookupValidator struct{}

func (v propagationPlanLookupValidator) Description(_ context.Context) string {
	return "Validates lookup arguments for a propagation plan."
}

func (v propagationPlanLookupValidator) MarkdownDescription(_ context.Context) string {
	return "Validates lookup arguments for a propagation plan."
}

func (v propagationPlanLookupValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var config customtypes.PropagationPlanModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, validationDiags := propagationPlanLookupModeFromValues(config.Id, config.Name)
	resp.Diagnostics.Append(validationDiags...)
}

type propagationPlanLookupMode int

const (
	propagationPlanLookupModeInvalid propagationPlanLookupMode = iota
	propagationPlanLookupModeName
	propagationPlanLookupModeId
)

func propagationPlanLookupModeFromValues(id, name types.String) (propagationPlanLookupMode, diag.Diagnostics) {
	var diags diag.Diagnostics

	nameSet := !name.IsNull() && !name.IsUnknown() && name.ValueString() != ""
	idSet := !id.IsNull() && !id.IsUnknown() && id.ValueString() != ""

	// Prefer name lookups if configured to avoid Optional+Computed id values being treated as conflicting.
	if nameSet {
		return propagationPlanLookupModeName, diags
	}

	if idSet {
		return propagationPlanLookupModeId, diags
	}

	diags.AddError(
		"Missing Required Arguments",
		"Configure either `id` or `name` to lookup a propagation plan.",
	)
	return propagationPlanLookupModeInvalid, diags
}

func (d *propagationPlanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *propagationPlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customtypes.PropagationPlanModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := state.EnvironmentId.ValueString()
	apiClient := d.client.API

	lookupMode, validationDiags := propagationPlanLookupModeFromValues(state.Id, state.Name)
	resp.Diagnostics.Append(validationDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch lookupMode {
	case propagationPlanLookupModeId:
		tflog.Info(ctx, "Reading propagation plan by ID", map[string]interface{}{
			"environment_id": environmentID,
			"id":             state.Id.ValueString(),
		})

		result, httpResp, err := apiClient.IdentityPropagationPlansApi.
			ReadOnePlan(ctx, environmentID, state.Id.ValueString()).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Propagation Plan",
				fmt.Sprintf("Could not read propagation plan: %s", utils.HandleSDKError(err, httpResp)),
			)
			return
		}

		state.Id = types.StringValue(result.GetId())
		state.Name = types.StringValue(result.GetName())
		state.EnvironmentId = types.StringValue(environmentID)
		state.Status = types.StringNull()
		if v, ok := result.GetStatusOk(); ok && v != nil {
			state.Status = types.StringValue(string(*v))
		}
	case propagationPlanLookupModeName:
		targetName := state.Name.ValueString()
		tflog.Info(ctx, "Reading propagation plan by Name", map[string]interface{}{
			"environment_id": environmentID,
			"name":           targetName,
		})

		iterator := apiClient.IdentityPropagationPlansApi.ReadAllPlans(ctx, environmentID).Execute()

		var matches []customtypes.PropagationPlanModel

		for cursor, iterErr := range iterator {
			if iterErr != nil {
				resp.Diagnostics.AddError(
					"Error Reading Propagation Plans",
					fmt.Sprintf("Could not iterate propagation plans: %s", iterErr),
				)
				return
			}
			if cursor.EntityArray == nil {
				continue
			}

			embedded, ok := cursor.EntityArray.GetEmbeddedOk()
			if !ok || embedded == nil {
				continue
			}

			plans, ok := embedded.GetPlansOk()
			if !ok {
				continue
			}

			for _, p := range plans {
				if p.GetName() != targetName {
					continue
				}

				model := customtypes.PropagationPlanModel{
					Id:            types.StringValue(p.GetId()),
					EnvironmentId: types.StringValue(environmentID),
					Name:          types.StringValue(p.GetName()),
					Status:        types.StringNull(),
				}
				if v, ok := p.GetStatusOk(); ok && v != nil {
					model.Status = types.StringValue(string(*v))
				}

				matches = append(matches, model)
			}
		}

		if len(matches) == 0 {
			resp.Diagnostics.AddError(
				"Propagation Plan Not Found",
				fmt.Sprintf("No propagation plan found with name %q in environment %q.", targetName, environmentID),
			)
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Propagation Plans Found",
				fmt.Sprintf("Found %d propagation plans with name %q in environment %q; refine your lookup.", len(matches), targetName, environmentID),
			)
			return
		}

		state = matches[0]
	default:
		resp.Diagnostics.AddError("Invalid Lookup Configuration", "Unable to determine lookup configuration for propagation plan.")
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
