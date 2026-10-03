package config

import (
	"path/filepath"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()

	// Override user home dir for test isolation
	t.Setenv("USERPROFILE", tempDir)
	t.Setenv("HOME", tempDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Agents) == 0 {
		t.Fatalf("expected default agents, got 0")
	}

	claude, ok := cfg.Agents["claude"]
	if !ok || claude.EnvVar != "CLAUDE_CONFIG_DIR" {
		t.Fatalf("expected claude agent with CLAUDE_CONFIG_DIR, got %+v", claude)
	}

	// Test adding a custom agent
	cfg.Agents["custom"] = AgentConfig{
		EnvVar:  "CUSTOM_CONFIG_DIR",
		HostDir: "~/.custom",
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Reload
	reloaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("reloading config failed: %v", err)
	}

	custom, err := reloaded.GetAgent("custom")
	if err != nil || custom.EnvVar != "CUSTOM_CONFIG_DIR" {
		t.Fatalf("expected custom agent to be saved and reloaded: %+v, err=%v", custom, err)
	}

	resolved, err := custom.ResolvedHostDir()
	if err != nil || !filepath.IsAbs(resolved) {
		t.Fatalf("expected resolved host dir to be absolute: %s", resolved)
	}
}

func TestValidatePresetName(t *testing.T) {
	valid := []string{"work", "work-backend", "team_1", "k8s-prod-2"}
	for _, v := range valid {
		if err := ValidatePresetName(v); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", v, err)
		}
	}

	invalid := []string{"", "work/backend", "work\\backend", "shared", "work backend", "test@123"}
	for _, inv := range invalid {
		if err := ValidatePresetName(inv); err == nil {
			t.Errorf("expected %q to be invalid, got nil error", inv)
		}
	}
}
