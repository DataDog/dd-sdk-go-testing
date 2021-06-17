// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package utils

import (
	"runtime"
	"strings"
)

func GetPackageAndName(pc uintptr) (suite string, name string) {
	funcFullName := runtime.FuncForPC(pc).Name()
	lastSlash := strings.LastIndexByte(funcFullName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	firstDot := strings.IndexByte(funcFullName[lastSlash:], '.') + lastSlash
	packName := funcFullName[:firstDot]
	funcName := funcFullName[firstDot+1:]
	return packName, funcName
}
