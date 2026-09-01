package goe2e

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
				assert.Equalf(t, expected, req.Method, "%s", requestDiagnostics(rh))
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
				assert.Equalf(t, expected, req.URL.Path, "%s", requestDiagnostics(rh))
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
				assert.Equalf(t, expected, req.Header.Get(key), "%s", requestDiagnostics(rh))
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
				assert.Equalf(t, expected, rh.Response.StatusCode, "%s", requestDiagnostics(rh))
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
				assert.Equalf(t, expected, rh.Response.Header.Get(key), "%s", requestDiagnostics(rh))
			}
		},
	}
}

// AssertResponseBodyContains creates a post-request statement that checks response body content.
func AssertResponseBodyContains(expected string) TestStatement {
	return TestStatement{
		Description: fmt.Sprintf("response body contains %q", expected),
		Statement: func(t *testing.T, rh *RequestHandler) {
			assert.Containsf(t, string(rh.ResponseBody), expected, "%s", requestDiagnostics(rh))
		},
	}
}

// AssertResponseJSONEquals creates a post-request statement that compares JSON documents semantically.
func AssertResponseJSONEquals(expected string) TestStatement {
	return TestStatement{
		Description: "response JSON matches expected document",
		Statement: func(t *testing.T, rh *RequestHandler) {
			assert.JSONEqf(t, expected, string(rh.ResponseBody), "%s", requestDiagnostics(rh))
		},
	}
}

// AssertResponseJSONPointer creates a post-request statement that checks a value selected by an RFC 6901 JSON Pointer.
// For example, /data/persons/0/name selects the name field of the first person.
func AssertResponseJSONPointer(pointer string, expected any) TestStatement {
	return TestStatement{
		Description: fmt.Sprintf("response JSON pointer %q matches expected value", pointer),
		Statement: func(t *testing.T, rh *RequestHandler) {
			var document any
			if !assert.NoErrorf(t, json.Unmarshal(rh.ResponseBody, &document), "%s", requestDiagnostics(rh)) {
				return
			}
			actual, err := jsonPointerValue(document, pointer)
			if !assert.NoErrorf(t, err, "%s", requestDiagnostics(rh)) {
				return
			}
			assert.EqualValuesf(t, expected, actual, "%s", requestDiagnostics(rh))
		},
	}
}

func jsonPointerValue(document any, pointer string) (any, error) {
	if pointer == "" {
		return document, nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("JSON pointer %q must start with /", pointer)
	}

	current := document
	for _, token := range strings.Split(pointer[1:], "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch value := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = value[token]
			if !ok {
				return nil, fmt.Errorf("JSON pointer %q does not exist", pointer)
			}
		case []any:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(value) {
				return nil, fmt.Errorf("JSON pointer %q has invalid array index %q", pointer, token)
			}
			current = value[index]
		default:
			return nil, fmt.Errorf("JSON pointer %q cannot descend into %T", pointer, current)
		}
	}
	return current, nil
}

// TestStatusCode is a shorthand for asserting a status code on a response.
// Deprecated: use AssertStatusCode in PostTestStatements.
func TestStatusCode(statusCode int) func(*testing.T, *RequestHandler) {
	return func(t *testing.T, rh *RequestHandler) {
		if assert.NotNil(t, rh.Response) {
			assert.Equalf(t, statusCode, rh.Response.StatusCode, "%s", requestDiagnostics(rh))
		}
	}
}
