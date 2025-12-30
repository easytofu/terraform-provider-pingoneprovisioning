package types

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PropagationStoreModel describes the shared resource data model.
type PropagationStoreModel struct {
	Id                              types.String                     `tfsdk:"id"`
	EnvironmentId                   types.String                     `tfsdk:"environment_id"`
	Name                            types.String                     `tfsdk:"name"`
	Description                     types.String                     `tfsdk:"description"`
	Type                            types.String                     `tfsdk:"type"`
	ImageId                         types.String                     `tfsdk:"image_id"`
	ImageHref                       types.String                     `tfsdk:"image_href"`
	Managed                         types.Bool                       `tfsdk:"managed"`
	Status                          types.String                     `tfsdk:"status"`
	SyncStatus                      types.Object                     `tfsdk:"sync_status"`
	ConfigurationAquera             *ConfigurationAquera             `tfsdk:"configuration_aquera"`
	ConfigurationAzureAdSamlV2      *ConfigurationAzureAdSamlV2      `tfsdk:"configuration_azure_ad_saml_v2"`
	ConfigurationGoogleApps         *ConfigurationGoogleApps         `tfsdk:"configuration_google_apps"`
	ConfigurationLdapGateway        *ConfigurationLdapGateway        `tfsdk:"configuration_ldap_gateway"`
	ConfigurationPingOne            *ConfigurationPingOne            `tfsdk:"configuration_ping_one"`
	ConfigurationSalesforce         *ConfigurationSalesforce         `tfsdk:"configuration_salesforce"`
	ConfigurationSalesforceContacts *ConfigurationSalesforceContacts `tfsdk:"configuration_salesforce_contacts"`
	ConfigurationScim               *ConfigurationScim               `tfsdk:"configuration_scim"`
	ScimConfiguration               *ConfigurationScim               `tfsdk:"scim_configuration"`
	ConfigurationServiceNow         *ConfigurationServiceNow         `tfsdk:"configuration_service_now"`
	ConfigurationSlack              *ConfigurationSlack              `tfsdk:"configuration_slack"`
	ConfigurationWorkday            *ConfigurationWorkday            `tfsdk:"configuration_workday"`
	ConfigurationZoom               *ConfigurationZoom               `tfsdk:"configuration_zoom"`
}

var SyncStatusAttrTypes = map[string]attr.Type{
	"last_sync_time": types.StringType,
	"next_sync_time": types.StringType,
	"status":         types.StringType,
	"details":        types.StringType,
}

// --- Configuration Structs ---

type ConfigurationAquera struct {
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	BaseUrl              types.String `tfsdk:"base_url"`
	ApiKey               types.String `tfsdk:"api_key"`
	ApiSecret            types.String `tfsdk:"api_secret"`
	BearerToken          types.String `tfsdk:"bearer_token"`
	Domain               types.String `tfsdk:"domain"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	GroupNameSource      types.String `tfsdk:"group_name_source"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
}

type ConfigurationAzureAdSamlV2 struct {
	BaseUrl          types.String `tfsdk:"base_url"`
	ScimUrl          types.String `tfsdk:"scim_url"`
	BearerToken      types.String `tfsdk:"bearer_token"`
	GroupNameSource  types.String `tfsdk:"group_name_source"`
	CreateUsers      types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers     types.Bool   `tfsdk:"disable_users"`
	RemoveAction     types.String `tfsdk:"remove_action"`
	UpdateUsers      types.Bool   `tfsdk:"update_users"`
}

type ConfigurationGoogleApps struct {
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	BaseUrl              types.String `tfsdk:"base_url"`
	OauthClientId        types.String `tfsdk:"oauth_client_id"`
	OauthClientSecret    types.String `tfsdk:"oauth_client_secret"`
	OauthRefreshToken    types.String `tfsdk:"oauth_refresh_token"`
	OauthTokenUrl        types.String `tfsdk:"oauth_token_url"`
	Domain               types.String `tfsdk:"domain"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers     types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	GroupNameSource      types.String `tfsdk:"group_name_source"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
}

type ConfigurationLdapGateway struct {
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	BaseUrl              types.String `tfsdk:"base_url"`
	LdapGatewayId        types.String `tfsdk:"ldap_gateway_id"`
	LdapGatewayRegion    types.String `tfsdk:"ldap_gateway_region"`
	ApiKey               types.String `tfsdk:"api_key"`
	ApiSecret            types.String `tfsdk:"api_secret"`
	BearerToken          types.String `tfsdk:"bearer_token"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers     types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	GroupNameSource      types.String `tfsdk:"group_name_source"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
}

type ConfigurationPingOne struct {
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	BaseUrl              types.String `tfsdk:"base_url"`
	ScimUrl              types.String `tfsdk:"scim_url"`
	BearerToken          types.String `tfsdk:"bearer_token"`
	OauthClientId        types.String `tfsdk:"oauth_client_id"`
	OauthClientSecret    types.String `tfsdk:"oauth_client_secret"`
	OauthTokenUrl        types.String `tfsdk:"oauth_token_url"`
	OauthRefreshToken    types.String `tfsdk:"oauth_refresh_token"`
	Domain               types.String `tfsdk:"domain"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers     types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	GroupNameSource      types.String `tfsdk:"group_name_source"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
}

type ConfigurationSalesforce struct {
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	BaseUrl              types.String `tfsdk:"base_url"`
	ScimUrl              types.String `tfsdk:"scim_url"`
	ConsumerKey          types.String `tfsdk:"consumer_key"`
	ConsumerSecret       types.String `tfsdk:"consumer_secret"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	SecurityToken        types.String `tfsdk:"security_token"`
	BearerToken          types.String `tfsdk:"bearer_token"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers     types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	GroupNameSource      types.String `tfsdk:"group_name_source"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
	RecordType           types.String `tfsdk:"record_type"`
}

type ConfigurationSalesforceContacts struct {
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	BaseUrl              types.String `tfsdk:"base_url"`
	ScimUrl              types.String `tfsdk:"scim_url"`
	ConsumerKey          types.String `tfsdk:"consumer_key"`
	ConsumerSecret       types.String `tfsdk:"consumer_secret"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	SecurityToken        types.String `tfsdk:"security_token"`
	BearerToken          types.String `tfsdk:"bearer_token"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers     types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	GroupNameSource      types.String `tfsdk:"group_name_source"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
	RecordType           types.String `tfsdk:"record_type"`
}

type ConfigurationScim struct {
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	AuthorizationType    types.String `tfsdk:"authorization_type"`
	BasicAuthUser        types.String `tfsdk:"basic_auth_user"`
	BasicAuthPassword    types.String `tfsdk:"basic_auth_password"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	GroupNameSource      types.String `tfsdk:"group_name_source"`
	GroupsResource       types.String `tfsdk:"groups_resource"`
	OauthAccessToken     types.String `tfsdk:"oauth_access_token"`
	OauthClientId        types.String `tfsdk:"oauth_client_id"`
	OauthClientSecret    types.String `tfsdk:"oauth_client_secret"`
	OauthTokenRequest    types.String `tfsdk:"oauth_token_request"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	ScimUrl              types.String `tfsdk:"scim_url"`
	ScimVersion          types.String `tfsdk:"scim_version"`
	UniqueUserIdentifier types.String `tfsdk:"unique_user_identifier"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
	UserFilter           types.String `tfsdk:"user_filter"`
	UsersResource        types.String `tfsdk:"users_resource"`
}

type ConfigurationServiceNow struct {
	BaseUrl          types.String `tfsdk:"base_url"`
	ScimUrl          types.String `tfsdk:"scim_url"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	CreateUsers      types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers     types.Bool   `tfsdk:"disable_users"`
	GroupNameSource  types.String `tfsdk:"group_name_source"`
	RemoveAction     types.String `tfsdk:"remove_action"`
	UpdateUsers      types.Bool   `tfsdk:"update_users"`
}

type ConfigurationSlack struct {
	BaseUrl          types.String `tfsdk:"base_url"`
	ScimUrl          types.String `tfsdk:"scim_url"`
	BearerToken      types.String `tfsdk:"bearer_token"`
	CreateUsers      types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers     types.Bool   `tfsdk:"disable_users"`
	GroupNameSource  types.String `tfsdk:"group_name_source"`
	RemoveAction     types.String `tfsdk:"remove_action"`
	UpdateUsers      types.Bool   `tfsdk:"update_users"`
}

type ConfigurationWorkday struct {
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	BaseUrl              types.String `tfsdk:"base_url"`
	ScimUrl              types.String `tfsdk:"scim_url"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	ClientId             types.String `tfsdk:"client_id"`
	ClientSecret         types.String `tfsdk:"client_secret"`
	TokenUrl             types.String `tfsdk:"token_url"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers     types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	GroupNameSource      types.String `tfsdk:"group_name_source"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
}

type ConfigurationZoom struct {
	ApiKey               types.String `tfsdk:"api_key"`
	ApiSecret            types.String `tfsdk:"api_secret"`
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	CreateUsers          types.Bool   `tfsdk:"create_users"`
	DeprovisionUsers     types.Bool   `tfsdk:"deprovision_users"`
	DisableUsers         types.Bool   `tfsdk:"disable_users"`
	OauthAccountId       types.String `tfsdk:"oauth_account_id"`
	OauthClientId        types.String `tfsdk:"oauth_client_id"`
	OauthClientSecret    types.String `tfsdk:"oauth_client_secret"`
	OauthTokenUrl        types.String `tfsdk:"oauth_token_url"`
	RemoveAction         types.String `tfsdk:"remove_action"`
	ScimUrl              types.String `tfsdk:"scim_url"`
	UpdateUsers          types.Bool   `tfsdk:"update_users"`
}

// PropagationStoreModelType returns the Terraform type definition for propagation store models.
func PropagationStoreModelType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":                                types.StringType,
			"environment_id":                    types.StringType,
			"name":                              types.StringType,
			"description":                       types.StringType,
			"type":                              types.StringType,
			"image_id":                          types.StringType,
			"image_href":                        types.StringType,
			"managed":                           types.BoolType,
			"status":                            types.StringType,
			"sync_status":                       types.ObjectType{AttrTypes: SyncStatusAttrTypes},
			"configuration_aquera":              configurationAqueraAttrType(),
			"configuration_azure_ad_saml_v2":    configurationAzureADSAMLAttrType(),
			"configuration_google_apps":         configurationGoogleAppsAttrType(),
			"configuration_ldap_gateway":        configurationLdapGatewayAttrType(),
			"configuration_ping_one":            configurationPingOneAttrType(),
			"configuration_salesforce":          configurationSalesforceAttrType(),
			"configuration_salesforce_contacts": configurationSalesforceAttrType(),
			"configuration_scim":                configurationScimAttrType(),
			"scim_configuration":                configurationScimAttrType(),
			"configuration_service_now":         configurationServiceNowAttrType(),
			"configuration_slack":               configurationSlackAttrType(),
			"configuration_workday":             configurationWorkdayAttrType(),
			"configuration_zoom":                configurationZoomAttrType(),
		},
	}
}

func configurationAqueraAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"authentication_method": types.StringType,
			"base_url":              types.StringType,
			"api_key":               types.StringType,
			"api_secret":            types.StringType,
			"bearer_token":          types.StringType,
			"domain":                types.StringType,
			"username":              types.StringType,
			"password":              types.StringType,
			"create_users":          types.BoolType,
			"disable_users":         types.BoolType,
			"group_name_source":     types.StringType,
			"remove_action":         types.StringType,
			"update_users":          types.BoolType,
		},
	}
}

func configurationAzureADSAMLAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"base_url":          types.StringType,
			"scim_url":          types.StringType,
			"bearer_token":      types.StringType,
			"group_name_source": types.StringType,
			"create_users":      types.BoolType,
			"deprovision_users": types.BoolType,
			"disable_users":     types.BoolType,
			"remove_action":     types.StringType,
			"update_users":      types.BoolType,
		},
	}
}

func configurationGoogleAppsAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"authentication_method": types.StringType,
			"base_url":              types.StringType,
			"oauth_client_id":       types.StringType,
			"oauth_client_secret":   types.StringType,
			"oauth_refresh_token":   types.StringType,
			"oauth_token_url":       types.StringType,
			"domain":                types.StringType,
			"create_users":          types.BoolType,
			"deprovision_users":     types.BoolType,
			"disable_users":         types.BoolType,
			"group_name_source":     types.StringType,
			"remove_action":         types.StringType,
			"update_users":          types.BoolType,
		},
	}
}

func configurationLdapGatewayAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"authentication_method": types.StringType,
			"base_url":              types.StringType,
			"ldap_gateway_id":       types.StringType,
			"ldap_gateway_region":   types.StringType,
			"api_key":               types.StringType,
			"api_secret":            types.StringType,
			"bearer_token":          types.StringType,
			"username":              types.StringType,
			"password":              types.StringType,
			"create_users":          types.BoolType,
			"deprovision_users":     types.BoolType,
			"disable_users":         types.BoolType,
			"group_name_source":     types.StringType,
			"remove_action":         types.StringType,
			"update_users":          types.BoolType,
		},
	}
}

func configurationPingOneAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"authentication_method": types.StringType,
			"base_url":              types.StringType,
			"scim_url":              types.StringType,
			"bearer_token":          types.StringType,
			"oauth_client_id":       types.StringType,
			"oauth_client_secret":   types.StringType,
			"oauth_token_url":       types.StringType,
			"oauth_refresh_token":   types.StringType,
			"domain":                types.StringType,
			"username":              types.StringType,
			"password":              types.StringType,
			"create_users":          types.BoolType,
			"deprovision_users":     types.BoolType,
			"disable_users":         types.BoolType,
			"group_name_source":     types.StringType,
			"remove_action":         types.StringType,
			"update_users":          types.BoolType,
		},
	}
}

func configurationSalesforceAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"authentication_method": types.StringType,
			"base_url":              types.StringType,
			"scim_url":              types.StringType,
			"consumer_key":          types.StringType,
			"consumer_secret":       types.StringType,
			"username":              types.StringType,
			"password":              types.StringType,
			"security_token":        types.StringType,
			"bearer_token":          types.StringType,
			"create_users":          types.BoolType,
			"deprovision_users":     types.BoolType,
			"disable_users":         types.BoolType,
			"group_name_source":     types.StringType,
			"remove_action":         types.StringType,
			"update_users":          types.BoolType,
			"record_type":           types.StringType,
		},
	}
}

func configurationScimAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"authentication_method":  types.StringType,
			"authorization_type":     types.StringType,
			"basic_auth_user":        types.StringType,
			"basic_auth_password":    types.StringType,
			"create_users":           types.BoolType,
			"disable_users":          types.BoolType,
			"group_name_source":      types.StringType,
			"groups_resource":        types.StringType,
			"oauth_access_token":     types.StringType,
			"oauth_client_id":        types.StringType,
			"oauth_client_secret":    types.StringType,
			"oauth_token_request":    types.StringType,
			"remove_action":          types.StringType,
			"scim_url":               types.StringType,
			"scim_version":           types.StringType,
			"unique_user_identifier": types.StringType,
			"update_users":           types.BoolType,
			"user_filter":            types.StringType,
			"users_resource":         types.StringType,
		},
	}
}

func configurationServiceNowAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"base_url":          types.StringType,
			"scim_url":          types.StringType,
			"username":          types.StringType,
			"password":          types.StringType,
			"create_users":      types.BoolType,
			"deprovision_users": types.BoolType,
			"disable_users":     types.BoolType,
			"group_name_source": types.StringType,
			"remove_action":     types.StringType,
			"update_users":      types.BoolType,
		},
	}
}

func configurationSlackAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"base_url":          types.StringType,
			"scim_url":          types.StringType,
			"bearer_token":      types.StringType,
			"create_users":      types.BoolType,
			"deprovision_users": types.BoolType,
			"disable_users":     types.BoolType,
			"group_name_source": types.StringType,
			"remove_action":     types.StringType,
			"update_users":      types.BoolType,
		},
	}
}

func configurationWorkdayAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"authentication_method": types.StringType,
			"base_url":              types.StringType,
			"scim_url":              types.StringType,
			"username":              types.StringType,
			"password":              types.StringType,
			"client_id":             types.StringType,
			"client_secret":         types.StringType,
			"token_url":             types.StringType,
			"create_users":          types.BoolType,
			"deprovision_users":     types.BoolType,
			"disable_users":         types.BoolType,
			"group_name_source":     types.StringType,
			"remove_action":         types.StringType,
			"update_users":          types.BoolType,
		},
	}
}

func configurationZoomAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"api_key":               types.StringType,
			"api_secret":            types.StringType,
			"authentication_method": types.StringType,
			"create_users":          types.BoolType,
			"deprovision_users":     types.BoolType,
			"disable_users":         types.BoolType,
			"oauth_account_id":      types.StringType,
			"oauth_client_id":       types.StringType,
			"oauth_client_secret":   types.StringType,
			"oauth_token_url":       types.StringType,
			"remove_action":         types.StringType,
			"scim_url":              types.StringType,
			"update_users":          types.BoolType,
		},
	}
}
