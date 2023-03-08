// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package utils

import (
	"os/exec"
	"strings"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
)

func OSVersion() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return constants.Unknown
	}
	return strings.Split(string(out), "-")[0]
}
