package goe2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// RequestConfig defines one HTTP request and its full assertion lifecycle.
type RequestConfig struct {
	// The name of the test.
	Name string
	// Context is applied to the request before pre-request statements run.
	// When nil, the request's existing context is used.
	Context context.Context
	// Timeout limits a single request without modifying HTTPClient. A zero timeout is unlimited.
	Timeout time.Duration
	// OnDiagnostic receives a safe, structured summary after the test request finishes.
	// It omits sensitive headers and response-body content.
	OnDiagnostic func(RequestDiagnostic)
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

// TestConfig combines one request with the client used to execute it.
// Supplying the client from httptest.NewTestServer keeps the request in memory.
type TestConfig struct {
	HTTPClient *http.Client
	Request    RequestConfig
}

type TestStatement struct {
	Description string
	Statement   func(*testing.T, *RequestHandler)
}

// TestRequest is the main routine for running an E2E test as a unit test.
// It executes the functions passed via the TestConfig with a fixed entry point for each of its field.
func TestRequest(t *testing.T, tc TestConfig) *RequestHandler {
	return runTestRequest(t, tc)
}

// runTestRequest executes a test configuration and returns its handler when setup and execution succeed.
// Assertion failures remain non-fatal, matching TestRequest's existing behavior.
func runTestRequest(t *testing.T, tc TestConfig) *RequestHandler {
	request := tc.Request
	// create request, checking for nil pointer
	rhOpts := []RequestHandlerOption{WithSpecOpts(request.SpecOpts...)}
	if tc.HTTPClient != nil {
		rhOpts = append(rhOpts, WithClient(tc.HTTPClient))
	}
	if request.Context != nil {
		rhOpts = append(rhOpts, WithContext(request.Context))
	}
	if request.Timeout != 0 {
		rhOpts = append(rhOpts, WithTimeout(request.Timeout))
	}
	rh, makeErr := NewRequestHandler(rhOpts...)
	if makeErr != nil {
		t.Errorf("request: %s \nGenerating request failed: %s", request.Name, makeErr.Error())
		return nil
	}
	if request.OnDiagnostic != nil {
		defer func() {
			request.OnDiagnostic(rh.Diagnostic())
		}()
	}
	// run request modfications
	modErr := rh.ModifyRequest(request.RequestMods...)
	if modErr != nil {
		t.Errorf("request: %s \n%s", request.Name, modErr.Error())
		return nil
	}
	// run pre-flight "script"
	if request.PreFunc != nil {
		preErr := request.PreFunc.Apply(rh)
		if preErr != nil {
			t.Errorf("request: %s \nPre-request function failed: %s", request.Name, preErr.Error())
			return nil
		}
	}
	// pre-flight checks
	for _, tt := range request.PreTestStatements {
		label := fmt.Sprintf("%s/[PRE]/%s", request.Name, tt.Description)
		t.Run(label, func(t *testing.T) {
			tt.Statement(t, rh)
		})
	}
	// run request
	runErr := rh.RunRequest()
	if runErr != nil {
		t.Errorf("request: %s\nRequest execution failed: %s\n%s", request.Name, runErr.Error(), requestDiagnostics(rh))
		return nil
	}
	// run response body modifications
	modBodyErr := rh.ModifyResponseBody(request.ResponseBodyMods...)
	if modBodyErr != nil {
		t.Errorf("request: %s\n%s\n%s", request.Name, modBodyErr.Error(), requestDiagnostics(rh))
		return nil
	}
	// run response modfications
	modRespErr := rh.ModifyResponse(request.ResponseMods...)
	if modRespErr != nil {
		t.Errorf("request: %s\n%s\n%s", request.Name, modRespErr.Error(), requestDiagnostics(rh))
		return nil
	}
	// run post-flight "script"
	if request.PostFunc != nil {
		postErr := request.PostFunc.Apply(rh)
		if postErr != nil {
			t.Errorf("request: %s \nPost-request function failed: %s", request.Name, postErr.Error())
			return nil
		}
	}
	// post-flight checks
	for _, tt := range request.PostTestStatements {
		label := fmt.Sprintf("%s/[POST]/%s", request.Name, tt.Description)
		t.Run(label, func(t *testing.T) {
			tt.Statement(t, rh)
		})
	}
	return rh
}
