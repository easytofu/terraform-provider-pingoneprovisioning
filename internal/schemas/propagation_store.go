package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
)

// Helper function to define an attribute that is Optional in Resource and Computed in DataSource
func optionalOrComputedString(isDataSource bool, sensitive bool) schema.Attribute {
	if isDataSource {
		return schema.StringAttribute{Computed: true, Sensitive: sensitive}
	}
	return schema.StringAttribute{Optional: true, Sensitive: sensitive}
}

func optionalOrComputedBool(isDataSource bool, defaultValue bool) schema.Attribute {
	if isDataSource {
		return schema.BoolAttribute{Computed: true}
	}
	return schema.BoolAttribute{Optional: true, Default: booldefault.StaticBool(defaultValue), Computed: true}
}

// CHANGED: For Resources, we strictly use Optional: true to avoid validation errors
// when the parent block is missing. Logic validation should handle missing required fields if the block is present.
func requiredOrComputedString(isDataSource bool, sensitive bool) schema.Attribute {
	if isDataSource {
		return schema.StringAttribute{Computed: true, Sensitive: sensitive}
	}
	return schema.StringAttribute{Optional: true, Sensitive: sensitive}
}

func requiredOrComputedBool(isDataSource bool) schema.Attribute {
	if isDataSource {
		return schema.BoolAttribute{Computed: true}
	}
	return schema.BoolAttribute{Optional: true}
}

// Aquera
func AqueraConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"authentication_method": requiredOrComputedString(isDataSource, false),
			"base_url":              requiredOrComputedString(isDataSource, false),
			"api_key":               optionalOrComputedString(isDataSource, true),
			"api_secret":            optionalOrComputedString(isDataSource, true),
			"bearer_token":          optionalOrComputedString(isDataSource, true),
			"domain":                optionalOrComputedString(isDataSource, false),
			"username":              optionalOrComputedString(isDataSource, false),
			"password":              optionalOrComputedString(isDataSource, true),
			"create_users":          optionalOrComputedBool(isDataSource, true),
			"disable_users":         optionalOrComputedBool(isDataSource, true),
			"group_name_source":     optionalOrComputedString(isDataSource, false),
			"remove_action":         optionalOrComputedString(isDataSource, false),
			"update_users":          optionalOrComputedBool(isDataSource, true),
		},
	}
}

// AzureADSAMLV2
func AzureAdSamlV2ConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"base_url":          requiredOrComputedString(isDataSource, false),
			"scim_url":          requiredOrComputedString(isDataSource, false),
			"bearer_token":      optionalOrComputedString(isDataSource, true),
			"group_name_source": optionalOrComputedString(isDataSource, false),
			"create_users":      optionalOrComputedBool(isDataSource, true),
			"deprovision_users": optionalOrComputedBool(isDataSource, true),
			"disable_users":     optionalOrComputedBool(isDataSource, true),
			"remove_action":     optionalOrComputedString(isDataSource, false),
			"update_users":      optionalOrComputedBool(isDataSource, true),
		},
	}
}

// GithubEmu
func GithubEmuConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"base_url":           requiredOrComputedString(isDataSource, false),
			"oauth_access_token": optionalOrComputedString(isDataSource, true),
			"create_users":       optionalOrComputedBool(isDataSource, true),
			"deprovision_users":  optionalOrComputedBool(isDataSource, true),
			"remove_action":      optionalOrComputedString(isDataSource, false),
			"update_users":       optionalOrComputedBool(isDataSource, true),
		},
	}
}

// GoogleApps
func GoogleAppsConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"authentication_method": requiredOrComputedString(isDataSource, false),
			"base_url":              requiredOrComputedString(isDataSource, false),
			"oauth_client_id":       optionalOrComputedString(isDataSource, false),
			"oauth_client_secret":   optionalOrComputedString(isDataSource, true),
			"oauth_refresh_token":   optionalOrComputedString(isDataSource, true),
			"oauth_token_url":       optionalOrComputedString(isDataSource, false),
			"domain":                optionalOrComputedString(isDataSource, false),
			"create_users":          optionalOrComputedBool(isDataSource, true),
			"deprovision_users":     optionalOrComputedBool(isDataSource, true),
			"disable_users":         optionalOrComputedBool(isDataSource, true),
			"group_name_source":     optionalOrComputedString(isDataSource, false),
			"remove_action":         optionalOrComputedString(isDataSource, false),
			"update_users":          optionalOrComputedBool(isDataSource, true),
		},
	}
}

// LDAPGateway
func LdapGatewayConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"authentication_method": requiredOrComputedString(isDataSource, false),
			"base_url":              requiredOrComputedString(isDataSource, false),
			"ldap_gateway_id":       requiredOrComputedString(isDataSource, false),
			"ldap_gateway_region":   requiredOrComputedString(isDataSource, false),
			"api_key":               optionalOrComputedString(isDataSource, true),
			"api_secret":            optionalOrComputedString(isDataSource, true),
			"bearer_token":          optionalOrComputedString(isDataSource, true),
			"username":              optionalOrComputedString(isDataSource, false),
			"password":              optionalOrComputedString(isDataSource, true),
			"create_users":          optionalOrComputedBool(isDataSource, true),
			"deprovision_users":     optionalOrComputedBool(isDataSource, true),
			"disable_users":         optionalOrComputedBool(isDataSource, true),
			"group_name_source":     optionalOrComputedString(isDataSource, false),
			"remove_action":         optionalOrComputedString(isDataSource, false),
			"update_users":          optionalOrComputedBool(isDataSource, true),
		},
	}
}

// PingOne
func PingOneConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"authentication_method": requiredOrComputedString(isDataSource, false),
			"base_url":              requiredOrComputedString(isDataSource, false),
			"scim_url":              requiredOrComputedString(isDataSource, false),
			"bearer_token":          optionalOrComputedString(isDataSource, true),
			"oauth_client_id":       optionalOrComputedString(isDataSource, false),
			"oauth_client_secret":   optionalOrComputedString(isDataSource, true),
			"oauth_token_url":       optionalOrComputedString(isDataSource, false),
			"oauth_refresh_token":   optionalOrComputedString(isDataSource, true),
			"domain":                optionalOrComputedString(isDataSource, false),
			"username":              optionalOrComputedString(isDataSource, false),
			"password":              optionalOrComputedString(isDataSource, true),
			"create_users":          optionalOrComputedBool(isDataSource, true),
			"deprovision_users":     optionalOrComputedBool(isDataSource, true),
			"disable_users":         optionalOrComputedBool(isDataSource, true),
			"group_name_source":     optionalOrComputedString(isDataSource, false),
			"remove_action":         optionalOrComputedString(isDataSource, false),
			"update_users":          optionalOrComputedBool(isDataSource, true),
		},
	}
}

// Salesforce
func SalesforceConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"authentication_method": requiredOrComputedString(isDataSource, false),
			"base_url":              requiredOrComputedString(isDataSource, false),
			"scim_url":              requiredOrComputedString(isDataSource, false),
			"consumer_key":          optionalOrComputedString(isDataSource, false),
			"consumer_secret":       optionalOrComputedString(isDataSource, true),
			"username":              optionalOrComputedString(isDataSource, false),
			"password":              optionalOrComputedString(isDataSource, true),
			"security_token":        optionalOrComputedString(isDataSource, true),
			"bearer_token":          optionalOrComputedString(isDataSource, true),
			"create_users":          optionalOrComputedBool(isDataSource, true),
			"deprovision_users":     optionalOrComputedBool(isDataSource, true),
			"disable_users":         optionalOrComputedBool(isDataSource, true),
			"group_name_source":     optionalOrComputedString(isDataSource, false),
			"remove_action":         optionalOrComputedString(isDataSource, false),
			"update_users":          optionalOrComputedBool(isDataSource, true),
			"record_type":           optionalOrComputedString(isDataSource, false),
		},
	}
}

// SalesforceContacts
func SalesforceContactsConfigSchema(isDataSource bool) schema.Block {
	return SalesforceConfigSchema(isDataSource)
}

// SCIM
func ScimConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: scimConfigAttributes(isDataSource),
	}
}

// ScimConfigAttributeSchema defines the SCIM configuration as an attribute (HCL assignment syntax),
// for compatibility with configurations that use `scim_configuration = { ... }`.
func ScimConfigAttributeSchema(isDataSource bool) schema.Attribute {
	if isDataSource {
		return schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: scimConfigAttributes(isDataSource),
		}
	}

	return schema.SingleNestedAttribute{
		Optional:   true,
		Attributes: scimConfigAttributes(isDataSource),
	}
}

func scimConfigAttributes(isDataSource bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"authentication_method":  requiredOrComputedString(isDataSource, false),
		"authorization_type":     requiredOrComputedString(isDataSource, false),
		"basic_auth_user":        optionalOrComputedString(isDataSource, false),
		"basic_auth_password":    optionalOrComputedString(isDataSource, true),
		"create_users":           optionalOrComputedBool(isDataSource, true),
		"disable_users":          optionalOrComputedBool(isDataSource, true),
		"group_name_source":      optionalOrComputedString(isDataSource, false),
		"groups_resource":        optionalOrComputedString(isDataSource, false),
		"oauth_access_token":     optionalOrComputedString(isDataSource, true),
		"oauth_client_id":        optionalOrComputedString(isDataSource, true),
		"oauth_client_secret":    optionalOrComputedString(isDataSource, true),
		"oauth_token_request":    optionalOrComputedString(isDataSource, false),
		"remove_action":          optionalOrComputedString(isDataSource, false),
		"scim_url":               requiredOrComputedString(isDataSource, false),
		"scim_version":           requiredOrComputedString(isDataSource, false),
		"unique_user_identifier": requiredOrComputedString(isDataSource, false),
		"update_users":           optionalOrComputedBool(isDataSource, true),
		"user_filter":            requiredOrComputedString(isDataSource, false),
		"users_resource":         requiredOrComputedString(isDataSource, false),
	}
}

// ServiceNow
func ServiceNowConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"base_url":          requiredOrComputedString(isDataSource, false),
			"scim_url":          requiredOrComputedString(isDataSource, false),
			"username":          optionalOrComputedString(isDataSource, false),
			"password":          optionalOrComputedString(isDataSource, true),
			"create_users":      optionalOrComputedBool(isDataSource, true),
			"deprovision_users": optionalOrComputedBool(isDataSource, true),
			"disable_users":     optionalOrComputedBool(isDataSource, true),
			"group_name_source": optionalOrComputedString(isDataSource, false),
			"remove_action":     optionalOrComputedString(isDataSource, false),
			"update_users":      optionalOrComputedBool(isDataSource, true),
		},
	}
}

// Slack
func SlackConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"base_url":          requiredOrComputedString(isDataSource, false),
			"scim_url":          requiredOrComputedString(isDataSource, false),
			"bearer_token":      optionalOrComputedString(isDataSource, true),
			"create_users":      optionalOrComputedBool(isDataSource, true),
			"deprovision_users": optionalOrComputedBool(isDataSource, true),
			"disable_users":     optionalOrComputedBool(isDataSource, true),
			"group_name_source": optionalOrComputedString(isDataSource, false),
			"remove_action":     optionalOrComputedString(isDataSource, false),
			"update_users":      optionalOrComputedBool(isDataSource, true),
		},
	}
}

// Workday
func WorkdayConfigSchema(isDataSource bool) schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"authentication_method": requiredOrComputedString(isDataSource, false),
			"base_url":              requiredOrComputedString(isDataSource, false),
			"scim_url":              requiredOrComputedString(isDataSource, false),
			"username":              optionalOrComputedString(isDataSource, false),
			"password":              optionalOrComputedString(isDataSource, true),
			"client_id":             optionalOrComputedString(isDataSource, false),
			"client_secret":         optionalOrComputedString(isDataSource, true),
			"token_url":             optionalOrComputedString(isDataSource, false),
			"create_users":          optionalOrComputedBool(isDataSource, true),
			"deprovision_users":     optionalOrComputedBool(isDataSource, true),
			"disable_users":         optionalOrComputedBool(isDataSource, true),
			"group_name_source":     optionalOrComputedString(isDataSource, false),
			"remove_action":         optionalOrComputedString(isDataSource, false),
			"update_users":          optionalOrComputedBool(isDataSource, true),
		},
	}
}

// Zoom
func ZoomConfigSchema(isDataSource bool) schema.Block {
	attrs := map[string]schema.Attribute{
		"api_key":             optionalOrComputedString(isDataSource, true),
		"api_secret":          optionalOrComputedString(isDataSource, true),
		"create_users":        optionalOrComputedBool(isDataSource, true),
		"deprovision_users":   optionalOrComputedBool(isDataSource, true),
		"disable_users":       optionalOrComputedBool(isDataSource, true),
		"oauth_account_id":    optionalOrComputedString(isDataSource, false),
		"oauth_client_id":     optionalOrComputedString(isDataSource, false),
		"oauth_client_secret": optionalOrComputedString(isDataSource, true),
		"oauth_token_url":     optionalOrComputedString(isDataSource, false),
		"remove_action":       optionalOrComputedString(isDataSource, false),
		"scim_url":            requiredOrComputedString(isDataSource, false),
		"update_users":        optionalOrComputedBool(isDataSource, true),
	}

	if isDataSource {
		attrs["authentication_method"] = schema.StringAttribute{Computed: true}
	} else {
		attrs["authentication_method"] = schema.StringAttribute{
			Optional: true,
			Default:  stringdefault.StaticString("JWT Bearer Token"),
			Computed: true,
		}
	}

	return schema.SingleNestedBlock{
		Attributes: attrs,
	}
}
