# API reference and migration guide

## Core test flow

`TestRequest(t, config)` executes one HTTP request through these phases:

1. Build the request from `SpecOpts`.
2. Apply `RequestMods`.
3. Run `PreFunc` and `PreTestStatements`.
4. Execute the request.
5. Run response body/response modifiers, `PostFunc`, and `PostTestStatements`.

Use `PreTestStatements` for request assertions and `PostTestStatements` for response assertions. Each item is a `TestStatement`, so custom domain-specific checks remain first-class.

## `RequestConfig`

| Field | Purpose |
| --- | --- |
| `SpecOpts` | Method, URL, body, and JSON request construction. |
| `RequestMods` | Headers, authentication, cookies, and request tracing. |
| `Context`, `Timeout` | Request-scoped cancellation controls. A timeout does not mutate a shared client. |
| `PreFunc`, `PostFunc` | Custom lifecycle hooks. |
| `PreTestStatements`, `PostTestStatements` | Ordered lifecycle-aware assertions. |
| `ResponseBodyMods`, `ResponseMods` | Ordered transformations of the received response. |
| `OnDiagnostic` | Optional safe, structured request summary for logging or metrics. |

## `TestConfig`

`TestConfig` supplies the execution client and a `RequestConfig`:

```go
goe2e.TestRequest(t, goe2e.TestConfig{
	HTTPClient: server.Client(),
	Request: goe2e.RequestConfig{
		Name:     "create person",
		SpecOpts: []goe2e.SpecOption{goe2e.WithURL(server.URL + "/persons")},
	},
})
```

## Common statement factories

Use these factories directly in the appropriate statement slice:

```go
PreTestStatements: []goe2e.TestStatement{
	goe2e.AssertRequestMethod(http.MethodPost),
	goe2e.AssertRequestPath("/persons"),
	goe2e.AssertRequestQuery("include", "address"),
	goe2e.AssertRequestHeader("Content-Type", goe2e.ContentHeaderJSON),
	goe2e.AssertRequestBodyEquals(`raw request body`),
	goe2e.AssertRequestJSONEquals(`{"name":"John"}`),
},
PostTestStatements: []goe2e.TestStatement{
	goe2e.AssertStatusCode(http.StatusCreated),
	goe2e.AssertResponseHeader("Location", "/persons/42"),
	goe2e.AssertResponseBodyContains("created"),
	goe2e.AssertResponseBodyEquals(`created`),
	goe2e.AssertResponseJSONEquals(`{"id":"42"}`),
	goe2e.AssertResponseJSONPointer("/id", "42"),
},
```

`AssertResponseJSONPointer` follows [RFC 6901 JSON Pointer](https://www.rfc-editor.org/rfc/rfc6901), including array indexes such as `/data/items/0/id`.
`AssertRequestBodyEquals` and `AssertResponseBodyEquals` compare bytes as text; use the JSON variants when object-key order should not matter.

## Request modifiers

`WithHeaders`, `WithContentType`, `WithBearerToken`, `WithBasicAuth`, and `WithCookies` compose in `RequestMods`. `WithTimeToFirstByte` attaches an `httptrace` timing trace to the request.

## Stateful sessions

`TestRequest` remains the concise API for one request. For a multi-step workflow, create a `Session` for each browser or user actor and call `Session.TestRequest`. A session clones the supplied `http.Client`, preserving its transport, and creates a cookie jar when the client does not already have one. It is not safe for concurrent use.

```go
session, err := goe2e.NewSession(goe2e.SessionConfig{
	HTTPClient: server.Client(),
	Defaults: goe2e.RequestConfig{
		RequestMods: []goe2e.RequestModifier{
			goe2e.WithHeaders(goe2e.D{"X-User-ID": userID}),
		},
	},
})
if err != nil {
	t.Fatal(err)
}

created := session.TestRequest(t, goe2e.RequestConfig{
	SpecOpts: []goe2e.SpecOption{
		goe2e.WithMethod(http.MethodPost),
		goe2e.WithURL(server.URL + "/resources"),
		goe2e.WithJSON(payload),
	},
})
if created == nil {
	t.Fatal("request failed")
}
id, err := created.ResponseJSONPointerString("/id")
```

Session defaults merge before request-specific values. Request-specific `Context` and non-zero `Timeout` override defaults; request-specific modifiers, statements, and response modifiers run after defaults. Default and request diagnostics both run, in that order. `Defaults.Name` and `Defaults.SpecOpts` must be empty because both always belong to an individual request. Set a custom cookie jar with `SessionConfig.CookieJar`; otherwise the session uses the supplied client's jar or creates one.

`ResponseJSONPointer` returns an `any` value selected by an RFC 6901 pointer; `ResponseJSONPointerString` additionally verifies that the selected value is a string.

## Migration notes

### Import from the module root

The public package now lives at the module root. Update existing imports:

```go
// Before
goe2e "github.com/J-Bockhofer/goe2e/pkg"

// After
goe2e "github.com/J-Bockhofer/goe2e"
```

### Separate request definitions from execution clients

`TestConfig` now contains `HTTPClient` and a `RequestConfig`. Use `RequestConfig` directly with `Session.TestRequest`:

```go
// Before
goe2e.TestRequest(t, &goe2e.TestConfig{
	HTTPClient: server.Client(),
	SpecOpts:   []goe2e.SpecOption{goe2e.WithURL(server.URL + "/health")},
})

// After
goe2e.TestRequest(t, goe2e.TestConfig{
	HTTPClient: server.Client(),
	Request: goe2e.RequestConfig{
		SpecOpts: []goe2e.SpecOption{goe2e.WithURL(server.URL + "/health")},
	},
})
```

### Move from externally running tests to in-memory handler tests

Keep the same request configuration and provide the server client:

```go
server := httptest.NewTestServer(t, app.Handler())

config.HTTPClient = server.Client()
config.Request.SpecOpts = append(config.Request.SpecOpts,
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
