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
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type (
	TestType int
	TestData struct {
		Type    TestType
		Suite   string
		Name    string
		Options []Option
	}

	Status         int
	TestFinishFunc func(status Status, r interface{})
)

const (
	spanKind      = "test"
	testFramework = "golang.org/pkg/testing"

	TypeTest      TestType = 0
	TypeBenchmark TestType = 1

	StatusPass Status = 0
	StatusFail Status = 1
	StatusSkip Status = 2
)

var (
	repoRegex     = regexp.MustCompile(`(?m)\/([a-zA-Z0-9\\\-_.]*)$`)
	initMutex     = sync.Mutex{}
	isInitialized = false
)

// Initialize the testing sdk, returns a func to defer the finalization
func Initialize(opts ...tracer.StartOption) func() {
	initMutex.Lock()
	defer initMutex.Unlock()

	// Checks if was already initialized
	if isInitialized {
		return func() {
			tracer.Flush()
		}
	}

	// Preload all CI and Git tags.
	loadTags()

	// Check if DD_SERVICE has been set; otherwise we default to repo name.
	if v := os.Getenv("DD_SERVICE"); v == "" {
		if repoUrl, ok := tags[constants.GitRepositoryURL]; ok {
			matches := repoRegex.FindStringSubmatch(repoUrl)
			if len(matches) > 1 {
				repoUrl = strings.TrimSuffix(matches[1], ".git")
			}
			opts = append(opts, tracer.WithService(repoUrl))
		}
	}

	// Initialize tracer
	tracer.Start(opts...)

	// Handle SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		tracer.Stop()
		os.Exit(0)
	}()

	return func() {
		tracer.Flush()
		tracer.Stop()
	}
}

// StartCustomTestOrBenchmark starts a new test or benchmark
func StartCustomTestOrBenchmark(ctx context.Context, testData TestData) (context.Context, TestFinishFunc) {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range testData.Options {
		fn(cfg)
	}

	fqn := fmt.Sprintf("%s.%s", testData.Suite, testData.Name)

	testOpts := []tracer.StartSpanOption{
		tracer.ResourceName(fqn),
		tracer.Tag(constants.TestName, testData.Name),
		tracer.Tag(constants.TestSuite, testData.Suite),
		tracer.Tag(constants.TestFramework, testFramework),
		tracer.Tag(constants.Origin, constants.CIAppTestOrigin),
	}

	if testData.Type == TypeTest {
		testOpts = append(testOpts, tracer.Tag(constants.TestType, constants.TestTypeTest))
	} else if testData.Type == TypeBenchmark {
		testOpts = append(testOpts, tracer.Tag(constants.TestType, constants.TestTypeBenchmark))
	}
	cfg.spanOpts = append(testOpts, cfg.spanOpts...)

	span, ctx := tracer.StartSpanFromContext(ctx, constants.SpanTypeTest, cfg.spanOpts...)

	return ctx, func(status Status, r interface{}) {
		if r == nil {
			r = recover()
		}

		if r != nil {
			// Panic handling
			span.SetTag(constants.TestStatus, constants.TestStatusFail)
			span.SetTag(ext.Error, true)
			span.SetTag(ext.ErrorMsg, fmt.Sprint(r))
			span.SetTag(ext.ErrorStack, getStacktrace(2))
			span.SetTag(ext.ErrorType, "panic")
		} else {
			// Normal finalization
			if status == StatusFail {
				span.SetTag(ext.Error, true)
				span.SetTag(constants.TestStatus, constants.TestStatusFail)
			} else if status == StatusSkip {
				span.SetTag(ext.Error, false)
				span.SetTag(constants.TestStatus, constants.TestStatusSkip)
			} else if status == StatusPass {
				span.SetTag(ext.Error, false)
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
