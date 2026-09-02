# goe2e roadmap

`goe2e` should make HTTP integration tests concise and deterministic while keeping the HTTP request lifecycle visible. It is not intended to replace Go's `testing`, `net/http`, or `net/http/httptest` packages.

## Design principles

- Keep assertions phase-aware: request assertions belong in `PreTestStatements`; response assertions belong in `PostTestStatements`.
- Support both in-memory handler tests and requests to an externally running service through the same `TestConfig`.
- Build on `net/http` and Testify's non-fatal `assert` package. Consumers can always add arbitrary `TestStatement` functions.
- Add convenience for repeated HTTP-test work, not a second test framework.

## Milestone 1: dependable core API

- [x] Allow callers to supply an `HTTPClient`, including the client returned by Go 1.27's `httptest.NewTestServer`.
- [x] Add lifecycle-aware statement factories for common request and response assertions.
- [x] Add context and per-request timeout configuration.
- [x] Improve request failures with method, URL, status, selected headers, and a safely truncated response body.
- [x] Make empty response/body modification lists explicit no-ops.

## Milestone 2: HTTP assertions that remove boilerplate

- [x] Request method, path, query, and header statements.
- [x] Response status, header, body-substring, and JSON-equality statements.
- [x] JSON-pointer/value statements using RFC 6901, avoiding a non-standard JSONPath dependency.
- [x] Authentication and cookie request modifiers.
- [x] Request and response JSON helpers with useful mismatch diffs.

## Milestone 3: deterministic integration-test workflows

- [x] Add a complete, runnable example for a standard `net/http` handler.
- [x] Add a complete Gin example without making it a library dependency.
- [x] Exercise HTTP, HTTPS, headers, query parameters, malformed JSON, and handler panics with `httptest.NewTestServer`.
- [x] Document when to use the in-memory client versus a real running service.
- [x] Offer opt-in structured, redacted request/response diagnostics.

## Milestone 4: library readiness

- [x] Compile runnable in-repository consumer examples in CI with `go test ./...`.
- [ ] Validate all README snippets in CI; framework-specific snippets need an example module or snippet extractor first.
- [x] Run CI with the Go version declared in `go.mod` and vet every package.
- [x] Document compatibility and release/versioning policy.
- [x] Publish a concise API reference and migration guide.

## Current preferred style

```go
server := httptest.NewTestServer(t, app.Router())

goe2e.TestRequest(t, &goe2e.TestConfig{
	Name:       "POST /persons",
	HTTPClient: server.Client(),
	SpecOpts: []goe2e.SpecOption{
		goe2e.WithMethod(http.MethodPost),
		goe2e.WithURL("https://app.test/persons"),
		goe2e.WithJSON(person),
	},
	PreTestStatements: []goe2e.TestStatement{
		goe2e.AssertRequestMethod(http.MethodPost),
		goe2e.AssertRequestHeader("Content-Type", goe2e.ContentHeaderJSON),
		goe2e.AssertRequestJSONEquals(`{"name":"John"}`),
	},
	PostTestStatements: []goe2e.TestStatement{
		goe2e.AssertStatusCode(http.StatusCreated),
		goe2e.AssertResponseHeader("Location", "/persons/42"),
		goe2e.AssertResponseJSONEquals(`{"id":"42","name":"John"}`),
		goe2e.AssertResponseJSONPointer("/id", "42"),
	},
})
```
