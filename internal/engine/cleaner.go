package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kkkklooooo/AgentEnv/internal/config"
	"github.com/kkkklooooo/AgentEnv/internal/platform"
)

// CleanRuntimes removes all synthesized viewports under ~/.agentenv/runtimes/
func CleanRuntimes() (int, error) {
	runtimesDir, err := config.RuntimesDir()
	if err != nil {
		return 0, err
	}

	entries, err := os.ReadDir(runtimesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read runtimes directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		target := filepath.Join(runtimesDir, entry.Name())
		if err := os.RemoveAll(target); err != nil {
			return count, fmt.Errorf("failed to remove %s: %w", target, err)
		}
		count++
	}

	return count, nil
}

// CleanDanglingSymlinks walks a directory and removes any symlinks pointing to non-existent targets
func CleanDanglingSymlinks(root string) (int, error) {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return 0, nil
	}

	removed := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip unreadable items
		}

		if d.Type()&os.ModeSymlink != 0 {
			if platform.IsDanglingSymlink(path) {
				if removeErr := os.Remove(path); removeErr == nil {
					removed++
				}
			}
		}
		return nil
	})

	return removed, err
}
