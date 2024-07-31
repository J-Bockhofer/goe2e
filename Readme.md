# goe2e - E2E testing in Go

Writing server in Go? Write E2E (End-to-End) tests in it, too!

Just add a test file to your source or an empty application and get going.

## How it works

Simply put, E2E tests work by sending a request to a running application.

Unit tests usually have great language tooling while E2E tests need another application like Postman or Insomnia to run.

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

	goe2e "github.com/J-Bockhofer/goe2e/pkg"

	"github.com/stretchr/testify/assert"
)

func TestPersonPost(t *testing.T) {
	p := model.Person{
		Name: "john",
		Age:  32,
	}
	rc := &goe2e.TestConfig{
		Name: "POST /persons",
		SpecOpts: []goe2e.RequestSpecOption{
			goe2e.WithMethod(http.MethodPost),
			goe2e.WithURL("http://localhost:8080/persons/"),
			goe2e.WithJSON(&p),
		},
		RequestMods: []goe2e.RequestModifier{
			goe2e.WithContentType(goe2e.ContentHeaderJSON),
		},
		PreFunc: func(r *goe2e.RequestHandler) error {
			return nil
		},
		PreTestStatements: []goe2e.TestStatement{
			{Description: "request not nil", Statement: func(t *testing.T, r *goe2e.RequestHandler) {
				assert.NotNil(t, r.Request())
			}},
		},
		PostFunc: func(r *goe2e.RequestHandler) error {
			return nil
		},
		PostTestStatements: []goe2e.TestStatement{
			{Description: "body not nil", Statement: func(t *testing.T, r *goe2e.RequestHandler) {
				assert.NotNil(t, r.ResponseBody)
			}},
			{Description: "status 202", Statement: func(t *testing.T, r *goe2e.RequestHandler) {
				assert.Equal(t, http.StatusAccepted, r.Response.StatusCode)
			}},
		},
	}
	goe2e.TestRequest(t, rc)
}
```

`One caveat:` We have to run the main application before we run the E2E test in unit test disguise.

This can be dealt with using environment variables that skip E2E tests / signal that the application is running and execute the tests.

Alternatively [testing.M](https://pkg.go.dev/testing#hdr-Main) provides a space for test setup and teardown functions.

That's it!


## Limitations & Future Plans

- Only build for the unit testing environment, with plans to adapt it for use in application code.

- No built-in solution for logging response times yet

- No convenience solution for "environments" yet (passing values between test contexts)

- Only has functions to deal with JSON encoding for now

- No built-in solution for test integration yet (skipping/env-var/testing.M)

- No built-in solution for run comparison/logging yet

- Missing many convenience functions (Auth Header, JSON Un-/Marshalling)

- Implementation is subject to change 