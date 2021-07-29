// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"context"
	"runtime"
	"testing"

	"github.com/DataDog/dd-sdk-go-testing/internal/utils"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// FinishFunc closes a started span and attaches test status information.
type FinishFunc func()

// Run is a helper function to run a `testing.M` object and gracefully stopping the tracer afterwards
func Run(m *testing.M, opts ...tracer.StartOption) int {
	// Initialize sdk
	finalizer := Initialize(opts...)
	defer finalizer()

	// Execute test suite
	return m.Run()
}

// StartTest returns a new span with the given testing.TB interface and options. It uses
// tracer.StartSpanFromContext function to start the span with automatically detected information.
func StartTest(tb testing.TB, opts ...Option) (context.Context, FinishFunc) {
	opts = append(opts, WithIncrementSkipFrame())
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

	testType := TypeTest
	switch tb.(type) {
	case *testing.T:
		testType = TypeTest
	case *testing.B:
		testType = TypeBenchmark
	}

	testData := TestData{
		Type:    testType,
		Suite:   suite,
		Name:    tb.Name(),
		Options: opts,
	}
	ctx, finishFunc := StartCustomTestOrBenchmark(ctx, testData)
	return ctx, func() {
		r := recover()
		if tb.Failed() {
			finishFunc(StatusFail, r)
		} else if tb.Skipped() {
			finishFunc(StatusSkip, r)
		} else {
			finishFunc(StatusPass, r)
		}
	}
}
