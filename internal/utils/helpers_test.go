package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestHandleSDKError(t *testing.T) {
	t.Parallel()

	t.Run("nil_response", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf("boom")
		got := HandleSDKError(err, nil)
		if got != "boom" {
			t.Fatalf("expected error string to be returned, got %q", got)
		}
	})

	t.Run("preserves_body", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf("boom")
		resp := &http.Response{
			Body: io.NopCloser(bytes.NewBufferString(`{"message":"bad"}`)),
		}

		got := HandleSDKError(err, resp)
		if got == "boom" || got == "" {
			t.Fatalf("expected error string to include response body, got %q", got)
		}

		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			t.Fatalf("failed reading response body after HandleSDKError: %v", readErr)
		}

		if string(bodyBytes) != `{"message":"bad"}` {
			t.Fatalf("expected response body to be preserved, got %q", string(bodyBytes))
		}
	})
}
