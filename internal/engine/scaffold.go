package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kkkklooooo/AgentEnv/internal/config"
)

// PresetSyncResult tracks folders created for a specific preset
type PresetSyncResult struct {
	PresetName  string
	CreatedDirs []string
}

// SyncReport aggregates all scaffolding actions across presets
type SyncReport struct {
	Results      []PresetSyncResult
	TotalCreated int
}

// SyncPresetAgentFolders scans all existing presets against the config.toml agent registry.
// If any registered agent is missing its scaffold directory in a preset (or if shared/ is missing),
// it automatically creates the missing directories to ensure framework consistency.
func SyncPresetAgentFolders(cfg *config.Config) (*SyncReport, error) {
	presetsDir, err := config.PresetsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(presetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &SyncReport{}, nil
		}
		return nil, fmt.Errorf("failed to read presets directory: %w", err)
	}

	report := &SyncReport{
		Results: make([]PresetSyncResult, 0),
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		presetName := entry.Name()
		if err := config.ValidatePresetName(presetName); err != nil {
			// Skip invalid directories (e.g. non-preset folders)
			continue
		}

		pDir := filepath.Join(presetsDir, presetName)
		var created []string

		// 1. Ensure presets/<preset>/shared/ exists
		sharedDir := filepath.Join(pDir, "shared")
		if _, err := os.Stat(sharedDir); os.IsNotExist(err) {
			if err := os.MkdirAll(sharedDir, 0755); err == nil {
				created = append(created, "shared")
			}
		}

		// 2. Ensure presets/<preset>/<agent>/ exists for every registered agent
		for agName := range cfg.Agents {
			agDir := filepath.Join(pDir, agName)
			if _, err := os.Stat(agDir); os.IsNotExist(err) {
				if err := os.MkdirAll(agDir, 0755); err == nil {
					created = append(created, agName)
				}
			}
		}

		if len(created) > 0 {
			report.Results = append(report.Results, PresetSyncResult{
				PresetName:  presetName,
				CreatedDirs: created,
			})
			report.TotalCreated += len(created)
		}
	}

	return report, nil
}
