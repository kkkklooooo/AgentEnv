//go:build windows

package platform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	// ERROR_PRIVILEGE_NOT_HELD corresponds to Windows error code 1314 (0x522)
	ERROR_PRIVILEGE_NOT_HELD syscall.Errno = 1314
)

// CreateSymlink creates a symbolic link on Windows with error 0x522 interception
func CreateSymlink(target, linkPath string) error {
	err := os.Symlink(target, linkPath)
	if err != nil {
		var sysErr syscall.Errno
		if errors.As(err, &sysErr) && sysErr == ERROR_PRIVILEGE_NOT_HELD {
			return fmt.Errorf("%w: privilege not held (error code 0x522)", ErrDeveloperModeRequired)
		}
		return err
	}
	return nil
}

// CheckDeveloperMode checks whether Windows Developer Mode is enabled via Registry
func CheckDeveloperMode() (bool, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock`, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AllowDevelopmentWithoutDevLicense")
	if err != nil {
		return false, err
	}

	return val == 1, nil
}

// CheckSymlinkCapability tests whether symlink creation is permitted without elevation
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
	if err := CreateSymlink(dummyTarget, dummyLink); err != nil {
		return err
	}

	return nil
}
