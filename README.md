# AgentEnv (`aenv`)

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-4D4D4D)](https://github.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**AgentEnv (`aenv`)** 是一个极简、极速、无侵入的通用 AI Coding Agent 预设隔离与环境运行时管理器（用 Go 实现）。

专为解决 **Claude Code**, **Codex CLI**, **Gemini CLI**, **OpenCode** 等终端 AI 工具在多终端、多项目、多身份切换下的**配置隔离、多端并发踩踏、MCP 冲突与规则污染**问题。

---

## 核心特性 (Key Features)

- ⚡ **零终端钩子侵入 (Zero-Hook Subshell)**：不碰 `.bashrc`、不改 PowerShell `$PROFILE`，无 `eval "$(aenv init)"`。按需拉起独立子 Shell，`exit` 即彻底还原。
- 🌲 **四层级联继承 (4-Tier Cascading Inheritance)**：有则覆写、无则穿透。预设只需存放差异文件，历史记录与未改动配置 100% 自动穿透宿主基底。
- 🔗 **纯原生软链农场 (Pure Native Symlink)**：内容无关（Content-Agnostic），不解析 JSON/TOML/SQLite，直接以文件系统拓扑为唯一事实源。
- ⏱️ **毫秒级增量缓存 (<1ms Cache)**：基于四层 `mtime` 的增量校验机制，未改动视口瞬间复用，无磁盘重写，多终端高并发零冲突。
- 🔄 **框架自动同步 (Framework Auto-Sync)**：新增 Agent 配置后，`aenv doctor` / `aenv shell` 自动为已有预设补全框架目录。

---

## 四层级联继承模型 (Cascading Inheritance)

视口合成引擎严格按照单向压制流水线合并生成影子运行目录（`~/.agentenv/runtimes/`）：

```text
[Layer 4 (Agent专属)]  presets/<preset>/<agent>/       (终极覆盖，如专属 settings.json)
        ▲
        │ 覆盖 / 穿透
[Layer 3 (预设共享)]    presets/<preset>/shared/        (预设内共享，如跨 Agent MCP 拓扑、AGENTS.md)
        ▲
        │ 覆盖 / 穿透
[Layer 2 (全局共享)]    ~/.agentenv/shared/             (全局通用 pool，如公共 skills/ 工具库)
        ▲
        │ 覆盖 / 穿透
[Layer 1 (宿主基底)]    ~/.<agent>/                     (宿主原生环境，穿透历史记录/鉴权会话)
```

---

## 快速安装 (Installation)

### 方式 1: 直接构建 (Go 1.22+)
```bash
git clone https://github.com/your-username/AgentEnv.git
cd AgentEnv
go build -o bin/aenv ./cmd/aenv
# 将 bin/ 添加到系统的 PATH 环境变量中
```

### 方式 2: Go Install
```bash
go install ./cmd/aenv
```

> **Windows 用户前置条件**：
> Windows 创建文件软链接需启用**开发者模式 (Developer Mode)**。以管理员身份在 PowerShell 中执行：
> ```powershell
> reg add "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock" /t REG_DWORD /f /v "AllowDevelopmentWithoutDevLicense" /d "1"
> ```
> 可通过 `aenv doctor` 命令自动检测系统状态。

---

## 目录结构拓扑 (Storage Layout)

数据均收敛在 `~/.agentenv/`（Windows 为 `%USERPROFILE%\.agentenv\`）：

```text
~/.agentenv/
├── config.toml                       # 全局注册表 (Agent 环境变量与基底映射)
├── shared/                           # 全局共享池 (Layer 2)
│   └── skills/                       # 全局通用工具与技能
├── presets/                          # 预设源码池（受 Git 追踪）
│   └── <preset_name>/
│       ├── shared/                   # 预设内跨 Agent 共享 (Layer 3)
│       ├── claude/                   # Claude 专属差量覆写 (Layer 4)
│       └── codex/                    # Codex 专属差量覆写 (Layer 4)
└── runtimes/                         # 易失影子运行视口（由 aenv 自动维护，可随时清空）
```

---

## 命令行完整手册 (CLI Usage)

```text
aenv [command] [flags]
```

### 1. 派生隔离子 Shell: `aenv shell` (别名: `activate`)
合成指定预设的影子视口，注入隔离环境变量并拉起当前终端同款交互式子 Shell。

```bash
# 激活预设中的单一指定 Agent
aenv shell work-backend/claude

# 激活预设（若包含多个 Agent，自动弹出 Bubbletea 交互式多选菜单）
aenv shell work-backend
```
*在子 Shell 中，命令提示符前缀将自动装饰为 `[(aenv:work-backend)]>`，输入 `exit` 即销毁子 Shell 恢复原状。*

### 2. 瞬时执行命令: `aenv run`
在隔离视口环境中一次性执行目标命令，不常驻子 Shell，**100% 透明转发子进程的标准流与退出码**。

```bash
# 语法：-- 之后的所有参数直接透传给目标程序
aenv run work-backend/claude -- claude --model sonnet

# 预览合成的环境变量与执行指令（不实际运行）
aenv run work-backend/claude --dry-run -- claude --version
```

### 3. 查看环境变量: `aenv env`
只读输出目标预设所需的环境变量语句（根据父终端 Shell 自动适配 `export` / `$env:` / `set` 语法）：

```bash
aenv env work-backend/claude
```

### 4. 预设管理: `aenv preset`

```bash
# 1. 初始化预设脚手架（自动创建 shared/ 及所有已注册 Agent 文件夹）
aenv preset create work-backend

# 2. 查看所有本地预设、生效目标及覆写文件列表
aenv preset list

# 3. 彩色 Diff：对比预设内的覆写文件与宿主原生基底文件的实际差异
aenv preset diff work-backend
```

### 5. 系统与运行时维护

```bash
# 系统与环境自检（检查权限、开发者模式、Shell 类型，并自动补齐新增 Agent 的预设框架）
aenv doctor

# 查看当前会话状态（仅在 aenv 子 Shell 内部有效）
aenv status

# 安全清空 runtimes/ 易失影子缓存
aenv clean

# 查看当前版本
aenv version
```

### 通用全局选项
- `--no-color`：禁用 ANSI 彩色输出。

---

## 配置规范 (`config.toml`)

位于 `~/.agentenv/config.toml`。首次运行会自动使用内置推荐配置，无需额外折腾：

```toml
[terminal]
# 可选：显式指定拉起子 Shell 使用的程序（"pwsh" | "powershell" | "cmd" | "bash" | "zsh" | "nu"）
# 留空或缺省时，走 Win32 Toolhelp32 / POSIX 智能父进程回溯探测
# shell = "pwsh"

[agents.claude]
env_var = "CLAUDE_CONFIG_DIR"
host_dir = "~/.claude"

[agents.codex]
env_var = "CODEX_CONFIG_DIR"
host_dir = "~/.codex"

[agents.gemini]
env_var = "GEMINI_CONFIG_DIR"
host_dir = "~/.gemini"

[agents.opencode]
env_var = "OPENCODE_HOME"
host_dir = "~/.opencode"
```

---

## 典型工作流示例 (Workflow Walkthrough)

以配置 `work-backend` 预设为例：

```bash
# 1. 创建预设框架
aenv preset create work-backend

# 2. 放入你的覆盖文件（在任意终端或资源管理器中）
# - 跨 Agent 共享的 MCP 配置:
#   ~/.agentenv/presets/work-backend/shared/.mcp.json
# - Claude 专属工作指令:
#   ~/.agentenv/presets/work-backend/claude/CLAUDE.md

# 3. 检查与宿主差异
aenv preset diff work-backend

# 4. 进入工作会话
aenv shell work-backend/claude
# 终端提示变为: [(aenv:work-backend)] ...>
claude  # 此时 Claude 自动读取隔离视口，共享宿主历史会话，但使用工作专有 MCP 与规约

# 5. 退出还原
exit
```

---

## 标准退出码 (Exit Codes)

| 代码 | 说明 | 场景 |
| :---: | :--- | :--- |
| **0** | `SUCCESS` | 正常执行完毕 |
| **1** | `ERR_BUSINESS` | 业务错误（预设或 Agent 未找到） |
| **2** | `ERR_USAGE` | 命令行参数或语法错误 |
| **3** | `ERR_SYSTEM` | 系统权限不足（Windows 未开启开发者模式） |
| **N** | `SUBPROCESS` | `aenv run` 透明返回目标子进程自身的退出码 |

---

## 许可证 (License)

[MIT License](LICENSE)
