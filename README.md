# Datadog SDK for Go testing
This SDK is part of Datadog's [CI Visibility](https://docs.datadoghq.com/continuous_integration/) product, currently in beta.

## Getting Started

### Installing
Installation of the Datadog Go testing SDK is done via `go get`:

```shell
go get -u github.com/DataDog/dd-sdk-go-testing
```

#### Requires:
- Go >= 1.12
- Datadog's Trace Agent >= 5.21.1

### Instrumenting your tests
To instrument tests that use Go's native 
[`testing`](https://golang.org/pkg/testing/) package, you have to 
call `ddtesting.Run(m)` in your `TestMain` function, and call
`ddtesting.StartTest(t)` or `ddtesting.StartTestWithContext(ctx, t)`
and `defer finish()` on each test.

For example:

```go
package go_sdk_sample

import (
	"os"
	"testing"

	ddtesting "github.com/DataDog/dd-sdk-go-testing"
)

func TestMain(m *testing.M) {
	os.Exit(ddtesting.Run(m))
}

// Simple test without `context` usage
func TestSimpleExample(t *testing.T) {
	_, finish := ddtesting.StartTest(t)
	defer finish()

	// Test code...
}

// Test with subtests using `StartTestWithContext`
func TestExampleWithSubTests(t *testing.T) {
	ctx, finish := ddtesting.StartTest(t)
	defer finish()

	testCases := []struct {
		name string
	}{
		{"Sub01"},
		{"Sub02"},
		{"Sub03"},
		{"Sub04"},
		{"Sub05"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, finish := ddtesting.StartTestWithContext(ctx, t)
			defer finish()

			// Test code ...
		})
	}
}
```

Note that after this, you can use `ctx` to refer to the context of the running test, which has information
about its trace. Use it when you make any external call to see the traces within the test span.

For example:

```go
package go_sdk_sample

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	ddtesting "github.com/DataDog/dd-sdk-go-testing"
	ddhttp "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func TestMain(m *testing.M) {
	os.Exit(ddtesting.Run(m))
}

func TestWithExternalCalls(t *testing.T) {
	ctx, finish := ddtesting.StartTest(t)
	defer finish()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}))
	defer s.Close()

	t.Run("default", func(t *testing.T) {
		ctx, finish := ddtesting.StartTestWithContext(ctx, t)
		defer finish()

		rt := ddhttp.WrapRoundTripper(http.DefaultTransport)
		client := &http.Client{
			Transport: rt,
		}

		req, err := http.NewRequest("GET", s.URL+"/hello/world", nil)
		if err != nil {
			t.FailNow()
		}

		req = req.WithContext(ctx)

		client.Do(req)
	})

	t.Run("custom-name", func(t *testing.T) {
		ctx, finish := ddtesting.StartTestWithContext(ctx, t)
		defer finish()

		span, _ := ddtracer.SpanFromContext(ctx)

		customNamer := func(req *http.Request) string {
			value := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
			span.SetTag("customNamer.Value", value)
			return value
		}

		rt := ddhttp.WrapRoundTripper(http.DefaultTransport, ddhttp.RTWithResourceNamer(customNamer))
		client := &http.Client{
			Transport: rt,
		}

		req, err := http.NewRequest("GET", s.URL+"/hello/world", nil)
		if err != nil {
			t.FailNow()
		}

		req = req.WithContext(ctx)

		client.Do(req)
	})
}
```

## License

This work is dual-licensed under Apache 2.0 or BSD3.

[Apache License, v2.0](LICENSE-APACHE)

[BSD, v3.0](LICENSE-BSD3)
