package goe2e_test

import (
	"encoding/json"
	"net/http"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e/pkg"

	"github.com/stretchr/testify/assert"
)

const (
	defaultUrl = "https://example.com"
)

var (
	env = goe2e.H{
		"personName": "jamie",
		"baseUrl":    "www.myurl.com",
	}
	keymap = goe2e.D{
		"personName": "name",
	}
)

func TestNewSpecWith(t *testing.T) {
	payload := goe2e.H{
		"name": "john",
	}
	payloadInBytes, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("could not marshal test payload: %s", err.Error())
	}
	type args struct {
		opts []goe2e.SpecOption
	}

	type expected struct {
		err    error
		method string
		url    string
		body   []byte
	}

	testCases := []struct {
		name string
		args
		expected
	}{
		{"Default", args{
			[]goe2e.SpecOption{},
		}, expected{
			nil,
			http.MethodGet,
			defaultUrl,
			[]byte{},
		}},
		{"WithJSON", args{
			[]goe2e.SpecOption{goe2e.WithJSON(&payload)},
		}, expected{
			nil,
			http.MethodGet,
			defaultUrl,
			payloadInBytes,
		}},
		{"WithUrl", args{
			[]goe2e.SpecOption{goe2e.WithUrl("gallo")},
		}, expected{
			nil,
			http.MethodGet,
			"gallo",
			[]byte{},
		}},
		{"WithMethod", args{
			[]goe2e.SpecOption{goe2e.WithMethod(http.MethodPost)},
		}, expected{
			nil,
			http.MethodPost,
			defaultUrl,
			[]byte{},
		}},
		{"WithBody", args{
			[]goe2e.SpecOption{goe2e.WithBody([]byte("hello"))},
		}, expected{
			nil,
			http.MethodGet,
			defaultUrl,
			[]byte("hello"),
		}},
		{"WithBody + WithMethod", args{
			[]goe2e.SpecOption{goe2e.WithBody([]byte("hello")), goe2e.WithMethod(http.MethodPost)},
		}, expected{
			nil,
			http.MethodPost,
			defaultUrl,
			[]byte("hello"),
		}},
		{"WithBaseUrlFromEnv", args{
			[]goe2e.SpecOption{goe2e.WithBaseURLFromEnv(env, "baseUrl", "persons")},
		}, expected{
			nil,
			http.MethodGet,
			"www.myurl.com/persons",
			[]byte{},
		}},
		{"WithBaseUrlFromEnv with /", args{
			[]goe2e.SpecOption{goe2e.WithBaseURLFromEnv(env, "baseUrl", "/persons")},
		}, expected{
			nil,
			http.MethodGet,
			"www.myurl.com/persons",
			[]byte{},
		}},
		{"WithRouteFromQueryMap", args{
			[]goe2e.SpecOption{goe2e.WithRouteFromQueryMap("www.myurl.com/persons", keymap)},
		}, expected{
			nil,
			http.MethodGet,
			"www.myurl.com/persons?personName=name",
			[]byte{},
		}},
		{"WithRouteFromQueryMap", args{
			[]goe2e.SpecOption{goe2e.AddQueryFromMap(keymap)},
		}, expected{
			nil,
			http.MethodGet,
			defaultUrl + "?personName=name",
			[]byte{},
		}},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := goe2e.NewSpec(tt.opts...)
			assert.Equal(t, tt.err, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.method, spec.Method)
			assert.Equal(t, tt.url, spec.Url)
			assert.Equal(t, tt.body, spec.Body)
			assert.NotNil(t, spec.Request)
		})
	}
}

func TestSpecWithEnv(t *testing.T) {
	payload := goe2e.H{
		"name": "john",
	}
	payloadInBytes, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("could not marshal test payload: %s", err.Error())
	}

	t.Run("WithBody+WithSetFromEnv", func(t *testing.T) {
		spec, err := goe2e.NewSpec(
			goe2e.WithBody(payloadInBytes), goe2e.WithSetFromEnv(env, keymap),
		)
		if err != nil {
			t.Errorf("failed to create spec: %s", err.Error())
		}
		var jsonBody goe2e.H
		err = json.Unmarshal(spec.Body, &jsonBody)
		if err != nil {
			t.Errorf("failed to unmarshal body: %s", err.Error())
		}
		assert.Equal(t, goe2e.H{"name": "jamie"}, jsonBody)
	})
}

func TestRequestModifier(t *testing.T) {
	spec, err := goe2e.NewSpec()
	if err != nil {
		t.Errorf("could not make a default spec: %s", err.Error())
	}

	t.Run("Modifiy content header", func(t *testing.T) {
		err := goe2e.WithContentType(goe2e.ContentHeaderJSON)(spec.Request)
		if err != nil {
			t.Errorf("Modifiy content header failed: %s", err.Error())
		}
		actual := spec.Request.Header.Get("Content-Type")
		assert.Equal(t, goe2e.ContentHeaderJSON, actual)
	})

	t.Run("WithHeaders", func(t *testing.T) {
		err := goe2e.WithHeaders(goe2e.D{"HELLO": "BYE"})(spec.Request)
		if err != nil {
			t.Errorf("With header failed: %s", err.Error())
		}
		actual := spec.Request.Header.Get("HELLO")
		assert.Equal(t, "BYE", actual)
	})
}
