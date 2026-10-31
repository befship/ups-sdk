package ups

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{Op: "rate", StatusCode: 400, Code: "111100", Message: "bad request"}
	got := err.Error()
	if got == "" || !containsAll(got, "rate", "400", "111100") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestResponseError_ParseUPSEnvelope(t *testing.T) {
	body := []byte(`{"response":{"errors":[{"code":"10400","message":"Invalid request"}]}}`)
	err := responseError("ship", 400, body)
	if err.Code != "10400" || err.Message != "Invalid request" {
		t.Fatalf("%+v", err)
	}
}

func TestTokenCache(t *testing.T) {
	c := newTokenCache()
	if _, ok := c.get(); ok {
		t.Fatal("expected empty")
	}
	c.set("tok", 3600)
	tok, ok := c.get()
	if !ok || tok != "tok" {
		t.Fatalf("got %q ok=%v", tok, ok)
	}
	c.invalidate()
	if _, ok := c.get(); ok {
		t.Fatal("expected invalidated")
	}
}

func TestBaseURL(t *testing.T) {
	if BaseURL(EnvProduction) != "https://onlinetools.ups.com" {
		t.Fatal(BaseURL(EnvProduction))
	}
	if BaseURL(EnvSandbox) != "https://wwwcie.ups.com" {
		t.Fatal(BaseURL(EnvSandbox))
	}
}

func TestServerURL_FromManifest(t *testing.T) {
	u := ServerURL(EnvSandbox, "rating")
	if u != "https://wwwcie.ups.com/api" {
		t.Fatalf("got %s", u)
	}
	u = ServerURL(EnvProduction, "oauthclientcredentials")
	if u != "https://onlinetools.ups.com" {
		t.Fatalf("got %s", u)
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	inner := errors.New("boom")
	err := &APIError{Op: "oauth", Err: inner}
	if !errors.Is(err, inner) {
		t.Fatal("unwrap")
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
