package goe2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfig holds all the necessary function handles to run a full end-2-end test as a unit test.
type TestConfig struct {
	// The name of the test.
	Name string
	// Request specific options, like url, method and body.
	// After applying the options the http.Request will we constructed.
	SpecOpts []SpecOption
	// RequestMods are used to modify the http.Request, like the headers.
	RequestMods []RequestModifier
	// Can be thought of as a general pre-request script.
	PreFunc RequestHandlerModifier
	// Named test statements for running tests before sending the request.
	PreTestStatements []TestStatement
	// Functions to run on the response body after the request was sent and the response received.
	ResponseBodyMods []ResponseBodyModifier
	// ResponseMods are used to modify the http.Response, so anything but the body.
	ResponseMods []ResponseModifier
	// Can be thought of as a general post-request script.
	PostFunc RequestHandlerModifier
	// Named test statements for running tests after the response has been received.
	PostTestStatements []TestStatement
}

type TestStatement struct {
	Description string
	Statement   func(*testing.T, *RequestHandler)
}

// TestStatusCode is a shorthand for asserting a status code on a response.
func TestStatusCode(statusCode int) func(*testing.T, *RequestHandler) {
	return func(t *testing.T, rh *RequestHandler) {
		assert.Equal(t, statusCode, rh.Response.StatusCode)
	}
}

// TestRequest is the main routine for running an E2E test as a unit test.
// It executes the functions passed via the TestConfig with a fixed entry point for each of its field.
func TestRequest(t *testing.T, tc *TestConfig) {
	// create request, checking for nil pointer
	rh, makeErr := NewRequestHandler(WithSpecOpts(tc.SpecOpts...))
	if makeErr != nil {
		t.Errorf("request: %s \nGenerating request failed: %s", tc.Name, makeErr.Error())
		return
	}
	// run request modfications
	modErr := rh.ModifyRequest(tc.RequestMods...)
	if modErr != nil {
		t.Errorf("request: %s \n%s", tc.Name, modErr.Error())
		return
	}
	// run pre-flight "script"
	if tc.PreFunc != nil {
		preErr := tc.PreFunc.Apply(rh)
		if preErr != nil {
			t.Errorf("request: %s \nPre-request function failed: %s", tc.Name, preErr.Error())
			return
		}
	}
	// pre-flight checks
	for _, tt := range tc.PreTestStatements {
		label := fmt.Sprintf("%s/[PRE]/%s", tc.Name, tt.Description)
		t.Run(label, func(t *testing.T) {
			tt.Statement(t, rh)
		})
	}
	// run request
	runErr := rh.RunRequest()
	if runErr != nil {
		t.Errorf("request: %s \nRequest execution failed: %s", tc.Name, runErr.Error())
		return
	}
	// run response body modifications
	modBodyErr := rh.ModifyResponseBody(tc.ResponseBodyMods...)
	if modBodyErr != nil {
		t.Errorf("request: %s \n%s\n", tc.Name, modBodyErr.Error())
		return
	}
	// run response modfications
	modRespErr := rh.ModifyResponse(tc.ResponseMods...)
	if modRespErr != nil {
		t.Errorf("request: %s \n%s", tc.Name, modRespErr.Error())
		return
	}
	// run post-flight "script"
	if tc.PostFunc != nil {
		postErr := tc.PostFunc.Apply(rh)
		if postErr != nil {
			t.Errorf("request: %s \nPost-request function failed: %s", tc.Name, postErr.Error())
			return
		}
	}
	// post-flight checks
	for _, tt := range tc.PostTestStatements {
		label := fmt.Sprintf("%s/[POST]/%s", tc.Name, tt.Description)
		t.Run(label, func(t *testing.T) {
			tt.Statement(t, rh)
		})
	}
}
