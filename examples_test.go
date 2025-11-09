// Tideland Go Worker - Examples Integration Test
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestExamplesBuild verifies that all example programs compile successfully.
func TestExamplesBuild(t *testing.T) {
	// Skip if short testing is enabled
	if testing.Short() {
		t.Skip("skipping examples build test in short mode")
	}

	examples := []struct {
		name string
		path string
	}{
		{
			name: "config",
			path: "./examples/config",
		},
		{
			name: "safety",
			path: "./examples/safety",
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			// Create a temporary directory for the build output
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, ex.name)
			if runtime.GOOS == "windows" {
				outputPath += ".exe"
			}

			// Build the example
			cmd := exec.Command("go", "build", "-o", outputPath, ex.path)
			cmd.Env = os.Environ()

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("Build output: %s", output)
				t.Fatalf("failed to build example %s: %v", ex.name, err)
			}

			// Verify the executable was created
			info, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("executable not created for %s: %v", ex.name, err)
			}
			if !info.Mode().IsRegular() {
				t.Errorf("output is not a regular file for %s", ex.name)
			}

			// On Unix-like systems, check if it's executable
			if runtime.GOOS != "windows" {
				if info.Mode()&0111 == 0 {
					t.Errorf("file is not executable for %s", ex.name)
				}
			}
		})
	}
}

// TestExamplesRun verifies that example programs run without panicking.
// This test is skipped by default as it actually runs the examples.
func TestExamplesRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping examples run test in short mode")
	}

	// Check for environment variable to enable this test
	if os.Getenv("TEST_EXAMPLES_RUN") != "1" {
		t.Skip("skipping examples run test (set TEST_EXAMPLES_RUN=1 to enable)")
	}

	examples := []struct {
		name string
		path string
	}{
		{
			name: "config",
			path: "./examples/config",
		},
		{
			name: "safety",
			path: "./examples/safety",
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			// Run the example with go run
			cmd := exec.Command("go", "run", ex.path)
			cmd.Env = os.Environ()

			output, err := cmd.CombinedOutput()
			t.Logf("Example %s output:\n%s", ex.name, output)

			// We expect these examples to run successfully
			if err != nil {
				t.Fatalf("example %s failed to run: %v", ex.name, err)
			}
		})
	}
}
