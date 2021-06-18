// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setEnvs(env map[string]string) func() {
	restore := map[string]*string{}
	for key, value := range env {
		oldValue, ok := os.LookupEnv(key)
		if ok {
			restore[key] = &oldValue
		} else {
			restore[key] = nil
		}
		os.Setenv(key, value)
	}
	return func() {
		for key, value := range restore {
			if value == nil {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, *value)
			}
		}
	}
}

// TestTags asserts that all tags are extracted from environment variables.
func TestTags(t *testing.T) {
	// Reset provider env key when running in CI
	resetProviders := map[string]string{}
	for key := range providers {
		if value, ok := os.LookupEnv(key); ok {
			resetProviders[key] = value
			os.Unsetenv(key)
		}
	}
	defer func() {
		for key, value := range resetProviders {
			os.Setenv(key, value)
		}
	}()

	paths, err := filepath.Glob("testdata/fixtures/*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		providerName := strings.TrimSuffix(filepath.Base(path), ".json")

		t.Run(providerName, func(t *testing.T) {
			fp, err := os.Open(fmt.Sprintf("testdata/fixtures/%s.json", providerName))
			if err != nil {
				t.Fatal(err)
			}

			data, err := ioutil.ReadAll(fp)
			if err != nil {
				t.Fatal(err)
			}

			var examples [][]map[string]string
			if err := json.Unmarshal(data, &examples); err != nil {
				t.Fatal(err)
			}

			for i, line := range examples {
				name := fmt.Sprintf("%d", i)
				env := line[0]
				tags := line[1]

				t.Run(name, func(t *testing.T) {
					reset := setEnvs(env)
					defer reset()
					providerTags := GetProviderTags()

					for expectedKey, expectedValue := range tags {
						if actualValue, ok := providerTags[expectedKey]; ok {
							if expectedValue != actualValue {
								t.Fatalf("Key: %s, the actual value (%s) is different to the expected value (%s)", expectedKey, actualValue, expectedValue)
							}
						} else {
							t.Fatalf("Key: %s, doesn't exist.", expectedKey)
						}
					}
				})
			}
		})
	}
}
