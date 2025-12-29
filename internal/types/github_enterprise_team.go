package types

import "github.com/hashicorp/terraform-plugin-framework/types"

// GithubEnterpriseTeamModel describes the Terraform model for a GitHub enterprise team.
type GithubEnterpriseTeamModel struct {
	Id                        types.String `tfsdk:"id"`
	Enterprise                types.String `tfsdk:"enterprise"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Slug                      types.String `tfsdk:"slug"`
	GroupId                   types.String `tfsdk:"group_id"`
	OrganizationSelectionType types.String `tfsdk:"organization_selection_type"`
	OrganizationSlugs         types.Set    `tfsdk:"organization_slugs"`
	URL                       types.String `tfsdk:"url"`
	HTMLURL                   types.String `tfsdk:"html_url"`
	MembersURL                types.String `tfsdk:"members_url"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
}
