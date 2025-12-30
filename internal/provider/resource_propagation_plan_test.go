package provider

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestIsPropagationPlanEnvironmentAlreadyHasPlanError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resp     *http.Response
		expected bool
	}{
		{
			name:     "nil response",
			resp:     nil,
			expected: false,
		},
		{
			name: "non-400/409 status",
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{"message":"environment contains an existing plan"}`)),
			},
			expected: false,
		},
		{
			name: "root message match",
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"message":"environment contains an existing plan"}`)),
			},
			expected: true,
		},
		{
			name: "details message match",
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body: io.NopCloser(strings.NewReader(`{
					"code":"INVALID_DATA",
					"details":[{"target":"environment","message":"environment contains an existing plan"}]
				}`)),
			},
			expected: true,
		},
		{
			name: "details target mismatch",
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body: io.NopCloser(strings.NewReader(`{
					"details":[{"target":"name","message":"environment contains an existing plan"}]
				}`)),
			},
			expected: false,
		},
		{
			name: "invalid JSON body",
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`not-json`)),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isPropagationPlanEnvironmentAlreadyHasPlanError(tt.resp); got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
