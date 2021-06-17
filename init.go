// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
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

// FinishFunc closes a started span and attaches test status information.
type FinishFunc func()

// Run is a helper function to run a `testing.M` object and gracefully stopping the tracer afterwards
func Run(m *testing.M, opts ...tracer.StartOption) int {
	loadTags()
	tracer.Start(opts...)
	defer tracer.Stop()

	// Handle SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		tracer.Stop()
		os.Exit(1)
	}()

	return m.Run()
}

// StartTest returns a new span with the given testing.TB interface and options. It uses
// tracer.StartSpanFromContext function to start the span with automatically detected information.
func StartTest(tb testing.TB, opts ...Option) (context.Context, FinishFunc) {
	return StartTestWithContext(context.Background(), tb, opts...)
}

// StartTestWithContext returns a new span with the given testing.TB interface and options. It uses
// tracer.StartSpanFromContext function to start the span with automatically detected information.
func StartTestWithContext(ctx context.Context, tb testing.TB, opts ...Option) (context.Context, FinishFunc) {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}

	pc, _, _, _ := runtime.Caller(cfg.skip)
	suite, _ := utils.GetPackageAndName(pc)
	name := tb.Name()
	fqn := fmt.Sprintf("%s.%s", suite, name)

	testOpts := []tracer.StartSpanOption{
		tracer.ResourceName(fqn),
		tracer.Tag(constants.TestName, name),
		tracer.Tag(constants.TestSuite, suite),
		tracer.Tag(constants.TestFramework, testFramework),
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
		span.SetTag(ext.Error, tb.Failed())
		if tb.Failed() {
			span.SetTag(constants.TestStatus, constants.TestStatusFail)
		} else if tb.Skipped() {
			span.SetTag(constants.TestStatus, constants.TestStatusSkip)
		} else {
			span.SetTag(constants.TestStatus, constants.TestStatusPass)
		}
		span.Finish(cfg.finishOpts...)
	}
}
