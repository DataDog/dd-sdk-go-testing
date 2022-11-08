package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	utils "github.com/DataDog/dd-sdk-go-testing/ginkgo/internal"
)

var _ = Describe("Suite", func() {
	var suiteTest *utils.SuiteTest

	BeforeEach(func() {
		suiteTest = utils.NewSuiteTest("test-framework")
		DeferCleanup(func() {
			Expect(utils.GetParent(suiteTest.Context())).To(BeNil())
			Expect(suiteTest.Close()).To(Succeed())
		})
	})

	Describe("entering a container", func() {
		It("should works", func() {
			Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
			Expect(suiteTest.LeaveContainer()).To(Succeed())
		})

		Context("entering another container and leaving it", func() {
			BeforeEach(func() {
				Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
				DeferCleanup(suiteTest.LeaveContainer)
			})

			It("should work", func() {
				Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
				Expect(suiteTest.LeaveContainer()).To(Succeed())
			})

			Context("after registering a test", func() {
				BeforeEach(func() {
					suiteTest.Snapshot().RegisterTest()
				})

				It("should work", func() {
					Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
					Expect(suiteTest.LeaveContainer()).To(Succeed())
				})
			})

			Context("after leaving a containers", func() {
				BeforeEach(func() {
					Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
					Expect(suiteTest.LeaveContainer()).To(Succeed())
				})

				It("should work", func() {
					Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
					Expect(suiteTest.LeaveContainer()).To(Succeed())
				})
			})
		})
	})

	Describe("Test", func() {
		var test utils.TestCase

		It("can be registered", func() {
			test = suiteTest.Snapshot().RegisterTest()
		})

		Context("after entering a container", func() {
			BeforeEach(func() {
				Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
				DeferCleanup(suiteTest.LeaveContainer)
			})

			When("not entering a test", func() {
				It("cannot leave it", func() {
					test = suiteTest.Snapshot().RegisterTest()
					Expect(test.Leave()).To(MatchError("test is not entered"))
				})
			})

			Context("after registering a test", func() {
				JustBeforeEach(func() {
					test = suiteTest.Snapshot().RegisterTest()
				})

				It("can be registered multiple time more", func() {
					test = suiteTest.Snapshot().RegisterTest()

					test = suiteTest.Snapshot().RegisterTest()
				})
			})
		})

		Describe("registering a test", func() {
			JustBeforeEach(func() {
				test = suiteTest.Snapshot().RegisterTest()
			})

			It("should work", func() {
				Expect(test.Enter("test", "test")).To(Succeed())
				Expect(test.Leave()).To(Succeed())
			})

			Context("After entering a container", func() {
				BeforeEach(func() {
					Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
					DeferCleanup(suiteTest.LeaveContainer)
				})

				Context("after registering a another test", func() {
					var test2 utils.TestCase

					BeforeEach(func() {
						test2 = suiteTest.Snapshot().RegisterTest()
					})

					It("should be possible to enter the test", func() {
						Expect(test.Enter("test", "test")).To(Succeed())
						Expect(test.Leave()).To(Succeed())
					})

					Context("after entering first test", func() {
						BeforeEach(func() {
							Expect(test.Enter("test", "test")).To(Succeed())
							DeferCleanup(test.Leave)
						})

						It("should be possible to enter the second one", func() {
							Expect(test2.Enter("test2", "test2")).To(Succeed())
							Expect(test2.Leave()).To(Succeed())
						})
					})
				})

				Context("entering another container and leaving it", func() {
					BeforeEach(func() {
						Expect(suiteTest.EnterContainer("test", "test")).To(Succeed())
						Expect(suiteTest.LeaveContainer()).To(Succeed())
					})

					It("should work", func() {
						Expect(test.Enter("test", "test")).To(Succeed())
					})
				})
			})
		})
	})
})
