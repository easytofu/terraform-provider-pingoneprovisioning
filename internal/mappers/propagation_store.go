// pingone/providers/pingone-propagation/internal/mappers/propagation_store.go
package mappers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	customtypes "github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/types"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/utils"
)

// ModelToConfigurationMap builds the configuration map to send to PingOne based on the propagation store type.
func ModelToConfigurationMap(m *customtypes.PropagationStoreModel) (map[string]interface{}, error) {
	storeType := m.Type.ValueString()

	switch storeType {
	case "Aquera":
		if m.ConfigurationAquera == nil {
			return nil, fmt.Errorf("configuration_aquera must be provided for type 'Aquera'")
		}
		return AqueraToMap(m.ConfigurationAquera), nil
	case "AzureADSAMLV2":
		if m.ConfigurationAzureAdSamlV2 == nil {
			return nil, fmt.Errorf("configuration_azure_ad_saml_v2 must be provided for type 'AzureADSAMLV2'")
		}
		return AzureAdSamlV2ToMap(m.ConfigurationAzureAdSamlV2), nil
	case "GithubEMU", "GitHubEMU":
		if m.ConfigurationGithubEmu == nil {
			return nil, fmt.Errorf("configuration_github_emu must be provided for type 'GithubEMU'")
		}
		return GithubEMUToMap(m.ConfigurationGithubEmu), nil
	case "GoogleApps":
		if m.ConfigurationGoogleApps == nil {
			return nil, fmt.Errorf("configuration_google_apps must be provided for type 'GoogleApps'")
		}
		return GoogleAppsToMap(m.ConfigurationGoogleApps), nil
	case "LDAPGateway":
		if m.ConfigurationLdapGateway == nil {
			return nil, fmt.Errorf("configuration_ldap_gateway must be provided for type 'LDAPGateway'")
		}
		return LdapGatewayToMap(m.ConfigurationLdapGateway), nil
	case "PingOne":
		if m.ConfigurationPingOne == nil {
			return nil, fmt.Errorf("configuration_ping_one must be provided for type 'PingOne'")
		}
		return PingOneToMap(m.ConfigurationPingOne), nil
	case "Salesforce":
		if m.ConfigurationSalesforce == nil {
			return nil, fmt.Errorf("configuration_salesforce must be provided for type 'Salesforce'")
		}
		return SalesforceToMap(m.ConfigurationSalesforce), nil
	case "SalesforceContacts":
		if m.ConfigurationSalesforceContacts == nil {
			return nil, fmt.Errorf("configuration_salesforce_contacts must be provided for type 'SalesforceContacts'")
		}
		return SalesforceContactsToMap(m.ConfigurationSalesforceContacts), nil
	case "SCIM", "scim":
		if m.ConfigurationScim != nil && m.ScimConfiguration != nil {
			return nil, fmt.Errorf("only one of configuration_scim or scim_configuration can be provided for type 'SCIM'")
		}

		scimConfig := m.ConfigurationScim
		if scimConfig == nil {
			scimConfig = m.ScimConfiguration
		}
		if scimConfig == nil {
			return nil, fmt.Errorf("configuration_scim (or scim_configuration) must be provided for type 'SCIM'")
		}
		return ScimToMap(scimConfig), nil
	case "ServiceNow":
		if m.ConfigurationServiceNow == nil {
			return nil, fmt.Errorf("configuration_service_now must be provided for type 'ServiceNow'")
		}
		return ServiceNowToMap(m.ConfigurationServiceNow), nil
	case "Slack":
		if m.ConfigurationSlack == nil {
			return nil, fmt.Errorf("configuration_slack must be provided for type 'Slack'")
		}
		return SlackToMap(m.ConfigurationSlack), nil
	case "Workday":
		if m.ConfigurationWorkday == nil {
			return nil, fmt.Errorf("configuration_workday must be provided for type 'Workday'")
		}
		return WorkdayToMap(m.ConfigurationWorkday), nil
	case "Zoom":
		if m.ConfigurationZoom == nil {
			return nil, fmt.Errorf("configuration_zoom must be provided for type 'Zoom'")
		}
		return ZoomToMap(m.ConfigurationZoom), nil
	default:
		return nil, fmt.Errorf("unsupported propagation store type: %s", storeType)
	}
}

// AqueraToMap maps Aquera configuration model to API configuration map.
func AqueraToMap(c *customtypes.ConfigurationAquera) map[string]interface{} {
	m := make(map[string]interface{})

	m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	m["BASE_URL"] = c.BaseUrl.ValueString()

	if !c.ApiKey.IsNull() {
		m["API_KEY"] = c.ApiKey.ValueString()
	}
	if !c.ApiSecret.IsNull() {
		m["API_SECRET"] = c.ApiSecret.ValueString()
	}
	if !c.BearerToken.IsNull() {
		m["BEARER_TOKEN"] = c.BearerToken.ValueString()
	}
	if !c.Domain.IsNull() {
		m["DOMAIN"] = c.Domain.ValueString()
	}
	if !c.Username.IsNull() {
		m["USERNAME"] = c.Username.ValueString()
	}
	if !c.Password.IsNull() {
		m["PASSWORD"] = c.Password.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// AqueraFromMap maps API configuration map to Aquera configuration model.
func AqueraFromMap(c *customtypes.ConfigurationAquera, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.AuthenticationMethod = utils.FromMapString(config, "AUTHENTICATION_METHOD")
	c.BaseUrl = utils.FromMapString(config, "BASE_URL")

	c.ApiKey = utils.FromMapString(config, "API_KEY")
	c.ApiSecret = utils.FromMapString(config, "API_SECRET")
	c.BearerToken = utils.FromMapString(config, "BEARER_TOKEN")
	c.Domain = utils.FromMapString(config, "DOMAIN")
	c.Username = utils.FromMapString(config, "USERNAME")
	c.Password = utils.FromMapString(config, "PASSWORD")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// AzureAdSamlV2ToMap maps Azure AD SAML v2 configuration model to API configuration map.
func AzureAdSamlV2ToMap(c *customtypes.ConfigurationAzureAdSamlV2) map[string]interface{} {
	m := make(map[string]interface{})

	m["BASE_URL"] = c.BaseUrl.ValueString()
	m["SCIM_URL"] = c.ScimUrl.ValueString()

	if !c.BearerToken.IsNull() {
		m["BEARER_TOKEN"] = c.BearerToken.ValueString()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// AzureAdSamlV2FromMap maps API configuration map to Azure AD SAML v2 configuration model.
func AzureAdSamlV2FromMap(c *customtypes.ConfigurationAzureAdSamlV2, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.BaseUrl = utils.FromMapString(config, "BASE_URL")
	c.ScimUrl = utils.FromMapString(config, "SCIM_URL")

	c.BearerToken = utils.FromMapString(config, "BEARER_TOKEN")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// GithubEMUToMap maps GitHub EMU configuration model to API configuration map.
func GithubEMUToMap(c *customtypes.ConfigurationGithubEmu) map[string]interface{} {
	m := make(map[string]interface{})

	m["BASE_URL"] = c.BaseUrl.ValueString()

	if !c.OauthAccessToken.IsNull() {
		m["OAUTH_ACCESS_TOKEN"] = c.OauthAccessToken.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// GithubEMUFromMap maps API configuration map to GitHub EMU configuration model.
func GithubEMUFromMap(c *customtypes.ConfigurationGithubEmu, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.BaseUrl = utils.FromMapString(config, "BASE_URL")

	c.OauthAccessToken = utils.FromMapString(config, "OAUTH_ACCESS_TOKEN")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// GoogleAppsToMap maps Google Apps configuration model to API configuration map.
func GoogleAppsToMap(c *customtypes.ConfigurationGoogleApps) map[string]interface{} {
	m := make(map[string]interface{})

	m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	m["BASE_URL"] = c.BaseUrl.ValueString()

	if !c.OauthClientId.IsNull() {
		m["OAUTH_CLIENT_ID"] = c.OauthClientId.ValueString()
	}
	if !c.OauthClientSecret.IsNull() {
		m["OAUTH_CLIENT_SECRET"] = c.OauthClientSecret.ValueString()
	}
	if !c.OauthRefreshToken.IsNull() {
		m["OAUTH_REFRESH_TOKEN"] = c.OauthRefreshToken.ValueString()
	}
	if !c.OauthTokenUrl.IsNull() {
		m["OAUTH_TOKEN_URL"] = c.OauthTokenUrl.ValueString()
	}
	if !c.Domain.IsNull() {
		m["DOMAIN"] = c.Domain.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// GoogleAppsFromMap maps API configuration map to Google Apps configuration model.
func GoogleAppsFromMap(c *customtypes.ConfigurationGoogleApps, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.AuthenticationMethod = utils.FromMapString(config, "AUTHENTICATION_METHOD")
	c.BaseUrl = utils.FromMapString(config, "BASE_URL")

	c.OauthClientId = utils.FromMapString(config, "OAUTH_CLIENT_ID")
	c.OauthClientSecret = utils.FromMapString(config, "OAUTH_CLIENT_SECRET")
	c.OauthRefreshToken = utils.FromMapString(config, "OAUTH_REFRESH_TOKEN")
	c.OauthTokenUrl = utils.FromMapString(config, "OAUTH_TOKEN_URL")
	c.Domain = utils.FromMapString(config, "DOMAIN")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// LdapGatewayToMap maps LDAP Gateway configuration model to API configuration map.
func LdapGatewayToMap(c *customtypes.ConfigurationLdapGateway) map[string]interface{} {
	m := make(map[string]interface{})

	m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	m["BASE_URL"] = c.BaseUrl.ValueString()
	m["LDAP_GATEWAY_ID"] = c.LdapGatewayId.ValueString()
	m["LDAP_GATEWAY_REGION"] = c.LdapGatewayRegion.ValueString()

	if !c.ApiKey.IsNull() {
		m["API_KEY"] = c.ApiKey.ValueString()
	}
	if !c.ApiSecret.IsNull() {
		m["API_SECRET"] = c.ApiSecret.ValueString()
	}
	if !c.BearerToken.IsNull() {
		m["BEARER_TOKEN"] = c.BearerToken.ValueString()
	}
	if !c.Username.IsNull() {
		m["USERNAME"] = c.Username.ValueString()
	}
	if !c.Password.IsNull() {
		m["PASSWORD"] = c.Password.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// LdapGatewayFromMap maps API configuration map to LDAP Gateway configuration model.
func LdapGatewayFromMap(c *customtypes.ConfigurationLdapGateway, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.AuthenticationMethod = utils.FromMapString(config, "AUTHENTICATION_METHOD")
	c.BaseUrl = utils.FromMapString(config, "BASE_URL")
	c.LdapGatewayId = utils.FromMapString(config, "LDAP_GATEWAY_ID")
	c.LdapGatewayRegion = utils.FromMapString(config, "LDAP_GATEWAY_REGION")

	c.ApiKey = utils.FromMapString(config, "API_KEY")
	c.ApiSecret = utils.FromMapString(config, "API_SECRET")
	c.BearerToken = utils.FromMapString(config, "BEARER_TOKEN")
	c.Username = utils.FromMapString(config, "USERNAME")
	c.Password = utils.FromMapString(config, "PASSWORD")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// PingOneToMap maps PingOne configuration model to API configuration map.
func PingOneToMap(c *customtypes.ConfigurationPingOne) map[string]interface{} {
	m := make(map[string]interface{})

	if !c.AuthenticationMethod.IsNull() && !c.AuthenticationMethod.IsUnknown() {
		m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	}
	if !c.BaseUrl.IsNull() && !c.BaseUrl.IsUnknown() {
		m["BASE_URL"] = c.BaseUrl.ValueString()
	}
	if !c.ScimUrl.IsNull() && !c.ScimUrl.IsUnknown() {
		m["SCIM_URL"] = c.ScimUrl.ValueString()
	}

	if !c.BearerToken.IsNull() {
		m["BEARER_TOKEN"] = c.BearerToken.ValueString()
	}
	if !c.OauthClientId.IsNull() {
		m["OAUTH_CLIENT_ID"] = c.OauthClientId.ValueString()
	}
	if !c.OauthClientSecret.IsNull() {
		m["OAUTH_CLIENT_SECRET"] = c.OauthClientSecret.ValueString()
	}
	if !c.OauthTokenUrl.IsNull() {
		m["OAUTH_TOKEN_URL"] = c.OauthTokenUrl.ValueString()
	}
	if !c.OauthRefreshToken.IsNull() {
		m["OAUTH_REFRESH_TOKEN"] = c.OauthRefreshToken.ValueString()
	}
	if !c.Domain.IsNull() {
		m["DOMAIN"] = c.Domain.ValueString()
	}
	if !c.Username.IsNull() {
		m["USERNAME"] = c.Username.ValueString()
	}
	if !c.Password.IsNull() {
		m["PASSWORD"] = c.Password.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// PingOneFromMap maps API configuration map to PingOne configuration model.
func PingOneFromMap(c *customtypes.ConfigurationPingOne, config map[string]interface{}) {
	if config == nil {
		return
	}

	// These fields are optional in Terraform for PingOne stores. When unset, we must
	// not "invent" empty strings or Terraform will raise "inconsistent result after apply".
	authenticationMethod := utils.FromMapString(config, "AUTHENTICATION_METHOD")
	if !authenticationMethod.IsNull() && authenticationMethod.ValueString() == "" {
		authenticationMethod = types.StringNull()
	}
	c.AuthenticationMethod = authenticationMethod

	baseUrl := utils.FromMapString(config, "BASE_URL")
	if !baseUrl.IsNull() && baseUrl.ValueString() == "" {
		baseUrl = types.StringNull()
	}
	c.BaseUrl = baseUrl

	scimUrl := utils.FromMapString(config, "SCIM_URL")
	if !scimUrl.IsNull() && scimUrl.ValueString() == "" {
		scimUrl = types.StringNull()
	}
	c.ScimUrl = scimUrl

	c.BearerToken = utils.FromMapString(config, "BEARER_TOKEN")
	c.OauthClientId = utils.FromMapString(config, "OAUTH_CLIENT_ID")
	c.OauthClientSecret = utils.FromMapString(config, "OAUTH_CLIENT_SECRET")
	c.OauthTokenUrl = utils.FromMapString(config, "OAUTH_TOKEN_URL")
	c.OauthRefreshToken = utils.FromMapString(config, "OAUTH_REFRESH_TOKEN")
	c.Domain = utils.FromMapString(config, "DOMAIN")
	c.Username = utils.FromMapString(config, "USERNAME")
	c.Password = utils.FromMapString(config, "PASSWORD")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// SalesforceToMap maps Salesforce configuration model to API configuration map.
func SalesforceToMap(c *customtypes.ConfigurationSalesforce) map[string]interface{} {
	m := make(map[string]interface{})

	m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	m["BASE_URL"] = c.BaseUrl.ValueString()
	m["SCIM_URL"] = c.ScimUrl.ValueString()

	if !c.ConsumerKey.IsNull() {
		m["CONSUMER_KEY"] = c.ConsumerKey.ValueString()
	}
	if !c.ConsumerSecret.IsNull() {
		m["CONSUMER_SECRET"] = c.ConsumerSecret.ValueString()
	}
	if !c.Username.IsNull() {
		m["USERNAME"] = c.Username.ValueString()
	}
	if !c.Password.IsNull() {
		m["PASSWORD"] = c.Password.ValueString()
	}
	if !c.SecurityToken.IsNull() {
		m["SECURITY_TOKEN"] = c.SecurityToken.ValueString()
	}
	if !c.BearerToken.IsNull() {
		m["BEARER_TOKEN"] = c.BearerToken.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// SalesforceFromMap maps API configuration map to Salesforce configuration model.
func SalesforceFromMap(c *customtypes.ConfigurationSalesforce, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.AuthenticationMethod = utils.FromMapString(config, "AUTHENTICATION_METHOD")
	c.BaseUrl = utils.FromMapString(config, "BASE_URL")
	c.ScimUrl = utils.FromMapString(config, "SCIM_URL")

	c.ConsumerKey = utils.FromMapString(config, "CONSUMER_KEY")
	c.ConsumerSecret = utils.FromMapString(config, "CONSUMER_SECRET")
	c.Username = utils.FromMapString(config, "USERNAME")
	c.Password = utils.FromMapString(config, "PASSWORD")
	c.SecurityToken = utils.FromMapString(config, "SECURITY_TOKEN")
	c.BearerToken = utils.FromMapString(config, "BEARER_TOKEN")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// SalesforceContactsToMap maps Salesforce Contacts configuration model to API configuration map.
func SalesforceContactsToMap(c *customtypes.ConfigurationSalesforceContacts) map[string]interface{} {
	m := make(map[string]interface{})

	m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	m["BASE_URL"] = c.BaseUrl.ValueString()
	m["SCIM_URL"] = c.ScimUrl.ValueString()

	if !c.ConsumerKey.IsNull() {
		m["CONSUMER_KEY"] = c.ConsumerKey.ValueString()
	}
	if !c.ConsumerSecret.IsNull() {
		m["CONSUMER_SECRET"] = c.ConsumerSecret.ValueString()
	}
	if !c.Username.IsNull() {
		m["USERNAME"] = c.Username.ValueString()
	}
	if !c.Password.IsNull() {
		m["PASSWORD"] = c.Password.ValueString()
	}
	if !c.SecurityToken.IsNull() {
		m["SECURITY_TOKEN"] = c.SecurityToken.ValueString()
	}
	if !c.BearerToken.IsNull() {
		m["BEARER_TOKEN"] = c.BearerToken.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// SalesforceContactsFromMap maps API configuration map to Salesforce Contacts configuration model.
func SalesforceContactsFromMap(c *customtypes.ConfigurationSalesforceContacts, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.AuthenticationMethod = utils.FromMapString(config, "AUTHENTICATION_METHOD")
	c.BaseUrl = utils.FromMapString(config, "BASE_URL")
	c.ScimUrl = utils.FromMapString(config, "SCIM_URL")

	c.ConsumerKey = utils.FromMapString(config, "CONSUMER_KEY")
	c.ConsumerSecret = utils.FromMapString(config, "CONSUMER_SECRET")
	c.Username = utils.FromMapString(config, "USERNAME")
	c.Password = utils.FromMapString(config, "PASSWORD")
	c.SecurityToken = utils.FromMapString(config, "SECURITY_TOKEN")
	c.BearerToken = utils.FromMapString(config, "BEARER_TOKEN")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// ScimToMap maps SCIM configuration model to API configuration map.
func ScimToMap(c *customtypes.ConfigurationScim) map[string]interface{} {
	m := make(map[string]interface{})

	m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	m["AUTHORIZATION_TYPE"] = c.AuthorizationType.ValueString()
	m["SCIM_URL"] = c.ScimUrl.ValueString()
	m["SCIM_VERSION"] = c.ScimVersion.ValueString()
	m["UNIQUE_USER_IDENTIFIER"] = c.UniqueUserIdentifier.ValueString()
	m["USER_FILTER"] = c.UserFilter.ValueString()
	m["USERS_RESOURCE"] = c.UsersResource.ValueString()

	if !c.BasicAuthUser.IsNull() {
		m["BASIC_AUTH_USER"] = c.BasicAuthUser.ValueString()
	}
	if !c.BasicAuthPassword.IsNull() {
		m["BASIC_AUTH_PASSWORD"] = c.BasicAuthPassword.ValueString()
	}
	if !c.GroupsResource.IsNull() {
		m["GROUPS_RESOURCE"] = c.GroupsResource.ValueString()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.OauthAccessToken.IsNull() {
		m["OAUTH_ACCESS_TOKEN"] = c.OauthAccessToken.ValueString()
	}
	if !c.OauthClientId.IsNull() {
		m["OAUTH_CLIENT_ID"] = c.OauthClientId.ValueString()
	}
	if !c.OauthClientSecret.IsNull() {
		m["OAUTH_CLIENT_SECRET"] = c.OauthClientSecret.ValueString()
	}
	if !c.OauthTokenRequest.IsNull() {
		m["OAUTH_TOKEN_REQUEST"] = c.OauthTokenRequest.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// ScimFromMap maps API configuration map to SCIM configuration model.
func ScimFromMap(c *customtypes.ConfigurationScim, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.AuthenticationMethod = utils.FromMapString(config, "AUTHENTICATION_METHOD")
	c.AuthorizationType = utils.FromMapString(config, "AUTHORIZATION_TYPE")
	c.BasicAuthUser = utils.FromMapString(config, "BASIC_AUTH_USER")
	c.BasicAuthPassword = utils.FromMapString(config, "BASIC_AUTH_PASSWORD")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.GroupsResource = utils.FromMapString(config, "GROUPS_RESOURCE")
	c.OauthAccessToken = utils.FromMapString(config, "OAUTH_ACCESS_TOKEN")
	c.OauthClientId = utils.FromMapString(config, "OAUTH_CLIENT_ID")
	c.OauthClientSecret = utils.FromMapString(config, "OAUTH_CLIENT_SECRET")
	c.OauthTokenRequest = utils.FromMapString(config, "OAUTH_TOKEN_REQUEST")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.ScimUrl = utils.FromMapString(config, "SCIM_URL")
	c.ScimVersion = utils.FromMapString(config, "SCIM_VERSION")
	c.UniqueUserIdentifier = utils.FromMapString(config, "UNIQUE_USER_IDENTIFIER")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
	c.UserFilter = utils.FromMapString(config, "USER_FILTER")
	c.UsersResource = utils.FromMapString(config, "USERS_RESOURCE")
}

// ServiceNowToMap maps ServiceNow configuration model to API configuration map.
func ServiceNowToMap(c *customtypes.ConfigurationServiceNow) map[string]interface{} {
	m := make(map[string]interface{})

	m["BASE_URL"] = c.BaseUrl.ValueString()
	m["SCIM_URL"] = c.ScimUrl.ValueString()

	if !c.Username.IsNull() {
		m["USERNAME"] = c.Username.ValueString()
	}
	if !c.Password.IsNull() {
		m["PASSWORD"] = c.Password.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// ServiceNowFromMap maps API configuration map to ServiceNow configuration model.
func ServiceNowFromMap(c *customtypes.ConfigurationServiceNow, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.BaseUrl = utils.FromMapString(config, "BASE_URL")
	c.ScimUrl = utils.FromMapString(config, "SCIM_URL")

	c.Username = utils.FromMapString(config, "USERNAME")
	c.Password = utils.FromMapString(config, "PASSWORD")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// SlackToMap maps Slack configuration model to API configuration map.
func SlackToMap(c *customtypes.ConfigurationSlack) map[string]interface{} {
	m := make(map[string]interface{})

	m["BASE_URL"] = c.BaseUrl.ValueString()
	m["SCIM_URL"] = c.ScimUrl.ValueString()

	if !c.BearerToken.IsNull() {
		m["BEARER_TOKEN"] = c.BearerToken.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// SlackFromMap maps API configuration map to Slack configuration model.
func SlackFromMap(c *customtypes.ConfigurationSlack, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.BaseUrl = utils.FromMapString(config, "BASE_URL")
	c.ScimUrl = utils.FromMapString(config, "SCIM_URL")

	c.BearerToken = utils.FromMapString(config, "BEARER_TOKEN")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// WorkdayToMap maps Workday configuration model to API configuration map.
func WorkdayToMap(c *customtypes.ConfigurationWorkday) map[string]interface{} {
	m := make(map[string]interface{})

	m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	m["BASE_URL"] = c.BaseUrl.ValueString()
	m["SCIM_URL"] = c.ScimUrl.ValueString()

	if !c.Username.IsNull() {
		m["USERNAME"] = c.Username.ValueString()
	}
	if !c.Password.IsNull() {
		m["PASSWORD"] = c.Password.ValueString()
	}
	if !c.ClientId.IsNull() {
		m["CLIENT_ID"] = c.ClientId.ValueString()
	}
	if !c.ClientSecret.IsNull() {
		m["CLIENT_SECRET"] = c.ClientSecret.ValueString()
	}
	if !c.TokenUrl.IsNull() {
		m["TOKEN_URL"] = c.TokenUrl.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.GroupNameSource.IsNull() {
		m["GROUP_NAME_SOURCE"] = c.GroupNameSource.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}

	return m
}

// WorkdayFromMap maps API configuration map to Workday configuration model.
func WorkdayFromMap(c *customtypes.ConfigurationWorkday, config map[string]interface{}) {
	if config == nil {
		return
	}

	c.AuthenticationMethod = utils.FromMapString(config, "AUTHENTICATION_METHOD")
	c.BaseUrl = utils.FromMapString(config, "BASE_URL")
	c.ScimUrl = utils.FromMapString(config, "SCIM_URL")

	c.Username = utils.FromMapString(config, "USERNAME")
	c.Password = utils.FromMapString(config, "PASSWORD")
	c.ClientId = utils.FromMapString(config, "CLIENT_ID")
	c.ClientSecret = utils.FromMapString(config, "CLIENT_SECRET")
	c.TokenUrl = utils.FromMapString(config, "TOKEN_URL")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.GroupNameSource = utils.FromMapString(config, "GROUP_NAME_SOURCE")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}

// ZoomToMap maps Zoom configuration model to API configuration map.
func ZoomToMap(c *customtypes.ConfigurationZoom) map[string]interface{} {
	m := make(map[string]interface{})

	m["SCIM_URL"] = c.ScimUrl.ValueString()

	if !c.ApiKey.IsNull() {
		m["API_KEY"] = c.ApiKey.ValueString()
	}
	if !c.ApiSecret.IsNull() {
		m["API_SECRET"] = c.ApiSecret.ValueString()
	}
	if !c.AuthenticationMethod.IsNull() {
		m["AUTHENTICATION_METHOD"] = c.AuthenticationMethod.ValueString()
	}
	if !c.CreateUsers.IsNull() {
		m["CREATE_USERS"] = c.CreateUsers.ValueBool()
	}
	if !c.DeprovisionUsers.IsNull() {
		m["DEPROVISION_USERS"] = c.DeprovisionUsers.ValueBool()
	}
	if !c.DisableUsers.IsNull() {
		m["DISABLE_USERS"] = c.DisableUsers.ValueBool()
	}
	if !c.OauthAccountId.IsNull() {
		m["OAUTH_ACCOUNT_ID"] = c.OauthAccountId.ValueString()
	}
	if !c.OauthClientId.IsNull() {
		m["OAUTH_CLIENT_ID"] = c.OauthClientId.ValueString()
	}
	if !c.OauthClientSecret.IsNull() {
		m["OAUTH_CLIENT_SECRET"] = c.OauthClientSecret.ValueString()
	}
	if !c.OauthTokenUrl.IsNull() {
		m["OAUTH_TOKEN_URL"] = c.OauthTokenUrl.ValueString()
	}
	if !c.RemoveAction.IsNull() {
		m["REMOVE_ACTION"] = c.RemoveAction.ValueString()
	}
	m["SCIM_URL"] = c.ScimUrl.ValueString()
	if !c.UpdateUsers.IsNull() {
		m["UPDATE_USERS"] = c.UpdateUsers.ValueBool()
	}
	return m
}

func ZoomFromMap(c *customtypes.ConfigurationZoom, config map[string]interface{}) {
	if config == nil {
		return
	}
	c.ScimUrl = utils.FromMapString(config, "SCIM_URL")

	c.ApiKey = utils.FromMapString(config, "API_KEY")
	c.ApiSecret = utils.FromMapString(config, "API_SECRET")
	c.AuthenticationMethod = utils.FromMapString(config, "AUTHENTICATION_METHOD")
	c.CreateUsers = utils.FromMapBool(config, "CREATE_USERS")
	c.DeprovisionUsers = utils.FromMapBool(config, "DEPROVISION_USERS")
	c.DisableUsers = utils.FromMapBool(config, "DISABLE_USERS")
	c.OauthAccountId = utils.FromMapString(config, "OAUTH_ACCOUNT_ID")
	c.OauthClientId = utils.FromMapString(config, "OAUTH_CLIENT_ID")
	c.OauthClientSecret = utils.FromMapString(config, "OAUTH_CLIENT_SECRET")
	c.OauthTokenUrl = utils.FromMapString(config, "OAUTH_TOKEN_URL")
	c.RemoveAction = utils.FromMapString(config, "REMOVE_ACTION")
	c.UpdateUsers = utils.FromMapBool(config, "UPDATE_USERS")
}
