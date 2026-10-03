//go:build !windows

package shell

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectActiveShell detects the user's shell on Unix/Linux systems
func DetectActiveShell() (*DetectedShell, error) {
	ppid := os.Getppid()

	// Method 1: Check $SHELL environment variable
	shellEnv := os.Getenv("SHELL")
	if shellEnv != "" {
		base := filepath.Base(shellEnv)
		return &DetectedShell{
			Type:       classifyUnixShell(base),
			BinaryPath: shellEnv,
			ParentPID:  ppid,
			ParentName: base,
		}, nil
	}

	// Method 2: Inspect /proc/<ppid>/exe on Linux
	procExe := fmt.Sprintf("/proc/%d/exe", ppid)
	if target, err := os.Readlink(procExe); err == nil && target != "" {
		base := filepath.Base(target)
		return &DetectedShell{
			Type:       classifyUnixShell(base),
			BinaryPath: target,
			ParentPID:  ppid,
			ParentName: base,
		}, nil
	}

	// Method 3: Fallback to common Unix shells in PATH
	for _, candidate := range []string{"bash", "zsh", "sh"} {
		if path, err := exec.LookPath(candidate); err == nil {
			return &DetectedShell{
				Type:       classifyUnixShell(candidate),
				BinaryPath: path,
				ParentPID:  ppid,
				ParentName: candidate,
			}, nil
		}
	}

	return &DetectedShell{
		Type:       ShellTypeBash,
		BinaryPath: "/bin/sh",
		ParentPID:  ppid,
		ParentName: "sh",
	}, nil
}

func classifyUnixShell(name string) ShellType {
	nameLower := strings.ToLower(name)
	switch {
	case strings.Contains(nameLower, "zsh"):
		return ShellTypeZsh
	case strings.Contains(nameLower, "bash"):
		return ShellTypeBash
	case strings.Contains(nameLower, "fish"):
		return ShellTypeFish
	case strings.Contains(nameLower, "nu"):
		return ShellTypeNushell
	default:
		return ShellTypeUnknown
	}
}

// SpawnSubshell spawns an interactive subshell with memory-injected prompt on Unix/Linux
func SpawnSubshell(shell *DetectedShell, cfg *SubshellConfig) (int, error) {
	var cmd *exec.Cmd

	switch shell.Type {
	case ShellTypeFish:
		// In-memory fish prompt
		promptFish := fmt.Sprintf(`function fish_prompt; set_color cyan; echo -n "[(aenv:%s)] "; set_color normal; echo -n (prompt_pwd) '> '; end`, cfg.PresetName)
		cmd = exec.Command(shell.BinaryPath, "-C", promptFish)

	default:
		// Bash, Zsh, and standard POSIX shells: interactive mode
		cmd = exec.Command(shell.BinaryPath, "-i")
	}

	// Forward stdio directly to host terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Construct environment variables
	env := os.Environ()
	env = append(env, cfg.FormatEnvSlice()...)

	// In-memory Prompt decoration for Bash and Zsh via PS1
	switch shell.Type {
	case ShellTypeZsh:
		env = append(env, fmt.Sprintf("PS1=[(aenv:%s)] %%~ %%# ", cfg.PresetName))
	case ShellTypeBash:
		env = append(env, fmt.Sprintf("PS1=[(aenv:%s)] \\u@\\h:\\w\\$ ", cfg.PresetName))
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
