package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolvePath expands ~ to user's home directory and returns an absolute, clean path
func ResolvePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		if path == "~" {
			path = home
		} else if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
			path = filepath.Join(home, path[2:])
		}
	}

	return filepath.Clean(path), nil
}

// AgentEnvDir returns the root ~/.agentenv directory
func AgentEnvDir() (string, error) {
	return ResolvePath("~/.agentenv")
}

// ConfigFile returns path to ~/.agentenv/config.toml
func ConfigFile() (string, error) {
	base, err := AgentEnvDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "config.toml"), nil
}

// PresetsDir returns path to ~/.agentenv/presets
func PresetsDir() (string, error) {
	base, err := AgentEnvDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "presets"), nil
}

// PresetDir returns path to ~/.agentenv/presets/<preset>
func PresetDir(name string) (string, error) {
	base, err := PresetsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, name), nil
}

// PresetSharedDir returns path to ~/.agentenv/presets/<preset>/shared
func PresetSharedDir(preset string) (string, error) {
	pDir, err := PresetDir(preset)
	if err != nil {
		return "", err
	}
	return filepath.Join(pDir, "shared"), nil
}

// PresetAgentDir returns path to ~/.agentenv/presets/<preset>/<agent>
func PresetAgentDir(preset, agent string) (string, error) {
	pDir, err := PresetDir(preset)
	if err != nil {
		return "", err
	}
	return filepath.Join(pDir, agent), nil
}

// GlobalSharedDir returns path to ~/.agentenv/shared
func GlobalSharedDir() (string, error) {
	base, err := AgentEnvDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "shared"), nil
}

// RuntimesDir returns path to ~/.agentenv/runtimes
func RuntimesDir() (string, error) {
	base, err := AgentEnvDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "runtimes"), nil
}

// RuntimePresetDir returns path to ~/.agentenv/runtimes/<preset>
func RuntimePresetDir(preset string) (string, error) {
	base, err := RuntimesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, preset), nil
}

// RuntimeAgentDir returns path to ~/.agentenv/runtimes/<preset>/<agent>
func RuntimeAgentDir(preset, agent string) (string, error) {
	base, err := RuntimePresetDir(preset)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, agent), nil
}

// EnsureBaseDirectories creates ~/.agentenv, presets, shared, runtimes
func EnsureBaseDirectories() error {
	dirs := []func() (string, error){
		AgentEnvDir,
		PresetsDir,
		GlobalSharedDir,
		RuntimesDir,
	}

	for _, fn := range dirs {
		d, err := fn()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	return nil
}
