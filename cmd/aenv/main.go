package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kkkklooooo/AgentEnv/internal/config"
	"github.com/kkkklooooo/AgentEnv/internal/engine"
	"github.com/kkkklooooo/AgentEnv/internal/platform"
	"github.com/kkkklooooo/AgentEnv/internal/shell"
	"github.com/kkkklooooo/AgentEnv/internal/ui"
)

const (
	Version = "v1.0.0"

	ExitSuccess  = 0
	ExitBusiness = 1
	ExitUsage    = 2
	ExitSystem   = 3
)

var (
	flagNoColor bool
	flagDryRun  bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "aenv",
		Short: "AgentEnv (aenv) - General AI Coding Agent Preset Isolation Runtime",
		Long: `AgentEnv (aenv) is a lightweight, zero-hook, multi-agent environment and preset isolation manager.
It provides instant shadow viewports with 4-tier cascading inheritance for Claude Code, Codex CLI, and more.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if flagNoColor {
				ui.DisableColors()
			}
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable ANSI color output")

	// Register subcommands
	rootCmd.AddCommand(newShellCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newEnvCmd())
	rootCmd.AddCommand(newPresetCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newCleanCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	if errors.Is(err, platform.ErrDeveloperModeRequired) {
		fmt.Fprintf(os.Stderr, "\n%s %v\n", ui.Bold(ui.Red("Error:")), err)
		if runtime.GOOS == "windows" {
			fmt.Fprintln(os.Stderr, ui.Yellow("\nTo enable Windows Developer Mode, run in Administrator PowerShell:"))
			fmt.Fprintln(os.Stderr, ui.Cyan(`reg add "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock" /t REG_DWORD /f /v "AllowDevelopmentWithoutDevLicense" /d "1"`))
		}
		os.Exit(ExitSystem)
	}

	fmt.Fprintf(os.Stderr, "%s %v\n", ui.Bold(ui.Red("Error:")), err)
	os.Exit(ExitBusiness)
}

// -------------------------------------------------------------
// aenv shell <preset>[/<agent>]
// -------------------------------------------------------------

func newShellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shell <preset>[/<agent>]",
		Aliases: []string{"activate"},
		Short:   "Spawn an isolated subshell for the specified preset",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			presetName, specificAgent := parsePresetAndAgent(args[0])
			if err := config.ValidatePresetName(presetName); err != nil {
				return err
			}

			presetDir, _ := config.PresetDir(presetName)
			if _, err := os.Stat(presetDir); os.IsNotExist(err) {
				return fmt.Errorf("preset %q does not exist in %s", presetName, presetDir)
			}

			// Ensure all registered agent folders are synchronized
			_, _ = engine.SyncPresetAgentFolders(cfg)

			var agentsToActivate []string
			if specificAgent != "" {
				if _, ok := cfg.Agents[specificAgent]; !ok {
					return fmt.Errorf("agent %q is not registered in config.toml", specificAgent)
				}
				agentsToActivate = []string{specificAgent}
			} else {
				available, err := engine.DetectPresetAgents(cfg, presetName)
				if err != nil {
					return err
				}
				selected, err := ui.SelectAgents(presetName, available)
				if err != nil {
					return err
				}
				agentsToActivate = selected
			}

			// Synthesize viewports for selected agents
			synth := engine.NewSynthesizer(cfg)
			envVars := map[string]string{
				"AENV_ACTIVE": "1",
				"AENV_PRESET": presetName,
				"AENV_AGENTS": strings.Join(agentsToActivate, ","),
			}

			for _, ag := range agentsToActivate {
				runtimeDir, err := synth.SynthesizeAgent(presetName, ag)
				if err != nil {
					return fmt.Errorf("failed to synthesize viewport for %s: %w", ag, err)
				}
				agCfg := cfg.Agents[ag]
				envVars[agCfg.EnvVar] = runtimeDir
			}

			// Detect parent shell
			detected, err := shell.DetectShellWithOverride(cfg.Terminal.Shell)
			if err != nil {
				return fmt.Errorf("failed to detect shell: %w", err)
			}

			subCfg := &shell.SubshellConfig{
				PresetName: presetName,
				EnvVars:    envVars,
			}

			// Launch subshell
			exitCode, err := shell.SpawnSubshell(detected, subCfg)
			if err != nil {
				return fmt.Errorf("subshell execution failed: %w", err)
			}

			os.Exit(exitCode)
			return nil
		},
	}
	return cmd
}

// -------------------------------------------------------------
// aenv run <preset>[/<agent>] [--dry-run] -- <command...>
// -------------------------------------------------------------

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <preset>[/<agent>] [flags] -- <command...>",
		Short: "Execute a command inside the preset runtime environment without entering a subshell",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find -- separator
			dashIdx := cmd.ArgsLenAtDash()
			if dashIdx == -1 || dashIdx == 0 {
				return fmt.Errorf("command to execute must be specified after '--', e.g.: aenv run %s -- claude", "work-backend")
			}

			targetArg := args[0]
			cmdToRun := args[dashIdx:]
			if len(cmdToRun) == 0 {
				return errors.New("no command specified after '--'")
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			presetName, specificAgent := parsePresetAndAgent(targetArg)
			if err := config.ValidatePresetName(presetName); err != nil {
				return err
			}

			presetDir, _ := config.PresetDir(presetName)
			if _, err := os.Stat(presetDir); os.IsNotExist(err) {
				return fmt.Errorf("preset %q does not exist in %s", presetName, presetDir)
			}

			// Ensure all registered agent folders are synchronized
			_, _ = engine.SyncPresetAgentFolders(cfg)

			var agentsToActivate []string
			if specificAgent != "" {
				if _, ok := cfg.Agents[specificAgent]; !ok {
					return fmt.Errorf("agent %q is not registered in config.toml", specificAgent)
				}
				agentsToActivate = []string{specificAgent}
			} else {
				available, err := engine.DetectPresetAgents(cfg, presetName)
				if err != nil {
					return err
				}
				agentsToActivate = available
			}

			synth := engine.NewSynthesizer(cfg)
			envVars := map[string]string{
				"AENV_ACTIVE": "1",
				"AENV_PRESET": presetName,
				"AENV_AGENTS": strings.Join(agentsToActivate, ","),
			}

			for _, ag := range agentsToActivate {
				runtimeDir, err := synth.SynthesizeAgent(presetName, ag)
				if err != nil {
					return fmt.Errorf("failed to synthesize viewport for %s: %w", ag, err)
				}
				agCfg := cfg.Agents[ag]
				envVars[agCfg.EnvVar] = runtimeDir
			}

			if flagDryRun {
				fmt.Println(ui.Bold(ui.Cyan("--- Dry-Run Mode ---")))
				fmt.Println(ui.Bold("Target Preset:"), presetName)
				fmt.Println(ui.Bold("Active Agents:"), strings.Join(agentsToActivate, ", "))
				fmt.Println(ui.Bold("\nInjected Environment Variables:"))
				for k, v := range envVars {
					fmt.Printf("  %s=%s\n", ui.Green(k), v)
				}
				fmt.Println(ui.Bold("\nCommand to Execute:"))
				fmt.Printf("  %s %s\n", ui.Yellow(cmdToRun[0]), strings.Join(cmdToRun[1:], " "))
				return nil
			}

			exitCode, err := shell.ExecCommand(cmdToRun[0], cmdToRun[1:], envVars)
			if err != nil {
				return fmt.Errorf("failed to execute %s: %w", cmdToRun[0], err)
			}

			os.Exit(exitCode)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Preview environment variables and command without executing")
	return cmd
}

// -------------------------------------------------------------
// aenv env <preset>[/<agent>]
// -------------------------------------------------------------

func newEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env <preset>[/<agent>]",
		Short: "Print environment variable export statements for the preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			presetName, specificAgent := parsePresetAndAgent(args[0])
			if err := config.ValidatePresetName(presetName); err != nil {
				return err
			}

			var agents []string
			if specificAgent != "" {
				agents = []string{specificAgent}
			} else {
				available, err := engine.DetectPresetAgents(cfg, presetName)
				if err != nil {
					return err
				}
				agents = available
			}

			synth := engine.NewSynthesizer(cfg)
			envVars := map[string]string{
				"AENV_ACTIVE": "1",
				"AENV_PRESET": presetName,
				"AENV_AGENTS": strings.Join(agents, ","),
			}

			for _, ag := range agents {
				runtimeDir, err := synth.SynthesizeAgent(presetName, ag)
				if err != nil {
					return err
				}
				agCfg := cfg.Agents[ag]
				envVars[agCfg.EnvVar] = runtimeDir
			}

			// Format exports according to current shell
			detected, _ := shell.DetectShellWithOverride(cfg.Terminal.Shell)
			isPowerShell := detected != nil && (detected.Type == shell.ShellTypePwsh || detected.Type == shell.ShellTypePowerShell)
			isCmd := detected != nil && detected.Type == shell.ShellTypeCmd

			for k, v := range envVars {
				if isPowerShell {
					fmt.Printf("$env:%s = %q\n", k, v)
				} else if isCmd {
					fmt.Printf("set %s=%s\n", k, v)
				} else {
					fmt.Printf("export %s=%q\n", k, v)
				}
			}

			return nil
		},
	}
	return cmd
}

// -------------------------------------------------------------
// aenv preset (list | create | diff)
// -------------------------------------------------------------

func newPresetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Manage presets and inspect override diffs",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all local presets and their overrides",
		RunE: func(cmd *cobra.Command, args []string) error {
			presetsDir, err := config.PresetsDir()
			if err != nil {
				return err
			}

			entries, err := os.ReadDir(presetsDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No presets found.")
					return nil
				}
				return err
			}

			table := ui.NewTable("PRESET", "TARGET", "ITEMS")
			hasPresets := false

			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				hasPresets = true
				pName := e.Name()
				pDir := filepath.Join(presetsDir, pName)

				subEntries, _ := os.ReadDir(pDir)
				first := true
				for _, sub := range subEntries {
					if !sub.IsDir() {
						continue
					}
					target := sub.Name()
					items := listItemsBrief(filepath.Join(pDir, target))

					if first {
						table.AddRow(ui.Bold(ui.Cyan(pName)), ui.Green(target), items)
						first = false
					} else {
						table.AddRow("", ui.Green(target), items)
					}
				}
				if first {
					table.AddRow(ui.Bold(ui.Cyan(pName)), ui.Gray("(empty)"), "-")
				}
			}

			if !hasPresets {
				fmt.Println("No presets found. Create one with: aenv preset create <name>")
				return nil
			}

			table.Render(os.Stdout)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create <name>",
		Short: "Scaffold a new preset directory structure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := config.ValidatePresetName(name); err != nil {
				return err
			}

			pDir, err := config.PresetDir(name)
			if err != nil {
				return err
			}

			if _, err := os.Stat(pDir); err == nil {
				return fmt.Errorf("preset %q already exists in %s", name, pDir)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			// Create shared directory
			sharedDir := filepath.Join(pDir, "shared")
			if err := os.MkdirAll(sharedDir, 0755); err != nil {
				return err
			}

			// Create subdirectories for registered agents
			for agName := range cfg.Agents {
				agDir := filepath.Join(pDir, agName)
				_ = os.MkdirAll(agDir, 0755)
			}

			fmt.Println(ui.Green("✓"), "Created preset scaffold:", ui.Bold(name))
			fmt.Println("\nDirectory structure:")
			fmt.Printf("  %s/\n", ui.Cyan(pDir))
			fmt.Printf("  ├── shared/        %s\n", ui.Gray("(rules/skills shared by all agents in this preset)"))
			for agName := range cfg.Agents {
				fmt.Printf("  ├── %s/        %s\n", agName, ui.Gray(fmt.Sprintf("(overrides specific to %s)", agName)))
			}
			fmt.Printf("\nPlace your override files directly in the folders above, then run:\n  %s\n", ui.Bold("aenv shell "+name))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "diff <preset>",
		Short: "Show unified diff between preset overrides and host baselines",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			presetName := args[0]
			diffText, err := engine.PresetDiff(cfg, presetName)
			if err != nil {
				return err
			}

			// Colorize diff output
			lines := strings.Split(diffText, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
					fmt.Println(ui.Bold(line))
				} else if strings.HasPrefix(line, "@@") {
					fmt.Println(ui.Cyan(line))
				} else if strings.HasPrefix(line, "+") {
					fmt.Println(ui.Green(line))
				} else if strings.HasPrefix(line, "-") {
					fmt.Println(ui.Red(line))
				} else {
					fmt.Println(line)
				}
			}

			return nil
		},
	})

	return cmd
}

func listItemsBrief(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return "-"
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return strings.Join(names, ", ")
}

// -------------------------------------------------------------
// aenv doctor
// -------------------------------------------------------------

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system permissions, Developer Mode, and agent configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(ui.Bold(ui.Cyan("AgentEnv (aenv) System Diagnostics:")))
			fmt.Println()

			table := ui.NewTable("CHECK", "STATUS", "DETAILS")

			// 1. OS & Architecture
			table.AddRow("Platform", ui.Green("OK"), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))

			// 2. Symlink & Developer Mode capability
			symlinkErr := platform.CheckSymlinkCapability()
			if symlinkErr == nil {
				table.AddRow("Symlinks", ui.Green("OK"), "Native symlinks enabled")
			} else {
				if runtime.GOOS == "windows" {
					devMode, _ := platform.CheckDeveloperMode()
					if !devMode {
						table.AddRow("Symlinks", ui.Red("FAIL"), "Developer Mode disabled (requires elevation/settings)")
					} else {
						table.AddRow("Symlinks", ui.Red("FAIL"), symlinkErr.Error())
					}
				} else {
					table.AddRow("Symlinks", ui.Red("FAIL"), symlinkErr.Error())
				}
			}

			// 3. Shell Detection
			cfg, _ := config.LoadConfig()
			var override string
			if cfg != nil {
				override = cfg.Terminal.Shell
			}
			detected, shellErr := shell.DetectShellWithOverride(override)
			if shellErr == nil {
				table.AddRow("Shell Detector", ui.Green("OK"), fmt.Sprintf("%s (%s)", detected.Type, detected.BinaryPath))
			} else {
				table.AddRow("Shell Detector", ui.Yellow("WARN"), shellErr.Error())
			}

			// 4. Config & Registered Agents
			if cfg != nil {
				var agSummaries []string
				for name, ag := range cfg.Agents {
					host, _ := ag.ResolvedHostDir()
					status := "found"
					if _, err := os.Stat(host); os.IsNotExist(err) {
						status = "uninitialized"
					}
					agSummaries = append(agSummaries, fmt.Sprintf("%s (%s)", name, status))
				}
				table.AddRow("Agents Registry", ui.Green("OK"), strings.Join(agSummaries, ", "))
			}

			// 5. Presets check
			pDir, _ := config.PresetsDir()
			pEntries, _ := os.ReadDir(pDir)
			table.AddRow("Presets Storage", ui.Green("OK"), fmt.Sprintf("%d presets in %s", len(pEntries), pDir))

			// 6. Preset Framework Synchronization (Check & auto-scaffold missing agent folders)
			syncReport, syncErr := engine.SyncPresetAgentFolders(cfg)
			if syncErr != nil {
				table.AddRow("Preset Framework", ui.Red("FAIL"), syncErr.Error())
			} else if syncReport.TotalCreated > 0 {
				table.AddRow("Preset Framework", ui.Green("SYNCED"), fmt.Sprintf("Auto-scaffolded %d missing agent folders across %d presets", syncReport.TotalCreated, len(syncReport.Results)))
			} else {
				table.AddRow("Preset Framework", ui.Green("OK"), "All presets synchronized with agent registry")
			}

			// 7. Runtimes Cache
			rDir, _ := config.RuntimesDir()
			rEntries, _ := os.ReadDir(rDir)
			table.AddRow("Runtimes Cache", ui.Green("OK"), fmt.Sprintf("%d active viewports in %s", len(rEntries), rDir))

			table.Render(os.Stdout)

			if syncReport != nil && syncReport.TotalCreated > 0 {
				fmt.Println()
				fmt.Println(ui.Green("✓"), ui.Bold("Framework Sync: Auto-scaffolded missing folders for newly registered agents:"))
				for _, res := range syncReport.Results {
					fmt.Printf("   * %s: created [%s]\n", ui.Bold(ui.Cyan(res.PresetName)), strings.Join(res.CreatedDirs, ", "))
				}
			}

			if symlinkErr != nil && runtime.GOOS == "windows" {
				fmt.Println()
				fmt.Println(ui.Yellow("Notice: Windows symlink creation failed."))
				fmt.Println("To enable Developer Mode, run Administrator PowerShell:")
				fmt.Println(ui.Cyan(`reg add "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock" /t REG_DWORD /f /v "AllowDevelopmentWithoutDevLicense" /d "1"`))
			}

			return nil
		},
	}
	return cmd
}

// -------------------------------------------------------------
// aenv clean
// -------------------------------------------------------------

func newCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clear all ephemeral runtime viewports",
		RunE: func(cmd *cobra.Command, args []string) error {
			count, err := engine.CleanRuntimes()
			if err != nil {
				return err
			}
			fmt.Printf("%s Cleaned %d runtime viewports from runtimes/ cache.\n", ui.Green("✓"), count)
			return nil
		},
	}
	return cmd
}

// -------------------------------------------------------------
// aenv status
// -------------------------------------------------------------

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show active preset and environment status",
		RunE: func(cmd *cobra.Command, args []string) error {
			active := os.Getenv("AENV_ACTIVE")
			if active != "1" {
				fmt.Println("Not currently inside an AgentEnv (aenv) subshell.")
				fmt.Println("Activate a preset with:", ui.Bold("aenv shell <preset>"))
				return nil
			}

			preset := os.Getenv("AENV_PRESET")
			agents := os.Getenv("AENV_AGENTS")

			fmt.Println(ui.Bold(ui.Cyan("AgentEnv (aenv) Active Session:")))
			fmt.Println()

			table := ui.NewTable("PROPERTY", "VALUE")
			table.AddRow("Preset", ui.Green(preset))
			table.AddRow("Agents", ui.Green(agents))

			cfg, _ := config.LoadConfig()
			if cfg != nil {
				for _, ag := range strings.Split(agents, ",") {
					ag = strings.TrimSpace(ag)
					if agCfg, ok := cfg.Agents[ag]; ok {
						val := os.Getenv(agCfg.EnvVar)
						table.AddRow(agCfg.EnvVar, val)
					}
				}
			}

			table.Render(os.Stdout)
			return nil
		},
	}
	return cmd
}

// -------------------------------------------------------------
// aenv version
// -------------------------------------------------------------

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print aenv version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("aenv %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
		},
	}
}

// -------------------------------------------------------------
// Helpers
// -------------------------------------------------------------

func parsePresetAndAgent(input string) (string, string) {
	input = strings.Trim(input, "/\\")
	if strings.Contains(input, "/") {
		parts := strings.SplitN(input, "/", 2)
		return parts[0], parts[1]
	}
	if strings.Contains(input, "\\") {
		parts := strings.SplitN(input, "\\", 2)
		return parts[0], parts[1]
	}
	return input, ""
}
