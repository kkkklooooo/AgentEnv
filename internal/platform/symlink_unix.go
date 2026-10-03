//go:build !windows

package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateSymlink creates a symbolic link on POSIX platforms
func CreateSymlink(target, linkPath string) error {
	return os.Symlink(target, linkPath)
}

// CheckDeveloperMode returns true on Unix/Linux as Developer Mode is a Windows-only restriction
func CheckDeveloperMode() (bool, error) {
	return true, nil
}

// CheckSymlinkCapability tests whether symlink creation is permitted
func CheckSymlinkCapability() error {
	tempDir, err := os.MkdirTemp("", "aenv-symlink-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for capability test: %w", err)
	}
	defer os.RemoveAll(tempDir)

	dummyTarget := filepath.Join(tempDir, "target.txt")
	if err := os.WriteFile(dummyTarget, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("failed to create dummy target: %w", err)
	}

	dummyLink := filepath.Join(tempDir, "link.txt")
	return os.Symlink(dummyTarget, dummyLink)
}
