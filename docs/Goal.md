# AI Coding Agent 预设隔离运行时 (agentenv)
## 1. 项目愿景与设计理念 (Vision & Philosophy)
### 1.1 核心痛点
主流终端 AI 编码助手（Claude Code, OpenCode, DeepSeek-harness, Codex CLI, Pi Agent 等）逐渐成为开发基础设施，但在多项目与多身份切换时存在明显的架构缺陷：

多身份与行为冲突：个人业余项目、公司业务代码、开源协同需要截然不同的 MCP 服务白名单、系统级 Prompt（如 CLAUDE.md / AGENTS.md）和模型行为偏好,Skills等等

全局覆写脆弱性：现有方案多依赖外部脚本粗暴替换 ~/.claude 等全局配置，多终端并发运行同一工具时必然导致 Race Condition、会话缓存踩踏与配置污染。


目录混杂泥潭：真实 Agent 的全局目录（如 ~/.claude/、~/.codex/）充斥着大型 SQLite 数据库、WAL 日志、Node 模块和全局会话记录。全量复制或持久化独立目录会导致数据严重割裂、同步噩梦与磁盘膨胀。

诉求本质是差量覆盖：用户切换预设，核心诉求仅是改变指令规则（Markdown）、工具拓扑（MCP）、模型/参数开关（Settings）以及按需注入的 Skills，其余的登录态、全局历史与底层缓存应当自然穿透与共享。


### 1.2 设计哲学
原生优先，拒绝二传手 DSL：不发明中间配置语言。每个 Preset 直接保存各 Agent 原生的配置文件结构，保持 100% 原生兼容与零理解成本。**文件系统作为事实唯一来源**

配置隔离，凭据穿透：行为规则、MCP 工具集、Prompt 与运行历史按 Preset 完全隔离；高频登录凭据（OAuth Token / API Keys）通过反向软链在全局单例共享，支持按需退化覆盖。

多模式按需使用：既支持显式环境激活（Conda 风格，直接执行原生命令），也支持非侵入式单次命令执行（run 模式）

轻量跨平台单二进制：Go 语言自包含实现，Shell 侧仅保留极薄的 Env-Printing Hook，无重型依赖。

最小差量声明 (Delta-Only Presets)：Preset 仓库绝不全量复制目录，只存放用户明确指定的覆写项（Override Files），没有定义的文件 100% 穿透继承宿主真实环境。

影子视口合成 (Shadow Viewport / Symlink Farm)：激活或运行时，动态构建一个毫秒级极轻量的虚拟视口目录。视口内对宿主真实目录做基底映射，再将 Preset 内部的 Override 文件强行覆盖在上层。


2. 核心架构与存储拓扑 (Architecture)
2.1 目录组织结构
所有配置与状态集中托管于 ~/.agentenv/，严格区分共享资产与独立预设：

~/.agentenv/
├── presets/                          # 用户真正维护的配置（只存差量，可纳入 Git 追踪）
│   ├── work-backend/
│   │   ├── claude/
│   │   │   ├── settings.json         # 覆盖：公司专用模型与白名单
│   │   │   └── CLAUDE.md             # 覆盖：后端开发规范
│   │   │   └── skills/               # preset 特定skills
│   │   │   └── .mcp.json             # preset 特定skills
│   │   ├── codex/
│   │   │   └── config.toml           # 覆盖：沙箱等级设置
│   │   │   └── ...
│   │   └── gemini/
│   │       └── GEMINI.md             # 覆盖：特定角色指令
│   │       └── ...
│   └── personal-frontend/
│
├── shared/                           # 可复用的通用组件池
│   └── skills/                       # 通用 Skill 脚本，按需注入到各 preset
│
└── runtimes/                         # 纯易失临时影子视口 (Shadow Viewports)
    └── work-backend/
        ├── claude/                   # 作为 CLAUDE_CONFIG_DIR 指向的目标
        │   ├── history.jsonl  ──────────> 指向宿主真实的 ~/.claude/history.jsonl
        │   ├── sessions/      ──────────> 指向宿主真实的 ~/.claude/sessions/
        │   ├── .credentials.json ───────> 指向宿主真实的 ~/.claude/.credentials.json
        │   ├── settings.json  ──────────> 指向 presets/work-backend/claude/settings.json [OVERRIDE]
        │   └── CLAUDE.md      ──────────> 指向 presets/work-backend/claude/CLAUDE.md   [OVERRIDE]
        └── codex/
            └── ... (同理合成)

### 2.2 核心机制详解


当执行激活或运行命令时，合成引擎执行以下毫秒级操作：

基底扫描：读取宿主原生的 Agent 全局目录（如 ~/.claude），遍历其所有一级子项。

基底映射：在 runtimes/<preset>/<agent>/ 影子目录下，为宿主子项批量创建软链接（Symlink / Junction）。

差量压制 (Override Injection)：检查 presets/<preset>/<agent>/，若存在同名文件或目录，直接解除基底软链，重新链接到 Preset 内部的覆盖项。

共享装配：将 Preset 声明需要的共享 Skills/Prompts 挂载进入影子视口的对应目录。

环境输出：将该影子目录绝对路径写入目标 Agent 的环境变量并暴露。

注：影子视口仅是轻量的软链网状结构（Symlink Farm），不占用实际存储；文件读写无拷贝开销，宿主状态即时同步。


B. 顽固工具的进程级 Fake HOME 注入(MVP暂时不做)
对于缺乏专用环境变量、硬编码读取 ~ 的 Agent（或单文件如 ~/.claude.json）：

严禁在当前终端执行全局 export HOME=...，避免污染 Git、SSH、Shell 提示符等宿主环境。

仅在 agentenv run 拉起子进程的瞬间，构造专用的 fake-home 视口，将宿主的核心基础设施（.gitconfig, .ssh, .zshrc 等）安全透传软链过去。

仅对该子进程临时注入 HOME=~/.agentenv/fake-homes/<preset>，退出后立即销毁或复用，对宿主终端零副作用。



C. 层级化命名 aenv activate <preset>/<agent>（命名空间模式）
把预设和 Agent 视作层级路径，类似 Git 分支或 Docker 镜像标签：

Bash
# 激活单一 Agent 的预设视口
aenv activate work-backend/claude

# 激活整个 work-backend 工作区
aenv activate work-backend


4.3 预设维护命令
aenv preset list：查看本地已配置的预设与已覆盖的文件清单。

aenv preset add <preset> <agent> <filename>：快速将宿主当前正在用的配置文件抓取并放入 Preset 作为 Override 起点。

aenv clean：清理所有 ~/.agentenv/runtimes/ 下生成的影子目录缓存。
