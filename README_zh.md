# AgentEnv (`aenv`)

<p align="left">
  <b><a href="README.md">English</a></b> | <b>简体中文</b>
</p>

**Stop juggling config files. Switch AI presets in seconds.**  
> **面向终端 AI 编码助手（Claude Code、Codex CLI、Gemini 等）的通用预设隔离与环境管理器。**  
> 一秒切换「工作 / 个人 / 开源」配置，多工具统一管理，终端零残留、零污染。

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-4D4D4D)](https://github.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 为什么需要 aenv？

```bash
# 以前：手动拷文件夹、改 settings.json，生怕把全局 Token 和历史搞丢
cp -r ~/.claude ~/.claude.work.bak   # 😰 切换麻烦，多窗口并发时极易互相踩踏

# 现在：一条命令进入隔离预设，用完 exit 瞬间复原
aenv shell work                      # ✅ 零侵入、零污染、全自动穿透
```

### 痛点对比

| 痛点 | `aenv` 的解决方式 |
| :--- | :--- |
| **多工具配置各管各的** | 统一管理 **Claude Code**, **Codex**, **Gemini**, **OpenCode** 等多个 Agent，一个预设即可联动所有工具。 |
| **覆盖全局配置容易踩踏** | **绝不硬写全局基底目录**。通过瞬时子 Shell 和软链视口隔离，多终端窗口同时跑不同预设互不干扰。 |
| **每次都要重复配置一套** | **有则覆盖，无则穿透**。预设里只放你想修改的文件（如 MCP 或规约），历史记录和登录态 100% 自动穿透共用。 |
| **Shell 钩子容易搞坏终端** | **零终端钩子 (Zero-Hook)**。不修改 `.bashrc` 或 PowerShell `$PROFILE`，退出即还原，对宿主环境零残留。 |

---

## 快速上手 (1 分钟)

### 1. 安装 (Go 1.22+)

```bash
# 直接安装
go install ./cmd/aenv

# 或克隆源码本地编译
git clone https://github.com/your-username/AgentEnv.git
cd AgentEnv && go build -o bin/aenv ./cmd/aenv
```

> **Windows 用户提示**：Windows 原生软链接需要启用系统**开发者模式 (Developer Mode)**。以管理员身份在 PowerShell 执行一次即可：
> ```powershell
> reg add "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock" /t REG_DWORD /f /v "AllowDevelopmentWithoutDevLicense" /d "1"
> ```
> 可随时运行 `aenv doctor` 检查环境就绪状态。

### 2. 三步日常工作流

```bash
# 1. 创建一个预设（例如名为 work）
aenv preset create work

# 2. 放入你要覆盖的配置文件
# 例如想单独给 Claude 换配置，把你的 settings.json 丢进：
# ~/.agentenv/presets/work/claude/settings.json

# 3. 激活预设进入终端
aenv shell work
# 终端提示符变成: [(aenv:work)]>
claude   # 此时 Claude 自动读取工作配置！

# 4. 用完打 exit 原样退出
exit
```

---

## 常用命令手册 (Usage)

### 核心命令

| 命令 | 说明与示例 |
| :--- | :--- |
| `aenv shell <preset>[/<agent>]` | **启动预设子终端**（别名 `activate`）。退出输入 `exit`。<br>• `aenv shell work`（多个 Agent 时自动唤起方向键多选菜单）<br>• `aenv shell work/claude`（直接指定单一 Agent） |
| `aenv run <preset>[/<agent>] -- <cmd>` | **单次执行命令**，不常驻子终端，100% 透明透传子进程退出码。<br>• `aenv run work/claude -- claude --model sonnet`<br>• `aenv run work/claude --dry-run -- claude`（仅预览环境变量与执行指令） |
| `aenv env <preset>[/<agent>]` | **只读输出环境变量**导出语句（供外部脚本调用，自动适配当前 Shell）。<br>• `aenv env work/claude` |

### 预设管理 (`aenv preset`)

| 命令 | 说明 |
| :--- | :--- |
| `aenv preset create <name>` | 创建新预设骨架（自动生成 `shared/` 以及各已注册 Agent 的专属文件夹） |
| `aenv preset list` | 查看所有预设及每个预设下正在生效的覆写文件列表 |
| `aenv preset diff <preset>` | 彩色比对预设内的文件与原生基底文件的差异（改动一目了然） |

### 系统维护

| 命令 | 说明 |
| :--- | :--- |
| `aenv doctor` | 一键自检系统权限、开发者模式、父 Shell，并**自动同步新增 Agent 到各已有预设** |
| `aenv status` | 查看当前子终端所处的预设名称与生效环境变量（在子 Shell 内使用） |
| `aenv clean` | 清除生成的运行时影子缓存（纯软链，随时可安全清理） |
| `aenv version` | 输出当前 aenv 版本 |

*支持全局选项 `--no-color` 关闭终端彩色高亮。*

---

## 怎么组织预设文件？

所有预设集中在 `~/.agentenv/presets/`（Windows 为 `%USERPROFILE%\.agentenv\presets\`）：

```text
~/.agentenv/presets/work/
├── shared/                   # 预设内所有 Agent 共享的文件（如通用的 .mcp.json, AGENTS.md）
├── claude/                   # Claude 专属覆写（如 settings.json, CLAUDE.md）
└── codex/                    # Codex 专属覆写（如 config.toml）
```

**简单规则，无需学习中间配置语法**：
1. **有则覆写**：你在 `claude/` 放了 `settings.json`，在预设环境里就会替换掉默认配置；
2. **无则穿透**：没放的文件（例如 `history.jsonl`、登录 token 等）全自动穿透使用宿主基底，无需重新登录；
3. **完全同构**：官方工具在 `~/.claude/` 怎么放，你在 `presets/work/claude/` 就怎么放。

---

## 扩展支持其他 Agent (`config.toml`)

默认内置支持 `claude`, `codex`, `gemini`, `opencode`。  
如果你想支持任何其他终端 AI 工具，在 `~/.agentenv/config.toml` 添加几行即可：

```toml
[agents.myagent]
env_var = "MYAGENT_CONFIG_DIR"        # 该工具读取配置目录的环境变量名
host_dir = "~/.myagent"               # 该工具默认的宿主配置目录
```

添加后运行一次 `aenv doctor`，它会自动为所有预设同步建好 `myagent/` 文件夹！

---

## 退出码规范 (Exit Codes)

| 退出码 | 说明 |
| :---: | :--- |
| `0` | 正常执行完毕 |
| `1` | 业务错误（如指定的预设或 Agent 不存在） |
| `2` | 命令行参数或语法错误 |
| `3` | 系统环境错误（如 Windows 未开启开发者模式） |
| `N` | `aenv run` 时直接透明返回目标子进程自身的退出码 |

---

## 致敬与灵感来源 (Acknowledgments)

本项目的设计灵感来源于优秀的开源项目 [claudectx](https://github.com/foxj77/claudectx)。

`claudectx` 最早探索了 Claude Code 多配置切换的便利体验。在此基础上，`aenv` 进一步演进并重构了运行模型：
- **从单工具到多 Agent**：不止服务 Claude，更以通用基础设施形态原生支持 Claude Code、Codex CLI、Gemini CLI 等任意终端 AI 助手；
- **从全局硬写到零钩子子终端**：不直接修改全局 `~/.claude` 或篡改 CLI 参数，而是通过**瞬时影子视口与进程级子 Shell**，实现真正的零残留、零踩踏与多端并发安全；
- **从整包复制到差量穿透**：引入智能分层继承与软链农场，只需维护几行差异文件，告别全盘文件冗余。

感谢 `claudectx` 为社区带来的启发！

---

## 许可证 (License)

[MIT License](LICENSE)
