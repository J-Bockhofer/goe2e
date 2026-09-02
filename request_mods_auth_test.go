package goe2e_test

import (
	"net/http"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationAndCookieModifiers(t *testing.T) {
	t.Run("bearer token", func(t *testing.T) {
		rh, err := goe2e.NewRequestHandler(goe2e.WithSpecOpts())
		require.NoError(t, err)
		require.NoError(t, rh.ModifyRequest(goe2e.WithBearerToken("token-123")))
		assert.Equal(t, "Bearer token-123", rh.Request().Header.Get("Authorization"))
	})

	t.Run("basic authentication", func(t *testing.T) {
		rh, err := goe2e.NewRequestHandler(goe2e.WithSpecOpts())
		require.NoError(t, err)
		require.NoError(t, rh.ModifyRequest(goe2e.WithBasicAuth("john", "s3cret")))

		username, password, ok := rh.Request().BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "john", username)
		assert.Equal(t, "s3cret", password)
	})

	t.Run("cookies", func(t *testing.T) {
		rh, err := goe2e.NewRequestHandler(goe2e.WithSpecOpts())
		require.NoError(t, err)
		require.NoError(t, rh.ModifyRequest(goe2e.WithCookies(
			&http.Cookie{Name: "session", Value: "abc"},
			&http.Cookie{Name: "theme", Value: "dark"},
		)))

		assert.Equal(t, []*http.Cookie{
			{Name: "session", Value: "abc"},
			{Name: "theme", Value: "dark"},
		}, rh.Request().Cookies())
	})

	t.Run("nil cookie is rejected", func(t *testing.T) {
		rh, err := goe2e.NewRequestHandler(goe2e.WithSpecOpts())
		require.NoError(t, err)
		assert.Error(t, rh.ModifyRequest(goe2e.WithCookies(nil)))
	})
}
