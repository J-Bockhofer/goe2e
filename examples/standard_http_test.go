package examples_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e/pkg"
)

func TestStandardLibraryHandler(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /persons", func(w http.ResponseWriter, r *http.Request) {
		var input person
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", goe2e.ContentHeaderJSON)
		w.Header().Set("Location", "/persons/42")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "42",
			"name": input.Name,
		})
	})

	server := httptest.NewTestServer(t, mux)
	payload := person{Name: "John", Age: 32}

	goe2e.TestRequest(t, &goe2e.TestConfig{
		Name:       "POST /persons",
		HTTPClient: server.Client(),
		SpecOpts: []goe2e.SpecOption{
			goe2e.WithMethod(http.MethodPost),
			goe2e.WithURL("https://app.test/persons"),
			goe2e.WithJSON(payload),
		},
		RequestMods: []goe2e.RequestModifier{
			goe2e.WithContentType(goe2e.ContentHeaderJSON),
		},
		PreTestStatements: []goe2e.TestStatement{
			goe2e.AssertRequestMethod(http.MethodPost),
			goe2e.AssertRequestPath("/persons"),
			goe2e.AssertRequestJSONEquals(`{"age":32,"name":"John"}`),
		},
		PostTestStatements: []goe2e.TestStatement{
			goe2e.AssertStatusCode(http.StatusCreated),
			goe2e.AssertResponseHeader("Location", "/persons/42"),
			goe2e.AssertResponseJSONPointer("/id", "42"),
		},
	})
}
