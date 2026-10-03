# AgentEnv (`aenv`)

> **给终端 AI 编码助手（Claude Code、Codex CLI 等）的多环境/预设隔离工具。**  
> 一秒切换「工作 / 个人 / 开源」预设，配置互不干扰，终端干净无污染。

---

## 它解决什么痛点？

在终端频繁使用 Claude Code、Codex 等 AI 编码工具时，经常会遇到这些麻烦：
- **MCP 与规则冲突**：公司项目有套私有 MCP 工具，个人项目有另一套，混在一起容易冲突、报错或污染上下文。
- **临时修改不敢存**：想临时换个 API Key、切个模型或加条专用规约，改了全局文件又容易忘记还原。
- **切换繁琐且容易踩踏**：手动备份复制 `~/.claude` 费时费力，多窗口并发时还容易互相覆盖丢会话。

**`aenv` 让你像管理 Python 虚拟环境一样管理 Agent 配置**：
- **进出随心**：`aenv shell work` 瞬间进入工作预设，提示符变为 `[(aenv:work)]>`；退出打 `exit`，终端恢复原状，不留任何痕迹。
- **单次运行**：`aenv run work -- claude` 跑完即走，适合单次任务或自动化脚本。
- **无需维护整套配置**：预设只放你想覆盖的差量文件（如 `settings.json`），其他文件（登录态、历史记录）全自动穿透共用，不用重复登录！

---

## 快速上手 (1 分钟)

### 1. 安装 (Go 1.22+)

```bash
# 直接安装
go install ./cmd/aenv

# 或克隆本地编译
git clone https://github.com/your-username/AgentEnv.git
cd AgentEnv && go build -o bin/aenv ./cmd/aenv
```

> **Windows 用户提示**：Windows 软链接需开启系统开发者模式。以管理员身份在 PowerShell 执行一次即可：
> ```powershell
> reg add "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock" /t REG_DWORD /f /v "AllowDevelopmentWithoutDevLicense" /d "1"
> ```
> 你也可以随时运行 `aenv doctor` 进行环境自检。

### 2. 日常工作流

```bash
# 1. 创建一个预设（例如名为 work）
aenv preset create work

# 2. 放入你要覆盖的配置文件
# 比如要给 Claude 单独换配置，把你的 settings.json 放进：
# ~/.agentenv/presets/work/claude/settings.json

# 3. 激活预设进入终端
aenv shell work
# 终端提示符变成: [(aenv:work)]>
claude   # 此时 Claude 自动读取工作配置！

# 4. 退出环境
exit
```

---

## 常用命令手册 (Usage)

### 核心命令

| 命令 | 说明与示例 |
| :--- | :--- |
| `aenv shell <preset>[/<agent>]` | **启动预设子终端**（别名 `activate`）。退出打 `exit`。<br>• `aenv shell work`（多个 Agent 时弹出方向键多选菜单）<br>• `aenv shell work/claude`（直接指定单一 Agent） |
| `aenv run <preset>[/<agent>] -- <cmd>` | **单次执行命令**，不进入子终端，运行完即销毁，100% 透传退出码。<br>• `aenv run work/claude -- claude --model sonnet`<br>• `aenv run work/claude --dry-run -- claude`（仅预览变量，不执行） |
| `aenv env <preset>[/<agent>]` | **只读输出环境变量**导出语句（供自定义脚本使用，自动适配当前 Shell）。<br>• `aenv env work/claude` |

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

*支持全局选项 `--no-color` 关闭彩色高亮。*

---

## 怎么组织预设文件？

所有预设集中存放在 `~/.agentenv/presets/`（Windows 为 `%USERPROFILE%\.agentenv\presets\`）：

```text
~/.agentenv/presets/work/
├── shared/                   # 预设内跨 Agent 共享文件（如统一的 .mcp.json, AGENTS.md）
├── claude/                   # Claude 专属覆写（如 settings.json, CLAUDE.md）
└── codex/                    # Codex 专属覆写（如 config.toml）
```

**简单规则，无需记忆多余语法**：
1. **有则覆写**：你在 `claude/` 放了 `settings.json`，在预设环境里就会替换掉默认配置；
2. **无则穿透**：没放的文件（例如 `history.jsonl`、登录 token、缓存等）全自动穿透使用宿主基底，无需反复登录；
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

## 退出码规范

| 退出码 | 说明 |
| :---: | :--- |
| `0` | 正常执行完毕 |
| `1` | 业务错误（如指定的预设或 Agent 不存在） |
| `2` | 命令行参数或语法错误 |
| `3` | 系统环境错误（如 Windows 未开启开发者模式） |
| `N` | `aenv run` 时直接透明返回目标子进程自身的退出码 |

---

## 许可证 (License)

[MIT License](LICENSE)
