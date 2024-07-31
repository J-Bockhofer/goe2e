package goe2e

import (
	"net/http"
	"strings"
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

// JoinAsRoute is a helper function to join two route components.
// It assures that they are indeed connected with a "/".
func JoinAsRoute(base, route string) string {
	const slash string = "/"
	if !strings.HasSuffix(base, slash) && !strings.HasPrefix(route, slash) {
		return base + slash + route
	}
	return base + route
}
