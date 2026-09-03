package goe2e_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e"
)

func TestTestRequest(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	tc := goe2e.TestConfig{
		HTTPClient: server.Client(),
		Request: goe2e.RequestConfig{
			Name: "GET /",
			SpecOpts: []goe2e.SpecOption{
				goe2e.WithURL("https://www.github.com"),
			},
			PostTestStatements: []goe2e.TestStatement{
				{"status 200", goe2e.TestStatusCode(200)},
			},
		},
	}
	goe2e.TestRequest(t, tc)
}

func TestTestRequestWithTimings(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	tc := goe2e.TestConfig{
		HTTPClient: server.Client(),
		Request: goe2e.RequestConfig{
			Name: "GET / with timings",
			SpecOpts: []goe2e.SpecOption{
				goe2e.WithURL("https://www.github.com"),
			},
			RequestMods: []goe2e.RequestModifier{
				goe2e.WithTimeToFirstByte(),
			},
			PostTestStatements: []goe2e.TestStatement{
				{"status 200", goe2e.TestStatusCode(200)},
			},
		},
	}
	goe2e.TestRequest(t, tc)
}
