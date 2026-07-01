// Package testprocess provides subprocess helpers for tests that need clean
// package-level state.
package testprocess

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"
)

// RunInFreshProcess reruns the current test in a new process gated by envName.
//
// Use it for assertions over package-level sync.Once state: the parent process
// may already have decoded data through another test, while the subprocess starts
// with clean package globals. The function returns true in the subprocess, where
// the caller should run the assertion.
func RunInFreshProcess(t testing.TB, envName string) bool {
	t.Helper()

	if os.Getenv(envName) == "1" {
		return true
	}

	testName := regexp.QuoteMeta(t.Name())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	//nolint:gosec // Test helper intentionally re-execs the current test binary with a fixed flag and env gate.
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+testName+"$", "-test.v")
	cmd.Env = append(os.Environ(), envName+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess %s failed: %v\n%s", t.Name(), err, out)
	}
	if !bytes.Contains(out, []byte("PASS")) {
		t.Fatalf("subprocess %s did not report PASS:\n%s", t.Name(), out)
	}
	return false
}
