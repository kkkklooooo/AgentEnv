package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kkkklooooo/AgentEnv/internal/config"
	"github.com/kkkklooooo/AgentEnv/internal/platform"
)

func TestFourTierCascadingSynthesis(t *testing.T) {
	tempRoot := t.TempDir()

	// Configure mock paths
	t.Setenv("USERPROFILE", tempRoot)
	t.Setenv("HOME", tempRoot)

	cfg := &config.Config{
		Agents: map[string]config.AgentConfig{
			"claude": {
				EnvVar:  "CLAUDE_CONFIG_DIR",
				HostDir: filepath.Join(tempRoot, "host_claude"),
			},
		},
	}

	hostDir, _ := cfg.Agents["claude"].ResolvedHostDir()
	_ = os.MkdirAll(hostDir, 0755)

	// Layer 1: Host items
	_ = os.WriteFile(filepath.Join(hostDir, "history.jsonl"), []byte("session-1"), 0644)
	_ = os.WriteFile(filepath.Join(hostDir, "settings.json"), []byte(`{"model":"host"}`), 0644)
	_ = os.MkdirAll(filepath.Join(hostDir, "skills"), 0755)
	_ = os.WriteFile(filepath.Join(hostDir, "skills", "host_skill.sh"), []byte("echo host"), 0644)

	// Layer 2: Global Shared
	globalShared, _ := config.GlobalSharedDir()
	_ = os.MkdirAll(globalShared, 0755)
	_ = os.WriteFile(filepath.Join(globalShared, "global.txt"), []byte("global-rule"), 0644)

	// Layer 3: Preset Shared
	presetName := "test-backend"
	presetShared, _ := config.PresetSharedDir(presetName)
	_ = os.MkdirAll(presetShared, 0755)
	_ = os.WriteFile(filepath.Join(presetShared, ".mcp.json"), []byte(`{"mcp":"shared"}`), 0644)

	// Layer 4: Preset Agent Specific
	presetAgent, _ := config.PresetAgentDir(presetName, "claude")
	_ = os.MkdirAll(presetAgent, 0755)
	_ = os.WriteFile(filepath.Join(presetAgent, "settings.json"), []byte(`{"model":"preset-override"}`), 0644)
	_ = os.MkdirAll(filepath.Join(presetAgent, "skills"), 0755)
	_ = os.WriteFile(filepath.Join(presetAgent, "skills", "preset_skill.sh"), []byte("echo preset"), 0644)

	// Execute Synthesizer
	synth := NewSynthesizer(cfg)
	runtimeDir, err := synth.SynthesizeAgent(presetName, "claude")
	if err != nil {
		t.Fatalf("SynthesizeAgent failed: %v", err)
	}

	// 1. Verify Layer 1 penetration: history.jsonl exists and matches host
	historyContent, err := os.ReadFile(filepath.Join(runtimeDir, "history.jsonl"))
	if err != nil || string(historyContent) != "session-1" {
		t.Fatalf("expected history.jsonl to penetrate from host: %s, err=%v", string(historyContent), err)
	}

	// 2. Verify Layer 2 penetration: global.txt exists
	globalContent, err := os.ReadFile(filepath.Join(runtimeDir, "global.txt"))
	if err != nil || string(globalContent) != "global-rule" {
		t.Fatalf("expected global.txt to penetrate from global shared: %s, err=%v", string(globalContent), err)
	}

	// 3. Verify Layer 3 penetration: .mcp.json exists from preset shared
	mcpContent, err := os.ReadFile(filepath.Join(runtimeDir, ".mcp.json"))
	if err != nil || string(mcpContent) != `{"mcp":"shared"}` {
		t.Fatalf("expected .mcp.json from preset shared: %s, err=%v", string(mcpContent), err)
	}

	// 4. Verify Layer 4 ultimate override: settings.json overridden by preset
	settingsContent, err := os.ReadFile(filepath.Join(runtimeDir, "settings.json"))
	if err != nil || string(settingsContent) != `{"model":"preset-override"}` {
		t.Fatalf("expected settings.json to be overridden by preset: %s, err=%v", string(settingsContent), err)
	}

	// 5. Verify Smart Directory Merging on skills/: both host_skill.sh and preset_skill.sh must exist!
	hostSkill, err := os.ReadFile(filepath.Join(runtimeDir, "skills", "host_skill.sh"))
	if err != nil || string(hostSkill) != "echo host" {
		t.Fatalf("expected host_skill.sh to penetrate inside merged skills dir: %s, err=%v", string(hostSkill), err)
	}

	presetSkill, err := os.ReadFile(filepath.Join(runtimeDir, "skills", "preset_skill.sh"))
	if err != nil || string(presetSkill) != "echo preset" {
		t.Fatalf("expected preset_skill.sh to exist inside merged skills dir: %s, err=%v", string(presetSkill), err)
	}

	// 6. Test Incremental Cache Hit
	metaBefore, err := ReadMeta(runtimeDir)
	if err != nil {
		t.Fatalf("failed to read meta: %v", err)
	}

	// Re-run synthesis immediately
	runtimeDir2, err := synth.SynthesizeAgent(presetName, "claude")
	if err != nil || runtimeDir2 != runtimeDir {
		t.Fatalf("expected identical runtime dir: %s, err=%v", runtimeDir2, err)
	}

	metaAfter, err := ReadMeta(runtimeDir)
	if err != nil || !metaAfter.SynthesizedAt.Equal(metaBefore.SynthesizedAt) {
		t.Fatalf("expected cache hit with identical SynthesizedAt timestamp")
	}

	// 7. Test Diff
	diffOutput, err := PresetDiff(cfg, presetName)
	if err != nil {
		t.Fatalf("PresetDiff failed: %v", err)
	}
	if diffOutput == "" || diffOutput == "No configuration differences detected between preset and host baseline.\n" {
		t.Fatalf("expected diff output, got: %s", diffOutput)
	}

	// 8. Test CleanRuntimes
	count, err := CleanRuntimes()
	if err != nil || count == 0 {
		t.Fatalf("expected CleanRuntimes to clean runtimes, count=%d, err=%v", count, err)
	}

	// Ensure runtime dir was removed
	if _, err := os.Stat(runtimeDir); !os.IsNotExist(err) {
		t.Fatalf("expected runtimeDir to be deleted after CleanRuntimes")
	}
}

func TestDanglingSymlinkCleanup(t *testing.T) {
	tempDir := t.TempDir()

	target := filepath.Join(tempDir, "real.txt")
	_ = os.WriteFile(target, []byte("data"), 0644)

	link := filepath.Join(tempDir, "link.txt")
	_ = platform.CreateSymlink(target, link)

	// Remove target to make it dangling
	_ = os.Remove(target)

	cleaned, err := CleanDanglingSymlinks(tempDir)
	if err != nil {
		t.Fatalf("CleanDanglingSymlinks failed: %v", err)
	}
	if cleaned != 1 {
		t.Fatalf("expected 1 dangling symlink removed, got %d", cleaned)
	}
}
