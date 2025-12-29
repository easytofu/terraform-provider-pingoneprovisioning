package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/githubapi"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/providerdata"
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
)

var (
	_ resource.Resource                = &githubEnterpriseTeamResource{}
	_ resource.ResourceWithConfigure   = &githubEnterpriseTeamResource{}
	_ resource.ResourceWithImportState = &githubEnterpriseTeamResource{}
)

type githubEnterpriseTeamResource struct {
	client *githubapi.Client
}

func NewGithubEnterpriseTeamResource() resource.Resource {
	return &githubEnterpriseTeamResource{}
}

func (r *githubEnterpriseTeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enterprise_team"
}

func (r *githubEnterpriseTeamResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a GitHub enterprise team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique ID of the enterprise team.",
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
			"name": schema.StringAttribute{
				Description: "The name of the team.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The team description.",
				Optional:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The team slug.",
				Computed:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "Optional IdP group ID used to manage membership for the enterprise team.",
				Optional:    true,
			},
			"organization_selection_type": schema.StringAttribute{
				Description: "Specifies which organizations in the enterprise should have access to this team. Allowed values: `disabled`, `selected`, `all`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "selected", "all"),
				},
			},
			"organization_slugs": schema.SetAttribute{
				Description: "Optional organization slugs to assign when `organization_selection_type` is `selected`.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"url": schema.StringAttribute{
				Description: "API URL for the enterprise team.",
				Computed:    true,
			},
			"html_url": schema.StringAttribute{
				Description: "Web URL for the enterprise team.",
				Computed:    true,
			},
			"members_url": schema.StringAttribute{
				Description: "API URL for team members.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Team creation timestamp.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Team update timestamp.",
				Computed:    true,
			},
		},
	}
}

func (r *githubEnterpriseTeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *githubEnterpriseTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customtypes.GithubEnterpriseTeamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(plan.Enterprise.ValueString())
	payload, selection, validationDiags := githubEnterpriseTeamPayload(ctx, &plan)
	resp.Diagnostics.Append(validationDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Do(ctx, http.MethodPost, enterpriseTeamsPath(enterprise), nil, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Enterprise Team",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Creating Enterprise Team",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}

	team, err := githubDecodeMap(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Decoding Enterprise Team",
			fmt.Sprintf("Could not parse response: %s", err),
		)
		return
	}

	state := plan
	applyGithubEnterpriseTeamToState(team, &state)
	if selection != "" {
		state.OrganizationSelectionType = types.StringValue(selection)
	}

	if !state.Slug.IsNull() && !state.Slug.IsUnknown() {
		state.Id = teamIDFromState(enterprise, &state)
	}

	manageOrgs := !plan.OrganizationSlugs.IsNull() && !plan.OrganizationSlugs.IsUnknown()
	if manageOrgs {
		desiredOrgs, listDiags := setToStrings(ctx, plan.OrganizationSlugs)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := syncEnterpriseTeamOrganizations(ctx, r.client, enterprise, state.Slug.ValueString(), desiredOrgs); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Enterprise Team Organizations",
				fmt.Sprintf("Could not reconcile organization assignments: %s", err),
			)
			return
		}

		actualOrgs, err := listEnterpriseTeamOrganizations(ctx, r.client, enterprise, state.Slug.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Enterprise Team Organizations",
				fmt.Sprintf("Could not read organization assignments: %s", err),
			)
			return
		}

		orgSet, orgDiags := stringsToSet(ctx, actualOrgs)
		resp.Diagnostics.Append(orgDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.OrganizationSlugs = orgSet
	} else {
		state.OrganizationSlugs = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *githubEnterpriseTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customtypes.GithubEnterpriseTeamModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	slug := strings.TrimSpace(state.Slug.ValueString())
	if slug == "" {
		resp.Diagnostics.AddError(
			"Missing Team Slug",
			"State is missing the team slug required to read the enterprise team.",
		)
		return
	}

	httpResp, err := r.client.Do(ctx, http.MethodGet, enterpriseTeamPath(enterprise, slug), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Enterprise Team",
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
			"Error Reading Enterprise Team",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}

	team, err := githubDecodeMap(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Decoding Enterprise Team",
			fmt.Sprintf("Could not parse response: %s", err),
		)
		return
	}

	applyGithubEnterpriseTeamToState(team, &state)

	if !state.OrganizationSlugs.IsNull() && !state.OrganizationSlugs.IsUnknown() {
		orgs, err := listEnterpriseTeamOrganizations(ctx, r.client, enterprise, state.Slug.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Enterprise Team Organizations",
				fmt.Sprintf("Could not read organization assignments: %s", err),
			)
			return
		}

		orgSet, orgDiags := stringsToSet(ctx, orgs)
		resp.Diagnostics.Append(orgDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.OrganizationSlugs = orgSet
	}

	state.Id = teamIDFromState(enterprise, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *githubEnterpriseTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customtypes.GithubEnterpriseTeamModel
	var state customtypes.GithubEnterpriseTeamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	slug := strings.TrimSpace(state.Slug.ValueString())
	if slug == "" {
		resp.Diagnostics.AddError(
			"Missing Team Slug",
			"State is missing the team slug required to update the enterprise team.",
		)
		return
	}

	payload, selection, validationDiags := githubEnterpriseTeamPayload(ctx, &plan)
	resp.Diagnostics.Append(validationDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Do(ctx, http.MethodPatch, enterpriseTeamPath(enterprise, slug), nil, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Enterprise Team",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Updating Enterprise Team",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}

	team, err := githubDecodeMap(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Decoding Enterprise Team",
			fmt.Sprintf("Could not parse response: %s", err),
		)
		return
	}

	newState := plan
	if plan.OrganizationSelectionType.IsNull() && !state.OrganizationSelectionType.IsNull() && !state.OrganizationSelectionType.IsUnknown() {
		newState.OrganizationSelectionType = state.OrganizationSelectionType
	}
	applyGithubEnterpriseTeamToState(team, &newState)
	if selection != "" {
		newState.OrganizationSelectionType = types.StringValue(selection)
	}

	manageOrgs := !plan.OrganizationSlugs.IsNull() && !plan.OrganizationSlugs.IsUnknown()
	if manageOrgs {
		desiredOrgs, listDiags := setToStrings(ctx, plan.OrganizationSlugs)
		resp.Diagnostics.Append(listDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := syncEnterpriseTeamOrganizations(ctx, r.client, enterprise, newState.Slug.ValueString(), desiredOrgs); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Enterprise Team Organizations",
				fmt.Sprintf("Could not reconcile organization assignments: %s", err),
			)
			return
		}

		actualOrgs, err := listEnterpriseTeamOrganizations(ctx, r.client, enterprise, newState.Slug.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Enterprise Team Organizations",
				fmt.Sprintf("Could not read organization assignments: %s", err),
			)
			return
		}

		orgSet, orgDiags := stringsToSet(ctx, actualOrgs)
		resp.Diagnostics.Append(orgDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		newState.OrganizationSlugs = orgSet
	} else {
		newState.OrganizationSlugs = types.SetNull(types.StringType)
	}

	newState.Id = teamIDFromState(enterprise, &newState)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *githubEnterpriseTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customtypes.GithubEnterpriseTeamModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, r.client) {
		return
	}

	enterprise := strings.TrimSpace(state.Enterprise.ValueString())
	slug := strings.TrimSpace(state.Slug.ValueString())
	if slug == "" {
		resp.Diagnostics.AddError(
			"Missing Team Slug",
			"State is missing the team slug required to delete the enterprise team.",
		)
		return
	}

	httpResp, err := r.client.Do(ctx, http.MethodDelete, enterpriseTeamPath(enterprise, slug), nil, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Enterprise Team",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Deleting Enterprise Team",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, r.client)),
		)
		return
	}
}

func (r *githubEnterpriseTeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Error Importing Enterprise Team",
			fmt.Sprintf("Unexpected import identifier format: %s. Expected '<enterprise>/<team_slug>'.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enterprise"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), idParts[1])...)
}

func githubEnterpriseTeamPayload(ctx context.Context, model *customtypes.GithubEnterpriseTeamModel) (map[string]interface{}, string, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := map[string]interface{}{
		"name": strings.TrimSpace(model.Name.ValueString()),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		payload["description"] = model.Description.ValueString()
	}

	if !model.GroupId.IsNull() && !model.GroupId.IsUnknown() {
		groupID := strings.TrimSpace(model.GroupId.ValueString())
		if groupID == "" {
			payload["group_id"] = nil
		} else {
			payload["group_id"] = groupID
		}
	}

	selection := ""
	if !model.OrganizationSelectionType.IsNull() && !model.OrganizationSelectionType.IsUnknown() {
		selection = strings.TrimSpace(model.OrganizationSelectionType.ValueString())
	}

	manageOrgs := !model.OrganizationSlugs.IsNull() && !model.OrganizationSlugs.IsUnknown()
	if manageOrgs && selection == "" {
		selection = "selected"
	}
	if selection != "" {
		payload["organization_selection_type"] = selection
	}

	if manageOrgs && selection != "" && selection != "selected" {
		diags.AddAttributeError(
			path.Root("organization_selection_type"),
			"Invalid Organization Selection",
			"`organization_selection_type` must be `selected` when `organization_slugs` is set.",
		)
	}

	return payload, selection, diags
}

func applyGithubEnterpriseTeamToState(team map[string]interface{}, state *customtypes.GithubEnterpriseTeamModel) {
	if team == nil || state == nil {
		return
	}

	if raw, ok := team["id"]; ok && raw != nil {
		state.Id = types.StringValue(mapString(team, "id"))
	}
	if raw, ok := team["name"]; ok && raw != nil {
		state.Name = types.StringValue(mapString(team, "name"))
	}
	if raw, ok := team["description"]; ok {
		if raw == nil {
			state.Description = types.StringNull()
		} else {
			state.Description = types.StringValue(mapString(team, "description"))
		}
	}
	if raw, ok := team["slug"]; ok && raw != nil {
		state.Slug = types.StringValue(mapString(team, "slug"))
	}
	if raw, ok := team["group_id"]; ok {
		if raw != nil {
			state.GroupId = types.StringValue(mapString(team, "group_id"))
		}
	}
	if raw, ok := team["url"]; ok && raw != nil {
		state.URL = types.StringValue(mapString(team, "url"))
	}
	if raw, ok := team["html_url"]; ok && raw != nil {
		state.HTMLURL = types.StringValue(mapString(team, "html_url"))
	}
	if raw, ok := team["members_url"]; ok && raw != nil {
		state.MembersURL = types.StringValue(mapString(team, "members_url"))
	}
	if raw, ok := team["created_at"]; ok && raw != nil {
		state.CreatedAt = types.StringValue(mapString(team, "created_at"))
	}
	if raw, ok := team["updated_at"]; ok && raw != nil {
		state.UpdatedAt = types.StringValue(mapString(team, "updated_at"))
	}
	if raw, ok := team["organization_selection_type"]; ok && raw != nil {
		state.OrganizationSelectionType = types.StringValue(mapString(team, "organization_selection_type"))
	}
}

func teamIDFromState(enterprise string, state *customtypes.GithubEnterpriseTeamModel) types.String {
	if state == nil {
		return types.StringNull()
	}
	if !state.Id.IsNull() && !state.Id.IsUnknown() && strings.TrimSpace(state.Id.ValueString()) != "" {
		return state.Id
	}
	if !state.Slug.IsNull() && !state.Slug.IsUnknown() {
		compound := fmt.Sprintf("%s/%s", strings.TrimSpace(enterprise), strings.TrimSpace(state.Slug.ValueString()))
		if strings.TrimSpace(compound) != "/" {
			return types.StringValue(compound)
		}
	}
	return types.StringNull()
}

func syncEnterpriseTeamOrganizations(ctx context.Context, client *githubapi.Client, enterprise string, teamSlug string, desired []string) error {
	existing, err := listEnterpriseTeamOrganizations(ctx, client, enterprise, teamSlug)
	if err != nil {
		return err
	}

	desired = normalizeStrings(desired)
	toAdd, toRemove := diffCaseInsensitive(desired, existing)

	if len(toAdd) > 0 {
		payload := map[string]interface{}{
			"organization_slugs": toAdd,
		}
		resp, err := client.Do(ctx, http.MethodPost, enterpriseTeamOrganizationsBulkAddPath(enterprise, teamSlug), nil, payload)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			return githubResponseErrorWithHint(resp, client)
		}
	}

	if len(toRemove) > 0 {
		payload := map[string]interface{}{
			"organization_slugs": toRemove,
		}
		resp, err := client.Do(ctx, http.MethodPost, enterpriseTeamOrganizationsBulkRemovePath(enterprise, teamSlug), nil, payload)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			return githubResponseErrorWithHint(resp, client)
		}
	}

	return nil
}
