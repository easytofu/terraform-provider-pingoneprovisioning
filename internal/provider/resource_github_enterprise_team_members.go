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
	_ resource.Resource                = &githubEnterpriseTeamMembersResource{}
	_ resource.ResourceWithConfigure   = &githubEnterpriseTeamMembersResource{}
	_ resource.ResourceWithImportState = &githubEnterpriseTeamMembersResource{}
)

type githubEnterpriseTeamMembersResource struct {
	client *githubapi.Client
}

func NewGithubEnterpriseTeamMembersResource() resource.Resource {
	return &githubEnterpriseTeamMembersResource{}
}

func (r *githubEnterpriseTeamMembersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enterprise_team_members"
}

func (r *githubEnterpriseTeamMembersResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages enterprise team membership in bulk.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID of the team membership set.",
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
			"usernames": schema.SetAttribute{
				Description: "GitHub usernames to ensure are members of the enterprise team.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *githubEnterpriseTeamMembersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *githubEnterpriseTeamMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.GithubEnterpriseTeamMembersModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	team := strings.TrimSpace(plan.TeamSlug.ValueString())

	desired, listDiags := setToStrings(ctx, plan.Usernames)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := syncEnterpriseTeamMembers(ctx, r.client, enterprise, team, desired); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Enterprise Team Members",
			fmt.Sprintf("Could not reconcile team members: %s", err),
		)
		return
	}

	actual, err := listEnterpriseTeamMembers(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Members",
			fmt.Sprintf("Could not read team members: %s", err),
		)
		return
	}

	userSet, userDiags := stringsToSet(ctx, actual)
	resp.Diagnostics.Append(userDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Usernames = userSet
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", enterprise, team))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *githubEnterpriseTeamMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.GithubEnterpriseTeamMembersModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	team := strings.TrimSpace(state.TeamSlug.ValueString())

	actual, err := listEnterpriseTeamMembers(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Members",
			fmt.Sprintf("Could not read team members: %s", err),
		)
		return
	}

	userSet, userDiags := stringsToSet(ctx, actual)
	resp.Diagnostics.Append(userDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Usernames = userSet
	state.Id = types.StringValue(fmt.Sprintf("%s/%s", enterprise, team))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *githubEnterpriseTeamMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.GithubEnterpriseTeamMembersModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	team := strings.TrimSpace(plan.TeamSlug.ValueString())

	desired, listDiags := setToStrings(ctx, plan.Usernames)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := syncEnterpriseTeamMembers(ctx, r.client, enterprise, team, desired); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Enterprise Team Members",
			fmt.Sprintf("Could not reconcile team members: %s", err),
		)
		return
	}

	actual, err := listEnterpriseTeamMembers(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Members",
			fmt.Sprintf("Could not read team members: %s", err),
		)
		return
	}

	userSet, userDiags := stringsToSet(ctx, actual)
	resp.Diagnostics.Append(userDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Usernames = userSet
	plan.Id = types.StringValue(fmt.Sprintf("%s/%s", enterprise, team))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *githubEnterpriseTeamMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customtypes.GithubEnterpriseTeamMembersModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	team := strings.TrimSpace(state.TeamSlug.ValueString())

	existing, err := listEnterpriseTeamMembers(ctx, r.client, enterprise, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team Members",
			fmt.Sprintf("Could not read team members: %s", err),
		)
		return
	}

	if len(existing) == 0 {
		return
	}

	payload := map[string]interface{}{
		"usernames": normalizeStrings(existing),
	}
	httpResp, err := r.client.Do(ctx, http.MethodPost, enterpriseTeamMembersBulkRemovePath(enterprise, team), nil, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Enterprise Team Members",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Removing Enterprise Team Members",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}
}

func (r *githubEnterpriseTeamMembersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Error Importing Enterprise Team Members",
			fmt.Sprintf("Unexpected import identifier format: %s. Expected '<enterprise>/<team_slug>'.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enterprise"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_slug"), idParts[1])...)
}

func syncEnterpriseTeamMembers(ctx context.Context, client *githubapi.Client, enterprise string, teamSlug string, desired []string) error {
	existing, err := listEnterpriseTeamMembers(ctx, client, enterprise, teamSlug)
	if err != nil {
		return err
	}

	desired = normalizeStrings(desired)
	toAdd, toRemove := diffCaseInsensitive(desired, existing)

	if len(toAdd) > 0 {
		payload := map[string]interface{}{
			"usernames": toAdd,
		}
		resp, err := client.Do(ctx, http.MethodPost, enterpriseTeamMembersBulkAddPath(enterprise, teamSlug), nil, payload)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			return githubResponseErrorWithHint(resp, client)
		}
	}

	if len(toRemove) > 0 {
		payload := map[string]interface{}{
			"usernames": toRemove,
		}
		resp, err := client.Do(ctx, http.MethodPost, enterpriseTeamMembersBulkRemovePath(enterprise, teamSlug), nil, payload)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			return githubResponseErrorWithHint(resp, client)
		}
	}

	return nil
}
