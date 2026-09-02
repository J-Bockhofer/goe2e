package goe2e_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e"
)

func TestLifecycleStatementHelpers(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/persons/42")
		w.Header().Set("Content-Type", goe2e.ContentHeaderJSON)
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprint(w, `{"name":"John","id":"42"}`)
	}))

	goe2e.TestRequest(t, &goe2e.TestConfig{
		Name:       "POST /persons",
		HTTPClient: server.Client(),
		SpecOpts: []goe2e.SpecOption{
			goe2e.WithMethod(http.MethodPost),
			goe2e.WithURL("https://service.test/persons?include=address"),
		},
		RequestMods: []goe2e.RequestModifier{
			goe2e.WithContentType(goe2e.ContentHeaderJSON),
		},
		PreTestStatements: []goe2e.TestStatement{
			goe2e.AssertRequestMethod(http.MethodPost),
			goe2e.AssertRequestPath("/persons"),
			goe2e.AssertRequestQuery("include", "address"),
			goe2e.AssertRequestHeader("Content-Type", goe2e.ContentHeaderJSON),
		},
		PostTestStatements: []goe2e.TestStatement{
			goe2e.AssertStatusCode(http.StatusCreated),
			goe2e.AssertResponseHeader("Location", "/persons/42"),
			goe2e.AssertResponseBodyContains(`"name":"John"`),
			goe2e.AssertResponseJSONEquals(`{"id":"42","name":"John"}`),
		},
	})
}
