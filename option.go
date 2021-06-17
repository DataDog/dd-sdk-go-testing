// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	"github.com/DataDog/dd-sdk-go-testing/internal/utils"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"runtime"
)

var (
	// tags contains information detected from CI/CD environment variables.
	tags map[string]string
)

type config struct {
	skip       int
	spanOpts   []ddtrace.StartSpanOption
	finishOpts []ddtrace.FinishOption
}

// Option represents an option that can be passed to NewServeMux or WrapHandler.
type Option func(*config)

func defaults(cfg *config) {
	// When StartSpanWithFinish is called directly from test function.
	cfg.skip = 1
	cfg.spanOpts = []ddtrace.StartSpanOption{
		tracer.SpanType(constants.SpanTypeTest),
		tracer.Tag(constants.SpanKind, spanKind),
		tracer.Tag(ext.ManualKeep, true),
	}

	// Load tags
	if tags == nil {
		loadTags()
	}

	for k, v := range tags {
		cfg.spanOpts = append(cfg.spanOpts, tracer.Tag(k, v))
	}

	cfg.finishOpts = []ddtrace.FinishOption{}
}

func loadTags() {
	tags = map[string]string{
		constants.OSPlatform:     utils.OSName(),
		constants.OSVersion:      utils.OSVersion(),
		constants.OSArchitecture: runtime.GOARCH,
		constants.RuntimeName:    runtime.Compiler,
		constants.RuntimeVersion: runtime.Version(),
	}
	/*
		tags = ci.Tags()

		// Guess Git metadata from a local Git repository otherwise.
		if _, ok := tags[ext.GitRepositoryURL]; !ok {
			tags[ext.GitRepositoryURL] = ext.LocalGitRepositoryURL()
		}
		if _, ok := tags[ext.GitCommitSHA]; !ok {
			tags[ext.GitCommitSHA] = ext.LocalGitCommitSHA()
		}
		if _, ok := tags[ext.GitBranch]; !ok {
			tags[ext.GitBranch] = ext.LocalGitBranch()
		}
	*/
}

// WithSpanOptions defines a set of additional ddtrace.StartSpanOption to be added
// to spans started by the integration.
func WithSpanOptions(opts ...ddtrace.StartSpanOption) Option {
	return func(cfg *config) {
		cfg.spanOpts = append(cfg.spanOpts, opts...)
	}
}

// WithSkipFrames defines a how many frames should be skipped for caller autodetection.
// The value should be changed if StartSpanWithFinish is called from a custom wrapper.
func WithSkipFrames(skip int) Option {
	return func(cfg *config) {
		cfg.skip = skip
	}
}
