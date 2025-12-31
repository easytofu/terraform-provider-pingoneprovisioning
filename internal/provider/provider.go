package provider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/client"
	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/githubapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &PingOneProvisioningProvider{}
)

// PingOneProvisioningProvider is the provider implementation.
type PingOneProvisioningProvider struct {
	Version string
}

// PingOneProvisioningProviderModel describes the provider data model.
type PingOneProvisioningProviderModel struct {
	ClientId         types.String `tfsdk:"client_id"`
	ClientSecret     types.String `tfsdk:"client_secret"`
	EnvironmentId    types.String `tfsdk:"environment_id"`
	Region           types.String `tfsdk:"region"`
	OauthTokenURL    types.String `tfsdk:"oauth_token_url"`
	APIBaseURL       types.String `tfsdk:"api_base_url"`
	GithubToken      types.String `tfsdk:"github_token"`
	GithubAPIBaseURL types.String `tfsdk:"github_api_base_url"`
	GithubAPIVersion types.String `tfsdk:"github_api_version"`
}

// New is a helper function to simplify the provider implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PingOneProvisioningProvider{
			Version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *PingOneProvisioningProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pingoneprovisioning"
	resp.Version = p.Version
}

// Schema defines the provider-level schema for configuration.
func (p *PingOneProvisioningProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The PingOne Provisioning provider is used to interact with PingOne Provisioning services.",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Description: "The Client ID for the worker application. Can also be set with the `PINGONE_CLIENT_ID` environment variable.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "The Client Secret for the worker application. Can also be set with the `PINGONE_CLIENT_SECRET` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The Environment ID where the worker application is defined. Can also be set with the `PINGONE_ENVIRONMENT_ID` environment variable.",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "The PingOne region to use. Short codes: `NA`, `EU`, `AP`, `CA`, `AU`, `SG`. Long codes: `NorthAmerica`, `Europe`, `AsiaPacific`, `Australia-AsiaPacific`, `Canada`, `Singapore`. Can also be set with the `PINGONE_REGION` environment variable. Default: `NA`",
				Optional:    true,
			},
			"oauth_token_url": schema.StringAttribute{
				Description: "Optional override for the OAuth token URL (example: `https://auth.pingone.com/<env_id>/as/token`). If unset, derived from `region` + `environment_id`.",
				Optional:    true,
			},
			"api_base_url": schema.StringAttribute{
				Description: "Optional override for the Management API base URL (example: `https://api.pingone.com/v1`). If unset, derived from `region`.",
				Optional:    true,
			},
			"github_token": schema.StringAttribute{
				Description: "GitHub classic personal access token for enterprise team APIs. Can also be set with the `GITHUB_TOKEN` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"github_api_base_url": schema.StringAttribute{
				Description: "Optional override for the GitHub API base URL (default: `https://api.github.com`). Can also be set with the `GITHUB_API_BASE_URL` environment variable.",
				Optional:    true,
			},
			"github_api_version": schema.StringAttribute{
				Description: "Optional override for the GitHub API version header (default: `2022-11-28`). Can also be set with the `GITHUB_API_VERSION` environment variable.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a PingOne SDK client for data sources and resources.
func (p *PingOneProvisioningProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config PingOneProvisioningProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values from environment variables if not set in config
	clientId := os.Getenv("PINGONE_CLIENT_ID")
	if !config.ClientId.IsNull() {
		clientId = config.ClientId.ValueString()
	}

	clientSecret := os.Getenv("PINGONE_CLIENT_SECRET")
	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	environmentId := os.Getenv("PINGONE_ENVIRONMENT_ID")
	if !config.EnvironmentId.IsNull() {
		environmentId = config.EnvironmentId.ValueString()
	}

	region := "NA" // Default
	if v := os.Getenv("PINGONE_REGION"); v != "" {
		region = v
	}
	if !config.Region.IsNull() {
		region = config.Region.ValueString()
	}

	oauthTokenURL := ""
	if !config.OauthTokenURL.IsNull() {
		oauthTokenURL = strings.TrimSpace(config.OauthTokenURL.ValueString())
	}

	apiBaseURL := ""
	if !config.APIBaseURL.IsNull() {
		apiBaseURL = strings.TrimSpace(config.APIBaseURL.ValueString())
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if !config.GithubToken.IsNull() {
		githubToken = config.GithubToken.ValueString()
	}

	githubAPIBaseURL := os.Getenv("GITHUB_API_BASE_URL")
	if !config.GithubAPIBaseURL.IsNull() {
		githubAPIBaseURL = strings.TrimSpace(config.GithubAPIBaseURL.ValueString())
	}

	githubAPIVersion := os.Getenv("GITHUB_API_VERSION")
	if !config.GithubAPIVersion.IsNull() {
		githubAPIVersion = strings.TrimSpace(config.GithubAPIVersion.ValueString())
	}

	// Map short codes (terraform standard) to Long Codes (SDK Requirement)
	mappedRegion := mapRegion(region)

	if clientId == "" || clientSecret == "" || environmentId == "" {
		resp.Diagnostics.AddError(
			"Missing Configuration",
			"The client_id, client_secret, and environment_id must be configured via the provider block or environment variables.",
		)
		return
	}

	apiClient, err := newManagementClient(ctx, p.Version, clientId, clientSecret, environmentId, mappedRegion, oauthTokenURL, apiBaseURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create PingOne Client",
			fmt.Sprintf("An error occurred when creating the PingOne client: %s", err),
		)
		return
	}

	// Create the provider client structure to pass to resources
	clientData := &client.Client{
		API: apiClient,
	}

	if githubToken != "" {
		userAgent := "terraform-provider-pingoneprovisioning"
		if p.Version != "" {
			userAgent = fmt.Sprintf("terraform-provider-pingoneprovisioning/%s", p.Version)
		}

		githubClient, ghErr := githubapi.NewClient(githubToken, githubAPIBaseURL, githubAPIVersion, userAgent, nil)
		if ghErr != nil {
			resp.Diagnostics.AddError(
				"Unable to create GitHub client",
				fmt.Sprintf("An error occurred when creating the GitHub client: %s", ghErr),
			)
			return
		}

		clientData.GitHub = githubClient
	}

	resp.DataSourceData = clientData
	resp.ResourceData = clientData
}

func newManagementClient(ctx context.Context, providerVersion string, clientID string, clientSecret string, authEnvironmentID string, region string, oauthTokenURL string, apiBaseURL string) (*management.APIClient, error) {
	regionSuffix, err := regionToURLSuffix(region)
	if err != nil {
		return nil, err
	}

	tokenURL := strings.TrimSpace(oauthTokenURL)
	if tokenURL == "" {
		tokenURL = fmt.Sprintf("https://auth.pingone.%s/%s/as/token", regionSuffix, authEnvironmentID)
	}

	tokenURL, err = normalizeURL(tokenURL, "https")
	if err != nil {
		return nil, fmt.Errorf("invalid oauth_token_url %q: %w", oauthTokenURL, err)
	}

	tokenCfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}

	baseRT := http.DefaultTransport
	if envBool("PINGONEPROVISIONING_DEBUG_HTTP") {
		baseRT = &loggingTransport{rt: baseRT}
	}

	// Do not use the provider Configure() request context for the token source.
	// That context is canceled after Configure returns, which would make all
	// subsequent token refreshes fail with `context canceled`.
	tokenHTTPClient := &http.Client{
		Transport: baseRT,
		Timeout:   90 * time.Second,
	}
	tokenCtx := context.WithValue(context.WithoutCancel(ctx), oauth2.HTTPClient, tokenHTTPClient)
	tokenSource := tokenCfg.TokenSource(tokenCtx)

	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.ReuseTokenSource(nil, tokenSource),
			Base:   baseRT,
		},
		Timeout: 90 * time.Second,
	}

	cfg := management.NewConfiguration()
	cfg.HTTPClient = httpClient
	if providerVersion != "" {
		cfg.AppendUserAgent(fmt.Sprintf("terraform-provider-pingoneprovisioning/%s", providerVersion))
	}

	baseURL := strings.TrimSpace(apiBaseURL)
	if baseURL != "" {
		baseURL, err = normalizeURL(baseURL, "https")
		if err != nil {
			return nil, fmt.Errorf("invalid api_base_url %q: %w", apiBaseURL, err)
		}

		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid api_base_url %q: %w", apiBaseURL, err)
		}
		if u.Host == "" {
			return nil, fmt.Errorf("invalid api_base_url %q: missing host", apiBaseURL)
		}

		cfg.SetDefaultServerIndex(1)
		if err := cfg.SetDefaultServerVariableDefaultValue("baseHostname", u.Host); err != nil {
			return nil, err
		}
		if u.Scheme != "" {
			if err := cfg.SetDefaultServerVariableDefaultValue("protocol", u.Scheme); err != nil {
				return nil, err
			}
		}
	} else {
		cfg.SetDefaultServerIndex(0)
		if err := cfg.SetDefaultServerVariableDefaultValue("suffix", regionSuffix); err != nil {
			return nil, err
		}
	}

	apiClient := management.NewAPIClient(cfg)
	if apiClient == nil {
		return nil, fmt.Errorf("failed to initialize PingOne management API client")
	}

	return apiClient, nil
}

type loggingTransport struct {
	rt http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	authVals := req.Header.Values("Authorization")
	for i, v := range authVals {
		fields := strings.Fields(v)
		scheme := "<empty>"
		if len(fields) > 0 {
			scheme = fields[0]
		}
		log.Printf("pingoneprovisioning: HTTP %s %s Authorization[%d]=%s sha256b64=%s", req.Method, req.URL.String(), i, scheme, sha256Base64(v))
	}
	if len(authVals) == 0 {
		log.Printf("pingoneprovisioning: HTTP %s %s Authorization=<none>", req.Method, req.URL.String())
	}

	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		log.Printf("pingoneprovisioning: -> error after %s: %v", time.Since(start).Round(time.Millisecond), err)
		return nil, err
	}

	log.Printf("pingoneprovisioning: -> %s in %s", resp.Status, time.Since(start).Round(time.Millisecond))
	return resp, nil
}

func sha256Base64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func envBool(key string) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return false
	}
	switch strings.ToLower(raw) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	default:
		return false
	}
}

func normalizeURL(raw string, defaultScheme string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty url")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		u, err = url.Parse(defaultScheme + "://" + raw)
		if err != nil {
			return "", err
		}
	}

	if u.Host == "" {
		return "", fmt.Errorf("missing host")
	}

	return u.String(), nil
}

func regionToURLSuffix(region string) (string, error) {
	switch strings.TrimSpace(region) {
	case "NorthAmerica":
		return "com", nil
	case "Europe":
		return "eu", nil
	case "AsiaPacific":
		return "asia", nil
	case "Australia-AsiaPacific":
		return "com.au", nil
	case "Canada":
		return "ca", nil
	case "Singapore":
		return "sg", nil
	default:
		return "", fmt.Errorf("invalid region %q (expected NA/EU/AP/CA/AU/SG or NorthAmerica/Europe/AsiaPacific/Australia-AsiaPacific/Canada/Singapore)", region)
	}
}

// mapRegion converts standard 2-char region codes to PingOne SDK specific region names
func mapRegion(code string) string {
	switch strings.ToUpper(code) {
	case "NA", "NORTHAMERICA":
		return "NorthAmerica"
	case "EU", "EUROPE":
		return "Europe"
	case "CA", "CANADA":
		return "Canada"
	case "AP", "ASIAPACIFIC":
		return "AsiaPacific"
	case "AU", "AUSTRALIA", "AUSTRALIA-ASIAPACIFIC":
		return "Australia-AsiaPacific"
	case "SG", "SINGAPORE":
		return "Singapore"
	default:
		return code // Return as-is if no match found
	}
}

// Resources defines the resources implemented in the provider.
func (p *PingOneProvisioningProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPropagationStoreResource,
		NewPropagationPlanResource,
		NewPropagationRuleResource,
		NewUserCustomAttributesResource,
		NewGithubEnterpriseTeamResource,
		NewGithubEnterpriseTeamMemberResource,
		NewGithubEnterpriseTeamMembersResource,
		NewGithubEnterpriseTeamOrganizationResource,
		NewGithubEnterpriseTeamOrganizationsResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *PingOneProvisioningProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPropagationStoreDataSource,
		NewPropagationStoresDataSource,
		NewPropagationPlanDataSource,
		NewPropagationRuleDataSource,
		NewGroupsDataSource,
		NewGithubEnterpriseTeamsDataSource,
		NewGithubScimGroupDataSource,
	}
}
