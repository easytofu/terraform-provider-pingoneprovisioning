package mappers

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	customtypes "github.com/easytofu/terraform-provider-pingoneprovisioning/internal/types"
)

func TestApplyPropagationStoreConfigurationFromMap(t *testing.T) {
	t.Parallel()

	t.Run("scim_prefers_scim_configuration_and_preserves_secrets", func(t *testing.T) {
		t.Parallel()

		model := customtypes.PropagationStoreModel{
			ConfigurationAquera: &customtypes.ConfigurationAquera{},
		}

		prior := &customtypes.PropagationStoreModel{
			ScimConfiguration: &customtypes.ConfigurationScim{
				OauthClientSecret: types.StringValue("secret"),
			},
		}

		config := map[string]interface{}{
			"AUTHENTICATION_METHOD":  "OAuth 2 Client Credentials",
			"AUTHORIZATION_TYPE":     "Bearer",
			"OAUTH_CLIENT_ID":        "client-id",
			"SCIM_URL":               "https://example/scim",
			"SCIM_VERSION":           "2.0",
			"UNIQUE_USER_IDENTIFIER": "userName",
			"USER_FILTER":            "active eq true",
			"USERS_RESOURCE":         "/Users",
			"GROUPS_RESOURCE":        "/Groups",
			"CREATE_USERS":           true,
			"DISABLE_USERS":          true,
			"UPDATE_USERS":           true,
		}

		ApplyPropagationStoreConfigurationFromMap(&model, "scim", config, prior)

		if model.ScimConfiguration == nil {
			t.Fatalf("expected scim_configuration to be set")
		}
		if model.ConfigurationScim != nil {
			t.Fatalf("expected configuration_scim to be nil when scim_configuration was configured")
		}
		if got := model.ScimConfiguration.OauthClientSecret.ValueString(); got != "secret" {
			t.Fatalf("expected oauth_client_secret to be preserved, got %q", got)
		}

		if model.ConfigurationAquera != nil ||
			model.ConfigurationAzureAdSamlV2 != nil ||
			model.ConfigurationGoogleApps != nil ||
			model.ConfigurationLdapGateway != nil ||
			model.ConfigurationPingOne != nil ||
			model.ConfigurationSalesforce != nil ||
			model.ConfigurationSalesforceContacts != nil ||
			model.ConfigurationServiceNow != nil ||
			model.ConfigurationSlack != nil ||
			model.ConfigurationWorkday != nil ||
			model.ConfigurationZoom != nil {
			t.Fatalf("expected non-matching configuration blocks to be nil")
		}
	})

	t.Run("unknown_type_clears_all_blocks", func(t *testing.T) {
		t.Parallel()

		model := customtypes.PropagationStoreModel{
			ConfigurationAquera: &customtypes.ConfigurationAquera{},
		}

		ApplyPropagationStoreConfigurationFromMap(&model, "directory", map[string]interface{}{"BASE_URL": "x"}, nil)

		if model.ConfigurationAquera != nil ||
			model.ConfigurationAzureAdSamlV2 != nil ||
			model.ConfigurationGoogleApps != nil ||
			model.ConfigurationLdapGateway != nil ||
			model.ConfigurationPingOne != nil ||
			model.ConfigurationSalesforce != nil ||
			model.ConfigurationSalesforceContacts != nil ||
			model.ConfigurationScim != nil ||
			model.ScimConfiguration != nil ||
			model.ConfigurationServiceNow != nil ||
			model.ConfigurationSlack != nil ||
			model.ConfigurationWorkday != nil ||
			model.ConfigurationZoom != nil {
			t.Fatalf("expected all configuration blocks to be nil")
		}
	})
}
