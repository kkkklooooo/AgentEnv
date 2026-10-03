# AGENTS.md - AgentEnv (aenv) AI Agent 指南与架构规约

> **致所有在此仓库工作的 AI Coding Agent**：  
> 本文件为本项目的核心操作守则与架构共识。在阅读、设计、修改或生成本仓库的任何代码前，必须严格遵循本文档所声明的架构铁律与设计规约。

---

## 1. 项目全景与定位 (Project Identity)

**AgentEnv (`aenv`)** 是一个**轻量、极速、无侵入、零依赖的通用 AI Coding Agent 预设隔离与环境运行时管理器**（用 Go 语言实现）。

它专为解决多终端、多项目、多身份切换下终端 AI 编码助手（如 Claude Code, Codex CLI, Gemini CLI, OpenCode, Pi Agent 等）的配置隔离、多端并发踩踏、MCP 冲突与规则污染问题。

---

## 2. 核心架构设计铁律 (Four Invariants)

在编写任何代码或设计新功能时，必须严格遵守以下四大不可违背的工程铁律：

### 铁律 1：文件系统作为唯一真理来源，拒绝内容微管理 (Filesystem Single Source of Truth)
- **“有即能用”原则**：Preset 目录中的文件拓扑与 Agent 原生主目录完全同构。Preset 中放了什么文件就覆盖什么文件，没放的文件 100% 穿透映射到宿主基底。
- **内容无关 (Content-Agnostic)**：`aenv` 是纯粹的基础设施层，**严禁解析各 Agent 配置文件的内部格式与语义**（不关心是 JSON/TOML/YAML/SQLite，不关心哪个字段存 Token 哪个存 Model）。
- **零中间 DSL**：坚决杜绝为了所谓“统一”而发明中间配置语言（严禁引入如 `preset.toml`、`skills.yaml` 等多余元数据文件）。

### 铁律 2：四层自动级联继承模型 (Four-Tier Cascading Inheritance)
视口合成引擎必须严格按照以下顺序执行单向压制流水线：
```text
[Layer 1 (底层基底)] 宿主原生基底 (Host Base: ~/.<agent>/)
        ▲
        │ 穿透 / 被自动覆盖
[Layer 2 (全局共享)] 全局共享池 (Global Shared: ~/.agentenv/shared/)
        ▲
        │ 穿透 / 被自动覆盖
[Layer 3 (预设共享)] 预设内部共享池 (Preset Shared: ~/.agentenv/presets/<preset>/shared/)
        ▲
        │ 终极覆盖
[Layer 4 (Agent专属)] Agent专属差量 (Agent Specific: ~/.agentenv/presets/<preset>/<agent>/)
```
- 预设根目录内部**只有子目录**：即 `shared/`（预设内部共享）和各注册的 Agent 专属目录（如 `claude/`, `codex/`），严禁散落松散根级文件。
- 跨层级同名目录采用**智能目录递归合并 (Smart Directory Merge)**，叶子文件逐层替换，独有文件向上穿透。

### 铁律 3：零终端钩子侵入，子 Shell 派生与瞬时运行 (Zero-Hook Subshell & Runner)
- **坚决不使用 Shell 钩子**：严禁采用 `eval "$(aenv init)"` 模式，严禁修改用户的 `.bashrc`、`.zshrc` 或 PowerShell `$PROFILE`。
- **Subshell 模式 (`aenv shell <preset>`)**：合成视口 -> 注入环境 -> 原地拉起目标子 Shell -> 退出时输入 `exit` 即恢复，对宿主终端零残留、零污染。
- **跨平台 Shell 探测**：在 Windows 上严禁依赖不存在的 `$SHELL` 或固定为 CMD 的 `%COMSPEC%`。必须通过 **Win32 Toolhelp32 API 回溯父进程 PID**，精准识别并拉起当前终端同款 Shell（`pwsh` / `powershell` / `cmd` / `nu` / `bash`）。

### 铁律 4：纯原生软链农场与增量缓存 (Pure Symlink Farm & Incremental Cache)
- **统一原生 Symlink**：在 Windows 下坚决统一采用原生软链接（要求开启 Developer Mode），**严禁使用 Hardlink**（避免跨卷断裂与原子写脱钩）或 Copy-on-Write 降级。
- **增量比对与缓存复用**：视口维护 `.synthesis_meta.json` 记录 4 层的最新 `mtime`。未变动时直接复用，耗时 `< 1ms`，零磁盘写操作，彻底杜绝 Windows 进程文件占用锁冲突。

---

## 3. 核心文档事实源索引 (Documentation Source of Truth)

修改代码或扩展设计前，请参考 [docs/](file:///D:/Projects/Go/AgentEnv/docs) 目录下的详尽文档：

- 📄 [docs/spec.md](file:///D:/Projects/Go/AgentEnv/docs/spec.md) - **唯一定量技术规范标准**（包含数据结构、CLI 命令集、退出码、`.synthesis_meta.json` 结构、Win32 探测流程与 Go 模块划分）。
- 📄 [docs/intend.md](file:///D:/Projects/Go/AgentEnv/docs/intend.md) - **架构设计深度意图与技术分支决策**（记录历次 Grill-Me 裁决过程与底层设计权衡推演）。
- 📄 [docs/Goal.md](file:///D:/Projects/Go/AgentEnv/docs/Goal.md) - **原始愿景与核心痛点定义**。
- 📄 [docs/claudectx_source_analysis.md](file:///D:/Projects/Go/AgentEnv/docs/claudectx_source_analysis.md) - **对标项目 `claudectx` 源码级竞品分析**（深入揭示其全局硬写与参数追加的致命缺陷）。
- 📂 `analyse_source/claudectx/` - 对标项目的参考源码（只读参考，已加入 `.gitignore`）。

---

## 4. 目标工程组织结构 (Go Project Layout)

后续在本项目根目录下创建和组织 Go 代码时，必须严格对齐 [docs/spec.md](file:///D:/Projects/Go/AgentEnv/docs/spec.md) 中的工程规划：

```text
D:\Projects\Go\AgentEnv\
├── cmd\
│   └── aenv\
│       └── main.go                   # CLI 入口 (基于 Cobra 构建命令树)
│
├── internal\
│   ├── config\                       # ~/.agentenv/config.toml 与注册表解析
│   │   ├── config.go                 # 全局配置读取与写入
│   │   ├── agent.go                  # Agent 注册表模型与内置列表
│   │   └── paths.go                  # 跨平台路径解析与规约
│   │
│   ├── engine\                       # 四层级联影子视口合成引擎
│   │   ├── synthesizer.go            # 核心四层 Farm Builder 与原子替换
│   │   ├── meta.go                   # .synthesis_meta.json 增量 mtime 校验
│   │   ├── cleaner.go                # 易失视口清理与死链剔除
│   │   └── diff.go                   # 彩色 Unified Diff 算法
│   │
│   ├── platform\                     # 跨平台文件系统原语抽象
│   │   ├── symlink.go                # 统一 Link 接口 (CreateSymlink, ValidateLink)
│   │   ├── symlink_unix.go           # POSIX 平台原生软链驱动
│   │   └── symlink_windows.go        # Windows 软链与 Developer Mode 状态探测
│   │
│   ├── shell\                        # 无钩子 Shell 探测与子 Shell 派生
│   │   ├── detect.go                 # 统一探测接口
│   │   ├── detect_windows.go         # Win32 Toolhelp32 父进程追踪器
│   │   ├── detect_unix.go            # POSIX $SHELL 与 /proc 探测器
│   │   ├── subshell.go               # 子进程启动与标准流/信号桥接
│   │   ├── prompt_windows.go         # PowerShell / CMD 内存级 Prompt 生成
│   │   └── prompt_unix.go            # Bash / Zsh 内存级 Prompt 生成
│   │
│   └── ui\                           # 终端交互式 UI
│       ├── menu.go                   # Bubbletea 交互式方向键多选菜单
│       ├── table.go                  # Doctor / Status 结构化表格打印
│       └── colors.go                 # ANSI 彩色样式定义
│
├── docs\                             # 设计意图、定量技术规范与调研文档
├── go.mod
└── go.sum
```

---

## 5. 编码规范与测试要求 (Development Standards)

1. **环境与依赖**：
   - 语言版本：**Go 1.22+**
   - 核心 CLI 框架：`github.com/spf13/cobra`
   - 终端 TUI 交互：`github.com/charmbracelet/bubbletea` / `huh`
   - 系统级 Windows API：`golang.org/x/sys/windows`
2. **跨平台构建保证**：
   - 涉及操作系统特性的代码必须通过 `_windows.go` 和 `_unix.go` 文件名后缀或 `//go:build` 条件编译分离。
   - 路径拼接必须始终使用 `filepath.Join`，严禁硬编码 `/` 或 `\`。
   - 用户目录解析必须使用 `os.UserHomeDir()`。
3. **统一标准化退出码**：
   - `0`: 成功
   - `1`: 业务错误（预设或 Agent 不存在）
   - `2`: 参数语法错误
   - `3`: 系统/环境错误（如 Windows 未开启开发者模式）
   - 目标子进程退出码：`aenv run` 必须 100% 透明转发子进程实际退出码。
4. **防御性编程**：
   - 捕获 Windows 错误码 `0x522` 并给出友好的 PowerShell 开启开发者模式指令；
   - 遍历外部目录时始终处理悬空死链（Dangling Symlink），绝不发生 Panic。

---

## 6. 禁止事项清单 (Anti-Patterns / Never Do)

- ❌ **严禁复制或直接修改宿主原生的 `~/.claude` 或 `~/.codex`**。一切修改只能发生在预设源码 `presets/` 或影子视口 `runtimes/` 中。
- ❌ **严禁编写任何 Shell Hook 初始化脚本**（如 `eval "$(aenv init)"`）。
- ❌ **严禁在 Preset 根目录下生成或要求散落配置文件**（必须放入 `presets/<preset>/shared/`）。
- ❌ **严禁引入中间层 DSL 清单文件**（如 `preset.yaml`），一切以文件系统真实存在的目录和文件为准。
