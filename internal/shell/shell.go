package shell

import (
	"fmt"
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
