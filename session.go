package goe2e

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

// SessionConfig configures a stateful HTTP Session.
type SessionConfig struct {
	HTTPClient *http.Client
	CookieJar  http.CookieJar
	Defaults   RequestConfig
}

// Session represents one stateful HTTP actor. It preserves the client's transport and cookie jar
// across requests. A Session is not safe for concurrent use.
type Session struct {
	client   *http.Client
	defaults RequestConfig
}

// NewSession creates a stateful client for one HTTP actor. It copies the supplied client so the
// caller's client is not mutated. CookieJar overrides the client's jar; when neither is supplied,
// the Session creates a new jar. A nil HTTPClient uses a copy of http.DefaultClient.
func NewSession(config SessionConfig) (*Session, error) {
	if config.Defaults.Name != "" {
		return nil, fmt.Errorf("session default request name must be empty")
	}
	if len(config.Defaults.SpecOpts) != 0 {
		return nil, fmt.Errorf("session default request SpecOpts must be empty")
	}
	client := config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	clientCopy := *client
	if config.CookieJar != nil {
		clientCopy.Jar = config.CookieJar
	} else if clientCopy.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("create session cookie jar: %w", err)
		}
		clientCopy.Jar = jar
	}
	return &Session{client: &clientCopy, defaults: cloneRequestConfig(config.Defaults)}, nil
}

// TestRequest executes request using the Session's persistent client. It returns nil when setup
// or request execution fails.
func (s *Session) TestRequest(t *testing.T, request RequestConfig) *RequestHandler {
	if s == nil {
		t.Error("session must not be nil")
		return nil
	}
	merged := mergeSessionRequestConfig(s.defaults, request)
	return runTestRequest(t, TestConfig{HTTPClient: s.client, Request: merged})
}

func mergeSessionRequestConfig(defaults, request RequestConfig) RequestConfig {
	merged := cloneRequestConfig(request)
	if merged.Context == nil {
		merged.Context = defaults.Context
	}
	if merged.Timeout == 0 {
		merged.Timeout = defaults.Timeout
	}
	merged.OnDiagnostic = combineDiagnostics(defaults.OnDiagnostic, request.OnDiagnostic)
	merged.RequestMods = appendSlices(defaults.RequestMods, request.RequestMods)
	merged.PreTestStatements = appendSlices(defaults.PreTestStatements, request.PreTestStatements)
	merged.ResponseBodyMods = appendSlices(defaults.ResponseBodyMods, request.ResponseBodyMods)
	merged.ResponseMods = appendSlices(defaults.ResponseMods, request.ResponseMods)
	merged.PostTestStatements = appendSlices(defaults.PostTestStatements, request.PostTestStatements)
	merged.PreFunc = combineHandlerModifiers(defaults.PreFunc, request.PreFunc)
	merged.PostFunc = combineHandlerModifiers(defaults.PostFunc, request.PostFunc)
	return merged
}

func cloneRequestConfig(config RequestConfig) RequestConfig {
	return RequestConfig{
		Name:               config.Name,
		Context:            config.Context,
		Timeout:            config.Timeout,
		OnDiagnostic:       config.OnDiagnostic,
		SpecOpts:           append([]SpecOption(nil), config.SpecOpts...),
		RequestMods:        append([]RequestModifier(nil), config.RequestMods...),
		PreFunc:            config.PreFunc,
		PreTestStatements:  append([]TestStatement(nil), config.PreTestStatements...),
		ResponseBodyMods:   append([]ResponseBodyModifier(nil), config.ResponseBodyMods...),
		ResponseMods:       append([]ResponseModifier(nil), config.ResponseMods...),
		PostFunc:           config.PostFunc,
		PostTestStatements: append([]TestStatement(nil), config.PostTestStatements...),
	}
}

func appendSlices[T any](first, second []T) []T {
	values := make([]T, 0, len(first)+len(second))
	values = append(values, first...)
	return append(values, second...)
}

func combineDiagnostics(first, second func(RequestDiagnostic)) func(RequestDiagnostic) {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	return func(diagnostic RequestDiagnostic) {
		first(diagnostic)
		second(diagnostic)
	}
}

func combineHandlerModifiers(first, second RequestHandlerModifier) RequestHandlerModifier {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	return RequestHandlerModFunc(func(handler *RequestHandler) error {
		if err := first.Apply(handler); err != nil {
			return err
		}
		return second.Apply(handler)
	})
}
