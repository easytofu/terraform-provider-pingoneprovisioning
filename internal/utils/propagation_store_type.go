// pingone/providers/pingone-propagation/internal/utils/propagation_store_type.go
package utils

import "strings"

// NormalizePropagationStoreTypeForAPI converts Terraform-facing propagation store type values
// into the exact type string expected by the PingOne Management API.
//
func NormalizePropagationStoreTypeForAPI(tfType string) string {
	// Trim whitespace defensively, then normalize case-sensitive enum values.
	t := strings.TrimSpace(tfType)

	if t == "" {
		return t
	}

	// Azure AD SAML v2
	if strings.EqualFold(t, "AzureADSAMLV2") || strings.EqualFold(t, "AzureActiveDirectorySAML2") {
		return "AzureActiveDirectorySAML2"
	}

	// SCIM
	if strings.EqualFold(t, "SCIM") || strings.EqualFold(t, "scim") {
		return "scim"
	}

	// LDAP Gateway
	if strings.EqualFold(t, "LDAPGateway") || strings.EqualFold(t, "LdapGateway") {
		return "LdapGateway"
	}

	return t
}

// NormalizePropagationStoreTypeForTerraform converts API propagation store type strings into the
// value that should be stored in Terraform state.
//
// If preferredTFType is provided (for example, from the planned/configured value), the returned
// value will match that spelling to avoid perpetual diffs.
func NormalizePropagationStoreTypeForTerraform(apiType string, preferredTFType string) string {
	apiT := strings.TrimSpace(apiType)
	pref := strings.TrimSpace(preferredTFType)

	if apiT == "" {
		return apiT
	}

	// Azure AD SAML v2
	if strings.EqualFold(apiT, "AzureActiveDirectorySAML2") || strings.EqualFold(apiT, "AzureADSAMLV2") {
		return "AzureADSAMLV2"
	}

	// SCIM
	if strings.EqualFold(apiT, "scim") || strings.EqualFold(apiT, "SCIM") {
		// Preserve configured spelling/casing when the configuration opted into a
		// non-default representation (e.g. `scim`).
		if pref != "" && strings.EqualFold(pref, "SCIM") {
			return pref
		}
		return "SCIM"
	}

	// LDAP Gateway
	if strings.EqualFold(apiT, "LdapGateway") || strings.EqualFold(apiT, "LDAPGateway") {
		if strings.EqualFold(pref, "LDAPGateway") {
			return "LDAPGateway"
		}
		return "LDAPGateway"
	}

	// Preserve configured casing/spelling when the API is case-insensitive for a value.
	if pref != "" && strings.EqualFold(apiT, pref) {
		return pref
	}

	return apiT
}
