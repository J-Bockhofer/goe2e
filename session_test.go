package goe2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type sessionRoundTripper func(*http.Request) (*http.Response, error)

func (f sessionRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSessionPersistsCookiesWithoutMutatingCallerClient(t *testing.T) {
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/sign-in":
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "alice", Path: "/"})
			w.WriteHeader(http.StatusNoContent)
		case "/account":
			cookie, err := request.Cookie("session")
			if err != nil || cookie.Value != "alice" {
				http.Error(w, "missing session cookie", http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, request)
		}
	}))

	client := server.Client()
	if client.Jar != nil {
		t.Fatal("test server client unexpectedly has a cookie jar")
	}
	session, err := NewSession(SessionConfig{HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	if client.Jar != nil {
		t.Fatal("NewSession mutated the caller's client")
	}
	if session.client.Jar == nil {
		t.Fatal("session client has no cookie jar")
	}

	if handler := session.TestRequest(t, RequestConfig{
		Name:     "sign in",
		SpecOpts: []SpecOption{WithURL(server.URL + "/sign-in")},
	}); handler == nil {
		t.Fatal("sign-in request failed")
	}
	account := session.TestRequest(t, RequestConfig{
		Name:     "load account",
		SpecOpts: []SpecOption{WithURL(server.URL + "/account")},
	})
	if account == nil || account.Response.StatusCode != http.StatusNoContent {
		t.Fatal("session cookie was not sent on the second request")
	}
}

func TestSessionPreservesSuppliedClientTransport(t *testing.T) {
	called := false
	client := &http.Client{Transport: sessionRoundTripper(func(request *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
	})}
	session, err := NewSession(SessionConfig{HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	handler := session.TestRequest(t, RequestConfig{SpecOpts: []SpecOption{WithURL("https://service.test/health")}})
	if handler == nil || !called {
		t.Fatal("session did not use the supplied client transport")
	}
}

func TestSessionMergesDefaultsInLifecycleOrder(t *testing.T) {
	var events []string
	record := func(event string) { events = append(events, event) }
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("X-Actor"); got != "request" {
			http.Error(w, "request modifier did not override default", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprint(w, "response")
	}))

	session, err := NewSession(SessionConfig{HTTPClient: server.Client(), Defaults: RequestConfig{
		RequestMods: []RequestModifier{func(request *http.Request) error {
			record("default request mod")
			request.Header.Set("X-Actor", "default")
			return nil
		}},
		PreFunc: RequestHandlerModFunc(func(*RequestHandler) error {
			record("default pre func")
			return nil
		}),
		PreTestStatements: []TestStatement{{Description: "default pre", Statement: func(*testing.T, *RequestHandler) {
			record("default pre statement")
		}}},
		ResponseBodyMods: []ResponseBodyModifier{func(body []byte) ([]byte, error) {
			record("default response body mod")
			return body, nil
		}},
		ResponseMods: []ResponseModifier{func(*http.Response) error {
			record("default response mod")
			return nil
		}},
		PostFunc: RequestHandlerModFunc(func(*RequestHandler) error {
			record("default post func")
			return nil
		}),
		PostTestStatements: []TestStatement{{Description: "default post", Statement: func(*testing.T, *RequestHandler) {
			record("default post statement")
		}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	handler := session.TestRequest(t, RequestConfig{
		Name:     "merged lifecycle",
		SpecOpts: []SpecOption{WithURL(server.URL)},
		RequestMods: []RequestModifier{func(request *http.Request) error {
			record("request request mod")
			request.Header.Set("X-Actor", "request")
			return nil
		}},
		PreFunc: RequestHandlerModFunc(func(*RequestHandler) error {
			record("request pre func")
			return nil
		}),
		PreTestStatements: []TestStatement{{Description: "request pre", Statement: func(*testing.T, *RequestHandler) {
			record("request pre statement")
		}}},
		ResponseBodyMods: []ResponseBodyModifier{func(body []byte) ([]byte, error) {
			record("request response body mod")
			return body, nil
		}},
		ResponseMods: []ResponseModifier{func(*http.Response) error {
			record("request response mod")
			return nil
		}},
		PostFunc: RequestHandlerModFunc(func(*RequestHandler) error {
			record("request post func")
			return nil
		}),
		PostTestStatements: []TestStatement{{Description: "request post", Statement: func(*testing.T, *RequestHandler) {
			record("request post statement")
		}}},
	})
	if handler == nil {
		t.Fatal("session request failed")
	}
	want := []string{
		"default request mod", "request request mod",
		"default pre func", "request pre func",
		"default pre statement", "request pre statement",
		"default response body mod", "request response body mod",
		"default response mod", "request response mod",
		"default post func", "request post func",
		"default post statement", "request post statement",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("unexpected lifecycle order:\n got: %v\nwant: %v", events, want)
	}
}

func TestSessionContextAndTimeoutPrecedence(t *testing.T) {
	type contextKey string
	key := contextKey("session")
	defaultContext := context.WithValue(context.Background(), key, "default")
	requestContext := context.WithValue(context.Background(), key, "request")
	client := &http.Client{Transport: sessionRoundTripper(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
	})}
	session, err := NewSession(SessionConfig{HTTPClient: client, Defaults: RequestConfig{Context: defaultContext, Timeout: time.Second}})
	if err != nil {
		t.Fatal(err)
	}
	defaultHandler := session.TestRequest(t, RequestConfig{SpecOpts: []SpecOption{WithURL("https://service.test/default")}})
	if defaultHandler == nil || defaultHandler.Request().Context().Value(key) != "default" || defaultHandler.Timeout != time.Second {
		t.Fatal("session defaults were not applied")
	}
	requestHandler := session.TestRequest(t, RequestConfig{
		Context:  requestContext,
		Timeout:  2 * time.Second,
		SpecOpts: []SpecOption{WithURL("https://service.test/request")},
	})
	if requestHandler == nil || requestHandler.Request().Context().Value(key) != "request" || requestHandler.Timeout != 2*time.Second {
		t.Fatal("request context and timeout did not override session defaults")
	}
}

func TestSessionDiagnosticsRunInOrderAndRemainSafe(t *testing.T) {
	var calls []string
	server := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Set-Cookie", "session=secret")
		_, _ = fmt.Fprint(w, "private response")
	}))
	checkDiagnostic := func(name string) func(RequestDiagnostic) {
		return func(diagnostic RequestDiagnostic) {
			calls = append(calls, name)
			if _, ok := diagnostic.RequestHeaders["Authorization"]; ok {
				t.Error("diagnostic exposed Authorization header")
			}
			if _, ok := diagnostic.ResponseHeaders["Set-Cookie"]; ok {
				t.Error("diagnostic exposed Set-Cookie header")
			}
			if diagnostic.ResponseBytes != len("private response") {
				t.Errorf("unexpected response byte count: %d", diagnostic.ResponseBytes)
			}
		}
	}
	session, err := NewSession(SessionConfig{HTTPClient: server.Client(), Defaults: RequestConfig{
		OnDiagnostic: checkDiagnostic("default"),
		RequestMods: []RequestModifier{func(request *http.Request) error {
			request.Header.Set("Authorization", "Bearer secret")
			return nil
		}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if handler := session.TestRequest(t, RequestConfig{
		OnDiagnostic: checkDiagnostic("request"),
		SpecOpts:     []SpecOption{WithURL(server.URL)},
	}); handler == nil {
		t.Fatal("session request failed")
	}
	if !reflect.DeepEqual(calls, []string{"default", "request"}) {
		t.Fatalf("unexpected diagnostic callbacks: %v", calls)
	}
}

func TestResponseJSONPointerExtraction(t *testing.T) {
	handler := &RequestHandler{ResponseBody: []byte(`{"id":"request-42","count":2}`)}
	value, err := handler.ResponseJSONPointer("/id")
	if err != nil || value != "request-42" {
		t.Fatalf("unexpected pointer value %v, error %v", value, err)
	}
	identifier, err := handler.ResponseJSONPointerString("/id")
	if err != nil || identifier != "request-42" {
		t.Fatalf("unexpected string pointer value %q, error %v", identifier, err)
	}
	for _, pointer := range []string{"/missing", "/count"} {
		if _, err := handler.ResponseJSONPointerString(pointer); err == nil {
			t.Errorf("expected pointer %q to fail", pointer)
		}
	}
	handler.ResponseBody = []byte(`{`)
	if _, err := handler.ResponseJSONPointer("/id"); err == nil || !strings.Contains(err.Error(), "parse response JSON") {
		t.Fatalf("expected useful malformed JSON error, got %v", err)
	}
	if _, err := (*RequestHandler)(nil).ResponseJSONPointer("/id"); err == nil {
		t.Fatal("nil handler did not return an error")
	}
}

func TestPackageAndSessionTestRequestShareLifecycle(t *testing.T) {
	newRequest := func(events *[]string) RequestConfig {
		return RequestConfig{
			SpecOpts: []SpecOption{WithURL("https://service.test/health")},
			RequestMods: []RequestModifier{func(*http.Request) error {
				*events = append(*events, "request mod")
				return nil
			}},
			PreFunc: RequestHandlerModFunc(func(*RequestHandler) error {
				*events = append(*events, "pre func")
				return nil
			}),
			PostFunc: RequestHandlerModFunc(func(*RequestHandler) error {
				*events = append(*events, "post func")
				return nil
			}),
		}
	}
	client := &http.Client{Transport: sessionRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
	})}
	var packageEvents []string
	TestRequest(t, TestConfig{HTTPClient: client, Request: newRequest(&packageEvents)})
	session, err := NewSession(SessionConfig{HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	var sessionEvents []string
	if handler := session.TestRequest(t, newRequest(&sessionEvents)); handler == nil {
		t.Fatal("session request failed")
	}
	if !reflect.DeepEqual(packageEvents, sessionEvents) {
		t.Fatalf("package and session lifecycle differ: %v != %v", packageEvents, sessionEvents)
	}
}
