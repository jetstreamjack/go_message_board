// Package helper provides shared utilities for HTTP request processing.
package helper

import (
	"encoding/json"
	"io"
	"net/http"
)

// jsonError is a custom error type for JSON-related failures.
type jsonError string

// Error implements the error interface for jsonError.
func (e jsonError) Error() string { return string(e) }

// ReadBody reads and returns the full request body, closing it afterwards.
// Returns an error if reading fails.
func ReadBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, jsonError("failed to read body")
	}
	return body, nil
}

// JSONDecode unmarshals JSON bytes into v.
// Returns an error if the data is not valid JSON.
func JSONDecode(data []byte, v any) error {
	return json.Unmarshal(data, &v)
}
