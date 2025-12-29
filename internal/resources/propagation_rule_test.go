package resources

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/patrickcping/pingone-go-sdk-v2/management"
)

func TestIsBadAuthorizationHeaderError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		resp     *http.Response
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			resp:     &http.Response{StatusCode: http.StatusForbidden, Body: io.NopCloser(strings.NewReader(`{"message":"Invalid key=value pair in Authorization header"}`))},
			expected: false,
		},
		{
			name:     "nil response",
			err:      errors.New("boom"),
			resp:     nil,
			expected: false,
		},
		{
			name:     "non-403 status",
			err:      errors.New("boom"),
			resp:     &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(`{"message":"Invalid key=value pair in Authorization header"}`))},
			expected: false,
		},
		{
			name:     "403 body matches",
			err:      errors.New("boom"),
			resp:     &http.Response{StatusCode: http.StatusForbidden, Body: io.NopCloser(strings.NewReader(`{"message":"Invalid key=value pair (missing equal-sign) in Authorization header"}`))},
			expected: true,
		},
		{
			name:     "403 unrelated body",
			err:      errors.New("boom"),
			resp:     &http.Response{StatusCode: http.StatusForbidden, Body: io.NopCloser(strings.NewReader(`{"message":"forbidden"}`))},
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isBadAuthorizationHeaderError(tt.err, tt.resp); got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestCreatePropagationRuleViaPlan_UsesPlanScopedEndpoint(t *testing.T) {
	t.Parallel()

	cfg := management.NewConfiguration()
	cfg.SetDefaultServerIndex(1)
	if err := cfg.SetDefaultServerVariableDefaultValue("baseHostname", "api.example"); err != nil {
		t.Fatalf("SetDefaultServerVariableDefaultValue: %v", err)
	}

	cfg.HTTPClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
			}

			if got := r.URL.String(); got != "https://api.example/v1/environments/env-id/propagation/plans/plan-id/rules" {
				t.Fatalf("url = %s, want %s", got, "https://api.example/v1/environments/env-id/propagation/plans/plan-id/rules")
			}

			return &http.Response{
				StatusCode: http.StatusCreated,
				Status:     "201 Created",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":"rule-123"}`)),
				Request:    r,
			}, nil
		}),
	}

	apiClient := management.NewAPIClient(cfg)

	ruleID, _, err := createPropagationRuleViaPlan(
		context.Background(),
		apiClient,
		"env-id",
		"plan-id",
		map[string]interface{}{
			"name": "test",
			"sourceStore": map[string]interface{}{
				"id": "source-id",
			},
			"targetStore": map[string]interface{}{
				"id": "target-id",
			},
		},
	)
	if err != nil {
		t.Fatalf("createPropagationRuleViaPlan error: %v", err)
	}
	if ruleID != "rule-123" {
		t.Fatalf("ruleID = %q, want %q", ruleID, "rule-123")
	}
}

func TestCreatePropagationRevisionWithFallback_UsesExpectedEndpoint(t *testing.T) {
	t.Parallel()

	cfg := management.NewConfiguration()
	cfg.SetDefaultServerIndex(1)
	if err := cfg.SetDefaultServerVariableDefaultValue("baseHostname", "api.example"); err != nil {
		t.Fatalf("SetDefaultServerVariableDefaultValue: %v", err)
	}

	cfg.HTTPClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
			}

			if got := r.URL.String(); got != "https://api.example/v1/environments/env-id/propagation/revisions" {
				t.Fatalf("url = %s, want %s", got, "https://api.example/v1/environments/env-id/propagation/revisions")
			}

			return &http.Response{
				StatusCode: http.StatusCreated,
				Status:     "201 Created",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(``)),
				Request:    r,
			}, nil
		}),
	}

	apiClient := management.NewAPIClient(cfg)

	if _, err := createPropagationRevisionWithFallback(context.Background(), apiClient, "env-id"); err != nil {
		t.Fatalf("createPropagationRevisionWithFallback error: %v", err)
	}
}

func TestDeletePropagationMapping_UsesMappingsEndpoint(t *testing.T) {
	t.Parallel()

	cfg := management.NewConfiguration()
	cfg.Servers = management.ServerConfigurations{
		{URL: "https://api.example/v1/propagation/mapping"},
	}
	cfg.SetDefaultServerIndex(0)

	cfg.HTTPClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodDelete {
				t.Fatalf("method = %s, want %s", r.Method, http.MethodDelete)
			}

			if got := r.URL.String(); got != "https://api.example/v1/environments/env-id/propagation/mappings/map-id" {
				t.Fatalf("url = %s, want %s", got, "https://api.example/v1/environments/env-id/propagation/mappings/map-id")
			}

			return &http.Response{
				StatusCode: http.StatusNoContent,
				Status:     "204 No Content",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(``)),
				Request:    r,
			}, nil
		}),
	}

	apiClient := management.NewAPIClient(cfg)

	if _, err := deletePropagationMapping(context.Background(), apiClient, "env-id", "map-id"); err != nil {
		t.Fatalf("deletePropagationMapping error: %v", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
