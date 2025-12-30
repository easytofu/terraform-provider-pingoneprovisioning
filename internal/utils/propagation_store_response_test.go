package utils

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestExtractPropagationStoreTypeStatus(t *testing.T) {
	t.Parallel()

	t.Run("nil_response", func(t *testing.T) {
		t.Parallel()

		gotType, gotStatus, err := ExtractPropagationStoreTypeStatus(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotType != "" || gotStatus != "" {
			t.Fatalf("expected empty values, got type=%q status=%q", gotType, gotStatus)
		}
	})

	t.Run("extracts_and_restores_body", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			Body: io.NopCloser(bytes.NewBufferString(`{"type":"PingOne","status":"ACTIVE"}`)),
		}

		gotType, gotStatus, err := ExtractPropagationStoreTypeStatus(resp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotType != "PingOne" || gotStatus != "ACTIVE" {
			t.Fatalf("unexpected values: type=%q status=%q", gotType, gotStatus)
		}

		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			t.Fatalf("failed reading restored body: %v", readErr)
		}
		if string(bodyBytes) != `{"type":"PingOne","status":"ACTIVE"}` {
			t.Fatalf("expected body to be restored, got %q", string(bodyBytes))
		}
	})

	t.Run("missing_status", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			Body: io.NopCloser(bytes.NewBufferString(`{"type":"directory"}`)),
		}

		gotType, gotStatus, err := ExtractPropagationStoreTypeStatus(resp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotType != "directory" || gotStatus != "" {
			t.Fatalf("unexpected values: type=%q status=%q", gotType, gotStatus)
		}
	})
}
