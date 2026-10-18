package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type stagedOutput struct {
	source string
	target string
}

// replaceOutputs owns backups separately from the caller's staging directory so
// failed recovery cannot lose its only old copy when the caller cleans staging.
func replaceOutputs(outDir string, outputs []stagedOutput, rename func(string, string) error) (retErr error) {
	if err := preflightOutputs(outputs); err != nil {
		return err
	}
	for _, output := range outputs {
		if err := os.MkdirAll(filepath.Dir(output.target), 0o750); err != nil {
			return err
		}
	}
	backupDir, err := os.MkdirTemp(outDir, ".go-intl-backup-")
	if err != nil {
		return err
	}
	retainBackup := false
	defer func() {
		if !retainBackup {
			if err := os.RemoveAll(backupDir); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("clean backups %s: %w", backupDir, err))
			}
		}
	}()
	type replacedOutput struct {
		target, backup string
		installed      bool
	}
	var replaced []replacedOutput
	restore := func(cause error) error {
		var recoveryErr error
		for i := len(replaced) - 1; i >= 0; i-- {
			previous := replaced[i]
			if previous.installed {
				if err := os.RemoveAll(previous.target); err != nil {
					recoveryErr = errors.Join(recoveryErr, fmt.Errorf("remove new output %s: %w", previous.target, err))
					continue
				}
			}
			if previous.backup != "" {
				if err := rename(previous.backup, previous.target); err != nil {
					recoveryErr = errors.Join(recoveryErr, fmt.Errorf("restore %s: %w", previous.target, err))
				}
			}
		}
		if recoveryErr != nil {
			retainBackup = true
			return errors.Join(cause, fmt.Errorf("recovery incomplete; backups retained at %s: %w", backupDir, recoveryErr))
		}
		return cause
	}
	for i, output := range outputs {
		previous := replacedOutput{target: output.target}
		if _, err := os.Lstat(output.target); err == nil {
			previous.backup = filepath.Join(backupDir, fmt.Sprint(i))
			if err := rename(output.target, previous.backup); err != nil {
				return restore(fmt.Errorf("retain %s: %w", output.target, err))
			}
		} else if !os.IsNotExist(err) {
			return restore(err)
		}
		replaced = append(replaced, previous)
		if err := rename(output.source, output.target); err != nil {
			return restore(fmt.Errorf("replace %s: %w", output.target, err))
		}
		replaced[len(replaced)-1].installed = true
	}
	return nil
}

func preflightOutputs(outputs []stagedOutput) error {
	for _, output := range outputs {
		source, err := os.Lstat(output.source)
		if err != nil {
			return fmt.Errorf("prepared output %s: %w", output.source, err)
		}
		target, err := os.Lstat(output.target)
		if err == nil && source.Mode().Type() != target.Mode().Type() {
			return fmt.Errorf("output %s: expected %s, got %s", output.target, source.Mode().Type(), target.Mode().Type())
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		for parent := filepath.Dir(output.target); ; parent = filepath.Dir(parent) {
			info, err := os.Stat(parent)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return fmt.Errorf("output parent %s is not a directory", parent)
			}
			break
		}
	}
	return nil
}
