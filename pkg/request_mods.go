package goe2e

import (
	"log"
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

// WithTimeToFirstByte will send a secondary request to precisely measure the time to first byte.
func WithTimeToFirstByte() RequestModifier {
	return func(r *http.Request) error {
		var start time.Time
		trace := &httptrace.ClientTrace{
			GotFirstResponseByte: func() {
				slog.Info("Time from start to first byte: " + time.Since(start).String())
			},
		}
		tr := r.Clone(httptrace.WithClientTrace(r.Context(), trace))
		start = time.Now()
		if _, err := http.DefaultTransport.RoundTrip(tr); err != nil {
			log.Fatal(err)
		}
		slog.Info("Total time: " + time.Since(start).String())
		return nil
	}
}
