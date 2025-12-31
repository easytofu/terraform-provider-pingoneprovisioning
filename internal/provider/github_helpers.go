package provider

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func requireGitHubClient(diags *diag.Diagnostics, client *client.GitHubClient) bool {
	if client == nil {
		diags.AddError(
			"Missing GitHub Configuration",
			"Configure `github_token` (or the `GITHUB_TOKEN` environment variable) to use GitHub SCIM data sources.",
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

func githubResponseErrorWithHint(resp *http.Response, client *client.GitHubClient) error {
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
