package goe2e

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

func TestRequestDiagnosticsIncludesSafeRequestAndResponseDetails(t *testing.T) {
	rh, err := NewRequestHandler(WithSpecOpts(WithURL("https://service.test/persons")))
	if err != nil {
		t.Fatal(err)
	}
	rh.Request().Header.Set("Content-Type", "application/json")
	rh.Request().Header.Set("Authorization", "Bearer secret")
	rh.Response = &http.Response{
		Status:     "404 Not Found",
		StatusCode: http.StatusNotFound,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Request-ID": []string{"req-42"},
			"Set-Cookie":   []string{"session=secret"},
		},
	}
	rh.ResponseBody = []byte(`{"error":"person not found"}`)

	diagnostics := requestDiagnostics(rh)
	for _, expected := range []string{
		"GET https://service.test/persons",
		"Content-Type: application/json",
		"status: 404 Not Found",
		"X-Request-ID: req-42",
		`body: {"error":"person not found"}`,
	} {
		if !strings.Contains(diagnostics, expected) {
			t.Fatalf("diagnostics missing %q:\n%s", expected, diagnostics)
		}
	}
	for _, unsafe := range []string{"Authorization", "Bearer secret", "Set-Cookie", "session=secret"} {
		if strings.Contains(diagnostics, unsafe) {
			t.Fatalf("diagnostics unexpectedly include %q:\n%s", unsafe, diagnostics)
		}
	}
}

func TestTruncateDiagnosticBody(t *testing.T) {
	body := bytes.Repeat([]byte("x"), maxDiagnosticBodyBytes+1)
	diagnostics := truncateDiagnosticBody(body)
	if !strings.Contains(diagnostics, "truncated; showing 4096 of 4097 bytes") {
		t.Fatalf("expected truncation details, got %q", diagnostics)
	}
}
