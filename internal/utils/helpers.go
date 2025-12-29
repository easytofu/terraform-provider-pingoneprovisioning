package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ReadError processes the HTTP response to extract error details
func ReadError(resp *http.Response, err error) string {
	if resp == nil {
		return err.Error()
	}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Sprintf("%s (failed to read response body: %s)", err, readErr)
	}

	return fmt.Sprintf("%s: %s", err, string(bodyBytes))
}

// SplitImportID is a helper to split import IDs
func SplitImportID(id string, expectedParts int) []string {
	parts := strings.Split(id, "/")
	if len(parts) != expectedParts {
		return nil
	}
	return parts
}

// --- SDK to Terraform Conversion Helpers ---

func StringOkToTF(v *string, ok bool) types.String {
	if !ok || v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func Int64OkToTF(v *int64, ok bool) types.Int64 {
	if !ok || v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

// Int32OkToTF converts SDK *int32 to Terraform types.Int64
func Int32OkToTF(v *int32, ok bool) types.Int64 {
	if !ok || v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}

func BoolOkToTF(v *bool, ok bool) types.Bool {
	if !ok || v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

func TimeOkToTF(v *time.Time, ok bool) types.String {
	if !ok || v == nil {
		return types.StringNull()
	}
	return types.StringValue(v.Format(time.RFC3339))
}

// --- Map (API Response) to Terraform Helpers ---

func FromMapString(m map[string]interface{}, key string) types.String {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return types.StringValue(s)
		}
	}
	return types.StringNull()
}

func FromMapBool(m map[string]interface{}, key string) types.Bool {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return types.BoolValue(b)
		}
	}
	return types.BoolNull()
}

// HandleSDKError provides a consistent error message from SDK failures.
func HandleSDKError(err error, resp *http.Response) string {
	if resp == nil {
		return err.Error()
	}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Sprintf("%s (failed to read response body: %s)", err, readErr)
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return fmt.Sprintf("%s: %s", err, string(bodyBytes))
}
