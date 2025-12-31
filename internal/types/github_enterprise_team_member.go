package types

import "github.com/hashicorp/terraform-plugin-framework/types"

// GithubEnterpriseTeamMemberModel describes a single enterprise team membership.
type GithubEnterpriseTeamMemberModel struct {
	Id         types.String `tfsdk:"id"`
	Enterprise types.String `tfsdk:"enterprise"`
	TeamSlug   types.String `tfsdk:"team_slug"`
	Username   types.String `tfsdk:"username"`
}
