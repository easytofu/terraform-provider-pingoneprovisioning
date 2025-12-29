package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestNewManagementClient_DoesNotUseCanceledContextForToken(t *testing.T) {
	var tokenRequests atomic.Int32
	var apiRequests atomic.Int32

	originalDefaultTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/as/token":
			tokenRequests.Add(1)

			if r.Method != http.MethodPost {
				t.Errorf("token endpoint method = %s, want %s", r.Method, http.MethodPost)
			}

			if err := r.ParseForm(); err != nil {
				t.Errorf("token endpoint ParseForm: %v", err)
			}

			if got := r.PostForm.Get("grant_type"); got != "client_credentials" {
				t.Errorf("token endpoint grant_type = %q, want %q", got, "client_credentials")
			}

			if user, pass, ok := r.BasicAuth(); ok {
				if user != "client-id" {
					t.Errorf("token endpoint basic auth username = %q, want %q", user, "client-id")
				}
				if pass != "client-secret" {
					t.Errorf("token endpoint basic auth password = %q, want %q", pass, "client-secret")
				}
			} else {
				if got := r.PostForm.Get("client_id"); got != "client-id" {
					t.Errorf("token endpoint client_id = %q, want %q", got, "client-id")
				}
				if got := r.PostForm.Get("client_secret"); got != "client-secret" {
					t.Errorf("token endpoint client_secret = %q, want %q", got, "client-secret")
				}
			}

			body := `{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    r,
			}, nil
		default:
			apiRequests.Add(1)

			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Errorf("api Authorization = %q, want %q", got, "Bearer test-token")
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("ok")),
				Request:    r,
			}, nil
		}
	})
	t.Cleanup(func() { http.DefaultTransport = originalDefaultTransport })

	configureCtx, cancel := context.WithCancel(context.Background())
	cancel()

	apiClient, err := newManagementClient(
		configureCtx,
		"test",
		"client-id",
		"client-secret",
		"auth-environment-id",
		"NorthAmerica",
		"https://auth.example/as/token",
		"",
	)
	if err != nil {
		t.Fatalf("newManagementClient error: %v", err)
	}

	resp, err := apiClient.GetConfig().HTTPClient.Get("https://api.example/ping")
	if err != nil {
		t.Fatalf("api request error: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("api status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if tokenRequests.Load() == 0 {
		t.Fatalf("expected at least one token request, got %d", tokenRequests.Load())
	}

	if apiRequests.Load() == 0 {
		t.Fatalf("expected at least one api request, got %d", apiRequests.Load())
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
