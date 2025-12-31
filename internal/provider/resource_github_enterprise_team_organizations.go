package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/githubapi"
	customtypes "github.com/easytofu/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &githubEnterpriseTeamOrganizationsResource{}
	_ resource.ResourceWithConfigure   = &githubEnterpriseTeamOrganizationsResource{}
	_ resource.ResourceWithImportState = &githubEnterpriseTeamOrganizationsResource{}
)

type githubEnterpriseTeamOrganizationsResource struct {
	client *githubapi.Client
}

func NewGithubEnterpriseTeamOrganizationsResource() resource.Resource {
	return &githubEnterpriseTeamOrganizationsResource{}
}

func (r *githubEnterpriseTeamOrganizationsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enterprise_team_organizations"
}

func (r *githubEnterpriseTeamOrganizationsResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages organization assignments for a GitHub enterprise team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID of the organization assignments.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enterprise": schema.StringAttribute{
				Description: "The enterprise slug.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_slug": schema.StringAttribute{
				Description: "The enterprise team slug or ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_slugs": schema.SetAttribute{
				Description: "Organization slugs to assign to the enterprise team.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *githubEnterpriseTeamOrganizationsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clientData, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = clientData.GitHub
}

func (r *githubEnterpriseTeamOrganizationsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.GithubEnterpriseTeamOrganizationsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	team := strings.TrimSpace(plan.TeamSlug.ValueString())

	desired, listDiags := setToStrings(ctx, plan.OrganizationSlugs)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := syncEnterpriseTeamOrganizations(ctx, r.client, enterprise, team, desired); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Enterprise Team Organizations",
			fmt.Sprintf("Could not reconcile organization assignments: %s", err),
		)
		return
	}

	actual, err := listEnterpriseTeamOrganizations(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Organizations",
			fmt.Sprintf("Could not read organization assignments: %s", err),
		)
		return
	}

	orgSet, orgDiags := stringsToSet(ctx, actual)
	resp.Diagnostics.Append(orgDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.OrganizationSlugs = orgSet
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", enterprise, team))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *githubEnterpriseTeamOrganizationsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.GithubEnterpriseTeamOrganizationsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	team := strings.TrimSpace(state.TeamSlug.ValueString())

	actual, err := listEnterpriseTeamOrganizations(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Organizations",
			fmt.Sprintf("Could not read organization assignments: %s", err),
		)
		return
	}

	orgSet, orgDiags := stringsToSet(ctx, actual)
	resp.Diagnostics.Append(orgDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.OrganizationSlugs = orgSet
	state.Id = types.StringValue(fmt.Sprintf("%s/%s", enterprise, team))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *githubEnterpriseTeamOrganizationsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.GithubEnterpriseTeamOrganizationsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	team := strings.TrimSpace(plan.TeamSlug.ValueString())

	desired, listDiags := setToStrings(ctx, plan.OrganizationSlugs)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := syncEnterpriseTeamOrganizations(ctx, r.client, enterprise, team, desired); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Enterprise Team Organizations",
			fmt.Sprintf("Could not reconcile organization assignments: %s", err),
		)
		return
	}

	actual, err := listEnterpriseTeamOrganizations(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Organizations",
			fmt.Sprintf("Could not read organization assignments: %s", err),
		)
		return
	}

	orgSet, orgDiags := stringsToSet(ctx, actual)
	resp.Diagnostics.Append(orgDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.OrganizationSlugs = orgSet
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", enterprise, team))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *githubEnterpriseTeamOrganizationsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customtypes.GithubEnterpriseTeamOrganizationsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	team := strings.TrimSpace(state.TeamSlug.ValueString())

	existing, err := listEnterpriseTeamOrganizations(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Organizations",
			fmt.Sprintf("Could not read organization assignments: %s", err),
		)
		return
	}

	for _, org := range existing {
		httpResp, err := r.client.Do(ctx, http.MethodDelete, enterpriseTeamOrganizationPath(enterprise, team, org), nil, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Removing Enterprise Team Organization",
				fmt.Sprintf("Request failed: %s", err),
			)
			return
		}
		if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode >= 300 {
			resp.Diagnostics.AddError(
				"Error Removing Enterprise Team Organization",
				fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
			)
			return
		}
	}
}

func (r *githubEnterpriseTeamOrganizationsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Error Importing Enterprise Team Organizations",
			fmt.Sprintf("Unexpected import identifier format: %s. Expected '<enterprise>/<team_slug>'.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enterprise"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_slug"), idParts[1])...)
}
