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
	_ resource.Resource                = &githubEnterpriseTeamMemberResource{}
	_ resource.ResourceWithConfigure   = &githubEnterpriseTeamMemberResource{}
	_ resource.ResourceWithImportState = &githubEnterpriseTeamMemberResource{}
)

type githubEnterpriseTeamMemberResource struct {
	client *githubapi.Client
}

func NewGithubEnterpriseTeamMemberResource() resource.Resource {
	return &githubEnterpriseTeamMemberResource{}
}

func (r *githubEnterpriseTeamMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enterprise_team_member"
}

func (r *githubEnterpriseTeamMemberResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single GitHub enterprise team membership.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID of the team membership.",
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
			"username": schema.StringAttribute{
				Description: "The GitHub username.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *githubEnterpriseTeamMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *githubEnterpriseTeamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.GithubEnterpriseTeamMemberModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	team := strings.TrimSpace(plan.TeamSlug.ValueString())
	username := strings.TrimSpace(plan.Username.ValueString())

	httpResp, err := r.client.Do(ctx, http.MethodPut, enterpriseTeamMemberPath(enterprise, team, username), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Enterprise Team Member",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Adding Enterprise Team Member",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}

	member, err := githubDecodeMap(httpResp)
	if err == nil && member != nil {
		if login := mapString(member, "login"); login != "" {
			plan.Username = types.StringValue(login)
		}
	}

	plan.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", enterprise, team, plan.Username.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *githubEnterpriseTeamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.GithubEnterpriseTeamMemberModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	team := strings.TrimSpace(state.TeamSlug.ValueString())
	username := strings.TrimSpace(state.Username.ValueString())

	httpResp, err := r.client.Do(ctx, http.MethodGet, enterpriseTeamMemberPath(enterprise, team, username), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Member",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Member",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}

	member, err := githubDecodeMap(httpResp)
	if err == nil && member != nil {
		if login := mapString(member, "login"); login != "" {
			state.Username = types.StringValue(login)
		}
	}

	state.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", enterprise, team, state.Username.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *githubEnterpriseTeamMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.GithubEnterpriseTeamMemberModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	team := strings.TrimSpace(plan.TeamSlug.ValueString())
	username := strings.TrimSpace(plan.Username.ValueString())

	httpResp, err := r.client.Do(ctx, http.MethodPut, enterpriseTeamMemberPath(enterprise, team, username), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Enterprise Team Member",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Adding Enterprise Team Member",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}

	member, err := githubDecodeMap(httpResp)
	if err == nil && member != nil {
		if login := mapString(member, "login"); login != "" {
			plan.Username = types.StringValue(login)
		}
	}

	plan.Id = types.StringValue(fmt.Sprintf("%s/%s/%s", enterprise, team, plan.Username.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *githubEnterpriseTeamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customtypes.GithubEnterpriseTeamMemberModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	team := strings.TrimSpace(state.TeamSlug.ValueString())
	username := strings.TrimSpace(state.Username.ValueString())

	httpResp, err := r.client.Do(ctx, http.MethodDelete, enterpriseTeamMemberPath(enterprise, team, username), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Enterprise Team Member",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Removing Enterprise Team Member",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}
}

func (r *githubEnterpriseTeamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 3 {
		resp.Diagnostics.AddError(
			"Error Importing Enterprise Team Member",
			fmt.Sprintf("Unexpected import identifier format: %s. Expected '<enterprise>/<team_slug>/<username>'.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enterprise"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_slug"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), idParts[2])...)
}
