package goe2e

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

// Session represents one stateful HTTP actor. It preserves the client's transport and cookie jar
// across requests. A Session is not safe for concurrent use.
type Session struct {
	client   *http.Client
	defaults TestConfig
}

type sessionConfig struct {
	defaults TestConfig
	jar      http.CookieJar
}

// SessionOption configures a Session.
type SessionOption func(*sessionConfig) error

// WithCookieJar sets the cookie jar used by a Session.
func WithCookieJar(jar http.CookieJar) SessionOption {
	return func(config *sessionConfig) error {
		if jar == nil {
			return fmt.Errorf("session cookie jar must not be nil")
		}
		config.jar = jar
		return nil
	}
}

// WithDefaults sets reusable TestConfig values for a Session. Request specification options and
// HTTPClient are request-specific and are not inherited from session defaults.
func WithDefaults(defaults TestConfig) SessionOption {
	return func(config *sessionConfig) error {
		config.defaults = cloneTestConfig(defaults)
		return nil
	}
}

// NewSession creates a stateful client for one HTTP actor. It copies the supplied client so the
// caller's client is not mutated. A nil client uses a copy of http.DefaultClient.
func NewSession(client *http.Client, options ...SessionOption) (*Session, error) {
	if client == nil {
		client = http.DefaultClient
	}
	clientCopy := *client
	config := sessionConfig{}
	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("session option must not be nil")
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}
	if config.jar != nil {
		clientCopy.Jar = config.jar
	} else if clientCopy.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("create session cookie jar: %w", err)
		}
		clientCopy.Jar = jar
	}
	return &Session{client: &clientCopy, defaults: config.defaults}, nil
}

// TestRequest executes config using the Session's persistent client. HTTPClient must not be set
// on config because it would bypass the Session's cookie jar and transport state. It returns nil
// when setup or request execution fails.
func (s *Session) TestRequest(t *testing.T, config *TestConfig) *RequestHandler {
	if s == nil {
		t.Error("session must not be nil")
		return nil
	}
	if config == nil {
		t.Error("session test request configuration must not be nil")
		return nil
	}
	if config.HTTPClient != nil {
		t.Error("TestConfig.HTTPClient must not be set when calling Session.TestRequest; the client belongs to the Session")
		return nil
	}
	merged := mergeSessionTestConfig(s.defaults, *config)
	merged.HTTPClient = s.client
	return runTestRequest(t, &merged)
}

func mergeSessionTestConfig(defaults, request TestConfig) TestConfig {
	merged := cloneTestConfig(request)
	merged.Name = request.Name
	merged.HTTPClient = nil
	merged.SpecOpts = append([]SpecOption(nil), request.SpecOpts...)
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

func cloneTestConfig(config TestConfig) TestConfig {
	return TestConfig{
		Name:               config.Name,
		HTTPClient:         config.HTTPClient,
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
