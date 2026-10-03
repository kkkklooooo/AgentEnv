package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSymlinkOperations(t *testing.T) {
	tempDir := t.TempDir()

	targetFile := filepath.Join(tempDir, "original.txt")
	if err := os.WriteFile(targetFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	linkFile := filepath.Join(tempDir, "link.txt")
	if err := CreateSymlink(targetFile, linkFile); err != nil {
		t.Skipf("skipping symlink test if permissions not held: %v", err)
	}

	isLink, err := IsSymlink(linkFile)
	if err != nil || !isLink {
		t.Fatalf("expected IsSymlink to be true, got %v, err=%v", isLink, err)
	}

	// Test read content through symlink
	content, err := os.ReadFile(linkFile)
	if err != nil || string(content) != "hello world" {
		t.Fatalf("unexpected content through symlink: %s, err=%v", string(content), err)
	}

	// Test dangling symlink
	if err := os.Remove(targetFile); err != nil {
		t.Fatalf("failed to remove target: %v", err)
	}

	if !IsDanglingSymlink(linkFile) {
		t.Fatalf("expected IsDanglingSymlink to be true after removing target")
	}

	// Test atomic update
	newTarget := filepath.Join(tempDir, "new_original.txt")
	if err := os.WriteFile(newTarget, []byte("new content"), 0644); err != nil {
		t.Fatalf("failed to create new target: %v", err)
	}

	if err := AtomicSymlink(newTarget, linkFile); err != nil {
		t.Fatalf("failed AtomicSymlink: %v", err)
	}

	content, err = os.ReadFile(linkFile)
	if err != nil || string(content) != "new content" {
		t.Fatalf("expected new content, got %s, err=%v", string(content), err)
	}
}
