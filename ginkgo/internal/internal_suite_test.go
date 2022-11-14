package utils_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ddtesting "github.com/DataDog/dd-sdk-go-testing"
)

func TestMain(m *testing.M) {
	os.Exit(ddtesting.Run(m))
}

func TestInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Internal Suite")
}
