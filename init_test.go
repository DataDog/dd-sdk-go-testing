// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(Run(m))
}

func TestRun(t *testing.T) {
	_, finish := StartTest(t)
	defer finish()

	t.Run("sub-test", func(t *testing.T) {
		_, finish := StartTest(t)
		defer finish()
	})
}
