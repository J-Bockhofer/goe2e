package goe2e_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e"

	"github.com/stretchr/testify/assert"
)

func TestDiagnosticCallback(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", goe2e.ContentHeaderJSON)
		w.Header().Set("X-Request-ID", "request-42")
		w.Header().Set("Set-Cookie", "session=secret")
		_, _ = fmt.Fprint(w, `{"ok":true}`)
	}))

	var received goe2e.RequestDiagnostic
	goe2e.TestRequest(t, &goe2e.TestConfig{
		Name:       "diagnostic callback",
		HTTPClient: server.Client(),
		SpecOpts:   []goe2e.SpecOption{goe2e.WithURL("https://app.test/diagnostic")},
		RequestMods: []goe2e.RequestModifier{
			goe2e.WithBearerToken("secret-token"),
			goe2e.WithContentType(goe2e.ContentHeaderJSON),
		},
		OnDiagnostic: func(diagnostic goe2e.RequestDiagnostic) {
			received = diagnostic
		},
	})

	assert.Equal(t, http.MethodGet, received.Method)
	assert.Equal(t, "https://app.test/diagnostic", received.URL)
	assert.Equal(t, goe2e.ContentHeaderJSON, received.RequestHeaders["Content-Type"])
	assert.Equal(t, http.StatusOK, received.ResponseCode)
	assert.Equal(t, "request-42", received.ResponseHeaders["X-Request-ID"])
	assert.Equal(t, len(`{"ok":true}`), received.ResponseBytes)
	assert.NotContains(t, received.RequestHeaders, "Authorization")
	assert.NotContains(t, received.ResponseHeaders, "Set-Cookie")
}
