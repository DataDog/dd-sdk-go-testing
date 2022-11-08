package utils

import (
	"runtime"
	"strings"

	"github.com/onsi/ginkgo/v2"
)

func IncrementOffset(additionalOffset ginkgo.Offset, args ...interface{}) []interface{} {
	for i, arg := range args {
		if v, ok := arg.(ginkgo.Offset); ok {
			args[i] = v + additionalOffset

			return args
		}
	}

	return append(args, additionalOffset)
}

func GetSuiteName() string {
	pc, _, _, _ := runtime.Caller(2)
	suite, _ := GetPackageAndName(pc)

	return suite
}

func GetPackageAndName(pc uintptr) (string, string) {
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
