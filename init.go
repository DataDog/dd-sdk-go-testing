// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	"github.com/DataDog/dd-sdk-go-testing/internal/utils"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	spanKind      = "test"
	testFramework = "golang.org/pkg/testing"
)

var repoRegex = regexp.MustCompile(`(?m)\/([a-zA-Z0-9\\\-_.]*)$`)

// FinishFunc closes a started span and attaches test status information.
type FinishFunc func()

// Run is a helper function to run a `testing.M` object and gracefully stopping the tracer afterwards
func Run(m *testing.M, opts ...tracer.StartOption) int {
	// Preload all CI and Git tags.
	ensureCITags()

	// Check if DD_SERVICE has been set; otherwise we default to repo name.
	if v := os.Getenv("DD_SERVICE"); v == "" {
		if repoUrl, ok := getFromCITags(constants.GitRepositoryURL); ok {
			matches := repoRegex.FindStringSubmatch(repoUrl)
			if len(matches) > 1 {
				repoUrl = strings.TrimSuffix(matches[1], ".git")
			}
			opts = append(opts, tracer.WithService(repoUrl))
		}
	}

	// Initialize tracer
	tracer.Start(opts...)
	exitFunc := func() {
		tracer.Flush()
		tracer.Stop()
	}
	defer exitFunc()

	// Handle SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		exitFunc()
		os.Exit(1)
	}()

	// Execute test suite
	return m.Run()
}

// TB is the minimal interface common to T and B.
type TB interface {
	Failed() bool
	Name() string
	Skipped() bool
}

var _ TB = (*testing.T)(nil)
var _ TB = (*testing.B)(nil)

// StartTest returns a new span with the given testing.TB interface and options. It uses
// tracer.StartSpanFromContext function to start the span with automatically detected information.
func StartTest(tb TB, opts ...Option) (context.Context, FinishFunc) {
	opts = append(opts, WithIncrementSkipFrame())
	return StartTestWithContext(context.Background(), tb, opts...)
}

// StartTestWithContext returns a new span with the given testing.TB interface and options. It uses
// tracer.StartSpanFromContext function to start the span with automatically detected information.
func StartTestWithContext(ctx context.Context, tb TB, opts ...Option) (context.Context, FinishFunc) {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}

	var pc uintptr
	if cfg.originalTestFunc == nil {
		pc, _, _, _ = runtime.Caller(cfg.skip)
	} else {
		pc = reflect.Indirect(reflect.ValueOf(cfg.originalTestFunc)).Pointer()
	}

	suite, _ := utils.GetPackageAndName(pc)
	name := tb.Name()
	fqn := fmt.Sprintf("%s.%s", suite, name)

	testOpts := []tracer.StartSpanOption{
		tracer.ResourceName(fqn),
		tracer.Tag(constants.TestName, name),
		tracer.Tag(constants.TestSuite, suite),
		tracer.Tag(constants.TestFramework, testFramework),
		tracer.Tag(constants.Origin, constants.CIAppTestOrigin),
	}

	switch tb.(type) {
	case *testing.T:
		testOpts = append(testOpts, tracer.Tag(constants.TestType, constants.TestTypeTest))
	case *testing.B:
		testOpts = append(testOpts, tracer.Tag(constants.TestType, constants.TestTypeBenchmark))
	}

	cfg.spanOpts = append(testOpts, cfg.spanOpts...)
	span, ctx := tracer.StartSpanFromContext(ctx, constants.SpanTypeTest, cfg.spanOpts...)

	return ctx, func() {
		var r interface{} = nil

		if r = recover(); r != nil {
			// Panic handling
			span.SetTag(constants.TestStatus, constants.TestStatusFail)
			span.SetTag(ext.Error, true)
			span.SetTag(ext.ErrorMsg, fmt.Sprint(r))
			span.SetTag(ext.ErrorStack, getStacktrace(2))
			span.SetTag(ext.ErrorType, "panic")
		} else {
			// Normal finalization
			span.SetTag(ext.Error, tb.Failed())

			if tb.Failed() {
				span.SetTag(constants.TestStatus, constants.TestStatusFail)
			} else if tb.Skipped() {
				span.SetTag(constants.TestStatus, constants.TestStatusSkip)
			} else {
				span.SetTag(constants.TestStatus, constants.TestStatusPass)
			}
		}

		span.Finish(cfg.finishOpts...)

		if r != nil {
			tracer.Flush()
			tracer.Stop()
			panic(r)
		}
	}
}

func getStacktrace(skip int) string {
	pcs := make([]uintptr, 256)
	total := runtime.Callers(skip+1, pcs)
	frames := runtime.CallersFrames(pcs[:total])
	buffer := new(bytes.Buffer)
	for {
		if frame, ok := frames.Next(); ok {
			fmt.Fprintf(buffer, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		} else {
			break
		}

	}
	return buffer.String()
}
