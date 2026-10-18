package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func preparedPublication(t *testing.T) (string, []stagedOutput) {
	t.Helper()
	root := t.TempDir()
	outputs := []stagedOutput{
		{source: filepath.Join(root, "stage", "first"), target: filepath.Join(root, "first")},
		{source: filepath.Join(root, "stage", "new"), target: filepath.Join(root, "new")},
		{source: filepath.Join(root, "stage", repositorySkipListFile), target: filepath.Join(root, repositorySkipListFile)},
	}
	mustWriteFile(t, filepath.Join(outputs[0].source, "fixture.json"), "new fixture")
	mustWriteFile(t, filepath.Join(outputs[0].target, "old.json"), "old fixture")
	mustWriteFile(t, outputs[1].source, "new file")
	mustWriteFile(t, outputs[2].source, "new skips")
	mustWriteFile(t, outputs[2].target, "old skips")
	return root, outputs
}

func TestPublicationRestoresCollectionOnFailure(t *testing.T) {
	t.Parallel()
	for _, duringBackup := range []bool{false, true} {
		t.Run(map[bool]string{false: "install skip", true: "retain skip"}[duringBackup], func(t *testing.T) {
			t.Parallel()
			root, outputs := preparedPublication(t)
			failure := errors.New("replacement failed")
			err := replaceOutputs(root, outputs, func(from, to string) error {
				if (!duringBackup && from == outputs[2].source) || (duringBackup && from == outputs[2].target) {
					return failure
				}
				return os.Rename(from, to)
			})
			if !errors.Is(err, failure) {
				t.Errorf("error = %v, want replacement failure", err)
			}
			for path, want := range map[string]string{filepath.Join(outputs[0].target, "old.json"): "old fixture", outputs[2].target: "old skips"} {
				got, readErr := os.ReadFile(path)
				if readErr != nil || string(got) != want {
					t.Errorf("restore %s: %q, %v", path, got, readErr)
				}
			}
			if _, err := os.Stat(outputs[1].target); !os.IsNotExist(err) {
				t.Errorf("new target remains after failure: %v", err)
			}
			if _, err := os.Stat(filepath.Join(outputs[0].target, "fixture.json")); !os.IsNotExist(err) {
				t.Errorf("new fixture remains after failure: %v", err)
			}
			assertNoPublicationBackups(t, root)
		})
	}
}

func TestPublicationSuccessReplacesCollectionAndCleansBackups(t *testing.T) {
	t.Parallel()
	root, outputs := preparedPublication(t)
	mustWriteFile(t, filepath.Join(root, "manual.json"), "manual")
	if err := replaceOutputs(root, outputs, os.Rename); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		filepath.Join(outputs[0].target, "fixture.json"): "new fixture",
		outputs[1].target:                  "new file",
		outputs[2].target:                  "new skips",
		filepath.Join(root, "manual.json"): "manual",
	} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Errorf("published %s = %q, %v", path, got, err)
		}
	}
	if _, err := os.Stat(filepath.Join(outputs[0].target, "old.json")); !os.IsNotExist(err) {
		t.Errorf("stale owned file remains: %v", err)
	}
	assertNoPublicationBackups(t, root)
}

func assertNoPublicationBackups(t *testing.T, root string) {
	t.Helper()
	backups, err := filepath.Glob(filepath.Join(root, ".go-intl-backup-*"))
	if err != nil || len(backups) != 0 {
		t.Errorf("publication backups remain: %v, %v", backups, err)
	}
}

func TestPublicationPreservesBackupWhenRecoveryFails(t *testing.T) {
	t.Parallel()
	root, outputs := preparedPublication(t)
	replacementFailure, recoveryFailure := errors.New("replacement failed"), errors.New("recovery failed")
	var backup string
	err := replaceOutputs(root, outputs, func(from, to string) error {
		if from == outputs[2].source {
			return replacementFailure
		}
		if to == outputs[0].target && from != outputs[0].source {
			backup = from
			return recoveryFailure
		}
		return os.Rename(from, to)
	})
	if !errors.Is(err, replacementFailure) || !errors.Is(err, recoveryFailure) {
		t.Errorf("error = %v, want both failures", err)
	}
	if backup == "" {
		t.Fatal("no retained backup was offered for recovery")
	}
	if !strings.Contains(err.Error(), filepath.Dir(backup)) {
		t.Errorf("error %v omits backup directory %s", err, filepath.Dir(backup))
	}
	if err := os.RemoveAll(filepath.Join(root, "stage")); err != nil {
		t.Fatal(err)
	}
	got, readErr := os.ReadFile(filepath.Join(backup, "old.json"))
	if readErr != nil || string(got) != "old fixture" {
		t.Fatalf("only old copy lost: %q, %v", got, readErr)
	}
}

func TestPublicationRequiresAllPreparedOutputs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	first := stagedOutput{source: filepath.Join(root, "stage", "first"), target: filepath.Join(root, "out", "first")}
	last := stagedOutput{source: filepath.Join(root, "stage", "last"), target: filepath.Join(root, "out", "last")}
	mustWriteFile(t, first.source, "new")
	mustWriteFile(t, first.target, "old")
	// An encoding failure must not leave a partial staged file eligible to publish.
	if err := writeJSON(last.source, make(chan int)); err == nil {
		t.Fatal("unsupported JSON value was encoded")
	}
	if err := replaceOutputs(root, []stagedOutput{first, last}, os.Rename); err == nil || !strings.Contains(err.Error(), last.source) {
		t.Fatalf("publication error = %v, want missing prepared output", err)
	}
	got, err := os.ReadFile(first.target)
	if err != nil || string(got) != "old" {
		t.Fatalf("first target changed: %q, %v", got, err)
	}
}
