package ups

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIError is a normalized UPS / transport error.
type APIError struct {
	Op         string // e.g. "rate", "ship", "track", "oauth"
	StatusCode int
	Code       string
	Message    string
	Body       string
	Err        error
}

func (e *APIError) Error() string {
	var b strings.Builder
	b.WriteString("ups")
	if e.Op != "" {
		b.WriteString(": ")
		b.WriteString(e.Op)
	}
	if e.StatusCode != 0 {
		fmt.Fprintf(&b, " HTTP %d", e.StatusCode)
	}
	if e.Code != "" {
		fmt.Fprintf(&b, " code=%s", e.Code)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	}
	if e.Err != nil {
		fmt.Fprintf(&b, ": %v", e.Err)
	}
	if e.Body != "" && e.Message == "" {
		fmt.Fprintf(&b, ": %s", truncate(e.Body, 512))
	}
	return b.String()
}

func (e *APIError) Unwrap() error { return e.Err }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// responseError builds an APIError from an HTTP status and body, best-effort parsing UPS error JSON.
func responseError(op string, status int, body []byte) *APIError {
	err := &APIError{Op: op, StatusCode: status, Body: string(body), Message: "request failed"}
	var envelope struct {
		Response struct {
			Errors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"response"`
	}
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Response.Errors) > 0 {
		err.Code = envelope.Response.Errors[0].Code
		err.Message = envelope.Response.Errors[0].Message
		return err
	}
	// Alternate shape seen in some UPS APIs
	var alt struct {
		Errors []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	if json.Unmarshal(body, &alt) == nil && len(alt.Errors) > 0 {
		err.Code = alt.Errors[0].Code
		err.Message = alt.Errors[0].Message
	}
	return err
}
