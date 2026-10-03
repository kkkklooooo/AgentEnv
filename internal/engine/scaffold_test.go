package engine

import (
	"os"
	"path/filepath"
	"testing"

	"agentenv/internal/config"
)

func TestSyncPresetAgentFolders(t *testing.T) {
	tempRoot := t.TempDir()

	t.Setenv("USERPROFILE", tempRoot)
	t.Setenv("HOME", tempRoot)

	cfg := &config.Config{
		Agents: map[string]config.AgentConfig{
			"claude":   {EnvVar: "CLAUDE_CONFIG_DIR", HostDir: "~/.claude"},
			"codex":    {EnvVar: "CODEX_CONFIG_DIR", HostDir: "~/.codex"},
			"gemini":   {EnvVar: "GEMINI_CONFIG_DIR", HostDir: "~/.gemini"},
			"opencode": {EnvVar: "OPENCODE_HOME", HostDir: "~/.opencode"},
		},
	}

	presetsDir, _ := config.PresetsDir()
	_ = os.MkdirAll(presetsDir, 0755)

	// Create a preset that only has claude
	presetA := filepath.Join(presetsDir, "preset-a")
	_ = os.MkdirAll(filepath.Join(presetA, "claude"), 0755)

	// Run SyncPresetAgentFolders
	report, err := SyncPresetAgentFolders(cfg)
	if err != nil {
		t.Fatalf("SyncPresetAgentFolders failed: %v", err)
	}

	if report.TotalCreated == 0 {
		t.Fatalf("expected missing folders to be created, got 0")
	}

	// Verify that shared, codex, gemini, and opencode were created
	for _, expected := range []string{"shared", "codex", "gemini", "opencode"} {
		path := filepath.Join(presetA, expected)
		if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
			t.Fatalf("expected folder %s to exist in preset-a", expected)
		}
	}

	// Re-run sync: should report 0 creations
	report2, err := SyncPresetAgentFolders(cfg)
	if err != nil {
		t.Fatalf("second SyncPresetAgentFolders failed: %v", err)
	}
	if report2.TotalCreated != 0 {
		t.Fatalf("expected 0 new creations on re-sync, got %d", report2.TotalCreated)
	}
}
