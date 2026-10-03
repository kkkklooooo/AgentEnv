package shell

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// ShellType represents the identified shell family
type ShellType string

const (
	ShellTypePwsh       ShellType = "PowerShell Core (pwsh)"
	ShellTypePowerShell ShellType = "Windows PowerShell 5.1"
	ShellTypeCmd        ShellType = "Command Prompt (cmd.exe)"
	ShellTypeBash       ShellType = "Bash"
	ShellTypeZsh        ShellType = "Zsh"
	ShellTypeFish       ShellType = "Fish"
	ShellTypeNushell    ShellType = "Nushell"
	ShellTypeUnknown    ShellType = "Unknown"
)

// DetectedShell contains metadata about the active terminal shell
type DetectedShell struct {
	Type       ShellType
	BinaryPath string
	ParentPID  int
	ParentName string
}

// SubshellConfig holds environment variables and parameters to pass into the subshell
type SubshellConfig struct {
	PresetName string
	EnvVars    map[string]string
}

// FormatEnvSlice converts map to KEY=VALUE slice
func (c *SubshellConfig) FormatEnvSlice() []string {
	res := make([]string, 0, len(c.EnvVars))
	for k, v := range c.EnvVars {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return res
}

// DetectShellWithOverride detects the shell, respecting an optional override from config
func DetectShellWithOverride(override string) (*DetectedShell, error) {
	if override != "" {
		path, err := exec.LookPath(override)
		if err != nil {
			path = override
		}
		return &DetectedShell{
			Type:       ShellType(fmt.Sprintf("Custom (%s)", override)),
			BinaryPath: path,
			ParentPID:  os.Getppid(),
			ParentName: override,
		}, nil
	}

	return DetectActiveShell()
}

// ExecCommand executes a command directly with injected environment variables, transparently forwarding stdio and exit code
func ExecCommand(cmdName string, cmdArgs []string, envVars map[string]string) (int, error) {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	env := os.Environ()
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}

	return 0, nil
}
