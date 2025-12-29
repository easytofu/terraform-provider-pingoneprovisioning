package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type propagationStoreResponseEnvelope struct {
	Type   string  `json:"type"`
	Status *string `json:"status,omitempty"`
}

// ExtractPropagationStoreTypeStatus reads and restores the HTTP response body, extracting
// the raw `type` and `status` values from the PingOne API response.
//
// This is required because the generated SDK enums coerce unknown values into "UNKNOWN",
// which breaks Terraform state consistency for newer propagation store types/statuses.
func ExtractPropagationStoreTypeStatus(resp *http.Response) (string, string, error) {
	if resp == nil || resp.Body == nil {
		return "", "", nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var envelope propagationStoreResponseEnvelope
	if err := json.Unmarshal(bodyBytes, &envelope); err != nil {
		return "", "", err
	}

	status := ""
	if envelope.Status != nil {
		status = *envelope.Status
	}

	return envelope.Type, status, nil
}
