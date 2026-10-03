# AgentEnv (`aenv`)

<p align="left">
  <b>English</b> | <a href="README_zh.md">简体中文</a>
</p>

**Stop juggling config files. Switch AI presets in seconds.**  
> **A lightweight, zero-hook environment and preset isolation manager for terminal AI coding agents (Claude Code, Codex CLI, Gemini CLI, OpenCode, and more).**  
> Switch between Work, Personal, and Client setups in seconds—unified multi-agent management, zero config collisions, zero terminal pollution.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-4D4D4D)](https://github.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Why aenv?

```bash
# Before: Manually backing up folders, editing settings.json, fearing lost tokens or history
cp -r ~/.claude ~/.claude.work.bak   # 😰 Tedious, and multi-window runs easily corrupt each other

# After: One command to enter an isolated preset; type exit to restore instantly
aenv shell work                      # ✅ Non-invasive, zero pollution, auto-penetrating
```

### Pain Points Comparison

| Pain Point | How `aenv` Solves It |
| :--- | :--- |
| **Fragmented multi-agent configs** | Unifies **Claude Code**, **Codex**, **Gemini**, and **OpenCode** under a single preset—switch once, isolate all. |
| **Global config collisions** | **Never overwrites your host home directory**. Uses ephemeral shadow viewports and subshells so multiple windows can run different presets safely. |
| **Reconfiguring everything from scratch** | **Overrides what exists, penetrates what doesn't**. Keep only diff files (like MCP configs or custom rules); auth tokens and session history penetrate seamlessly. |
| **Shell hooks mess up terminals** | **Zero-hook design**. No tampering with `.bashrc`, `.zshrc`, or PowerShell `$PROFILE`. Type `exit` to cleanly restore. |

---

## Quick Start (1 Minute)

### 1. Installation (Go 1.22+)

```bash
# Install directly
go install ./cmd/aenv

# Or build from source
git clone https://github.com/your-username/AgentEnv.git
cd AgentEnv && go build -o bin/aenv ./cmd/aenv
```

> **Windows Notice**: Windows native symlinks require **Developer Mode** enabled. Run once in Administrator PowerShell:
> ```powershell
> reg add "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock" /t REG_DWORD /f /v "AllowDevelopmentWithoutDevLicense" /d "1"
> ```
> You can run `aenv doctor` at any time to verify system readiness.

### 2. Three-Step Daily Workflow

```bash
# 1. Create a preset (e.g., named "work")
aenv preset create work

# 2. Drop your override files into the corresponding folder
# E.g., to override settings for Claude Code, place your settings.json into:
# ~/.agentenv/presets/work/claude/settings.json

# 3. Launch the preset environment
aenv shell work
# Terminal prompt changes to: [(aenv:work)]>
claude   # Claude now runs with your work configuration!

# 4. Exit anytime
exit
```

---

## Command-Line Usage (CLI)

### Core Commands

| Command | Description & Example |
| :--- | :--- |
| `aenv shell <preset>[/<agent>]` | **Spawn an isolated subshell** (alias: `activate`). Type `exit` to leave.<br>• `aenv shell work` (interactive arrow-key menu if multiple agents exist)<br>• `aenv shell work/claude` (target a single agent directly) |
| `aenv run <preset>[/<agent>] -- <cmd>` | **Execute a one-off command** in the preset environment without keeping a subshell. 100% transparent exit code pass-through.<br>• `aenv run work/claude -- claude --model sonnet`<br>• `aenv run work/claude --dry-run -- claude` (preview env vars without executing) |
| `aenv env <preset>[/<agent>]` | **Print environment variable exports** (read-only, formatted for your current shell).<br>• `aenv env work/claude` |

### Preset Management (`aenv preset`)

| Command | Description |
| :--- | :--- |
| `aenv preset create <name>` | Scaffold a new preset directory (creates `shared/` and registered agent folders) |
| `aenv preset list` | List all presets, target agents, and active override files |
| `aenv preset diff <preset>` | Show colorized diff between preset overrides and baseline host configs |

### System Maintenance

| Command | Description |
| :--- | :--- |
| `aenv doctor` | Run system diagnostics (permissions, Developer Mode, shell detection, and **auto-sync new agents to existing presets**) |
| `aenv status` | Show active preset and environment variables (use inside an active subshell) |
| `aenv clean` | Purge ephemeral runtime cache in `runtimes/` (safe to run anytime) |
| `aenv version` | Print current aenv version |

*Global flag `--no-color` is supported to disable ANSI colors.*

---

## How Presets Work

Presets are stored in `~/.agentenv/presets/` (Windows: `%USERPROFILE%\.agentenv\presets\`):

```text
~/.agentenv/presets/work/
├── shared/                   # Files shared across all agents in this preset (e.g. .mcp.json, AGENTS.md)
├── claude/                   # Claude-specific overrides (e.g. settings.json, CLAUDE.md)
└── codex/                    # Codex-specific overrides (e.g. config.toml)
```

**Three Simple Principles**:
1. **Override what exists**: Put `settings.json` in `claude/`, and it replaces the default config in the preset runtime.
2. **Penetrate what doesn't**: Anything omitted (e.g. `history.jsonl`, auth sessions) penetrates directly from your host baseline—no re-logging in.
3. **100% Isomorphic**: The directory structure mirrors native agent home folders (`~/.claude/`). No intermediate DSLs to learn.

---

## Registering Custom Agents (`config.toml`)

`claude`, `codex`, `gemini`, and `opencode` are built-in by default.  
To support any other CLI agent, simply add a few lines to `~/.agentenv/config.toml`:

```toml
[agents.myagent]
env_var = "MYAGENT_CONFIG_DIR"        # Env variable the tool reads for its config dir
host_dir = "~/.myagent"               # Host baseline directory
```

Then run `aenv doctor`, and it will automatically scaffold `myagent/` across all your existing presets!

---

## Exit Codes

| Exit Code | Description |
| :---: | :--- |
| `0` | Success |
| `1` | Business error (preset or agent not found) |
| `2` | Usage or CLI argument syntax error |
| `3` | System error (Windows Developer Mode / symlink privilege not held) |
| `N` | `aenv run` transparently forwards the target process's exit code |

---

## Acknowledgments & Inspiration

This project was inspired by the excellent open-source tool [claudectx](https://github.com/foxj77/claudectx).

`claudectx` pioneered the concept of quick profile switching for Claude Code. `aenv` builds upon this idea and evolves the paradigm:
- **From Single-Tool to Multi-Agent**: Seamless native support not just for Claude, but across Claude Code, Codex CLI, Gemini CLI, and any terminal AI agent.
- **From Global Overwriting to Zero-Hook Subshells**: Instead of writing to global `~/.claude` or hacking CLI flags, `aenv` uses **ephemeral shadow viewports and process-level subshells** for zero pollution, zero collisions, and safe multi-terminal concurrency.
- **From Full-Copy to Diff-Penetration**: Smart cascading inheritance and symlink farms ensure you only maintain the diffs you care about, eliminating configuration bloat.

Kudos to `claudectx` for inspiring the community!

---

## License

[MIT License](LICENSE)
