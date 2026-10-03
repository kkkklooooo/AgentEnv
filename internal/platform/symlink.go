package platform

import (
	"errors"
	"fmt"
	"os"
)

// ErrDeveloperModeRequired is returned when Windows symlinks fail due to missing Developer Mode
var ErrDeveloperModeRequired = errors.New("windows developer mode is required to create symlinks")

// IsSymlink returns true if the specified path is a symbolic link
func IsSymlink(path string) (bool, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return fi.Mode()&os.ModeSymlink != 0, nil
}

// IsDanglingSymlink returns true if path is a symlink but its target does not exist
func IsDanglingSymlink(path string) bool {
	isLink, err := IsSymlink(path)
	if err != nil || !isLink {
		return false
	}
	// os.Stat resolves the symlink target; if target doesn't exist, stat returns IsNotExist
	_, err = os.Stat(path)
	return os.IsNotExist(err)
}

// RemoveIfSymlinkOrFile safely removes a path if it exists (whether link, file, or empty dir)
func RemoveIfSymlinkOrFile(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return os.Remove(path)
	}

	// If it's a real directory, remove all its contents
	return os.RemoveAll(path)
}

// AtomicSymlink creates or updates a symlink atomically, replacing any existing item
func AtomicSymlink(target, linkPath string) error {
	// If linkPath already exists, remove it first
	if err := RemoveIfSymlinkOrFile(linkPath); err != nil {
		return fmt.Errorf("failed to remove existing target before linking %s: %w", linkPath, err)
	}

	return CreateSymlink(target, linkPath)
}
