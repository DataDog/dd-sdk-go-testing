// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"runtime"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	"github.com/DataDog/dd-sdk-go-testing/internal/utils"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
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
	tags = utils.GetProviderTags()
	tags[constants.OSPlatform] = utils.OSName()
	tags[constants.OSVersion] = utils.OSVersion()
	tags[constants.OSArchitecture] = runtime.GOARCH
	tags[constants.RuntimeName] = runtime.Compiler
	tags[constants.RuntimeVersion] = runtime.Version()

	gitData, _ := utils.LocalGetGitData()

	// Guess Git metadata from a local Git repository otherwise.
	if _, ok := tags[constants.CIWorkspacePath]; !ok {
		tags[constants.CIWorkspacePath] = gitData.SourceRoot
	}
	if _, ok := tags[constants.GitRepositoryURL]; !ok {
		tags[constants.GitRepositoryURL] = gitData.RepositoryUrl
	}
	if _, ok := tags[constants.GitCommitSHA]; !ok {
		tags[constants.GitCommitSHA] = gitData.CommitSha
	}
	if _, ok := tags[constants.GitBranch]; !ok {
		tags[constants.GitBranch] = gitData.Branch
	}

	if tags[constants.GitCommitSHA] == gitData.CommitSha {
		if _, ok := tags[constants.GitCommitAuthorDate]; !ok {
			tags[constants.GitCommitAuthorDate] = gitData.AuthorDate.String()
		}
		if _, ok := tags[constants.GitCommitAuthorName]; !ok {
			tags[constants.GitCommitAuthorName] = gitData.AuthorName
		}
		if _, ok := tags[constants.GitCommitAuthorEmail]; !ok {
			tags[constants.GitCommitAuthorEmail] = gitData.AuthorEmail
		}
		if _, ok := tags[constants.GitCommitCommitterDate]; !ok {
			tags[constants.GitCommitCommitterDate] = gitData.CommitterDate.String()
		}
		if _, ok := tags[constants.GitCommitCommitterName]; !ok {
			tags[constants.GitCommitCommitterName] = gitData.CommitterName
		}
		if _, ok := tags[constants.GitCommitCommitterEmail]; !ok {
			tags[constants.GitCommitCommitterEmail] = gitData.CommitterEmail
		}
		if _, ok := tags[constants.GitCommitMessage]; !ok {
			tags[constants.GitCommitMessage] = gitData.CommitMessage
		}
	}
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

// WithIncrementSkipFrame increments how many frames should be skipped for caller by 1.
func WithIncrementSkipFrame() Option {
	return func(cfg *config) {
		cfg.skip = cfg.skip + 1
	}
}

// WithFinishOptions defines a set of additional ddtrace.FinishOption to be added
// to spans started by the integration.
func WithFinishOptions(opts ...ddtrace.FinishOption) Option {
	return func(cfg *config) {
		cfg.finishOpts = append(cfg.finishOpts, opts...)
	}
}
