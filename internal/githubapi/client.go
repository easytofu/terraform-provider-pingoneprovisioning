package githubapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultBaseURL    = "https://api.github.com"
	defaultAPIVersion = "2022-11-28"
	defaultAccept     = "application/vnd.github+json"
	scimAccept        = "application/scim+json"
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
	APIVersion string
	UserAgent  string
}

func NewClient(token string, baseURL string, apiVersion string, userAgent string, httpClient *http.Client) (*Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, nil
	}

	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	normalizedBaseURL, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	apiVersion = strings.TrimSpace(apiVersion)
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}

	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		userAgent = "terraform-provider-pingoneprovisioning"
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 90 * time.Second}
	}

	return &Client{
		HTTPClient: httpClient,
		BaseURL:    normalizedBaseURL,
		Token:      token,
		APIVersion: apiVersion,
		UserAgent:  userAgent,
	}, nil
}

func (c *Client) Do(ctx context.Context, method string, path string, query url.Values, payload any) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("nil github client")
	}
	if c.HTTPClient == nil {
		return nil, fmt.Errorf("github client has nil http client")
	}

	endpoint, err := c.buildURL(path, query)
	if err != nil {
		return nil, err
	}

	var bodyBytes []byte
	var bodyReader *bytes.Reader
	if payload != nil {
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	acceptHeader := defaultAccept
	if isScimPath(path) {
		acceptHeader = scimAccept
	}
	req.Header.Set("Accept", acceptHeader)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-GitHub-Api-Version", c.APIVersion)
	req.Header.Set("User-Agent", c.UserAgent)
	if payload != nil {
		contentType := "application/json"
		if acceptHeader == scimAccept {
			contentType = scimAccept
		}
		req.Header.Set("Content-Type", contentType)
	}

	if githubDebugEnabled() {
		tflog.Debug(ctx, "github enterprise api request", map[string]interface{}{
			"method":  req.Method,
			"url":     req.URL.String(),
			"headers": redactHeaders(req.Header),
			"body":    string(bodyBytes),
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if githubDebugEnabled() {
			tflog.Debug(ctx, "github enterprise api request error", map[string]interface{}{
				"method": req.Method,
				"url":    req.URL.String(),
				"error":  err.Error(),
			})
		}
		return nil, err
	}

	if githubDebugEnabled() {
		respBody, _ := utils.ReadAndRestoreResponseBody(resp)
		tflog.Debug(ctx, "github enterprise api response", map[string]interface{}{
			"status":  resp.StatusCode,
			"headers": redactHeaders(resp.Header),
			"body":    string(respBody),
		})
	}

	return resp, nil
}

func (c *Client) buildURL(path string, query url.Values) (string, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}

	cleanPath := "/" + strings.TrimLeft(path, "/")
	basePath := strings.TrimRight(base.Path, "/")
	if basePath == "" {
		base.Path = cleanPath
	} else {
		base.Path = basePath + cleanPath
	}

	if query != nil && len(query) > 0 {
		base.RawQuery = query.Encode()
	}

	return base.String(), nil
}

func normalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty base url")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u, err = url.Parse("https://" + raw)
		if err != nil {
			return "", err
		}
	}
	if u.Host == "" {
		return "", fmt.Errorf("base url missing host")
	}

	u.Path = strings.TrimRight(u.Path, "/")
	return u.String(), nil
}

func githubDebugEnabled() bool {
	raw := strings.TrimSpace(os.Getenv("TF_LOG"))
	if raw == "" {
		return false
	}
	switch strings.ToLower(raw) {
	case "debug", "trace":
		return true
	default:
		return false
	}
}

func isScimPath(path string) bool {
	trimmed := strings.TrimSpace(path)
	trimmed = strings.TrimLeft(trimmed, "/")
	return strings.HasPrefix(strings.ToLower(trimmed), "scim/")
}

func redactHeaders(headers http.Header) map[string]string {
	out := make(map[string]string, len(headers))
	for k, vals := range headers {
		if strings.EqualFold(k, "Authorization") {
			out[k] = redactAuth(vals)
			continue
		}
		out[k] = strings.Join(vals, ",")
	}
	return out
}

func redactAuth(vals []string) string {
	if len(vals) == 0 {
		return "<empty>"
	}
	fields := strings.Fields(vals[0])
	scheme := "<empty>"
	if len(fields) > 0 {
		scheme = fields[0]
	}
	return scheme + " sha256b64=" + sha256Base64(vals[0])
}

func sha256Base64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}
