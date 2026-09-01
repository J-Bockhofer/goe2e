package goe2e

import (
	"fmt"
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

// WithBearerToken sets the Authorization header using the Bearer scheme.
func WithBearerToken(token string) RequestModifier {
	return func(r *http.Request) error {
		r.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

// WithBasicAuth sets the Authorization header using HTTP Basic authentication.
func WithBasicAuth(username, password string) RequestModifier {
	return func(r *http.Request) error {
		r.SetBasicAuth(username, password)
		return nil
	}
}

// WithCookies adds cookies to the request.
func WithCookies(cookies ...*http.Cookie) RequestModifier {
	return func(r *http.Request) error {
		for _, cookie := range cookies {
			if cookie == nil {
				return fmt.Errorf("request cookie must not be nil")
			}
			r.AddCookie(cookie)
		}
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
