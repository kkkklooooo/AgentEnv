package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const MetaFileName = ".synthesis_meta.json"

// SynthesisMeta records modification timestamps for 4-tier incremental validation
type SynthesisMeta struct {
	Version           int       `json:"version"`
	Preset            string    `json:"preset"`
	Agent             string    `json:"agent"`
	SynthesizedAt     time.Time `json:"synthesized_at"`
	HostDirMtime      int64     `json:"host_dir_mtime"`
	GlobalSharedMtime int64     `json:"global_shared_mtime"`
	PresetSharedMtime int64     `json:"preset_shared_mtime"`
	PresetAgentMtime  int64     `json:"preset_agent_mtime"`
}

// CurrentMtimes aggregates timestamps for all 4 layers
type CurrentMtimes struct {
	HostDir      int64
	GlobalShared int64
	PresetShared int64
	PresetAgent  int64
}

// GetPathMtime returns the modification time (unix seconds) of a path, or 0 if it does not exist
func GetPathMtime(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.ModTime().Unix()
}

// ReadMeta reads .synthesis_meta.json from the runtime viewport directory
func ReadMeta(viewportPath string) (*SynthesisMeta, error) {
	metaPath := filepath.Join(viewportPath, MetaFileName)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta SynthesisMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// WriteMeta writes .synthesis_meta.json into the runtime viewport directory
func WriteMeta(viewportPath string, meta *SynthesisMeta) error {
	metaPath := filepath.Join(viewportPath, MetaFileName)
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, append(data, '\n'), 0644)
}

// IsUpToDate checks whether all 4 layer mtimes match the saved meta
func IsUpToDate(viewportPath string, curr CurrentMtimes) bool {
	meta, err := ReadMeta(viewportPath)
	if err != nil {
		return false
	}

	return meta.HostDirMtime == curr.HostDir &&
		meta.GlobalSharedMtime == curr.GlobalShared &&
		meta.PresetSharedMtime == curr.PresetShared &&
		meta.PresetAgentMtime == curr.PresetAgent
}
