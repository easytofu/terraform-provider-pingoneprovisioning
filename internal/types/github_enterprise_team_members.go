package types

import "github.com/hashicorp/terraform-plugin-framework/types"

// GithubEnterpriseTeamMembersModel describes bulk enterprise team memberships.
type GithubEnterpriseTeamMembersModel struct {
	Id         types.String `tfsdk:"id"`
	Enterprise types.String `tfsdk:"enterprise"`
	TeamSlug   types.String `tfsdk:"team_slug"`
	Usernames  types.Set    `tfsdk:"usernames"`
}
