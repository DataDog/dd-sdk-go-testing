package ginkgo_test

import (
	"fmt"
	"testing"

	. "github.com/DataDog/dd-sdk-go-testing/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/example/books"
)

func ExampleRunSpecs() {
	t := &testing.T{} // Get t from parameter

	Describe("Books", func() {
		It("can guess the category", func() {
			book := &books.Book{
				Title:  "Les Miserables",
				Author: "Victor Hugo",
				Pages:  2783,
			}

			Expect(book.CategoryByLength()).To(Equal("NOVEL"))

			fmt.Println("This is a novel!")
		})
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Example Suite")
	// Output: This is a novel!
}
