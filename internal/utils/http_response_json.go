package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// ReadAndRestoreResponseBody reads the response body and restores it so it can be read again.
func ReadAndRestoreResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}

// DecodeResponseJSON reads, restores, and JSON-decodes the response body.
//
// If the response has an empty body, this returns (nil, nil).
func DecodeResponseJSON(resp *http.Response) (any, error) {
	bodyBytes, err := ReadAndRestoreResponseBody(resp)
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		return nil, nil
	}

	var decoded any
	if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

// ExtractEmbeddedArray attempts to extract a list payload from common PingOne list response shapes.
//
// Supports:
// - `{"_embedded": {"<key>": [ ... ]}}` (for any of the provided keys)
// - `[ ... ]` (root array)
func ExtractEmbeddedArray(decoded any, embeddedKeys ...string) ([]interface{}, error) {
	if decoded == nil {
		return nil, nil
	}

	if list, ok := decoded.([]interface{}); ok {
		return list, nil
	}

	root, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected JSON response shape")
	}

	embeddedRaw, ok := root["_embedded"]
	if !ok || embeddedRaw == nil {
		return nil, errors.New("missing _embedded in response")
	}

	embedded, ok := embeddedRaw.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid _embedded in response")
	}

	for _, key := range embeddedKeys {
		if v, ok := embedded[key]; ok && v != nil {
			if list, ok := v.([]interface{}); ok {
				return list, nil
			}
		}
	}

	return nil, errors.New("embedded list not found in response")
}

func NestedString(m map[string]interface{}, keys ...string) (string, bool) {
	if len(keys) == 0 {
		return "", false
	}

	var current any = m
	for _, key := range keys {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}

		val, ok := obj[key]
		if !ok || val == nil {
			return "", false
		}
		current = val
	}

	s, ok := current.(string)
	return s, ok
}
