package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// TerminalConfig declares terminal-specific preferences
type TerminalConfig struct {
	Shell string `toml:"shell,omitempty"` // empty for auto-detect
}

// Config represents the ~/.agentenv/config.toml file
type Config struct {
	Terminal TerminalConfig           `toml:"terminal"`
	Agents   map[string]AgentConfig   `toml:"agents"`
}

// NewDefaultConfig returns a new Config with default agents populated
func NewDefaultConfig() *Config {
	return &Config{
		Terminal: TerminalConfig{},
		Agents:   DefaultAgents(),
	}
}

// LoadConfig loads the global config, or bootstraps a default one if absent
func LoadConfig() (*Config, error) {
	if err := EnsureBaseDirectories(); err != nil {
		return nil, fmt.Errorf("failed to ensure base directories: %w", err)
	}

	confPath, err := ConfigFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(confPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Bootstrap default config
			cfg := NewDefaultConfig()
			if err := SaveConfig(cfg); err != nil {
				return nil, fmt.Errorf("failed to bootstrap default config: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", confPath, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML from %s: %w", confPath, err)
	}

	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentConfig)
	}

	// Validate agents
	for name, ag := range cfg.Agents {
		if err := ag.Validate(name); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

// SaveConfig saves the given config to ~/.agentenv/config.toml
func SaveConfig(cfg *Config) error {
	confPath, err := ConfigFile()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to TOML: %w", err)
	}

	// Add comments/header
	header := "# AgentEnv (aenv) Configuration File\n# Generated automatically. Edit with caution.\n\n"
	content := append([]byte(header), data...)

	return os.WriteFile(confPath, content, 0644)
}

// GetAgent retrieves agent config by name, returning an error if not registered
func (c *Config) GetAgent(name string) (*AgentConfig, error) {
	ag, ok := c.Agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q is not registered in ~/.agentenv/config.toml", name)
	}
	return &ag, nil
}
