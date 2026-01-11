package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/easytofu/terraform-provider-pingoneprovisioning/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// Retry configuration for rate limiting
	maxRetryTimeout = 5 * time.Minute
	initialBackoff  = 1 * time.Second
	maxBackoff      = 60 * time.Second
	backoffFactor   = 2.0
)

const (
	defaultBaseURL    = "https://api.github.com"
	defaultAPIVersion = "2022-11-28"
	defaultAccept     = "application/vnd.github+json"
	scimAccept        = "application/scim+json"
)

type GitHubClient struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
	APIVersion string
	UserAgent  string
}

func NewGitHubClient(token string, baseURL string, apiVersion string, userAgent string, httpClient *http.Client) (*GitHubClient, error) {
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

	return &GitHubClient{
		HTTPClient: httpClient,
		BaseURL:    normalizedBaseURL,
		Token:      token,
		APIVersion: apiVersion,
		UserAgent:  userAgent,
	}, nil
}

func (c *GitHubClient) Do(ctx context.Context, method string, path string, query url.Values, payload any) (*http.Response, error) {
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
	if payload != nil {
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	acceptHeader := defaultAccept
	if isScimPath(path) {
		acceptHeader = scimAccept
	}

	contentType := "application/json"
	if acceptHeader == scimAccept {
		contentType = scimAccept
	}

	// Retry loop with exponential backoff
	deadline := time.Now().Add(maxRetryTimeout)
	attempt := 0
	backoff := initialBackoff

	for {
		// Create fresh request for each attempt
		var bodyReader *bytes.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		} else {
			bodyReader = bytes.NewReader(nil)
		}

		req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", acceptHeader)
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("X-GitHub-Api-Version", c.APIVersion)
		req.Header.Set("User-Agent", c.UserAgent)
		if payload != nil {
			req.Header.Set("Content-Type", contentType)
		}

		if githubDebugEnabled() {
			tflog.Debug(ctx, "github enterprise api request", map[string]interface{}{
				"method":  req.Method,
				"url":     req.URL.String(),
				"headers": redactHeaders(req.Header),
				"body":    string(bodyBytes),
				"attempt": attempt + 1,
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

		// Check if we should retry
		if !shouldRetryStatus(resp.StatusCode) {
			return resp, nil
		}

		// Check if we've exceeded the deadline
		if time.Now().After(deadline) {
			log.Printf("pingoneprovisioning: github retry timeout exceeded after %d attempts for %s %s",
				attempt+1, method, endpoint)
			return resp, nil
		}

		// Calculate sleep duration
		sleepDuration := calculateRetryBackoff(resp, backoff)

		// Don't sleep longer than the remaining time
		remaining := time.Until(deadline)
		if sleepDuration > remaining {
			sleepDuration = remaining
		}

		// Drain and close the response body before retrying
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		log.Printf("pingoneprovisioning: github received %d for %s %s, retrying in %s (attempt %d)",
			resp.StatusCode, method, endpoint, sleepDuration.Round(time.Millisecond), attempt+1)

		// Sleep before retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleepDuration):
		}

		// Increase backoff for next iteration
		backoff = nextRetryBackoff(backoff)
		attempt++
	}
}

// shouldRetryStatus determines if the HTTP status code indicates a retryable error.
func shouldRetryStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout,      // 504
		http.StatusBadGateway:          // 502
		return true
	default:
		return false
	}
}

// calculateRetryBackoff determines how long to wait before the next retry.
func calculateRetryBackoff(resp *http.Response, currentBackoff time.Duration) time.Duration {
	// Check for Retry-After header
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		// Try parsing as seconds first
		if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
			duration := time.Duration(seconds) * time.Second
			if duration > maxBackoff {
				duration = maxBackoff
			}
			return duration
		}

		// Try parsing as HTTP date
		if retryTime, err := http.ParseTime(retryAfter); err == nil {
			duration := time.Until(retryTime)
			if duration < 0 {
				duration = initialBackoff
			}
			if duration > maxBackoff {
				duration = maxBackoff
			}
			return duration
		}
	}

	// Use exponential backoff with jitter
	jitter := time.Duration(rand.Float64() * 0.25 * float64(currentBackoff))
	return currentBackoff + jitter
}

// nextRetryBackoff calculates the next backoff duration using exponential growth.
func nextRetryBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * backoffFactor)
	if next > maxBackoff {
		next = maxBackoff
	}
	return next
}

func (c *GitHubClient) buildURL(path string, query url.Values) (string, error) {
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
