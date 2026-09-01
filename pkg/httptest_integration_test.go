package goe2e_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e/pkg"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryServerSupportsHTTPS(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			t.Error("expected an HTTPS request to use TLS")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	goe2e.TestRequest(t, &goe2e.TestConfig{
		Name:       "HTTPS request",
		HTTPClient: server.Client(),
		SpecOpts:   []goe2e.SpecOption{goe2e.WithURL("https://app.test/health")},
		PostTestStatements: []goe2e.TestStatement{
			goe2e.AssertStatusCode(http.StatusNoContent),
		},
	})
}

func TestInMemoryServerHandlesMalformedJSON(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	goe2e.TestRequest(t, &goe2e.TestConfig{
		Name:       "invalid JSON payload",
		HTTPClient: server.Client(),
		SpecOpts: []goe2e.SpecOption{
			goe2e.WithMethod(http.MethodPost),
			goe2e.WithURL("https://app.test/persons"),
			goe2e.WithBody([]byte(`{"name":`)),
		},
		RequestMods: []goe2e.RequestModifier{
			goe2e.WithContentType(goe2e.ContentHeaderJSON),
		},
		PostTestStatements: []goe2e.TestStatement{
			goe2e.AssertStatusCode(http.StatusBadRequest),
			goe2e.AssertResponseBodyContains("invalid JSON"),
		},
	})
}

func TestInMemoryServerReportsHandlerPanic(t *testing.T) {
	if os.Getenv("GOE2E_PANIC_HELPER") == "1" {
		server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("intentional handler panic")
		}))
		_, _ = server.Client().Get("https://app.test/panic")
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestInMemoryServerReportsHandlerPanic$")
	command.Env = append(os.Environ(), "GOE2E_PANIC_HELPER=1")
	output, err := command.CombinedOutput()
	assert.Error(t, err, "the helper test should fail when its handler panics")
	assert.True(t, strings.Contains(string(output), "httptest: panic in server handler"), string(output))
}
