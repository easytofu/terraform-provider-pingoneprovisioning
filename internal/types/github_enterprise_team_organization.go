package types

import "github.com/hashicorp/terraform-plugin-framework/types"

// GithubEnterpriseTeamOrganizationModel describes a single enterprise team organization assignment.
type GithubEnterpriseTeamOrganizationModel struct {
	Id           types.String `tfsdk:"id"`
	Enterprise   types.String `tfsdk:"enterprise"`
	TeamSlug     types.String `tfsdk:"team_slug"`
	Organization types.String `tfsdk:"organization"`
}
