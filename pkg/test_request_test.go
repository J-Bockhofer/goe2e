package goe2e_test

import (
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e/pkg"
)

func TestTestRequest(t *testing.T) {
	tc := &goe2e.TestConfig{
		Name: "",
		SpecOpts: []goe2e.SpecOption{
			goe2e.WithUrl("https://www.github.com"),
		},
		PostTestStatements: []goe2e.TestStatement{
			{"status 200", goe2e.TestStatusCode(200)},
		},
	}
	goe2e.TestRequest(t, tc)
}

func TestTestRequestWithTimings(t *testing.T) {
	tc := &goe2e.TestConfig{
		Name: "",
		SpecOpts: []goe2e.SpecOption{
			goe2e.WithUrl("https://www.github.com"),
		},
		RequestMods: []goe2e.RequestModifier{
			goe2e.WithTimeToFirstByte(),
		},
		PostTestStatements: []goe2e.TestStatement{
			{"status 200", goe2e.TestStatusCode(200)},
		},
	}
	goe2e.TestRequest(t, tc)
}
