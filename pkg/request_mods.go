package goe2e

import (
	"log/slog"
	"net/http"
	"net/http/httptrace"
	"time"
)

// RequestModifer is used to modify the http.Request.
type RequestModifier func(*http.Request) error

// WithHeaders set the http.Request.Header from a map.
func WithHeaders(headers D) RequestModifier {
	return func(r *http.Request) error {
		for k, v := range headers {
			r.Header.Set(k, v)
		}
		return nil
	}
}

const (
	ContentHeaderJSON string = "application/json"
)

// WithContentType sets the ContentType request header.
func WithContentType(contentType string) RequestModifier {
	return func(r *http.Request) error {
		r.Header.Set("Content-Type", contentType)
		return nil
	}
}

// WithTimeToFirstByte attaches an httptrace to the request that measures the time
// from when the request is fully written to the wire until the first response byte arrives (TTFB).
func WithTimeToFirstByte() RequestModifier {
	return func(r *http.Request) error {
		var start time.Time
		trace := &httptrace.ClientTrace{
			WroteRequest: func(_ httptrace.WroteRequestInfo) {
				start = time.Now()
			},
			GotFirstResponseByte: func() {
				slog.Info("TTFB: " + time.Since(start).String())
			},
		}
		*r = *r.WithContext(httptrace.WithClientTrace(r.Context(), trace))
		return nil
	}
}
