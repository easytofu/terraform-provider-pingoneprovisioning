package utils

import "testing"

func TestNormalizePropagationStoreTypeForAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "githubemu_legacy", in: "GithubEMU", want: "GitHubEMU"},
		{name: "githubemu_canonical", in: "GitHubEMU", want: "GitHubEMU"},
		{name: "scim_upper", in: "SCIM", want: "scim"},
		{name: "scim_lower", in: "scim", want: "scim"},
		{name: "ldapgateway_upper", in: "LDAPGateway", want: "LdapGateway"},
		{name: "ldapgateway_api", in: "LdapGateway", want: "LdapGateway"},
		{name: "azure_tf", in: "AzureADSAMLV2", want: "AzureActiveDirectorySAML2"},
		{name: "azure_api", in: "AzureActiveDirectorySAML2", want: "AzureActiveDirectorySAML2"},
		{name: "passthrough", in: "PingOne", want: "PingOne"},
		{name: "trims_whitespace", in: "  GitHubEMU ", want: "GitHubEMU"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := NormalizePropagationStoreTypeForAPI(tt.in); got != tt.want {
				t.Fatalf("NormalizePropagationStoreTypeForAPI(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizePropagationStoreTypeForTerraform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		api  string
		pref string
		want string
	}{
		{name: "empty", api: "", pref: "", want: ""},
		{name: "github_defaults_to_legacy_spelling", api: "GitHubEMU", pref: "", want: "GithubEMU"},
		{name: "github_keeps_legacy_spelling_when_configured", api: "GitHubEMU", pref: "GithubEMU", want: "GithubEMU"},
		{name: "github_preserves_configured_spelling", api: "GitHubEMU", pref: "GitHubEMU", want: "GitHubEMU"},
		{name: "scim_normalizes", api: "scim", pref: "", want: "SCIM"},
		{name: "scim_preserves_configured_lowercase", api: "scim", pref: "scim", want: "scim"},
		{name: "ldapgateway_normalizes", api: "LdapGateway", pref: "", want: "LDAPGateway"},
		{name: "azure_normalizes", api: "AzureActiveDirectorySAML2", pref: "", want: "AzureADSAMLV2"},
		{name: "case_insensitive_pref_preserved", api: "pingone", pref: "PingOne", want: "PingOne"},
		{name: "trims_whitespace", api: "  scim ", pref: "", want: "SCIM"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := NormalizePropagationStoreTypeForTerraform(tt.api, tt.pref); got != tt.want {
				t.Fatalf("NormalizePropagationStoreTypeForTerraform(%q, %q) = %q, want %q", tt.api, tt.pref, got, tt.want)
			}
		})
	}
}
