package goe2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertRequestMethod creates a pre-request statement that checks the request method.
func AssertRequestMethod(expected string) TestStatement {
	return TestStatement{
		Description: fmt.Sprintf("request method is %s", expected),
		Statement: func(t *testing.T, rh *RequestHandler) {
			req := rh.Request()
			if assert.NotNil(t, req) {
				assert.Equal(t, expected, req.Method)
			}
		},
	}
}

// AssertRequestPath creates a pre-request statement that checks the request URL path.
func AssertRequestPath(expected string) TestStatement {
	return TestStatement{
		Description: fmt.Sprintf("request path is %s", expected),
		Statement: func(t *testing.T, rh *RequestHandler) {
			req := rh.Request()
			if assert.NotNil(t, req) {
				assert.Equal(t, expected, req.URL.Path)
			}
		},
	}
}

// AssertRequestHeader creates a pre-request statement that checks a request header.
func AssertRequestHeader(key, expected string) TestStatement {
	return TestStatement{
		Description: fmt.Sprintf("request header %s is %q", key, expected),
		Statement: func(t *testing.T, rh *RequestHandler) {
			req := rh.Request()
			if assert.NotNil(t, req) {
				assert.Equal(t, expected, req.Header.Get(key))
			}
		},
	}
}

// AssertStatusCode creates a post-request statement that checks the response status code.
func AssertStatusCode(expected int) TestStatement {
	return TestStatement{
		Description: fmt.Sprintf("response status is %d", expected),
		Statement: func(t *testing.T, rh *RequestHandler) {
			if assert.NotNil(t, rh.Response) {
				assert.Equal(t, expected, rh.Response.StatusCode)
			}
		},
	}
}

// AssertResponseHeader creates a post-request statement that checks a response header.
func AssertResponseHeader(key, expected string) TestStatement {
	return TestStatement{
		Description: fmt.Sprintf("response header %s is %q", key, expected),
		Statement: func(t *testing.T, rh *RequestHandler) {
			if assert.NotNil(t, rh.Response) {
				assert.Equal(t, expected, rh.Response.Header.Get(key))
			}
		},
	}
}

// AssertResponseBodyContains creates a post-request statement that checks response body content.
func AssertResponseBodyContains(expected string) TestStatement {
	return TestStatement{
		Description: fmt.Sprintf("response body contains %q", expected),
		Statement: func(t *testing.T, rh *RequestHandler) {
			assert.Contains(t, string(rh.ResponseBody), expected)
		},
	}
}

// AssertResponseJSONEquals creates a post-request statement that compares JSON documents semantically.
func AssertResponseJSONEquals(expected string) TestStatement {
	return TestStatement{
		Description: "response JSON matches expected document",
		Statement: func(t *testing.T, rh *RequestHandler) {
			assert.JSONEq(t, expected, string(rh.ResponseBody))
		},
	}
}

// TestStatusCode is a shorthand for asserting a status code on a response.
// Deprecated: use AssertStatusCode in PostTestStatements.
func TestStatusCode(statusCode int) func(*testing.T, *RequestHandler) {
	return func(t *testing.T, rh *RequestHandler) {
		if assert.NotNil(t, rh.Response) {
			assert.Equal(t, statusCode, rh.Response.StatusCode)
		}
	}
}
