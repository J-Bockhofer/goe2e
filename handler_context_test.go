package goe2e_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	goe2e "github.com/J-Bockhofer/goe2e"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestRequestHandlerContext(t *testing.T) {
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("request-id"), "abc-123")
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "abc-123", r.Context().Value(contextKey("request-id")))
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
	})}

	rh, err := goe2e.NewRequestHandler(
		goe2e.WithSpecOpts(goe2e.WithURL("https://service.test/health")),
		goe2e.WithClient(client),
		goe2e.WithContext(ctx),
	)
	require.NoError(t, err)
	require.NoError(t, rh.RunRequest())
	assert.Equal(t, http.StatusNoContent, rh.Response.StatusCode)
}

func TestRequestHandlerTimeout(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}

	rh, err := goe2e.NewRequestHandler(
		goe2e.WithSpecOpts(goe2e.WithURL("https://service.test/slow")),
		goe2e.WithClient(client),
		goe2e.WithTimeout(10*time.Millisecond),
	)
	require.NoError(t, err)
	assert.ErrorIs(t, rh.RunRequest(), context.DeadlineExceeded)
}

func TestRequestHandlerRejectsInvalidContextAndTimeout(t *testing.T) {
	_, err := goe2e.NewRequestHandler(goe2e.WithContext(nil))
	assert.Error(t, err)

	_, err = goe2e.NewRequestHandler(goe2e.WithTimeout(-time.Second))
	assert.Error(t, err)
}
