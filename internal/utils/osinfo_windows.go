// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package utils

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	"golang.org/x/sys/windows/registry"
)

func OSName() string {
	return runtime.GOOS
}

func OSVersion() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return constants.Unknown
	}
	defer k.Close()

	var version strings.Builder

	maj, _, err := k.GetIntegerValue("CurrentMajorVersionNumber")
	if err == nil {
		version.WriteString(fmt.Sprintf("%d", maj))
		min, _, err := k.GetIntegerValue("CurrentMinorVersionNumber")
		if err == nil {
			version.WriteString(fmt.Sprintf(".%d", min))
		}
	} else {
		version.WriteString(constants.Unknown)
	}

	ed, _, err := k.GetStringValue("EditionID")
	if err == nil {
		version.WriteString(" " + ed)
	} else {
		version.WriteString(" Unknown Edition")
	}

	build, _, err := k.GetStringValue("CurrentBuild")
	if err == nil {
		version.WriteString(" Build " + build)
	} else {
		version.WriteString(" Unknown Build")
	}
	return version.String()
}
