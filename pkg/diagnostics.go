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

// RequestDiagnostic is a safe, structured summary of a request execution.
// Header maps contain only the diagnostic header allowlists and never include credentials or cookies.
type RequestDiagnostic struct {
	Method          string
	URL             string
	RequestHeaders  map[string]string
	ResponseStatus  string
	ResponseCode    int
	ResponseHeaders map[string]string
	ResponseBytes   int
}

// Diagnostic returns a safe, structured summary suitable for logging or metrics.
func (rh *RequestHandler) Diagnostic() RequestDiagnostic {
	diagnostic := RequestDiagnostic{}
	if rh == nil || rh.spec == nil || rh.spec.Request == nil {
		return diagnostic
	}

	req := rh.spec.Request
	diagnostic.Method = req.Method
	diagnostic.URL = req.URL.String()
	diagnostic.RequestHeaders = diagnosticHeaders(req.Header, diagnosticRequestHeaders)
	if rh.Response == nil {
		return diagnostic
	}

	diagnostic.ResponseStatus = rh.Response.Status
	diagnostic.ResponseCode = rh.Response.StatusCode
	diagnostic.ResponseHeaders = diagnosticHeaders(rh.Response.Header, diagnosticResponseHeaders)
	diagnostic.ResponseBytes = len(rh.ResponseBody)
	return diagnostic
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
	values := diagnosticHeaders(headers, keys)
	if len(values) == 0 {
		return
	}
	lines := make([]string, 0, len(values))
	for _, key := range keys {
		if value, ok := values[key]; ok {
			lines = append(lines, fmt.Sprintf("%s: %s", key, value))
		}
	}
	fmt.Fprintf(b, "\n%s:\n    %s", label, strings.Join(lines, "\n    "))
}

func diagnosticHeaders(headers http.Header, keys []string) map[string]string {
	values := make(map[string]string)
	for _, key := range keys {
		if value := diagnosticHeaderValue(headers, key); value != "" {
			values[key] = value
		}
	}
	return values
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
