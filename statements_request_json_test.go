package goe2e_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e"
)

func TestAssertRequestJSONEqualsRestoresRequestBody(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil || payload["name"] != "John" {
			t.Errorf("server received unexpected request body %q: %v", body, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))

	goe2e.TestRequest(t, &goe2e.TestConfig{
		Name:       "POST /persons with JSON body",
		HTTPClient: server.Client(),
		SpecOpts: []goe2e.SpecOption{
			goe2e.WithMethod(http.MethodPost),
			goe2e.WithURL("https://service.test/persons"),
			goe2e.WithJSON(goe2e.H{"name": "John", "age": 32}),
		},
		PreTestStatements: []goe2e.TestStatement{
			goe2e.AssertRequestJSONEquals(`{"age":32,"name":"John"}`),
		},
		PostTestStatements: []goe2e.TestStatement{
			goe2e.AssertStatusCode(http.StatusCreated),
		},
	})
}
