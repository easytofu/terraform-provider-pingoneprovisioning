package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/githubapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &githubEnterpriseTeamsDataSource{}
	_ datasource.DataSourceWithConfigure = &githubEnterpriseTeamsDataSource{}
)

type githubEnterpriseTeamsDataSource struct {
	client *githubapi.Client
}

type githubEnterpriseTeamsDataSourceModel struct {
	Enterprise types.String                    `tfsdk:"enterprise"`
	Teams      []githubEnterpriseTeamDataModel `tfsdk:"teams"`
}

type githubEnterpriseTeamDataModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Slug        types.String `tfsdk:"slug"`
	GroupId     types.String `tfsdk:"group_id"`
	URL         types.String `tfsdk:"url"`
	HTMLURL     types.String `tfsdk:"html_url"`
	MembersURL  types.String `tfsdk:"members_url"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewGithubEnterpriseTeamsDataSource() datasource.DataSource {
	return &githubEnterpriseTeamsDataSource{}
}

func (d *githubEnterpriseTeamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enterprise_teams"
}

func (d *githubEnterpriseTeamsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists GitHub enterprise teams.",
		Attributes: map[string]schema.Attribute{
			"enterprise": schema.StringAttribute{
				Description: "The enterprise slug.",
				Required:    true,
			},
			"teams": schema.ListNestedAttribute{
				Description: "Enterprise teams.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Team ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Team name.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Team description.",
							Computed:    true,
						},
						"slug": schema.StringAttribute{
							Description: "Team slug.",
							Computed:    true,
						},
						"group_id": schema.StringAttribute{
							Description: "IdP group ID.",
							Computed:    true,
						},
						"url": schema.StringAttribute{
							Description: "API URL for the team.",
							Computed:    true,
						},
						"html_url": schema.StringAttribute{
							Description: "Web URL for the team.",
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
				},
			},
		},
	}
}

func (d *githubEnterpriseTeamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *githubEnterpriseTeamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config githubEnterpriseTeamsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !requireGitHubClient(&resp.Diagnostics, d.client) {
		return
	}

	enterprise := strings.TrimSpace(config.Enterprise.ValueString())
	items, err := githubListAll(ctx, d.client, enterpriseTeamsPath(enterprise))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Enterprise Teams",
			fmt.Sprintf("Could not list enterprise teams: %s", err),
		)
		return
	}

	var teams []githubEnterpriseTeamDataModel
	for _, item := range items {
		team := githubEnterpriseTeamDataModel{
			Id:         types.StringValue(mapString(item, "id")),
			Name:       types.StringValue(mapString(item, "name")),
			Slug:       types.StringValue(mapString(item, "slug")),
			URL:        types.StringValue(mapString(item, "url")),
			HTMLURL:    types.StringValue(mapString(item, "html_url")),
			MembersURL: types.StringValue(mapString(item, "members_url")),
			CreatedAt:  types.StringValue(mapString(item, "created_at")),
			UpdatedAt:  types.StringValue(mapString(item, "updated_at")),
		}
		if raw := mapString(item, "description"); raw != "" || item["description"] != nil {
			if item["description"] == nil {
				team.Description = types.StringNull()
			} else {
				team.Description = types.StringValue(raw)
			}
		}
		if raw := mapString(item, "group_id"); raw != "" || item["group_id"] != nil {
			if item["group_id"] == nil {
				team.GroupId = types.StringNull()
			} else {
				team.GroupId = types.StringValue(raw)
			}
		}
		teams = append(teams, team)
	}

	config.Teams = teams
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
