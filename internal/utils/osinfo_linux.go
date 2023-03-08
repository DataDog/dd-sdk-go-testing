// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package utils

import (
	"bufio"
	"os"
	"strings"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
)

func OSVersion() string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return constants.Unknown
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	version := constants.Unknown
	for s.Scan() {
		parts := strings.SplitN(s.Text(), "=", 2)
		switch parts[0] {
		case "VERSION":
			version = strings.Trim(parts[1], "\"")
		case "VERSION_ID":
			if version == "" {
				version = strings.Trim(parts[1], "\"")
			}
		}
	}
	return version
}
