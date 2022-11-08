# Datadog SDK for Ginkgo testing
This SDK is part of Datadog's [CI Visibility sdk](https://github.com/DataDog/dd-sdk-go-testing).

## Getting Started

### Installing
Installation of the Datadog Go testing SDK is done via `go get`:

```shell
go get -u github.com/DataDog/dd-sdk-go-testing/ginkgo
```

### Instrumenting your tests
To instrument tests that use 
[Ginkgo v2](https://pkg.go.dev/github.com/onsi/ginkgo/v2) package, you have to 
call `Run(m)` in your `TestMain` function, and replace the `github.com/onsi/ginkgo/v2` import with `github.com/DataDog/dd-sdk-go-testing/ginkgo`:

For example:

```go
package go_sdk_sample

import (
	"os"
	"testing"

	. "github.com/DataDog/dd-sdk-go-testing/ginkgo"
)

func TestMain(m *testing.M) {
	os.Exit(Run(m))
}

// This test will be instrumented
func TestSimpleExample(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reader Suite")
}

// Test example from https://onsi.github.io/ginkgo/#spec-subjects-it
var _ = Describe("Books", func() {
  It("can extract the author's last name", func() {
    book = &books.Book{
      Title: "Les Miserables",
      Author: "Victor Hugo",
      Pages: 2783,
    }

    Expect(book.AuthorLastName()).To(Equal("Hugo"))
  })
})
```

## Configuration

The sdk can be configured the same way as the [Datadog Go tracing library](github.com/DataDog/dd-sdk-go-testing).

## License

This work is part of the Datadog [Go testing SDK](https://github.com/DataDog/dd-sdk-go-testing) and inherits its [license](https://github.com/DataDog/dd-sdk-go-testing/#license).