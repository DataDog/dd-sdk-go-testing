package utils_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	utils "github.com/DataDog/dd-sdk-go-testing/ginkgo/internal"
)

var _ = Describe("State", func() {
	var state *utils.State

	BeforeEach(func() {
		state = utils.NewState(context.Background())
	})

	Describe("get contexts", func() {
		It("should return the correct contexts", func() {
			Expect(state.GetContexts()).To(HaveLen(1))
		})

		Context("after entering a container", func() {
			BeforeEach(func() {
				state.Push(context.Background())
				DeferCleanup(state.Pop)
			})

			It("should return the correct contexts", func() {
				Expect(state.GetContexts()).To(HaveLen(2))
			})
		})
	})
})
