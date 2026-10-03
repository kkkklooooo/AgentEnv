package shell

import (
	"strings"
	"testing"
)

func TestShellDetection(t *testing.T) {
	detected, err := DetectActiveShell()
	if err != nil {
		t.Fatalf("DetectActiveShell failed: %v", err)
	}

	if detected.BinaryPath == "" {
		t.Fatalf("expected detected.BinaryPath to be non-empty")
	}

	if detected.Type == ShellTypeUnknown && detected.ParentName == "" {
		t.Fatalf("expected detected shell to have Type or ParentName")
	}
}

func TestSubshellConfigFormat(t *testing.T) {
	cfg := &SubshellConfig{
		PresetName: "test",
		EnvVars: map[string]string{
			"FOO": "BAR",
			"BAZ": "QUX",
		},
	}

	slice := cfg.FormatEnvSlice()
	if len(slice) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(slice))
	}

	foundFoo := false
	for _, env := range slice {
		if strings.HasPrefix(env, "FOO=BAR") {
			foundFoo = true
		}
	}
	if !foundFoo {
		t.Fatalf("expected FOO=BAR in slice")
	}
}
