// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package utils

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
)

func OSName() string {
	return runtime.GOOS
}

func OSVersion() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return constants.Unknown
	}
	return strings.Trim(string(out), "\n")
}
