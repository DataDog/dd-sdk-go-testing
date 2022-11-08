package ginkgo

import (
	"fmt"
	"sync"
	"testing"

	ddtesting "github.com/DataDog/dd-sdk-go-testing"
	utils "github.com/DataDog/dd-sdk-go-testing/ginkgo/internal"
	"github.com/onsi/ginkgo/v2"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// https://pkg.go.dev/github.com/onsi/ginkgo/v2@v2.1.6/dsl/core

const GINKGO_VERSION = ginkgo.GINKGO_VERSION + "+dd-trace-go"

type GinkgoWriterInterface = ginkgo.GinkgoWriterInterface
type GinkgoTestingT = ginkgo.GinkgoTestingT
type GinkgoTInterface = ginkgo.GinkgoTInterface

var GinkgoWriter = ginkgo.GinkgoWriter
var GinkgoConfiguration = ginkgo.GinkgoConfiguration
var GinkgoRandomSeed = ginkgo.GinkgoRandomSeed
var GinkgoParallelProcess = ginkgo.GinkgoParallelProcess
var PauseOutputInterception = ginkgo.PauseOutputInterception
var ResumeOutputInterception = ginkgo.ResumeOutputInterception
var Skip = ginkgo.Skip
var Fail = ginkgo.Fail
var AbortSuite = ginkgo.AbortSuite
var GinkgoRecover = ginkgo.GinkgoRecover
var GinkgoT = ginkgo.GinkgoT

/******** Branches **********/

func Describe(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.Describe(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("Describe", text) }, currentSuite.LeaveContainer, callback...)...)
}

func FDescribe(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.FDescribe(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("FDescribe", text) }, currentSuite.LeaveContainer, callback...)...)
}

func PDescribe(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.PDescribe(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("PDescribe", text) }, currentSuite.LeaveContainer, callback...)...)
}

func XDescribe(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.XDescribe(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("XDescribe", text) }, currentSuite.LeaveContainer, callback...)...)
}

func Context(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.Context(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("Context", text) }, currentSuite.LeaveContainer, callback...)...)
}

func FContext(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.FContext(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("FContext", text) }, currentSuite.LeaveContainer, callback...)...)
}

func PContext(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.PContext(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("PContext", text) }, currentSuite.LeaveContainer, callback...)...)
}

func XContext(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.XContext(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("XContext", text) }, currentSuite.LeaveContainer, callback...)...)
}

func When(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.When(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("When", text) }, currentSuite.LeaveContainer, callback...)...)
}

func FWhen(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.FWhen(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("FWhen", text) }, currentSuite.LeaveContainer, callback...)...)
}

func PWhen(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.PWhen(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("PWhen", text) }, currentSuite.LeaveContainer, callback...)...)
}

func XWhen(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.XWhen(text, utils.PatchArgsNoArgs(func() error { return currentSuite.EnterContainer("XWhen", text) }, currentSuite.LeaveContainer, callback...)...)
}

/******** Leafs **********/

func It(text string, callback ...interface{}) bool {
	initSetup()

	test := currentSuite.Snapshot().RegisterTest()

	suiteName := utils.GetSuiteName()

	return ginkgo.It(text, utils.PatchArgsNoArgs(func() error { return test.Enter("It", suiteName) }, test.Leave, callback...)...)
}

func FIt(text string, callback ...interface{}) bool {
	initSetup()

	test := currentSuite.Snapshot().RegisterTest()

	suiteName := utils.GetSuiteName()

	return ginkgo.FIt(text, utils.PatchArgsNoArgs(func() error { return test.Enter("FIt", suiteName) }, test.Leave, callback...)...)
}

func PIt(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.PIt(text, callback...)
}

func XIt(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.XIt(text, callback...)
}

func Specify(text string, callback ...interface{}) bool {
	initSetup()

	test := currentSuite.Snapshot().RegisterTest()

	suiteName := utils.GetSuiteName()

	return ginkgo.Specify(text, utils.PatchArgsNoArgs(func() error { return test.Enter("Specify", suiteName) }, test.Leave, callback...)...)
}

func FSpecify(text string, callback ...interface{}) bool {
	initSetup()

	test := currentSuite.Snapshot().RegisterTest()

	suiteName := utils.GetSuiteName()

	return ginkgo.FSpecify(text, utils.PatchArgsNoArgs(func() error { return test.Enter("FSpecify", suiteName) }, test.Leave, callback...)...)
}

func noop() error { return nil }

func PSpecify(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.PSpecify(text, utils.PatchArgsNoArgs(noop, noop, callback...)...)
}

func XSpecify(text string, callback ...interface{}) bool {
	initSetup()

	return ginkgo.XSpecify(text, utils.PatchArgsNoArgs(noop, noop, callback...)...)
}

func By(text string, callback ...func()) {
	initSetup()

	test := currentSuite.Snapshot().RegisterTest()

	if len(callback) == 0 {
		callback = []func(){func() {}}
	}

	suiteName := utils.GetSuiteName()

	ginkgo.By(text, utils.PatchFuncs(func() error { return test.Enter("By", suiteName) }, test.Leave, callback...)...)
}

func SynchronizedBeforeSuite(process1Body func() []byte, allProcessBody func([]byte)) bool {
	initSetup()

	snapshot := currentSuite.Snapshot()
	test1 := snapshot.RegisterTest()

	var test2 utils.TestCase

	suiteName := utils.GetSuiteName()

	return ginkgo.SynchronizedBeforeSuite(func() []byte {
		if err := test1.Enter("SynchronizedBeforeSuite.process1Body", suiteName); err != nil {
			panic(fmt.Errorf("enter %s test: %w", "SynchronizedBeforeSuite.process1Body", err))
		}
		defer func() {
			if err := test1.Leave(); err != nil {
				panic(fmt.Errorf("leave %s test: %w", "SynchronizedBeforeSuite.process1Body", err))
			}
		}()

		body := process1Body()

		test2 = snapshot.RegisterTest()

		return body
	}, func(data []byte) {
		if err := test2.Enter("SynchronizedBeforeSuite.allProcessBody", suiteName); err != nil {
			panic(fmt.Errorf("enter %s test: %w", "SynchronizedBeforeSuite.allProcessBody", err))
		}
		defer func() {
			if err := test2.Leave(); err != nil {
				panic(fmt.Errorf("leave %s test: %w", "SynchronizedBeforeSuite.allProcessBody", err))
			}
		}()

		allProcessBody(data)
	})
}

func SynchronizedAfterSuite(allProcessBody func(), process1Body func()) bool {
	initSetup()

	snapshot := currentSuite.Snapshot()
	test1 := snapshot.RegisterTest()

	var test2 utils.TestCase

	suiteName := utils.GetSuiteName()

	return ginkgo.SynchronizedAfterSuite(func() {
		if err := test1.Enter("SynchronizedAfterSuite.allProcessBody", suiteName); err != nil {
			panic(fmt.Errorf("enter %s test: %w", "SynchronizedAfterSuite.allProcessBody", err))
		}
		defer func() {
			if err := test1.Leave(); err != nil {
				panic(fmt.Errorf("leave %s test: %w", "SynchronizedAfterSuite.allProcessBody", err))
			}
		}()

		allProcessBody()

		test2 = snapshot.RegisterTest()
	}, func() {
		if err := test2.Enter("SynchronizedAfterSuite.process1Body", suiteName); err != nil {
			panic(fmt.Errorf("enter 2nd %s test: %w", "SynchronizedAfterSuite.process1Body", err))
		}
		defer func() {
			if err := test2.Leave(); err != nil {
				panic(fmt.Errorf("leave 2nd %s test: %w", "SynchronizedAfterSuite.process1Body", err))
			}
		}()

		process1Body()
	})
}

/*
 * TODO: add support for *Each* functions
 *
 * ```go
 * test := currentSuite.Snapshot().RegisterTest()
 *
 * suiteName := utils.GetSuiteName()
 * return ginkgo.BeforeEach(utils.PatchArgsNoArgs(func() error { return test.Enter("BeforeEach", suiteName) }, test.Leave, args...)...)
 * ```
 */

func BeforeEach(args ...interface{}) bool {
	initSetup()

	return ginkgo.BeforeEach(args...)
}

func JustBeforeEach(args ...interface{}) bool {
	initSetup()

	return ginkgo.JustBeforeEach(args...)
}

func AfterEach(args ...interface{}) bool {
	initSetup()

	return ginkgo.AfterEach(args...)
}

func JustAfterEach(args ...interface{}) bool {
	initSetup()

	return ginkgo.JustAfterEach(args...)
}

func BeforeAll(args ...interface{}) bool {
	initSetup()

	return ginkgo.BeforeAll(args...)
}

func AfterAll(args ...interface{}) bool {
	initSetup()

	return ginkgo.AfterAll(args...)
}

func DeferCleanup(args ...interface{}) {
	// initSetup() // no need to init setup here, because it is called in methods above

	ginkgo.DeferCleanup(args...)
}

func BeforeSuite(userFunc func()) bool {
	initSetup()

	test := currentSuite.Snapshot().RegisterTest()

	suiteName := utils.GetSuiteName()

	return ginkgo.BeforeSuite(func() {
		if err := test.Enter("BeforeSuite", suiteName); err != nil {
			panic(fmt.Errorf("enter %s test: %w", "BeforeSuite", err))
		}
		defer func() {
			if err := test.Leave(); err != nil {
				panic(fmt.Errorf("leave %s test: %w", "BeforeSuite", err))
			}
		}()

		userFunc()
	})
}

func AfterSuite(userFunc func()) bool {
	initSetup()

	test := currentSuite.Snapshot().RegisterTest()

	suiteName := utils.GetSuiteName()

	return ginkgo.AfterSuite(func() {
		if err := test.Enter("AfterSuite", suiteName); err != nil {
			panic(fmt.Errorf("enter %s test: %w", "AfterSuite", err))
		}
		defer func() {
			if err := test.Leave(); err != nil {
				panic(fmt.Errorf("leave %s test: %w", "AfterSuite", err))
			}
		}()

		userFunc()
	})
}

func RunSpecs(t ginkgo.GinkgoTestingT, description string, args ...interface{}) bool {
	initSetup()

	if err := currentSuite.LeaveContainer(); err != nil { // exit initialization step
		panic(fmt.Errorf("leave container: %w", err))
	}

	if err := currentSuite.EnterContainer("RunSpecs", description); err != nil {
		panic(fmt.Errorf("enter %s container: %w", "RunSpecs", err))
	}

	defer currentSuite.Close()

	return ginkgo.RunSpecs(t, description, args...)
}

const TestFrameworkName = "github.com/onsi/ginkgo/v2"

var (
	initOnce     sync.Once
	currentSuite *utils.SuiteTest
)

func initSetup() {
	initOnce.Do(func() {
		currentSuite = utils.NewSuiteTest(TestFrameworkName)
		if err := currentSuite.EnterContainer("PhaseBuildTree", "init"); err != nil {
			panic(fmt.Errorf("enter %s container: %w", "PhaseBuildTree", err))
		}
	})
}

// Run is a helper function to run a `testing.M` object and gracefully stopping the tracer afterwards
func Run(m *testing.M, opts ...tracer.StartOption) int {
	initSetup()

	defer func() {
		if err := currentSuite.Close(); err != nil {
			panic(fmt.Errorf("failed to close suite: %w", err))
		}

		initOnce = sync.Once{}

		initSetup()
	}()

	return ddtesting.Run(m, opts...)
}
