# GitAI —— AI 时代的下一代版本控制系统

> 古早 Review：查代码安全与效率
> 未来 Review：查 Prompt 意图与 Token 成本

**Git 兼容 · 轻量 CLI · 纯本地运行 · 单人 7 天 Vibe Coding 落地**

---

## 🔥 一句话定位

GitAI 是**AI 原生版本控制系统**，在传统 Git 之上，直接管理：

- Prompt 版本
- Token 消耗
- 模型调用成本
- AI 工程师新 KPI

让 PR 从「审代码」变成「审意图、审成本、审ROI」。

---

## ✨ 核心能力

- ✅ 完全兼容现有 Git 仓库，不入侵、不迁移
- ✅ `.prompt` 文件结构化管理与版本对比
- ✅ **纯本地 Token 计算**，不联网、不调接口
- ✅ `gitai diff`：对比 Prompt 意图 + Token 变化
- ✅ `gitai cost`：按文件 / 分支 / 开发者统计成本（USD/CNY）
- ✅ `gitai review`：自动检查 Prompt 清晰度与冗余
- ✅ 支持多模型：OpenAI / Qwen / Wenxin

---

## 🚀 快速开始

```bash
# 1. 任意 Git 仓库内初始化
gitai init

# 2. 开始管理你的 Prompt
gitai track prompts/agent.prompt

# 3. 查看版本与 Token 变化
gitai diff
gitai log
gitai cost --user
gitai cost --branch
```

---

## 📦 安装

### 从源码编译

```bash
git clone https://github.com/YuanyuanMa03/gitai.git
cd gitai
go build -o gitai ./cmd/gitai
sudo mv gitai /usr/local/bin/
```

### 使用 Go Install

```bash
go install github.com/YuanyuanMa03/gitai/cmd/gitai@latest
```

---

## 💡 为什么需要 GitAI？

### 问题背景

AI 时代，软件开发已发生根本性变化：

| 传统时代 | AI 时代 |
|---------|---------|
| 代码 = 逻辑 | 代码 = Prompts |
| Review = 安全/性能 | Review = 意图/成本 |
| KPI = 代码行数 | KPI = Token 效率 |
| 版本控制 = Git | 版本控制 = ？？？ |

Git 完美管理代码变更，但 **Prompt 文件需要专门的版本控制**：

- **意图追踪**：要求 AI 做什么？
- **Token 记账**：这次变更花了多少钱？
- **成本优化**：哪些 Prompt 最贵？
- **ROI 分析**：复杂度是否值得？

### 解决方案

GitAI 作为轻量级覆盖层，**与 Git 并存**：

```
your-project/
├── .git/              # 传统 Git（代码）
├── .gitai/            # GitAI（prompts、tokens、成本）
├── src/
└── prompts/
    ├── agent.prompt
    └── reviewer.prompt
```

**无数据库。无网页界面。纯 CLI。零网络依赖。**

---

## 📖 使用指南

### 初始化 GitAI

```bash
$ gitai init
✓ Initialized .gitai/ in /path/to/repo
  Repository: /path/to/repo
  Branch: main
  Head: a1b2c3d4e5f6
```

### 追踪 Prompt 文件

创建结构化的 `.prompt` 文件：

```markdown
# role: 高级 Go 开发者

你是一名经验丰富的 Go 开发者，擅长编写简洁高效的代码。

## Task:
帮助用户完成 Go 编程任务，包括：
- 编写地道的 Go 代码
- 调试和性能优化
- 代码审查和最佳实践

## Constraints:
- 始终使用 Go 1.21+ 特性
- 遵循 Go 最佳实践和惯用法
- 包含恰当的错误处理

## Examples:
示例 1：创建 REST API 服务器
示例 2：编写并发 goroutine

## Input:
用户的 Go 编程问题。

## Output:
清晰、解释充分的 Go 代码，带有注释。
```

追踪文件：

```bash
$ gitai track prompts/developer.prompt
✓ Tracked: prompts/developer.prompt
  Role: 高级 Go 开发者
  Total Tokens: 126
  Hash: 337f78512cf1
  Sections:
    task: 38 tokens
    constraints: 32 tokens
    examples: 28 tokens
    input: 12 tokens
    output: 16 tokens
```

### 查看已追踪文件

```bash
$ gitai status
File              Role                 Tokens  Hash
----              ----                 ------  ----
developer.prompt  高级 Go 开发者        126     337f7851
coder.prompt      全栈开发者            103     df636864

Total: 2 files, 229 tokens

Token breakdown:
  Task: 72 tokens
  Constraints: 55 tokens
  Examples: 55 tokens
```

### 显示详细信息

```bash
$ gitai show developer.prompt
File: prompts/developer.prompt
  Role: 高级 Go 开发者
  Total Tokens: 126
  Added: 2026-03-01 02:09:11
  Modified: 2026-03-01 02:09:10
  Hash: 337f78512cf1...

  Token Breakdown:
    Task: 38 tokens
    Constraints: 32 tokens
    Examples: 28 tokens
    Input: 12 tokens
    Output: 16 tokens

  Task:
    帮助用户完成 Go 编程任务，包括：
    - 编写地道的 Go 代码
    - 调试和性能优化
    ...
```

---

## 🗺️ 开发路线图

### 阶段 1：基础建设（Day 1-2）✅
- [x] 项目骨架 + Git 集成
- [x] `.gitai/` 目录存储
- [x] 追踪 `.prompt` 文件
- [x] 本地 Token 估算（中英混合）

### 阶段 2：版本控制（Day 3-4）
- [ ] `gitai diff`：对比 Prompt 版本（意图 + Token）
- [ ] `gitai log`：显示提交历史与 Token 增量
- [ ] `gitai cost`：按文件/分支/用户计算成本
- [ ] 多币种支持（USD/CNY）

### 阶段 3：配置管理（Day 5-6）
- [ ] `.gitai.toml` 配置文件
- [ ] 模型切换（OpenAI/Qwen/Wenxin）
- [ ] 自定义单价
- [ ] 汇率配置

### 阶段 4：智能优化（Day 7）
- [ ] `gitai review`：分析 Prompt 质量
- [ ] Token 效率建议
- [ ] 冗余检测
- [ ] 清晰度评分

---

## 🏗️ 项目结构

```
gitai/
├── cmd/gitai/           # CLI 入口
│   └── main.go
├── internal/gitai/      # 核心命令
│   ├── git.go           # Git 集成（go-git）
│   ├── storage.go       # .gitai/ 存储
│   ├── init.go          # init 命令
│   ├── track.go         # track 命令
│   ├── status.go        # status 命令
│   └── diff.go          # diff 命令（TODO）
├── pkg/token/           # Token 计算
│   ├── estimator.go     # 纯 Go Token 估算
│   └── parser.go        # Prompt 文件解析
└── go.mod
```

### 技术栈

| 组件 | 技术选型 |
|------|---------|
| 语言 | Go 1.21+ |
| Git 集成 | [go-git](https://github.com/go-git/go-git) |
| CLI 框架 | [urfave/cli/v2](https://github.com/urfave/cli) |
| 存储 | JSON（`.gitai/tracked.json`）|
| 配置 | TOML（`.gitai/config.toml`）|

### 核心设计决策

1. **无数据库**：JSON 文件存储 —— 简单、可移植、可被 Git 追踪
2. **无 Web UI**：纯 CLI —— 快速、可脚本化、开发者友好
3. **无网络**：纯本地运行 —— 私密、快速、零 API 成本
4. **Git 兼容**：寄生式设计 —— 适用于任何现有仓库
5. **标准库优先**：最少依赖 —— 易于维护

---

## 📊 Token 估算算法

GitAI 使用纯 Go 实现 Token 估算：

| 字符类型 | Token 比率 |
|----------|-----------|
| 中文（CJK）| ~0.7 tokens/字 |
| 英文字母 | ~0.25 tokens/字符 |
| 数字 | ~0.3 tokens/字符 |
| 标点符号 | ~0.5 tokens/字符 |

**为什么使用本地估算？**
- 零网络延迟
- 零 API 成本
- 离线可用
- 保护隐私
- 成本规划精度足够（~10% 误差内）

---

## 🤝 贡献指南

GitAI 正在积极开发中（7 天 Vibe Coding 挑战）。

欢迎贡献！请随时提交 Issue 和 Pull Request。

### 开发环境搭建

```bash
git clone https://github.com/YuanyuanMa03/gitai.git
cd gitai
go mod tidy
go build -o gitai ./cmd/gitai
./gitai --help
```

### 运行测试

```bash
# 单元测试（TODO）
go test ./...

# 集成测试
cd /tmp && mkdir test-repo && cd test-repo
git init
echo "test" > README.md && git add . && git commit -m "init"
/path/to/gitai init
```

---

## 📝 开源协议

MIT License - 详见 [LICENSE](LICENSE)

---

## 🙏 致谢

- [go-git](https://github.com/go-git/go-git) - Git 纯 Go 实现
- [urfave/cli](https://github.com/urfave/cli) - 命令行界面框架

---

## 📮 联系方式

- GitHub: [@YuanyuanMa03](https://github.com/YuanyuanMa03)
- Issues: [GitHub Issues](https://github.com/YuanyuanMa03/gitai/issues)

---

**7 天 Vibe Coding 之爱与激情构建 ❤️**
