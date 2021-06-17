// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package utils

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type LocalGitData struct {
	SourceRoot     string
	RepositoryUrl  string
	Branch         string
	CommitSha      string
	AuthorDate     time.Time
	AuthorName     string
	AuthorEmail    string
	CommitterDate  time.Time
	CommitterName  string
	CommitterEmail string
	CommitMessage  string
}

// LocalGetGitData get the git data from the HEAD in Git repository
func LocalGetGitData() (LocalGitData, error) {
	gitData := LocalGitData{}

	// Extract git working folder
	out, err := exec.Command("git", "rev-parse", "--absolute-git-dir").Output()
	if err != nil {
		return gitData, err
	}
	gitData.SourceRoot = strings.ReplaceAll(strings.Trim(string(out), "\n"), ".git", "")

	// Extract repository data
	out, err = exec.Command("git", "ls-remote", "--get-url").Output()
	if err != nil {
		return gitData, err
	}
	gitData.RepositoryUrl = strings.Trim(string(out), "\n")

	// Extract the branch name
	out, err = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return gitData, err
	}
	gitData.Branch = strings.Trim(string(out), "\n")

	// Get remaining data from the git log command: git log -1 --pretty='%H","%aI","%an","%ae","%cI","%cn","%ce","%B'
	out, err = exec.Command("git", "log", "-1", "--pretty=%H\",\"%at\",\"%an\",\"%ae\",\"%ct\",\"%cn\",\"%ce\",\"%B").Output()
	if err != nil {
		return gitData, err
	}
	outArray := strings.Split(string(out), "\",\"")
	authorUnixDate, _ := strconv.ParseInt(outArray[1], 10, 64)
	committerUnixDate, _ := strconv.ParseInt(outArray[4], 10, 64)

	gitData.CommitSha = outArray[0]
	gitData.AuthorDate = time.Unix(authorUnixDate, 0)
	gitData.AuthorName = outArray[2]
	gitData.AuthorEmail = outArray[3]
	gitData.CommitterDate = time.Unix(committerUnixDate, 0)
	gitData.CommitterName = outArray[5]
	gitData.CommitterEmail = outArray[6]
	gitData.CommitMessage = strings.Trim(outArray[7], "\n")

	fmt.Println(gitData)
	return gitData, nil
}
