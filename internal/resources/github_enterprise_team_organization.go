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
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/githubapi"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/providerdata"
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
)

var (
	_ resource.Resource                = &githubEnterpriseTeamOrganizationResource{}
	_ resource.ResourceWithConfigure   = &githubEnterpriseTeamOrganizationResource{}
	_ resource.ResourceWithImportState = &githubEnterpriseTeamOrganizationResource{}
)

type githubEnterpriseTeamOrganizationResource struct {
	client *githubapi.Client
}

func NewGithubEnterpriseTeamOrganizationResource() resource.Resource {
	return &githubEnterpriseTeamOrganizationResource{}
}

func (r *githubEnterpriseTeamOrganizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enterprise_team_organization"
}

func (r *githubEnterpriseTeamOrganizationResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an organization assignment for a GitHub enterprise team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID of the assignment.",
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
			"organization": schema.StringAttribute{
				Description: "The organization slug.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *githubEnterpriseTeamOrganizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clientData, ok := req.ProviderData.(*providerdata.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerdata.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = clientData.GitHub
}

func (r *githubEnterpriseTeamOrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.GithubEnterpriseTeamOrganizationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	team := strings.TrimSpace(plan.TeamSlug.ValueString())
	org := strings.TrimSpace(plan.Organization.ValueString())

	httpResp, err := r.client.Do(ctx, http.MethodPut, enterpriseTeamOrganizationPath(enterprise, team, org), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Assigning Enterprise Team Organization",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Assigning Enterprise Team Organization",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}

	plan.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", enterprise, team, org))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *githubEnterpriseTeamOrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.GithubEnterpriseTeamOrganizationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	team := strings.TrimSpace(state.TeamSlug.ValueString())
	org := strings.TrimSpace(state.Organization.ValueString())

	orgs, err := listEnterpriseTeamOrganizations(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Organizations",
			fmt.Sprintf("Could not read organization assignments: %s", err),
		)
		return
	}

	if !containsCaseInsensitive(orgs, org) {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", enterprise, team, org))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *githubEnterpriseTeamOrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.GithubEnterpriseTeamOrganizationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	team := strings.TrimSpace(plan.TeamSlug.ValueString())
	org := strings.TrimSpace(plan.Organization.ValueString())

	httpResp, err := r.client.Do(ctx, http.MethodPut, enterpriseTeamOrganizationPath(enterprise, team, org), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Assigning Enterprise Team Organization",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Assigning Enterprise Team Organization",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}

	plan.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", enterprise, team, org))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *githubEnterpriseTeamOrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customtypes.GithubEnterpriseTeamOrganizationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	team := strings.TrimSpace(state.TeamSlug.ValueString())
	org := strings.TrimSpace(state.Organization.ValueString())

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

func (r *githubEnterpriseTeamOrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 3 {
		resp.Diagnostics.AddError(
			"Error Importing Enterprise Team Organization",
			fmt.Sprintf("Unexpected import identifier format: %s. Expected '<enterprise>/<team_slug>/<org>'.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enterprise"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_slug"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), idParts[2])...)
}

func containsCaseInsensitive(values []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return false
	}
	for _, v := range values {
		if strings.ToLower(strings.TrimSpace(v)) == target {
			return true
		}
	}
	return false
}
