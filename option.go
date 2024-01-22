// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"runtime"
	"sync"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	"github.com/DataDog/dd-sdk-go-testing/internal/utils"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
)

var (
	// tags contains information detected from CI/CD environment variables.
	tags     map[string]string
	tagsOnce sync.Once
)

type config struct {
	skip       int
	spanOpts   []tracer.StartSpanOption
	finishOpts []tracer.FinishOption
}

// Option represents an option that can be passed to NewServeMux or WrapHandler.
type Option func(*config)

func defaults(cfg *config) {
	// When StartSpanWithFinish is called directly from test function.
	cfg.skip = 1
	cfg.spanOpts = []tracer.StartSpanOption{
		tracer.SpanType(constants.SpanTypeTest),
		tracer.Tag(constants.SpanKind, spanKind),
		tracer.Tag(ext.ManualKeep, true),
	}

	// Ensure CI tags
	ensureCITags()
	forEachCITags(func(k, v string) {
		cfg.spanOpts = append(cfg.spanOpts, tracer.Tag(k, v))
	})

	cfg.finishOpts = []tracer.FinishOption{}
}

func ensureCITags() {
	tagsOnce.Do(ensureCITagsLocked)
}

func ensureCITagsLocked() {
	localTags := utils.GetProviderTags()
	localTags[constants.OSPlatform] = runtime.GOOS
	localTags[constants.OSVersion] = utils.OSVersion()
	localTags[constants.OSArchitecture] = runtime.GOARCH
	localTags[constants.RuntimeName] = runtime.Compiler
	localTags[constants.RuntimeVersion] = runtime.Version()

	gitData, _ := utils.LocalGetGitData()

	// Guess Git metadata from a local Git repository otherwise.
	if _, ok := localTags[constants.CIWorkspacePath]; !ok {
		localTags[constants.CIWorkspacePath] = gitData.SourceRoot
	}
	if _, ok := localTags[constants.GitRepositoryURL]; !ok {
		localTags[constants.GitRepositoryURL] = gitData.RepositoryUrl
	}
	if _, ok := localTags[constants.GitCommitSHA]; !ok {
		localTags[constants.GitCommitSHA] = gitData.CommitSha
	}
	if _, ok := localTags[constants.GitBranch]; !ok {
		localTags[constants.GitBranch] = gitData.Branch
	}

	if localTags[constants.GitCommitSHA] == gitData.CommitSha {
		if _, ok := localTags[constants.GitCommitAuthorDate]; !ok {
			localTags[constants.GitCommitAuthorDate] = gitData.AuthorDate.String()
		}
		if _, ok := localTags[constants.GitCommitAuthorName]; !ok {
			localTags[constants.GitCommitAuthorName] = gitData.AuthorName
		}
		if _, ok := localTags[constants.GitCommitAuthorEmail]; !ok {
			localTags[constants.GitCommitAuthorEmail] = gitData.AuthorEmail
		}
		if _, ok := localTags[constants.GitCommitCommitterDate]; !ok {
			localTags[constants.GitCommitCommitterDate] = gitData.CommitterDate.String()
		}
		if _, ok := localTags[constants.GitCommitCommitterName]; !ok {
			localTags[constants.GitCommitCommitterName] = gitData.CommitterName
		}
		if _, ok := localTags[constants.GitCommitCommitterEmail]; !ok {
			localTags[constants.GitCommitCommitterEmail] = gitData.CommitterEmail
		}
		if _, ok := localTags[constants.GitCommitMessage]; !ok {
			localTags[constants.GitCommitMessage] = gitData.CommitMessage
		}
	}

	// Replace global tags with local copy
	tags = localTags
}

func getFromCITags(key string) (string, bool) {
	if value, ok := tags[key]; ok {
		return value, ok
	}

	return "", false
}

func forEachCITags(itemFunc func(string, string)) {
	for k, v := range tags {
		itemFunc(k, v)
	}
}

// WithSpanOptions defines a set of additional tracer.StartSpanOption to be added
// to spans started by the integration.
func WithSpanOptions(opts ...tracer.StartSpanOption) Option {
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
