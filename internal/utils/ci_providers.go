// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package utils

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	homedir "github.com/mitchellh/go-homedir"
)

type providerType = func() map[string]string

var providers = map[string]providerType{
	"APPVEYOR":           extractAppveyor,
	"TF_BUILD":           extractAzurePipelines,
	"BITBUCKET_COMMIT":   extractBitbucket,
	"BUILDKITE":          extractBuildkite,
	"CIRCLECI":           extractCircleCI,
	"GITHUB_SHA":         extractGithubActions,
	"GITLAB_CI":          extractGitlab,
	"JENKINS_URL":        extractJenkins,
	"TEAMCITY_VERSION":   extractTeamcity,
	"TRAVIS":             extractTravis,
	"BITRISE_BUILD_SLUG": extractBitrise,
}

// GetProviderTags extracts CI information from environment variables.
func GetProviderTags() map[string]string {
	tags := map[string]string{}
	for key, provider := range providers {
		if _, ok := os.LookupEnv(key); !ok {
			continue
		}
		tags = provider()

		if tag, ok := tags[constants.GitTag]; ok && tag != "" {
			tags[constants.GitTag] = normalizeRef(tag)
			delete(tags, constants.GitBranch)
		}
		if tag, ok := tags[constants.GitBranch]; ok && tag != "" {
			tags[constants.GitBranch] = normalizeRef(tag)
		}
		if tag, ok := tags[constants.GitRepositoryURL]; ok && tag != "" {
			tags[constants.GitRepositoryURL] = filterSensitiveInfo(tag)
		}

		// Expand ~
		if tag, ok := tags[constants.CIWorkspacePath]; ok && tag != "" {
			homedir.Reset()
			if value, err := homedir.Expand(tag); err == nil {
				tags[constants.CIWorkspacePath] = value
			}
		}
	}

	// replace with user specific tags
	replaceWithUserSpecificTags(tags)

	// remove empty values
	for tag, value := range tags {
		if value == "" {
			delete(tags, tag)
		}
	}

	return tags
}

func replaceWithUserSpecificTags(tags map[string]string) {

	replace := func(tagName, envName string) {
		tags[tagName] = getEnvironmentVariableIfIsNotEmpty(envName, tags[tagName])
	}

	replace(constants.GitBranch, "DD_GIT_BRANCH")
	replace(constants.GitTag, "DD_GIT_TAG")
	replace(constants.GitRepositoryURL, "DD_GIT_REPOSITORY_URL")
	replace(constants.GitCommitSHA, "DD_GIT_COMMIT_SHA")
	replace(constants.GitCommitMessage, "DD_GIT_COMMIT_MESSAGE")
	replace(constants.GitCommitAuthorName, "DD_GIT_COMMIT_AUTHOR_NAME")
	replace(constants.GitCommitAuthorEmail, "DD_GIT_COMMIT_AUTHOR_EMAIL")
	replace(constants.GitCommitAuthorDate, "DD_GIT_COMMIT_AUTHOR_DATE")
	replace(constants.GitCommitCommitterName, "DD_GIT_COMMIT_COMMITTER_NAME")
	replace(constants.GitCommitCommitterEmail, "DD_GIT_COMMIT_COMMITTER_EMAIL")
	replace(constants.GitCommitCommitterDate, "DD_GIT_COMMIT_COMMITTER_DATE")
}

func getEnvironmentVariableIfIsNotEmpty(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	} else {
		return defaultValue
	}
}

func normalizeRef(name string) string {
	empty := []byte("")
	refs := regexp.MustCompile("^refs/(heads/)?")
	origin := regexp.MustCompile("^origin/")
	tags := regexp.MustCompile("^tags/")
	return string(tags.ReplaceAll(origin.ReplaceAll(refs.ReplaceAll([]byte(name), empty), empty), empty)[:])
}

func filterSensitiveInfo(url string) string {
	return string(regexp.MustCompile("(https?://)[^/]*@").ReplaceAll([]byte(url), []byte("$1"))[:])
}

func lookupEnvs(keys ...string) ([]string, bool) {
	values := make([]string, len(keys))
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if !ok {
			return nil, false
		}
		values = append(values, value)
	}
	return values, true
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			if value != "" {
				return value
			}
		}
	}
	return ""
}

func extractAppveyor() map[string]string {
	tags := map[string]string{}
	url := fmt.Sprintf("https://ci.appveyor.com/project/%s/builds/%s", os.Getenv("APPVEYOR_REPO_NAME"), os.Getenv("APPVEYOR_BUILD_ID"))
	tags[constants.CIProviderName] = "appveyor"
	if os.Getenv("APPVEYOR_REPO_PROVIDER") == "github" {
		tags[constants.GitRepositoryURL] = fmt.Sprintf("https://github.com/%s.git", os.Getenv("APPVEYOR_REPO_NAME"))
	} else {
		tags[constants.GitRepositoryURL] = os.Getenv("APPVEYOR_REPO_NAME")
	}

	tags[constants.GitCommitSHA] = os.Getenv("APPVEYOR_REPO_COMMIT")
	tags[constants.GitBranch] = firstEnv("APPVEYOR_PULL_REQUEST_HEAD_REPO_BRANCH", "APPVEYOR_REPO_BRANCH")
	tags[constants.GitTag] = os.Getenv("APPVEYOR_REPO_TAG_NAME")

	tags[constants.CIWorkspacePath] = os.Getenv("APPVEYOR_BUILD_FOLDER")
	tags[constants.CIPipelineID] = os.Getenv("APPVEYOR_BUILD_ID")
	tags[constants.CIPipelineName] = os.Getenv("APPVEYOR_REPO_NAME")
	tags[constants.CIPipelineNumber] = os.Getenv("APPVEYOR_BUILD_NUMBER")
	tags[constants.CIPipelineURL] = url
	tags[constants.CIJobURL] = url
	tags[constants.GitCommitMessage] = os.Getenv("APPVEYOR_REPO_COMMIT_MESSAGE_EXTENDED")
	tags[constants.GitCommitAuthorName] = os.Getenv("APPVEYOR_REPO_COMMIT_AUTHOR")
	tags[constants.GitCommitAuthorEmail] = os.Getenv("APPVEYOR_REPO_COMMIT_AUTHOR_EMAIL")
	return tags
}

func extractAzurePipelines() map[string]string {
	tags := map[string]string{}
	baseURL := fmt.Sprintf("%s%s/_build/results?buildId=%s", os.Getenv("SYSTEM_TEAMFOUNDATIONSERVERURI"), os.Getenv("SYSTEM_TEAMPROJECTID"), os.Getenv("BUILD_BUILDID"))
	pipelineURL := baseURL
	jobURL := fmt.Sprintf("%s&view=logs&j=%s&t=%s", baseURL, os.Getenv("SYSTEM_JOBID"), os.Getenv("SYSTEM_TASKINSTANCEID"))
	branchOrTag := firstEnv("SYSTEM_PULLREQUEST_SOURCEBRANCH", "BUILD_SOURCEBRANCH", "BUILD_SOURCEBRANCHNAME")
	branch := ""
	tag := ""
	if strings.Contains(branchOrTag, "tags/") {
		tag = branchOrTag
	} else {
		branch = branchOrTag
	}
	tags[constants.CIProviderName] = "azurepipelines"
	tags[constants.CIWorkspacePath] = os.Getenv("BUILD_SOURCESDIRECTORY")

	tags[constants.CIPipelineID] = os.Getenv("BUILD_BUILDID")
	tags[constants.CIPipelineName] = os.Getenv("BUILD_DEFINITIONNAME")
	tags[constants.CIPipelineNumber] = os.Getenv("BUILD_BUILDID")
	tags[constants.CIPipelineURL] = pipelineURL

	tags[constants.CIStageName] = os.Getenv("SYSTEM_STAGEDISPLAYNAME")

	tags[constants.CIJobName] = os.Getenv("SYSTEM_JOBDISPLAYNAME")
	tags[constants.CIJobURL] = jobURL

	tags[constants.GitRepositoryURL] = firstEnv("SYSTEM_PULLREQUEST_SOURCEREPOSITORYURI", "BUILD_REPOSITORY_URI")
	tags[constants.GitCommitSHA] = firstEnv("SYSTEM_PULLREQUEST_SOURCECOMMITID", "BUILD_SOURCEVERSION")
	tags[constants.GitBranch] = branch
	tags[constants.GitTag] = tag
	tags[constants.GitCommitMessage] = os.Getenv("BUILD_SOURCEVERSIONMESSAGE")
	tags[constants.GitCommitAuthorName] = os.Getenv("BUILD_REQUESTEDFORID")
	tags[constants.GitCommitAuthorEmail] = os.Getenv("BUILD_REQUESTEDFOREMAIL")
	return tags
}

func extractBitrise() map[string]string {
	tags := map[string]string{}
	tags[constants.CIProviderName] = "bitrise"
	tags[constants.GitRepositoryURL] = os.Getenv("GIT_REPOSITORY_URL")
	tags[constants.GitCommitSHA] = firstEnv("BITRISE_GIT_COMMIT", "GIT_CLONE_COMMIT_HASH")
	tags[constants.GitBranch] = firstEnv("BITRISEIO_GIT_BRANCH_DEST", "BITRISE_GIT_BRANCH")
	tags[constants.GitTag] = os.Getenv("BITRISE_GIT_TAG")
	tags[constants.CIWorkspacePath] = os.Getenv("BITRISE_SOURCE_DIR")
	tags[constants.CIPipelineID] = os.Getenv("BITRISE_BUILD_SLUG")
	tags[constants.CIPipelineName] = os.Getenv("BITRISE_TRIGGERED_WORKFLOW_ID")
	tags[constants.CIPipelineNumber] = os.Getenv("BITRISE_BUILD_NUMBER")
	tags[constants.CIPipelineURL] = os.Getenv("BITRISE_BUILD_URL")
	tags[constants.GitCommitMessage] = os.Getenv("BITRISE_GIT_MESSAGE")
	return tags
}

func extractBitbucket() map[string]string {
	tags := map[string]string{}
	url := fmt.Sprintf("https://bitbucket.org/%s/addon/pipelines/home#!/results/%s", os.Getenv("BITBUCKET_REPO_FULL_NAME"), os.Getenv("BITBUCKET_BUILD_NUMBER"))
	tags[constants.CIProviderName] = "bitbucket"
	tags[constants.GitRepositoryURL] = os.Getenv("BITBUCKET_GIT_SSH_ORIGIN")
	tags[constants.GitCommitSHA] = os.Getenv("BITBUCKET_COMMIT")
	tags[constants.GitBranch] = os.Getenv("BITBUCKET_BRANCH")
	tags[constants.GitTag] = os.Getenv("BITBUCKET_TAG")
	tags[constants.CIWorkspacePath] = os.Getenv("BITBUCKET_CLONE_DIR")
	tags[constants.CIPipelineID] = strings.Trim(os.Getenv("BITBUCKET_PIPELINE_UUID"), "{}")
	tags[constants.CIPipelineNumber] = os.Getenv("BITBUCKET_BUILD_NUMBER")
	tags[constants.CIPipelineName] = os.Getenv("BITBUCKET_REPO_FULL_NAME")
	tags[constants.CIPipelineURL] = url
	tags[constants.CIJobURL] = url
	return tags
}

func extractBuildkite() map[string]string {
	tags := map[string]string{}
	tags[constants.GitBranch] = os.Getenv("BUILDKITE_BRANCH")
	tags[constants.GitCommitSHA] = os.Getenv("BUILDKITE_COMMIT")
	tags[constants.GitRepositoryURL] = os.Getenv("BUILDKITE_REPO")
	tags[constants.GitTag] = os.Getenv("BUILDKITE_TAG")
	tags[constants.CIPipelineID] = os.Getenv("BUILDKITE_BUILD_ID")
	tags[constants.CIPipelineName] = os.Getenv("BUILDKITE_PIPELINE_SLUG")
	tags[constants.CIPipelineNumber] = os.Getenv("BUILDKITE_BUILD_NUMBER")
	tags[constants.CIPipelineURL] = os.Getenv("BUILDKITE_BUILD_URL")
	tags[constants.CIJobURL] = fmt.Sprintf("%s#%s", os.Getenv("BUILDKITE_BUILD_URL"), os.Getenv("BUILDKITE_JOB_ID"))
	tags[constants.CIProviderName] = "buildkite"
	tags[constants.CIWorkspacePath] = os.Getenv("BUILDKITE_BUILD_CHECKOUT_PATH")
	tags[constants.GitCommitMessage] = os.Getenv("BUILDKITE_MESSAGE")
	tags[constants.GitCommitAuthorName] = os.Getenv("BUILDKITE_BUILD_AUTHOR")
	tags[constants.GitCommitAuthorEmail] = os.Getenv("BUILDKITE_BUILD_AUTHOR_EMAIL")
	return tags
}

func extractCircleCI() map[string]string {
	tags := map[string]string{}
	tags[constants.CIProviderName] = "circleci"
	tags[constants.GitRepositoryURL] = os.Getenv("CIRCLE_REPOSITORY_URL")
	tags[constants.GitCommitSHA] = os.Getenv("CIRCLE_SHA1")
	tags[constants.GitTag] = os.Getenv("CIRCLE_TAG")
	tags[constants.GitBranch] = os.Getenv("CIRCLE_BRANCH")
	tags[constants.CIWorkspacePath] = os.Getenv("CIRCLE_WORKING_DIRECTORY")
	tags[constants.CIPipelineID] = os.Getenv("CIRCLE_WORKFLOW_ID")
	tags[constants.CIPipelineName] = os.Getenv("CIRCLE_PROJECT_REPONAME")
	tags[constants.CIPipelineNumber] = os.Getenv("CIRCLE_BUILD_NUM")
	tags[constants.CIPipelineURL] = fmt.Sprintf("https://app.circleci.com/pipelines/workflows/%s", os.Getenv("CIRCLE_WORKFLOW_ID"))
	tags[constants.CIJobName] = os.Getenv("CIRCLE_JOB")
	tags[constants.CIJobURL] = os.Getenv("CIRCLE_BUILD_URL")
	return tags
}

func extractGithubActions() map[string]string {
	tags := map[string]string{}
	branchOrTag := firstEnv("GITHUB_HEAD_REF", "GITHUB_REF")
	tag := ""
	branch := ""
	if strings.Contains(branchOrTag, "tags/") {
		tag = branchOrTag
	} else {
		branch = branchOrTag
	}

	url := fmt.Sprintf("https://github.com/%s/commit/%s/checks", os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_SHA"))
	tags[constants.CIProviderName] = "github"
	tags[constants.GitRepositoryURL] = fmt.Sprintf("https://github.com/%s.git", os.Getenv("GITHUB_REPOSITORY"))
	tags[constants.GitCommitSHA] = os.Getenv("GITHUB_SHA")
	tags[constants.GitBranch] = branch
	tags[constants.GitTag] = tag
	tags[constants.CIWorkspacePath] = os.Getenv("GITHUB_WORKSPACE")
	tags[constants.CIPipelineID] = os.Getenv("GITHUB_RUN_ID")
	tags[constants.CIPipelineNumber] = os.Getenv("GITHUB_RUN_NUMBER")
	tags[constants.CIPipelineName] = os.Getenv("GITHUB_WORKFLOW")
	tags[constants.CIPipelineURL] = url
	tags[constants.CIJobURL] = url
	return tags
}

func extractGitlab() map[string]string {
	tags := map[string]string{}
	url := os.Getenv("CI_PIPELINE_URL")
	url = string(regexp.MustCompile("/-/pipelines/").ReplaceAll([]byte(url), []byte("/pipelines/"))[:])
	url = strings.ReplaceAll(url, "/-/pipelines/", "/pipelines/")

	tags[constants.CIProviderName] = "gitlab"
	tags[constants.GitRepositoryURL] = os.Getenv("CI_REPOSITORY_URL")
	tags[constants.GitCommitSHA] = os.Getenv("CI_COMMIT_SHA")
	tags[constants.GitBranch] = firstEnv("CI_COMMIT_BRANCH", "CI_COMMIT_REF_NAME")
	tags[constants.GitTag] = os.Getenv("CI_COMMIT_TAG")
	tags[constants.CIWorkspacePath] = os.Getenv("CI_PROJECT_DIR")
	tags[constants.CIPipelineID] = os.Getenv("CI_PIPELINE_ID")
	tags[constants.CIPipelineName] = os.Getenv("CI_PROJECT_PATH")
	tags[constants.CIPipelineNumber] = os.Getenv("CI_PIPELINE_IID")
	tags[constants.CIPipelineURL] = url
	tags[constants.CIJobURL] = os.Getenv("CI_JOB_URL")
	tags[constants.CIJobName] = os.Getenv("CI_JOB_NAME")
	tags[constants.CIStageName] = os.Getenv("CI_JOB_STAGE")
	tags[constants.GitCommitMessage] = os.Getenv("CI_COMMIT_MESSAGE")

	author := os.Getenv("CI_COMMIT_AUTHOR")
	authorArray := strings.FieldsFunc(author, func(s rune) bool {
		return s == '<' || s == '>'
	})
	tags[constants.GitCommitAuthorName] = strings.TrimSpace(authorArray[0])
	tags[constants.GitCommitAuthorEmail] = strings.TrimSpace(authorArray[1])
	tags[constants.GitCommitAuthorDate] = os.Getenv("CI_COMMIT_TIMESTAMP")
	return tags
}

func extractJenkins() map[string]string {
	tags := map[string]string{}
	tags[constants.CIProviderName] = "jenkins"
	tags[constants.GitRepositoryURL] = firstEnv("GIT_URL", "GIT_URL_1")
	tags[constants.GitCommitSHA] = os.Getenv("GIT_COMMIT")

	branchOrTag := os.Getenv("GIT_BRANCH")
	empty := []byte("")
	name, hasName := os.LookupEnv("JOB_NAME")

	if strings.Contains(branchOrTag, "tags/") {
		tags[constants.GitTag] = branchOrTag
	} else {
		tags[constants.GitBranch] = branchOrTag
		// remove branch for job name
		removeBranch := regexp.MustCompile(fmt.Sprintf("/%s", normalizeRef(branchOrTag)))
		name = string(removeBranch.ReplaceAll([]byte(name), empty))
	}

	if hasName {
		removeVars := regexp.MustCompile("/[^/]+=[^/]*")
		name = string(removeVars.ReplaceAll([]byte(name), empty))
	}

	tags[constants.CIWorkspacePath] = os.Getenv("WORKSPACE")
	tags[constants.CIPipelineID] = os.Getenv("BUILD_TAG")
	tags[constants.CIPipelineNumber] = os.Getenv("BUILD_NUMBER")
	tags[constants.CIPipelineName] = name
	tags[constants.CIPipelineURL] = os.Getenv("BUILD_URL")
	return tags
}

func extractTeamcity() map[string]string {
	tags := map[string]string{}
	tags[constants.CIProviderName] = "teamcity"
	tags[constants.GitRepositoryURL] = os.Getenv("BUILD_VCS_URL")
	tags[constants.GitCommitSHA] = os.Getenv("BUILD_VCS_NUMBER")
	tags[constants.CIWorkspacePath] = os.Getenv("BUILD_CHECKOUTDIR")
	tags[constants.CIPipelineID] = os.Getenv("BUILD_ID")
	tags[constants.CIPipelineNumber] = os.Getenv("BUILD_NUMBER")
	tags[constants.CIPipelineURL] = fmt.Sprintf("%s/viewLog.html?buildId=%s", os.Getenv("SERVER_URL"), os.Getenv("BUILD_ID"))
	return tags
}

func extractTravis() map[string]string {
	tags := map[string]string{}
	prSlug := os.Getenv("TRAVIS_PULL_REQUEST_SLUG")
	repoSlug := prSlug
	if strings.TrimSpace(repoSlug) == "" {
		repoSlug = os.Getenv("TRAVIS_REPO_SLUG")
	}
	tags[constants.CIProviderName] = "travisci"
	tags[constants.GitRepositoryURL] = fmt.Sprintf("https://github.com/%s.git", repoSlug)
	tags[constants.GitCommitSHA] = os.Getenv("TRAVIS_COMMIT")
	tags[constants.GitTag] = os.Getenv("TRAVIS_TAG")
	tags[constants.GitBranch] = firstEnv("TRAVIS_PULL_REQUEST_BRANCH", "TRAVIS_BRANCH")
	tags[constants.CIWorkspacePath] = os.Getenv("TRAVIS_BUILD_DIR")
	tags[constants.CIPipelineID] = os.Getenv("TRAVIS_BUILD_ID")
	tags[constants.CIPipelineNumber] = os.Getenv("TRAVIS_BUILD_NUMBER")
	tags[constants.CIPipelineName] = repoSlug
	tags[constants.CIPipelineURL] = os.Getenv("TRAVIS_BUILD_WEB_URL")
	tags[constants.CIJobURL] = os.Getenv("TRAVIS_JOB_WEB_URL")
	tags[constants.GitCommitMessage] = os.Getenv("TRAVIS_COMMIT_MESSAGE")
	return tags
}
