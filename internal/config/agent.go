package config

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	presetNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
)

// AgentConfig declares environment variable and host base path for an agent
type AgentConfig struct {
	EnvVar  string `toml:"env_var"`
	HostDir string `toml:"host_dir"`
}

// Validate checks if the agent config has mandatory fields
func (a AgentConfig) Validate(name string) error {
	if a.EnvVar == "" {
		return fmt.Errorf("agent %q: env_var cannot be empty", name)
	}
	if a.HostDir == "" {
		return fmt.Errorf("agent %q: host_dir cannot be empty", name)
	}
	return nil
}

// ResolvedHostDir returns the cleaned, expanded host directory path
func (a AgentConfig) ResolvedHostDir() (string, error) {
	return ResolvePath(a.HostDir)
}

// DefaultAgents returns the initial built-in agent definitions
func DefaultAgents() map[string]AgentConfig {
	return map[string]AgentConfig{
		"claude": {
			EnvVar:  "CLAUDE_CONFIG_DIR",
			HostDir: "~/.claude",
		},
		"codex": {
			EnvVar:  "CODEX_CONFIG_DIR",
			HostDir: "~/.codex",
		},
		"gemini": {
			EnvVar:  "GEMINI_CONFIG_DIR",
			HostDir: "~/.gemini",
		},
	}
}

// ValidatePresetName checks if a preset name matches ^[a-zA-Z0-9_\-]+$
func ValidatePresetName(name string) error {
	if name == "" {
		return errors.New("preset name cannot be empty")
	}
	if !presetNameRegex.MatchString(name) {
		return fmt.Errorf("invalid preset name %q: must only contain letters, numbers, underscores and hyphens (^[a-zA-Z0-9_\\-]+$)", name)
	}
	if name == "shared" {
		return errors.New("preset name 'shared' is reserved and cannot be used")
	}
	return nil
}
