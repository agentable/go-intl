package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

const runCLIEnv = "GO_INTL_CHECK_GENERATED_DATA_RUN_CLI"

func TestMainCommand(t *testing.T) {
	committed := t.TempDir()
	generated := t.TempDir()
	const path = "cldr/number/data.go"
	contents := generatedTestHeader + "package number\n"
	writeGeneratedTestFile(t, committed, path, contents)
	writeGeneratedTestFile(t, generated, path, contents)

	tests := []struct {
		name       string
		args       []string
		wantExit   int
		wantStderr string
	}{
		{
			name:     "identical roots",
			args:     []string{"-committed", committed, "-generated", generated},
			wantExit: 0,
		},
		{
			name:       "invalid arguments",
			wantExit:   1,
			wantStderr: "usage: check-generated-data",
		},
		{
			name:       "different generated data",
			args:       []string{"-committed", committed, "-generated", t.TempDir()},
			wantExit:   1,
			wantStderr: path,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestMainProcess$") // #nosec G204 -- re-executes this test binary.
			cmd.Env = append(os.Environ(), runCLIEnv+"=1")
			cmd.Args = append(cmd.Args, "--")
			cmd.Args = append(cmd.Args, tc.args...)
			stderr, err := cmd.CombinedOutput()
			if got := cmd.ProcessState.ExitCode(); got != tc.wantExit {
				t.Fatalf("exit code = %d, want %d; error = %v; stderr = %q", got, tc.wantExit, err, stderr)
			}
			if got := string(stderr); !strings.Contains(got, tc.wantStderr) {
				t.Fatalf("stderr = %q, want substring %q", got, tc.wantStderr)
			}
		})
	}
}

func TestMainProcess(t *testing.T) {
	if os.Getenv(runCLIEnv) != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{os.Args[0]}, os.Args[i+1:]...)
			main()
			return
		}
	}
	t.Fatal("CLI arguments separator is missing")
}

func TestRunComparesGeneratedRoots(t *testing.T) {
	t.Parallel()

	committed := t.TempDir()
	generated := t.TempDir()
	const path = "cldr/number/data.go"
	contents := generatedTestHeader + "package number\n"
	writeGeneratedTestFile(t, committed, path, contents)
	writeGeneratedTestFile(t, generated, path, contents)

	if err := run([]string{"-committed", committed, "-generated", generated}); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing roots", want: "usage:"},
		{name: "missing generated root", args: []string{"-committed", t.TempDir()}, want: "usage:"},
		{name: "extra argument", args: []string{"-committed", t.TempDir(), "-generated", t.TempDir(), "extra"}, want: "usage:"},
		{name: "unknown flag", args: []string{"-unknown"}, want: "flag provided but not defined"},
		{name: "missing committed directory", args: []string{"-committed", t.TempDir() + "/missing", "-generated", t.TempDir()}, want: "open generated files root"},
		{name: "missing generated directory", args: []string{"-committed", t.TempDir(), "-generated", t.TempDir() + "/missing"}, want: "open generated files root"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := run(tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("run(%q) error = %v, want error containing %q", tc.args, err, tc.want)
			}
		})
	}
}
