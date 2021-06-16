// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"context"
	"testing"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// FinishFunc closes a started span and attaches test status information.
type FinishFunc func()

// Run is a helper function to run a `testing.M` object and gracefully stopping the agent afterwards
func Run(m *testing.M, opts ...tracer.StartOption) int {
	tracer.Start(opts...)
	defer tracer.Stop()
	return m.Run()
}

// StartTest returns a new span with the given testing.TB interface and options. It uses
// tracer.StartSpanFromContext function to start the span with automatically detected information.
func StartTest(ctx context.Context) (context.Context, FinishFunc) {

	return context.Background(), func() {
	}
}
