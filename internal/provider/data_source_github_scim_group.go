package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &githubScimGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &githubScimGroupDataSource{}
)

type githubScimGroupDataSource struct {
	client *client.GitHubClient
}

type githubScimGroupDataSourceModel struct {
	Enterprise  types.String                 `tfsdk:"enterprise"`
	DisplayName types.String                 `tfsdk:"display_name"`
	Id          types.String                 `tfsdk:"id"`
	ExternalId  types.String                 `tfsdk:"external_id"`
	Members     []githubScimGroupMemberModel `tfsdk:"members"`
}

type githubScimGroupMemberModel struct {
	Value   types.String `tfsdk:"value"`
	Display types.String `tfsdk:"display"`
	Type    types.String `tfsdk:"type"`
	Ref     types.String `tfsdk:"ref"`
}

type githubScimGroupListResponse struct {
	Resources []githubScimGroupResponse `json:"Resources"`
}

type githubScimGroupResponse struct {
	Id          string                          `json:"id"`
	DisplayName string                          `json:"displayName"`
	ExternalId  string                          `json:"externalId"`
	Members     []githubScimGroupMemberResponse `json:"members"`
}

type githubScimGroupMemberResponse struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Type    string `json:"type"`
	Ref     string `json:"$ref"`
}

func NewGithubScimGroupDataSource() datasource.DataSource {
	return &githubScimGroupDataSource{}
}

func (d *githubScimGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_github_scim_group"
}

func (d *githubScimGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a GitHub SCIM group by display name.",
		Attributes: map[string]schema.Attribute{
			"enterprise": schema.StringAttribute{
				Description: "The enterprise slug.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The SCIM group display name to look up.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The SCIM group ID.",
				Computed:    true,
			},
			"external_id": schema.StringAttribute{
				Description: "The external ID for the SCIM group.",
				Computed:    true,
			},
			"members": schema.ListNestedAttribute{
				Description: "Members returned by the SCIM group query.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Description: "Member ID.",
							Computed:    true,
						},
						"display": schema.StringAttribute{
							Description: "Member display name.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Member type.",
							Computed:    true,
						},
						"ref": schema.StringAttribute{
							Description: "Member reference URL.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *githubScimGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clientData, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = clientData.GitHub
}

func (d *githubScimGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config githubScimGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, d.client) {
		return
	}

	enterprise := strings.TrimSpace(config.Enterprise.ValueString())
	displayName := strings.TrimSpace(config.DisplayName.ValueString())
	if enterprise == "" {
		resp.Diagnostics.AddError(
			"Missing Enterprise",
			"enterprise must be provided to look up a GitHub SCIM group.",
		)
		return
	}
	if displayName == "" {
		resp.Diagnostics.AddError(
			"Missing SCIM Group Name",
			"display_name must be provided to look up a GitHub SCIM group.",
		)
		return
	}

	query := url.Values{}
	query.Set("filter", fmt.Sprintf("displayName eq \"%s\"", escapeScimFilterValue(displayName)))

	httpResp, err := d.client.Do(ctx, http.MethodGet, enterpriseScimGroupsPath(enterprise), query, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SCIM Group",
			fmt.Sprintf("Request failed: %s", err),
		)
		return
	}
	if httpResp.StatusCode >= 300 {
		resp.Diagnostics.AddError(
			"Error Reading SCIM Group",
			fmt.Sprintf("API error: %s", githubResponseErrorWithHint(httpResp, d.client)),
		)
		return
	}

	bodyBytes, err := utils.ReadAndRestoreResponseBody(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SCIM Group",
			fmt.Sprintf("Could not read response: %s", err),
		)
		return
	}

	var payload githubScimGroupListResponse
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing SCIM Group",
			fmt.Sprintf("Could not parse response: %s", err),
		)
		return
	}

	if len(payload.Resources) == 0 {
		resp.Diagnostics.AddError(
			"SCIM Group Not Found",
			fmt.Sprintf("No SCIM group found with display_name %q.", displayName),
		)
		return
	}
	if len(payload.Resources) > 1 {
		resp.Diagnostics.AddError(
			"Multiple SCIM Groups Found",
			fmt.Sprintf("More than one SCIM group matched display_name %q.", displayName),
		)
		return
	}

	group := payload.Resources[0]
	state := config
	state.Id = types.StringValue(group.Id)
	state.DisplayName = stringValueOrNull(group.DisplayName, displayName)
	state.ExternalId = stringValueOrNull(group.ExternalId, "")

	members := make([]githubScimGroupMemberModel, 0, len(group.Members))
	for _, member := range group.Members {
		members = append(members, githubScimGroupMemberModel{
			Value:   stringValueOrNull(member.Value, ""),
			Display: stringValueOrNull(member.Display, ""),
			Type:    stringValueOrNull(member.Type, ""),
			Ref:     stringValueOrNull(member.Ref, ""),
		})
	}
	state.Members = members

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func enterpriseScimGroupsPath(enterprise string) string {
	return fmt.Sprintf("/scim/v2/enterprises/%s/Groups", url.PathEscape(strings.TrimSpace(enterprise)))
}

func escapeScimFilterValue(value string) string {
	return strings.ReplaceAll(value, "\"", "\\\"")
}

func stringValueOrNull(value string, fallback string) types.String {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = strings.TrimSpace(fallback)
	}
	if trimmed == "" {
		return types.StringNull()
	}
	return types.StringValue(trimmed)
}
