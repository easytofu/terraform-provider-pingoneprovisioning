package mappers

import (
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
)

// ApplyPropagationStoreConfigurationFromMap populates exactly one configuration_* block on the
// model based on the propagation store type and clears all others.
//
// Many propagation store types share configuration keys (e.g. BASE_URL, CREATE_USERS). If we
// populate every configuration block from the same map, Terraform sees blocks "appear" after
// apply, causing state inconsistencies.
func ApplyPropagationStoreConfigurationFromMap(model *customtypes.PropagationStoreModel, tfType string, config map[string]interface{}, prior *customtypes.PropagationStoreModel) {
	model.ConfigurationAquera = nil
	model.ConfigurationAzureAdSamlV2 = nil
	model.ConfigurationGithubEmu = nil
	model.ConfigurationGoogleApps = nil
	model.ConfigurationLdapGateway = nil
	model.ConfigurationPingOne = nil
	model.ConfigurationSalesforce = nil
	model.ConfigurationSalesforceContacts = nil
	model.ConfigurationScim = nil
	model.ScimConfiguration = nil
	model.ConfigurationServiceNow = nil
	model.ConfigurationSlack = nil
	model.ConfigurationWorkday = nil
	model.ConfigurationZoom = nil

	switch tfType {
	case "Aquera":
		model.ConfigurationAquera = &customtypes.ConfigurationAquera{}
		AqueraFromMap(model.ConfigurationAquera, config)
	case "AzureADSAMLV2":
		model.ConfigurationAzureAdSamlV2 = &customtypes.ConfigurationAzureAdSamlV2{}
		AzureAdSamlV2FromMap(model.ConfigurationAzureAdSamlV2, config)
	case "GithubEMU", "GitHubEMU":
		model.ConfigurationGithubEmu = &customtypes.ConfigurationGithubEmu{}
		GithubEMUFromMap(model.ConfigurationGithubEmu, config)

		// GitHub EMU tokens are typically write-only; preserve the configured value when the
		// API response does not include it.
		if prior != nil && prior.ConfigurationGithubEmu != nil {
			if (model.ConfigurationGithubEmu.OauthAccessToken.IsNull() || model.ConfigurationGithubEmu.OauthAccessToken.IsUnknown()) &&
				!prior.ConfigurationGithubEmu.OauthAccessToken.IsNull() && !prior.ConfigurationGithubEmu.OauthAccessToken.IsUnknown() {
				model.ConfigurationGithubEmu.OauthAccessToken = prior.ConfigurationGithubEmu.OauthAccessToken
			}
		}
	case "GoogleApps":
		model.ConfigurationGoogleApps = &customtypes.ConfigurationGoogleApps{}
		GoogleAppsFromMap(model.ConfigurationGoogleApps, config)
	case "LDAPGateway":
		model.ConfigurationLdapGateway = &customtypes.ConfigurationLdapGateway{}
		LdapGatewayFromMap(model.ConfigurationLdapGateway, config)
	case "PingOne":
		model.ConfigurationPingOne = &customtypes.ConfigurationPingOne{}
		PingOneFromMap(model.ConfigurationPingOne, config)
	case "Salesforce":
		model.ConfigurationSalesforce = &customtypes.ConfigurationSalesforce{}
		SalesforceFromMap(model.ConfigurationSalesforce, config)
	case "SalesforceContacts":
		model.ConfigurationSalesforceContacts = &customtypes.ConfigurationSalesforceContacts{}
		SalesforceContactsFromMap(model.ConfigurationSalesforceContacts, config)
	case "SCIM", "scim":
		useScimConfiguration := prior != nil && prior.ScimConfiguration != nil
		if useScimConfiguration {
			model.ScimConfiguration = &customtypes.ConfigurationScim{}
			ScimFromMap(model.ScimConfiguration, config)
		} else {
			model.ConfigurationScim = &customtypes.ConfigurationScim{}
			ScimFromMap(model.ConfigurationScim, config)
		}

		// SCIM secrets are typically write-only; preserve configured values when the API
		// response does not include them.
		if prior != nil {
			var priorScim *customtypes.ConfigurationScim
			if prior.ScimConfiguration != nil {
				priorScim = prior.ScimConfiguration
			} else if prior.ConfigurationScim != nil {
				priorScim = prior.ConfigurationScim
			}

			var currentScim *customtypes.ConfigurationScim
			if model.ScimConfiguration != nil {
				currentScim = model.ScimConfiguration
			} else {
				currentScim = model.ConfigurationScim
			}

			if priorScim != nil && currentScim != nil {
				if (currentScim.BasicAuthPassword.IsNull() || currentScim.BasicAuthPassword.IsUnknown()) &&
					!priorScim.BasicAuthPassword.IsNull() && !priorScim.BasicAuthPassword.IsUnknown() {
					currentScim.BasicAuthPassword = priorScim.BasicAuthPassword
				}
				if (currentScim.OauthAccessToken.IsNull() || currentScim.OauthAccessToken.IsUnknown()) &&
					!priorScim.OauthAccessToken.IsNull() && !priorScim.OauthAccessToken.IsUnknown() {
					currentScim.OauthAccessToken = priorScim.OauthAccessToken
				}
				if (currentScim.OauthClientSecret.IsNull() || currentScim.OauthClientSecret.IsUnknown()) &&
					!priorScim.OauthClientSecret.IsNull() && !priorScim.OauthClientSecret.IsUnknown() {
					currentScim.OauthClientSecret = priorScim.OauthClientSecret
				}
			}
		}
	case "ServiceNow":
		model.ConfigurationServiceNow = &customtypes.ConfigurationServiceNow{}
		ServiceNowFromMap(model.ConfigurationServiceNow, config)
	case "Slack":
		model.ConfigurationSlack = &customtypes.ConfigurationSlack{}
		SlackFromMap(model.ConfigurationSlack, config)
	case "Workday":
		model.ConfigurationWorkday = &customtypes.ConfigurationWorkday{}
		WorkdayFromMap(model.ConfigurationWorkday, config)
	case "Zoom":
		model.ConfigurationZoom = &customtypes.ConfigurationZoom{}
		ZoomFromMap(model.ConfigurationZoom, config)
	}
}
