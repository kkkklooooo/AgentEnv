package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kkkklooooo/AgentEnv/internal/shell"
)

func main() {
	preset := "demo-test"
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		preset = os.Args[1]
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// 1. Detect parent terminal shell
	detected, err := shell.DetectActiveShell()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting shell: %v\n", err)
		os.Exit(1)
	}

	// 2. Prepare mock viewport runtime paths & env vars
	mockClaudeRuntime := filepath.Join(homeDir, ".agentenv", "runtimes", preset, "claude")
	mockCodexRuntime := filepath.Join(homeDir, ".agentenv", "runtimes", preset, "codex")

	cfg := &shell.SubshellConfig{
		PresetName: preset,
		EnvVars: map[string]string{
			"AENV_ACTIVE":       "1",
			"AENV_PRESET":       preset,
			"AENV_AGENTS":       "claude,codex",
			"CLAUDE_CONFIG_DIR": mockClaudeRuntime,
			"CODEX_CONFIG_DIR":  mockCodexRuntime,
		},
	}

	// 3. Print diagnostic banner
	fmt.Println("=================================================================")
	fmt.Println("   AgentEnv (aenv) - Zero-Hook Subshell Verification POC        ")
	fmt.Println("=================================================================")
	fmt.Printf(" [Platform]       %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf(" [Detected Shell] %s\n", detected.Type)
	fmt.Printf(" [Binary Path]    %s\n", detected.BinaryPath)
	fmt.Printf(" [Parent Process] PID: %d (%s)\n", detected.ParentPID, detected.ParentName)
	fmt.Printf(" [Active Preset]  %s\n", preset)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println(" [Injected Environment Variables]:")
	for k, v := range cfg.EnvVars {
		fmt.Printf("   * %s = %s\n", k, v)
	}
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println(" [Verification Guide]:")
	fmt.Println("   1. Notice the prompt prefix: [(aenv:" + preset + ")]")
	fmt.Println("   2. Inspect env vars:")
	if runtime.GOOS == "windows" {
		fmt.Println("      PowerShell : $env:CLAUDE_CONFIG_DIR")
		fmt.Println("      CMD        : echo %CLAUDE_CONFIG_DIR%")
	} else {
		fmt.Println("      Bash/Zsh   : echo $CLAUDE_CONFIG_DIR")
	}
	fmt.Println("   3. When finished, type 'exit' to restore host terminal.")
	fmt.Println("=================================================================")
	fmt.Println()

	// 4. Launch subshell
	exitCode, err := shell.SpawnSubshell(detected, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running subshell: %v\n", err)
		os.Exit(1)
	}

	// 5. Post-exit clean feedback
	fmt.Println()
	fmt.Println("=================================================================")
	fmt.Printf(" [aenv-poc] Subshell session ended cleanly (Exit Code: %d).\n", exitCode)
	fmt.Println(" [aenv-poc] Host terminal state is 100% restored. Zero residue.")
	fmt.Println("=================================================================")
	os.Exit(exitCode)
}
