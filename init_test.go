// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"fmt"
	"os"
	"testing"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/mocktracer"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
)

func TestMain(m *testing.M) {
	os.Exit(Run(m))
}

func TestStatus(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	t.Run("pass", func(t *testing.T) {
		ctx, finish := StartTest(t)
		defer finish()

		span, _ := tracer.SpanFromContext(ctx)
		span.SetTag("k", "1")
	})

	t.Run("skip", func(t *testing.T) {
		ctx, finish := StartTest(t)
		defer finish()

		span, _ := tracer.SpanFromContext(ctx)
		span.SetTag("k", "2")
		t.Skip("good reason")
	})

	spans := mt.FinishedSpans()
	if len(spans) != 2 {
		t.FailNow()
	}

	const suiteName string = "github.com/DataDog/dd-sdk-go-testing"
	const framework string = "golang.org/pkg/testing"

	s := spans[0]
	assertEqual("test", s.OperationName())
	assertEqual("TestStatus/pass", s.Tag(constants.TestName).(string))
	assertEqual(suiteName, s.Tag(constants.TestSuite).(string))
	assertEqual(fmt.Sprintf("%s.%s", suiteName, "TestStatus/pass"), s.Tag(ext.ResourceName).(string))
	assertEqual(framework, s.Tag(constants.TestFramework).(string))
	assertEqual(constants.TestStatusPass, s.Tag(constants.TestStatus).(string))
	commonEqualCheck(s)
	commonNotEmptyCheck(s)
	fmt.Println(s)

	s = spans[1]
	assertEqual("test", s.OperationName())
	assertEqual("TestStatus/skip", s.Tag(constants.TestName).(string))
	assertEqual(suiteName, s.Tag(constants.TestSuite).(string))
	assertEqual(fmt.Sprintf("%s.%s", suiteName, "TestStatus/skip"), s.Tag(ext.ResourceName).(string))
	assertEqual(framework, s.Tag(constants.TestFramework).(string))
	assertEqual(constants.TestStatusSkip, s.Tag(constants.TestStatus).(string))
	commonEqualCheck(s)
	commonNotEmptyCheck(s)
	fmt.Println(s)
}

func TestPanic(t *testing.T) {
	var span *tracer.Span

	if _, ok := tracer.GetGlobalTracer().(*tracer.NoopTracer); ok {
		// Global tracer is set to noop by previous test. To make sure
		// it works, we need to start the standard tracer again.
		tracer.Start()
	}

	pf := func() {
		defer func() {
			// recover panic to finish the subtest successfully
			recover()
		}()

		_, finish := StartTest(
			t,
			withSpansCapture(func(sp *tracer.Span) {
				span = sp
				t.Log("span", span)
			}),
		)
		// This cancel function alters the global tracer, that's why we
		// need to capture the finished spans in the option when calling it.
		// This is also the reason we can't use mocktracer.
		defer finish()

		panic("forced panic")
	}

	pf()
	if span == nil {
		t.FailNow()
	}

	s := span.AsMap()
	assertEqual("forced panic", s[ext.ErrorMsg].(string))
	assertEqual("panic", s[ext.ErrorType].(string))
	assertEqual("1", fmt.Sprint(s[ext.MapSpanError]))
	assertNotEmpty(s[ext.ErrorStack].(string))
}

func commonEqualCheck(s *mocktracer.Span) {
	assertEqual(constants.SpanTypeTest, s.Tag(ext.SpanType).(string))
	assertEqual(constants.SpanTypeTest, s.Tag(constants.SpanKind).(string))
	assertEqual(constants.TestTypeTest, s.Tag(constants.TestType).(string))
	assertEqual(constants.CIAppTestOrigin, s.Tag(constants.Origin).(string))
}

func commonNotEmptyCheck(s *mocktracer.Span) {
	assertNotEmpty(s.Tag(constants.GitCommitAuthorDate).(string))
	assertNotEmpty(s.Tag(constants.GitCommitAuthorEmail).(string))
	assertNotEmpty(s.Tag(constants.GitCommitAuthorName).(string))
	assertNotEmpty(s.Tag(constants.GitCommitCommitterDate).(string))
	assertNotEmpty(s.Tag(constants.GitCommitCommitterEmail).(string))
	assertNotEmpty(s.Tag(constants.GitCommitCommitterName).(string))
	assertNotEmpty(s.Tag(constants.GitCommitMessage).(string))
	assertNotEmpty(s.Tag(constants.GitCommitSHA).(string))
	assertNotEmpty(s.Tag(constants.GitRepositoryURL).(string))
	assertNotEmpty(s.Tag(constants.CIWorkspacePath).(string))
	assertNotEmpty(s.Tag(constants.OSArchitecture).(string))
	assertNotEmpty(s.Tag(constants.OSPlatform).(string))
	assertNotEmpty(s.Tag(constants.OSVersion).(string))
}

func assertEqual(expected string, actual string) {
	if expected != actual {
		panic(fmt.Sprintf("Value expected: %s, Actual: %s", expected, actual))
	}
}

func assertNotEmpty(actual string) {
	if actual == "" {
		panic("Value is empty")
	}
}
