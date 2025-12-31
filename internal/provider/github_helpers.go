package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/githubapi"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const githubPerPage = 100

func requireGitHubClient(diags *diag.Diagnostics, client *githubapi.Client) bool {
	if client == nil {
		diags.AddError(
			"Missing GitHub Configuration",
			"Configure `github_token` (or the `GITHUB_TOKEN` environment variable) to use GitHub enterprise data sources or resources.",
		)
		return false
	}
	return true
}

func githubResponseError(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("empty response")
	}
	bodyBytes, _ := utils.ReadAndRestoreResponseBody(resp)
	body := strings.TrimSpace(string(bodyBytes))
	if body == "" {
		body = resp.Status
	}
	return fmt.Errorf("%s: %s", resp.Status, body)
}

func githubResponseErrorWithHint(resp *http.Response, client *githubapi.Client) error {
	err := githubResponseError(resp)
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		return err
	}

	baseURL := "<unknown>"
	apiVersion := "<unknown>"
	if client != nil {
		if strings.TrimSpace(client.BaseURL) != "" {
			baseURL = client.BaseURL
		}
		if strings.TrimSpace(client.APIVersion) != "" {
			apiVersion = client.APIVersion
		}
	}

	return fmt.Errorf("%s (hint: verify github_token/GITHUB_TOKEN is a classic PAT with admin:enterprise; github_api_base_url=%s github_api_version=%s)", err, baseURL, apiVersion)
}

func githubDecodeMap(resp *http.Response) (map[string]interface{}, error) {
	decoded, err := utils.DecodeResponseJSON(resp)
	if err != nil {
		return nil, err
	}
	if decoded == nil {
		return nil, nil
	}

	switch v := decoded.(type) {
	case map[string]interface{}:
		return v, nil
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		if m, ok := v[0].(map[string]interface{}); ok {
			return m, nil
		}
	}

	return nil, fmt.Errorf("unexpected json response shape")
}

func githubDecodeList(resp *http.Response) ([]map[string]interface{}, error) {
	decoded, err := utils.DecodeResponseJSON(resp)
	if err != nil {
		return nil, err
	}
	if decoded == nil {
		return nil, nil
	}

	switch v := decoded.(type) {
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			out = append(out, m)
		}
		return out, nil
	case map[string]interface{}:
		return []map[string]interface{}{v}, nil
	default:
		return nil, fmt.Errorf("unexpected json response shape")
	}
}

func githubListAll(ctx context.Context, client *githubapi.Client, path string) ([]map[string]interface{}, error) {
	var out []map[string]interface{}
	page := 1
	for {
		query := url.Values{}
		query.Set("per_page", strconv.Itoa(githubPerPage))
		query.Set("page", strconv.Itoa(page))

		resp, err := client.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 300 {
			return nil, githubResponseErrorWithHint(resp, client)
		}

		items, err := githubDecodeList(resp)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}
		out = append(out, items...)
		if len(items) < githubPerPage {
			break
		}
		page++
	}
	return out, nil
}

func normalizeStrings(values []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, raw := range values {
		val := strings.TrimSpace(raw)
		if val == "" {
			continue
		}
		key := strings.ToLower(val)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, val)
	}
	sort.Strings(out)
	return out
}

func diffCaseInsensitive(desired []string, existing []string) ([]string, []string) {
	desiredMap := make(map[string]string)
	for _, raw := range desired {
		val := strings.TrimSpace(raw)
		if val == "" {
			continue
		}
		key := strings.ToLower(val)
		if _, ok := desiredMap[key]; ok {
			continue
		}
		desiredMap[key] = val
	}

	existingMap := make(map[string]string)
	for _, raw := range existing {
		val := strings.TrimSpace(raw)
		if val == "" {
			continue
		}
		key := strings.ToLower(val)
		if _, ok := existingMap[key]; ok {
			continue
		}
		existingMap[key] = val
	}

	var toAdd []string
	for key, val := range desiredMap {
		if _, ok := existingMap[key]; !ok {
			toAdd = append(toAdd, val)
		}
	}

	var toRemove []string
	for key, val := range existingMap {
		if _, ok := desiredMap[key]; !ok {
			toRemove = append(toRemove, val)
		}
	}

	sort.Strings(toAdd)
	sort.Strings(toRemove)
	return toAdd, toRemove
}

func setToStrings(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return nil, diags
	}
	var values []string
	diags.Append(set.ElementsAs(ctx, &values, false)...)
	return values, diags
}

func stringsToSet(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	normalized := normalizeStrings(values)
	return types.SetValueFrom(ctx, types.StringType, normalized)
}

func mapString(m map[string]interface{}, key string) string {
	raw, ok := m[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func enterpriseTeamsPath(enterprise string) string {
	return fmt.Sprintf("/enterprises/%s/teams", url.PathEscape(strings.TrimSpace(enterprise)))
}

func enterpriseTeamPath(enterprise string, teamSlug string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)))
}

func enterpriseTeamOrganizationsPath(enterprise string, teamSlug string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s/organizations", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)))
}

func enterpriseTeamOrganizationsBulkAddPath(enterprise string, teamSlug string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s/organizations/add", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)))
}

func enterpriseTeamOrganizationsBulkRemovePath(enterprise string, teamSlug string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s/organizations/remove", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)))
}

func enterpriseTeamOrganizationPath(enterprise string, teamSlug string, org string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s/organizations/%s", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)), url.PathEscape(strings.TrimSpace(org)))
}

func enterpriseTeamMembersPath(enterprise string, teamSlug string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s/memberships", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)))
}

func enterpriseTeamMemberPath(enterprise string, teamSlug string, username string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s/memberships/%s", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)), url.PathEscape(strings.TrimSpace(username)))
}

func enterpriseTeamMembersBulkAddPath(enterprise string, teamSlug string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s/memberships/add", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)))
}

func enterpriseTeamMembersBulkRemovePath(enterprise string, teamSlug string) string {
	return fmt.Sprintf("/enterprises/%s/teams/%s/memberships/remove", url.PathEscape(strings.TrimSpace(enterprise)), url.PathEscape(strings.TrimSpace(teamSlug)))
}

func listEnterpriseTeamOrganizations(ctx context.Context, client *githubapi.Client, enterprise string, teamSlug string) ([]string, error) {
	items, err := githubListAll(ctx, client, enterpriseTeamOrganizationsPath(enterprise, teamSlug))
	if err != nil {
		return nil, err
	}
	var orgs []string
	for _, item := range items {
		if login := mapString(item, "login"); login != "" {
			orgs = append(orgs, login)
		}
	}
	return normalizeStrings(orgs), nil
}

func listEnterpriseTeamMembers(ctx context.Context, client *githubapi.Client, enterprise string, teamSlug string) ([]string, error) {
	items, err := githubListAll(ctx, client, enterpriseTeamMembersPath(enterprise, teamSlug))
	if err != nil {
		return nil, err
	}
	var users []string
	for _, item := range items {
		if login := mapString(item, "login"); login != "" {
			users = append(users, login)
		}
	}
	return normalizeStrings(users), nil
}
