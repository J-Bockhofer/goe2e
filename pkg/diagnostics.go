package goe2e

import (
	"fmt"
	"net/http"
	"strings"
)

const maxDiagnosticBodyBytes = 4 << 10

var diagnosticRequestHeaders = []string{
	"Accept",
	"Content-Type",
	"User-Agent",
}

var diagnosticResponseHeaders = []string{
	"Content-Type",
	"Content-Length",
	"Location",
	"Retry-After",
	"WWW-Authenticate",
	"X-Request-ID",
}

func requestDiagnostics(rh *RequestHandler) string {
	var b strings.Builder
	b.WriteString("request:")
	if rh == nil || rh.spec == nil || rh.spec.Request == nil {
		b.WriteString(" <unavailable>")
		return b.String()
	}

	req := rh.spec.Request
	fmt.Fprintf(&b, "\n  %s %s", req.Method, req.URL)
	writeDiagnosticHeaders(&b, "  headers", req.Header, diagnosticRequestHeaders)

	if rh.Response == nil {
		return b.String()
	}

	b.WriteString("\nresponse:")
	if rh.Response.Status != "" {
		fmt.Fprintf(&b, "\n  status: %s", rh.Response.Status)
	} else {
		fmt.Fprintf(&b, "\n  status: %d", rh.Response.StatusCode)
	}
	writeDiagnosticHeaders(&b, "  headers", rh.Response.Header, diagnosticResponseHeaders)
	if len(rh.ResponseBody) > 0 {
		fmt.Fprintf(&b, "\n  body: %s", truncateDiagnosticBody(rh.ResponseBody))
	}
	return b.String()
}

func writeDiagnosticHeaders(b *strings.Builder, label string, headers http.Header, keys []string) {
	var lines []string
	for _, key := range keys {
		if value := diagnosticHeaderValue(headers, key); value != "" {
			lines = append(lines, fmt.Sprintf("%s: %s", key, value))
		}
	}
	if len(lines) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s:\n    %s", label, strings.Join(lines, "\n    "))
}

func diagnosticHeaderValue(headers http.Header, key string) string {
	if value := headers.Get(key); value != "" {
		return value
	}
	for headerKey, values := range headers {
		if strings.EqualFold(headerKey, key) {
			return strings.Join(values, ", ")
		}
	}
	return ""
}

func truncateDiagnosticBody(body []byte) string {
	if len(body) <= maxDiagnosticBodyBytes {
		return string(body)
	}
	return fmt.Sprintf("%s… (truncated; showing %d of %d bytes)", body[:maxDiagnosticBodyBytes], maxDiagnosticBodyBytes, len(body))
}
