package types

import "github.com/hashicorp/terraform-plugin-framework/types"

// GithubEnterpriseTeamOrganizationsModel describes bulk enterprise team organization assignments.
type GithubEnterpriseTeamOrganizationsModel struct {
	Id                types.String `tfsdk:"id"`
	Enterprise        types.String `tfsdk:"enterprise"`
	TeamSlug          types.String `tfsdk:"team_slug"`
	OrganizationSlugs types.Set    `tfsdk:"organization_slugs"`
}
