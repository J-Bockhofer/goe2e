# goe2e - E2E testing in Go

Writing server in Go? Write E2E (End-to-End) tests in it, too!

Just add a test file to your source or an empty application and get going.

## How it works

Simply put, E2E tests work by sending a request to a running application.

This proof-of-concept package presents a way to combine the IDE, build pipeline support and relative ease of writing of unit tests with flexible E2E testing.

It works as a convenience wrapper/function adapter around http requests and responses.

The whole configuration is based on typed function handles, so you can write your own functions to feed into the testing frame.

Say we have a simple [gin](https://github.com/gin-gonic/gin) application like this:

```go
package main

import (
	"net/http"
	"goe2e-example/model"

	"github.com/gin-gonic/gin"
)

func runServer() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/persons", func(c *gin.Context) {
		var p model.Person
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bad request",
			})
            return
		}
		c.JSON(http.StatusAccepted, p)
	})
	r.Run()
}

func main() {
	runServer()
}
```

We define the model that we want the POST request to send.

```go
package model

type Person struct {
	Name string `json:"name" binding:"required,min=2"`
	Age  int    `json:"age" binding:"required"`
}
```

Now we can write an e2e test like this:

```go
package model_test

import (
	"encoding/json"
	"goe2e-example/model"
	"net/http"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e"

	"github.com/stretchr/testify/assert"
)

func TestPersonPost(t *testing.T) {
	p := model.Person{
		Name: "john",
		Age:  32,
	}
	rc := goe2e.TestConfig{Request: goe2e.RequestConfig{
		Name: "POST /persons",
		SpecOpts: []goe2e.SpecOption{
			goe2e.WithMethod(http.MethodPost),
			goe2e.WithURL("http://localhost:8080/persons/"),
			goe2e.WithJSON(&p),
		},
		RequestMods: []goe2e.RequestModifier{
			goe2e.WithContentType(goe2e.ContentHeaderJSON),
		},
		PreTestStatements: []goe2e.TestStatement{
			{Description: "request not nil", Statement: func(t *testing.T, r *goe2e.RequestHandler) {
				assert.NotNil(t, r.Request())
			}},
		},
		PostTestStatements: []goe2e.TestStatement{
			{Description: "body not nil", Statement: func(t *testing.T, r *goe2e.RequestHandler) {
				assert.NotNil(t, r.ResponseBody)
			}},
			{Description: "status 202", Statement: func(t *testing.T, r *goe2e.RequestHandler) {
				assert.Equal(t, http.StatusAccepted, r.Response.StatusCode)
			}},
		},
	}}
	goe2e.TestRequest(t, rc)
}
```

`One caveat:` We have to run the main application before we run the E2E test in unit test disguise.

This can be dealt with using environment variables that skip E2E tests / signal that the application is running and execute the tests.

Alternatively [testing.M](https://pkg.go.dev/testing#hdr-Main) provides a space for test setup and teardown functions.

### In-memory server tests (Go 1.27+)

For deterministic tests without a running application, create an in-memory test server and pass its client to `TestConfig`. The client sends requests to the test handler without opening a port or resolving DNS.

```go
func TestPersonPost(t *testing.T) {
	server := httptest.NewTestServer(t, app.Router())

	goe2e.TestRequest(t, goe2e.TestConfig{
		HTTPClient: server.Client(),
		Request: goe2e.RequestConfig{
			Name: "POST /persons",
			SpecOpts: []goe2e.SpecOption{
			goe2e.WithMethod(http.MethodPost),
			goe2e.WithURL("https://service.test/persons"),
			goe2e.WithJSON(&model.Person{Name: "john", Age: 32}),
		},
			PostTestStatements: []goe2e.TestStatement{
			{"status 202", goe2e.TestStatusCode(http.StatusAccepted)},
			},
		},
	})
}
```

`httptest.NewTestServer` registers its own cleanup, so no `defer server.Close()` is necessary. Omitting `HTTPClient` retains the real-network behavior.

For a complete, runnable example using only `net/http`, see [examples/standard_http_test.go](examples/standard_http_test.go).

### Multi-step workflows

Use a `Session` when one test actor makes multiple requests. It keeps the supplied client's transport and cookie jar, so a session created from `httptest.NewTestServer(t).Client()` remains in-memory and carries cookies between calls. Create one session per browser or user actor.

```go
alice, err := goe2e.NewSession(goe2e.SessionConfig{
	HTTPClient: server.Client(),
	Defaults: goe2e.RequestConfig{RequestMods: []goe2e.RequestModifier{
		goe2e.WithHeaders(goe2e.D{"X-User-ID": aliceID}),
	}},
})
if err != nil {
	t.Fatal(err)
}

created := alice.TestRequest(t, goe2e.RequestConfig{
	Name: "create interest request",
	SpecOpts: []goe2e.SpecOption{
		goe2e.WithMethod(http.MethodPost),
		goe2e.WithURL(server.URL + "/interest-requests/" + bobID),
		goe2e.WithJSON(map[string]string{"message": "Hello"}),
	},
	PostTestStatements: []goe2e.TestStatement{
		goe2e.AssertStatusCode(http.StatusCreated),
	},
})
if created == nil {
	t.Fatal("create interest request failed")
}
requestID, err := created.ResponseJSONPointerString("/id")
if err != nil {
	t.Fatal(err)
}

bob, err := goe2e.NewSession(goe2e.SessionConfig{
	HTTPClient: server.Client(),
	Defaults: goe2e.RequestConfig{RequestMods: []goe2e.RequestModifier{
		goe2e.WithHeaders(goe2e.D{"X-User-ID": bobID}),
	}},
})
if err != nil {
	t.Fatal(err)
}
bob.TestRequest(t, goe2e.RequestConfig{
	Name: "accept interest request",
	SpecOpts: []goe2e.SpecOption{
		goe2e.WithMethod(http.MethodPost),
		goe2e.WithURL(server.URL + "/interest-requests/" + requestID + "/response"),
		goe2e.WithJSON(map[string]string{"action": "accepted"}),
	},
	PostTestStatements: []goe2e.TestStatement{
		goe2e.AssertStatusCode(http.StatusOK),
	},
})
```

`Session.TestRequest` accepts `RequestConfig`, so the session-owned client cannot be accidentally replaced. Sessions are not safe for concurrent use; share one deliberately only when testing session switching or cookie-precedence behavior.

For browser-style cookie tests, make the sign-in request and the follow-up request through the same session:

```go
server := httptest.NewTestServer(t, app.Handler()) // /sign-in sets a session cookie
browser, err := goe2e.NewSession(goe2e.SessionConfig{HTTPClient: server.Client()})
if err != nil {
	t.Fatal(err)
}
browser.TestRequest(t, goe2e.RequestConfig{SpecOpts: []goe2e.SpecOption{
	goe2e.WithMethod(http.MethodPost), goe2e.WithURL(server.URL + "/sign-in"),
}})
account := browser.TestRequest(t, goe2e.RequestConfig{SpecOpts: []goe2e.SpecOption{
	goe2e.WithURL(server.URL + "/account"),
}}) // receives the stored cookie automatically
_ = account
```

### Gin without a running port

Gin's `*gin.Engine` implements `http.Handler`, so it can be passed straight to `httptest.NewTestServer`. This keeps Gin an application dependency only; `goe2e` does not depend on it.

```go
func TestPersonPost(t *testing.T) {
	router := gin.New()
	router.POST("/persons", func(c *gin.Context) {
		var person model.Person
		if err := c.ShouldBindJSON(&person); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.JSON(http.StatusCreated, person)
	})

	server := httptest.NewTestServer(t, router)
	goe2e.TestRequest(t, goe2e.TestConfig{
		HTTPClient: server.Client(),
		Request: goe2e.RequestConfig{
			Name: "POST /persons",
			SpecOpts: []goe2e.SpecOption{
			goe2e.WithMethod(http.MethodPost),
			goe2e.WithURL("https://app.test/persons"),
			goe2e.WithJSON(model.Person{Name: "John", Age: 32}),
		},
			RequestMods: []goe2e.RequestModifier{
			goe2e.WithContentType(goe2e.ContentHeaderJSON),
		},
			PostTestStatements: []goe2e.TestStatement{
			goe2e.AssertStatusCode(http.StatusCreated),
			goe2e.AssertResponseJSONPointer("/name", "John"),
			},
		},
	})
}
```

### Choosing an in-memory or real service target

Use `httptest.NewTestServer` and `HTTPClient: server.Client()` when testing handlers, middleware, validation, and response contracts. It is deterministic, does not open a port, and routes the configured client to the test handler even when the request URL uses a readable hostname such as `https://app.test`.

Omit `HTTPClient` when a test must reach a separately running service. This is appropriate for deployment configuration, DNS, real TLS/network behavior, or dependencies that cannot be represented by a local handler. Set `Context` or `Timeout` on `RequestConfig` for an explicit cancellation boundary in those tests.

### Structured diagnostics

Set `OnDiagnostic` to send a safe request summary to your logger or metrics system. It includes the method, URL, status, selected response metadata, and byte count, while omitting credentials, cookies, and response-body content.

```go
OnDiagnostic: func(d goe2e.RequestDiagnostic) {
	logger.Info("HTTP test request", "method", d.Method, "url", d.URL,
		"status", d.ResponseCode, "response_bytes", d.ResponseBytes)
},
```

That's it!

## Compatibility and releases

The module currently requires Go 1.27. Its public API and versioning policy are documented in [COMPATIBILITY.md](COMPATIBILITY.md).

For the lifecycle flow, common helpers, and migration notes, see [API.md](API.md).


## Limitations

- Only build for the unit testing environment, might adapt it for use in application code.

- No built-in solution for logging response times yet

- Only has functions to deal with JSON encoding for now

- No built-in solution for test integration yet (skipping/env-var/testing.M)

- No built-in solution for run comparison/logging yet

- Missing convenience functions (Auth Header)

- Implementation is subject to change
