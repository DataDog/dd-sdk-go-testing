# Datadog SDK for Go testing
> Datadog Test Instrumentation framework for Go

## Getting Started

### Installing
Installation of the DataDog go testing sdk is done via `go get`:

```shell
go get -u github.com/DataDog/dd-sdk-go-testing
```

#### Requires:
- Go >= 1.12
- Datadog's Trace Agent >= 5.21.1

### Instrumenting your tests
In order to instrument your tests that use Go's native 
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
about its trace. Use it when you make any external calls in order to see the traces within the test span.

## License

This work is dual-licensed under Apache 2.0 or BSD3.

[Apache License, v2.0](LICENSE-APACHE)

[BSD, v3.0](LICENSE-BSD3)