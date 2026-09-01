# API reference and migration guide

## Core test flow

`TestRequest(t, config)` executes one HTTP request through these phases:

1. Build the request from `SpecOpts`.
2. Apply `RequestMods`.
3. Run `PreFunc` and `PreTestStatements`.
4. Execute the request.
5. Run response body/response modifiers, `PostFunc`, and `PostTestStatements`.

Use `PreTestStatements` for request assertions and `PostTestStatements` for response assertions. Each item is a `TestStatement`, so custom domain-specific checks remain first-class.

## `TestConfig`

| Field | Purpose |
| --- | --- |
| `SpecOpts` | Method, URL, body, and JSON request construction. |
| `RequestMods` | Headers, authentication, cookies, and request tracing. |
| `HTTPClient` | Optional client used to execute the request; use `httptest.NewTestServer(t, handler).Client()` for in-memory tests. |
| `Context`, `Timeout` | Request-scoped cancellation controls. A timeout does not mutate a shared client. |
| `PreFunc`, `PostFunc` | Custom lifecycle hooks. |
| `PreTestStatements`, `PostTestStatements` | Ordered lifecycle-aware assertions. |
| `ResponseBodyMods`, `ResponseMods` | Ordered transformations of the received response. |
| `OnDiagnostic` | Optional safe, structured request summary for logging or metrics. |

## Common statement factories

Use these factories directly in the appropriate statement slice:

```go
PreTestStatements: []goe2e.TestStatement{
	goe2e.AssertRequestMethod(http.MethodPost),
	goe2e.AssertRequestPath("/persons"),
	goe2e.AssertRequestQuery("include", "address"),
	goe2e.AssertRequestHeader("Content-Type", goe2e.ContentHeaderJSON),
	goe2e.AssertRequestJSONEquals(`{"name":"John"}`),
},
PostTestStatements: []goe2e.TestStatement{
	goe2e.AssertStatusCode(http.StatusCreated),
	goe2e.AssertResponseHeader("Location", "/persons/42"),
	goe2e.AssertResponseBodyContains("created"),
	goe2e.AssertResponseJSONEquals(`{"id":"42"}`),
	goe2e.AssertResponseJSONPointer("/id", "42"),
},
```

`AssertResponseJSONPointer` follows [RFC 6901 JSON Pointer](https://www.rfc-editor.org/rfc/rfc6901), including array indexes such as `/data/items/0/id`.

## Request modifiers

`WithHeaders`, `WithContentType`, `WithBearerToken`, `WithBasicAuth`, and `WithCookies` compose in `RequestMods`. `WithTimeToFirstByte` attaches an `httptrace` timing trace to the request.

## Migration notes

### Move from externally running tests to in-memory handler tests

Keep the same request configuration and provide the server client:

```go
server := httptest.NewTestServer(t, app.Handler())

config.HTTPClient = server.Client()
config.SpecOpts = append(config.SpecOpts,
	goe2e.WithURL("https://app.test/persons"),
)
```

Do not reuse `server.Client()` for the application’s own outbound dependencies. Inject those clients/endpoints separately, including Testcontainers endpoints.

### Replace `TestStatusCode`

`TestStatusCode` remains available for compatibility, but new code should use the phase-aware factory:

```go
// Before
PostTestStatements: []goe2e.TestStatement{
	{"status 201", goe2e.TestStatusCode(http.StatusCreated)},
},

// After
PostTestStatements: []goe2e.TestStatement{
	goe2e.AssertStatusCode(http.StatusCreated),
},
```

### Rename stale documentation references

Use `SpecOption`, not the obsolete README spelling `RequestSpecOption`.
