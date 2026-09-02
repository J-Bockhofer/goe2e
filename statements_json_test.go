package goe2e_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e"
)

func TestAssertResponseJSONPointer(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", goe2e.ContentHeaderJSON)
		_, _ = fmt.Fprint(w, `{"data":{"people":[{"name":"John","age":32}],"a/b":"slash","a~b":"tilde"}}`)
	}))

	goe2e.TestRequest(t, &goe2e.TestConfig{
		Name:       "GET nested response",
		HTTPClient: server.Client(),
		SpecOpts:   []goe2e.SpecOption{goe2e.WithURL("https://service.test/people")},
		PostTestStatements: []goe2e.TestStatement{
			goe2e.AssertResponseJSONPointer("/data/people/0/name", "John"),
			goe2e.AssertResponseJSONPointer("/data/people/0/age", 32),
			goe2e.AssertResponseJSONPointer("/data/a~1b", "slash"),
			goe2e.AssertResponseJSONPointer("/data/a~0b", "tilde"),
		},
	})
}
