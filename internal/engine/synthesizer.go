package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kkkklooooo/AgentEnv/internal/config"
	"github.com/kkkklooooo/AgentEnv/internal/platform"
)

// Synthesizer coordinates 4-tier cascading viewport synthesis
type Synthesizer struct {
	Config *config.Config
}

// NewSynthesizer creates a new Synthesizer instance
func NewSynthesizer(cfg *config.Config) *Synthesizer {
	return &Synthesizer{Config: cfg}
}

// SynthesizeAgent builds or reuses the shadow viewport for a specific agent in a preset
func (s *Synthesizer) SynthesizeAgent(presetName, agentName string) (string, error) {
	if err := config.ValidatePresetName(presetName); err != nil {
		return "", err
	}

	agCfg, err := s.Config.GetAgent(agentName)
	if err != nil {
		return "", err
	}

	// 1. Resolve all directory paths
	presetDir, err := config.PresetDir(presetName)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(presetDir); os.IsNotExist(err) {
		return "", fmt.Errorf("preset %q does not exist in %s", presetName, presetDir)
	}

	hostDir, err := agCfg.ResolvedHostDir()
	if err != nil {
		return "", err
	}

	// Lazy bootstrap of host base if it does not exist
	if _, err := os.Stat(hostDir); os.IsNotExist(err) {
		if err := os.MkdirAll(hostDir, 0755); err != nil {
			return "", fmt.Errorf("failed to bootstrap host directory %s: %w", hostDir, err)
		}
	}

	globalSharedDir, err := config.GlobalSharedDir()
	if err != nil {
		return "", err
	}

	presetSharedDir, err := config.PresetSharedDir(presetName)
	if err != nil {
		return "", err
	}

	presetAgentDir, err := config.PresetAgentDir(presetName, agentName)
	if err != nil {
		return "", err
	}

	runtimeDir, err := config.RuntimeAgentDir(presetName, agentName)
	if err != nil {
		return "", err
	}

	// 2. Incremental Cache Check: compare mtimes of all 4 layers
	currMtimes := CurrentMtimes{
		HostDir:      GetPathMtime(hostDir),
		GlobalShared: GetPathMtime(globalSharedDir),
		PresetShared: GetPathMtime(presetSharedDir),
		PresetAgent:  GetPathMtime(presetAgentDir),
	}

	if IsUpToDate(runtimeDir, currMtimes) {
		// Cache hit! Return immediately in <1ms without disk mutation
		return runtimeDir, nil
	}

	// 3. Cache Miss: Perform 4-Tier Cascading Assembly
	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create runtime viewport directory %s: %w", runtimeDir, err)
	}

	// Pre-flight: clean any existing dangling symlinks
	_, _ = CleanDanglingSymlinks(runtimeDir)

	// Layer 1: Host Base Projection
	hostEntries, err := os.ReadDir(hostDir)
	if err != nil {
		return "", fmt.Errorf("failed to read host directory %s: %w", hostDir, err)
	}
	for _, entry := range hostEntries {
		name := entry.Name()
		if name == MetaFileName {
			continue
		}
		target := filepath.Join(hostDir, name)
		dest := filepath.Join(runtimeDir, name)
		if err := platform.AtomicSymlink(target, dest); err != nil {
			return "", fmt.Errorf("failed to link host base item %s: %w", name, err)
		}
	}

	// Layer 2: Global Shared Merge
	if info, err := os.Stat(globalSharedDir); err == nil && info.IsDir() {
		if err := mergeTree(globalSharedDir, runtimeDir); err != nil {
			return "", fmt.Errorf("failed to merge global shared layer: %w", err)
		}
	}

	// Layer 3: Preset Shared Merge
	if info, err := os.Stat(presetSharedDir); err == nil && info.IsDir() {
		if err := mergeTree(presetSharedDir, runtimeDir); err != nil {
			return "", fmt.Errorf("failed to merge preset shared layer: %w", err)
		}
	}

	// Layer 4: Preset Agent Overrides (Final Override)
	if info, err := os.Stat(presetAgentDir); err == nil && info.IsDir() {
		if err := mergeTree(presetAgentDir, runtimeDir); err != nil {
			return "", fmt.Errorf("failed to merge preset agent layer: %w", err)
		}
	}

	// 4. Update .synthesis_meta.json
	meta := &SynthesisMeta{
		Version:           1,
		Preset:            presetName,
		Agent:             agentName,
		SynthesizedAt:     time.Now().UTC(),
		HostDirMtime:      currMtimes.HostDir,
		GlobalSharedMtime: currMtimes.GlobalShared,
		PresetSharedMtime: currMtimes.PresetShared,
		PresetAgentMtime:  currMtimes.PresetAgent,
	}
	if err := WriteMeta(runtimeDir, meta); err != nil {
		return "", fmt.Errorf("failed to write synthesis metadata: %w", err)
	}

	return runtimeDir, nil
}

// mergeTree overlays sourceDir onto destDir with Smart Directory Merging
func mergeTree(sourceDir, destDir string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == MetaFileName {
			continue
		}

		srcItem := filepath.Join(sourceDir, name)
		destItem := filepath.Join(destDir, name)

		if entry.IsDir() {
			// Directory handling
			isLink, _ := platform.IsSymlink(destItem)
			destFi, destErr := os.Stat(destItem)

			if destErr == nil && destFi.IsDir() {
				// destItem already exists as a directory
				if isLink {
					// Expand directory symlink into a real directory with individual leaf symlinks
					target, err := filepath.EvalSymlinks(destItem)
					if err != nil {
						return fmt.Errorf("failed to resolve symlink target for %s: %w", destItem, err)
					}
					// Remove the symlink
					if err := os.Remove(destItem); err != nil {
						return fmt.Errorf("failed to remove directory symlink %s: %w", destItem, err)
					}
					// Create real directory
					if err := os.MkdirAll(destItem, 0755); err != nil {
						return fmt.Errorf("failed to create real directory %s: %w", destItem, err)
					}
					// Re-link previous items individually
					subEntries, _ := os.ReadDir(target)
					for _, sub := range subEntries {
						subTarget := filepath.Join(target, sub.Name())
						subDest := filepath.Join(destItem, sub.Name())
						_ = platform.AtomicSymlink(subTarget, subDest)
					}
				}
				// Recurse into directory
				if err := mergeTree(srcItem, destItem); err != nil {
					return err
				}
			} else {
				// destItem does not exist yet: link directly as a directory symlink
				if err := platform.AtomicSymlink(srcItem, destItem); err != nil {
					return err
				}
			}
		} else {
			// Leaf file or symlink: atomic override
			if err := platform.AtomicSymlink(srcItem, destItem); err != nil {
				return err
			}
		}
	}

	return nil
}

// DetectPresetAgents scans a preset directory and returns all configured agents within it
func DetectPresetAgents(cfg *config.Config, presetName string) ([]string, error) {
	presetDir, err := config.PresetDir(presetName)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(presetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("preset %q does not exist", presetName)
		}
		return nil, err
	}

	var agents []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "shared" {
			continue // shared is not an agent
		}
		if _, ok := cfg.Agents[name]; ok {
			agents = append(agents, name)
		}
	}

	// If no agent-specific directories exist but shared exists or preset exists,
	// default to all registered agents that can inherit shared configurations
	if len(agents) == 0 {
		for name := range cfg.Agents {
			agents = append(agents, name)
		}
	}

	return agents, nil
}
