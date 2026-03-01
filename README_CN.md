# GitAI —— AI 时代的下一代版本控制系统

> **传统代码审查**：检查代码安全与效率
> **AI 时代代码审查**：检查 Prompt 意图、Token 成本与 ROI

**Git 兼容 · 轻量 CLI · 纯本地 · 100% Go**

---

## 🎯 设计理念

### AI 原生开发的范式转变

AI 时代让软件开发发生了根本性变化：

| 维度 | 传统时代 | AI 时代 |
|------|---------|--------|
| **编写内容** | 代码 | Prompts |
| **审查重点** | 安全性、性能 | 意图、成本、效果 |
| **考核指标** | 代码行数 | Token 效率、ROI |
| **版本控制** | Git 管理代码 | 谁来管理 Prompts？ |

**GitAI** 填补了这一空白 —— 它是 **AI 原生版本控制系统**，将 **Prompts 提升为代码之外的一等公民**。

### 核心原则

1. **意图作为第一手数据**
   - 每次提交记录：*为什么*要做、*用了* 哪些 prompts、*花了* 多少成本
   - 将模糊的代码提交转化为可追踪的意图链

2. **Prompts 是代码，但有所不同**
   - Prompts 有版本、依赖关系和质量指标
   - Token 消耗是真实成本，需要追踪
   - Prompt 的变更可能对代码库产生连锁反应

3. **双向可追溯性**
   - 正向：意图 → Prompts → 代码 → 行为
   - 反向：代码 → Prompts → 意图 → 决策
   - AI 开发闭环中，没有任何信息丢失

4. **持续学习**
   - 追踪什么有效、什么无效
   - 从反馈中构建组织知识
   - 基于真实结果优化 Prompts

---

## ✨ 核心功能

### 📋 Prompt 版本控制
- 与代码一起追踪 `.prompt` 文件
- 查看版本间的 Token 变化（`gitai diff`）
- 查看提交历史与 Token 增量（`gitai log`）
- 结构化 Prompt 格式：role、task、constraints、examples、input、output

### 💰 成本智能
- 按文件/章节/角色/用户/分支计算成本
- 多模型支持：OpenAI、Qwen、文心、自定义模型
- 货币转换：USD/CNY 可配置汇率
- 可视化成本趋势（`gitai trend`）

### 🔍 影响分析
- 查看修改 Prompt 的影响范围（`gitai impact <prompt>`）
- 自动检测 Prompt→代码依赖关系（`gitai scan-deps`）
- 追踪哪些 Prompts 影响哪些文件
- 查找共享受影响文件的 Prompts

### 🧪 质量保证
- 自动化 Prompt 质量评分（清晰度、简洁性、效率）
- Pre-commit hooks 拦截低质量 Prompts
- 改进建议
- 质量仪表板（`gitai review`、`gitai quality`）

### 🎯 意图追踪
- 每次提交记录开发者意图
- 自动分类检测（fix/feature/refactor/ai-gen/experimental）
- 将提交与生成它们的 Prompts 关联
- 完整意图历史（`gitai intent-log`）

### 📚 组织管理
- **模板**：创建可复用的 Prompt 模板（`gitai template`）
- **标签**：用自定义标签组织 Prompts（`gitai tag`）
- **搜索**：跨所有 Prompts 全文搜索
- **导出/导入**：备份和恢复 Prompt 库

### 🧠 反馈闭环
- 记录结果：good/bad/neutral 反馈（`gitai feedback`）
- 追踪随时间变化的满意度指标
- 从结果中学习优化 Prompts
- AI 生成的洞察（`gitai insights`）

### 🔗 完全可追溯性
- 追踪意图 → Prompts → 受影响文件（`gitai trace <intent>`）
- 追踪文件 → Prompts → 意图（`gitai trace <file>`）
- 从想法到实现的完整链路
- 理解变更的下游影响

### ⚙️ 自动化
- Git hooks 自动质量检查
- Pre-commit：提交前验证 Prompt 质量
- Post-commit：自动记录意图
- 零手动开销 - 自动工作

---

## 🚀 快速开始

```bash
# 1. 在 Git 仓库中初始化
git init
gitai init

# 2. 追踪你的第一个 Prompt
gitai track prompts/agent.prompt

# 3. 查看变化
gitai diff                    # Token 变化
gitai log                     # 提交历史
gitai cost --user             # 按开发者统计成本
gitai impact prompts/agent.prompt  # 影响分析

# 4. 从数据中学习
gitai insights                # AI 生成的洞察
gitai trace <intent-id>       # 完整追踪链路
```

---

## 📖 命令参考

### 入门命令

| 命令 | 描述 |
|------|------|
| `gitai init` | 在当前仓库初始化 `.gitai/` |
| `gitai track <file>` | 追踪 `.prompt` 文件 |
| `gitai status` | 显示已追踪文件状态 |
| `gitai show <file>` | 显示文件详细信息 |

### 版本控制

| 命令 | 描述 |
|------|------|
| `gitai diff <file>` | 对比 Prompt 版本与 Token 变化 |
| `gitai log` | 显示提交历史与 Token 增量 |
| `gitai commit` | 创建带意图追踪的提交 |

### 成本分析

| 命令 | 描述 |
|------|------|
| `gitai cost` | 显示成本明细（按文件/章节/角色） |
| `gitai cost --user` | 按开发者统计成本 |
| `gitai cost --branch` | 按分支统计成本 |
| `gitai price` | 列出模型定价 |
| `gitai stats` | 综合统计仪表板 |
| `gitai trend` | Token 增长趋势 |

### 质量与审查

| 命令 | 描述 |
|------|------|
| `gitai review` | 质量分析与改进建议 |
| `gitai quality` | 快速质量概览 |
| `gitai hooks install` | 安装自动质量检查 |

### 组织管理

| 命令 | 描述 |
|------|------|
| `gitai template create` | 创建 Prompt 模板 |
| `gitai template use` | 使用模板创建文件 |
| `gitai tag add <file> <tag>` | 为文件添加标签 |
| `gitai search <query>` | 搜索 Prompts |
| `gitai export` | 导出为 JSON/CSV |
| `gitai import` | 从导出文件导入 |

### AI 时代功能

| 命令 | 描述 |
|------|------|
| `gitai impact <prompt>` | 分析修改 Prompt 的影响 |
| `gitai scan-deps` | 自动检测 Prompt 依赖 |
| `gitai show-deps` | 显示依赖图 |
| `gitai feedback [good\|bad]` | 记录提交反馈 |
| `gitai insights` | 显示 AI 生成的洞察 |
| `gitai trace <id\|file>` | 追踪意图↔代码关系 |

---

## 💡 使用场景

### 1. 了解你的 AI 支出

```bash
$ gitai cost --user
Cost breakdown by user (model: gpt-4o)

User          Commits  Files  Tokens    USD      CNY
----          -------  -----  ------    ---      ---
alice         15       23     45,230    $0.12    ¥0.86
bob           8        12     12,400    $0.03    ¥0.22

TOTAL         23       35     57,630    $0.15    ¥1.08
```

### 2. 从想法追溯实现

```bash
$ gitai trace intent-abc123456
🔍 Trace: intent-abc123456

Commit: abc1234
Message: feat: add user authentication

📝 Prompts Used (2):
  → prompts/auth-api.prompt
  → prompts/security-check.prompt

📁 Affected Files (5):
  → api/auth.go
  → api/middleware.go
  → models/user.go
  → utils/crypto.go
  → tests/auth_test.go
```

### 3. 变更前的影响分析

```bash
$ gitai impact prompts/api-design.prompt
📊 Impact Analysis: prompts/api-design.prompt

Affected Files: 3 files
  → api/handlers.go
  → api/routes.go
  → models/types.go

⚠️  Related Prompts (2)
These prompts share affected files:
  • prompts/api-docs.prompt
  • prompts/api-tests.prompt
```

### 4. 从数据中学习

```bash
$ gitai insights
💡 GitAI Insights

Total Commits Tracked:  47
Total Intents:           42
Total Feedback:          15
Average Score:           78.5/100

📊 Activity by Category:
  feature: 18
  fix: 12
  refactor: 8
  ai-gen: 4

✨ Key Learnings:
  • Most active category: feature (18 intents)
  • High satisfaction rate - prompts are working well
  • Consider standardizing common patterns

🔝 Most Used Prompts:
  1. prompts/api-helper.prompt (12 uses)
  2. prompts/code-reviewer.prompt (8 uses)
  3. prompts/test-generator.prompt (5 uses)
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

## 🏗️ 项目架构

```
gitai/
├── cmd/gitai/              # CLI 入口
├── internal/gitai/         # 核心命令
│   ├── git.go              # Git 集成
│   ├── storage.go          # 存储层
│   ├── intent.go           # 意图追踪
│   ├── feedback.go          # 反馈系统
│   ├── dependency.go       # 依赖图
│   ├── hooks.go            # Git hooks
│   ├── insights.go         # 分析与追踪
│   └── ...                 # 其他命令
├── pkg/token/              # Token 计算
│   ├── estimator.go        # 纯 Go 估算
│   └── parser.go           # Prompt 解析器
└── go.mod
```

### 技术栈

| 组件 | 技术选型 |
|------|---------|
| 语言 | Go 1.21+ |
| Git 集成 | [go-git](https://github.com/go-git/go-git) |
| CLI 框架 | [urfave/cli/v2](https://github.com/urfave/cli) |
| 存储 | JSON（`.gitai/*.json`）|
| 配置 | TOML（`.gitai/config.toml`）|

### 设计原则

1. **无数据库** - JSON 文件存储，简单可移植
2. **无 Web UI** - 纯 CLI，快速可脚本化
3. **无网络** - 100% 本地，私密快速
4. **Git 兼容** - 适用于任何现有 Git 仓库
5. **最少依赖** - 易于维护和审计

---

## 📊 Token 估算算法

GitAI 使用纯 Go 实现进行精确的 Token 估算：

| 字符类型 | Token 比率 |
|----------|-----------|
| 中文（CJK）| ~0.7 tokens/字 |
| 英文字母 | ~0.25 tokens/字符 |
| 数字 | ~0.3 tokens/字符 |
| 标点符号 | ~0.5 tokens/字符 |

**为什么使用本地估算？**
- ✅ 零网络延迟
- ✅ 零 API 成本
- ✅ 离线可用
- ✅ 保护隐私
- ✅ 成本规划精度足够（~10% 误差）

---

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行带覆盖率的测试
go test -cover ./...

# 运行特定测试
go test -v ./internal/gitai -run TestIntent
```

GitAI 采用 **测试驱动开发（TDD）** 构建 —— 所有功能都有全面的测试覆盖。

---

## 📝 开源协议

MIT License - 详见 [LICENSE](LICENSE)

---

## 🤝 贡献指南

欢迎贡献！请随时提交 Issue 和 Pull Request。

### 开发环境搭建

```bash
git clone https://github.com/YuanyuanMa03/gitai.git
cd gitai
go mod tidy
go build -o gitai ./cmd/gitai
./gitai --help
```

---

## 📮 联系方式

- GitHub: [@YuanyuanMa03](https://github.com/YuanyuanMa03)
- Issues: [GitHub Issues](https://github.com/YuanyuanMa03/gitai/issues)

---

## 🙏 致谢

- [go-git](https://github.com/go-git/go-git) - Git 纯 Go 实现
- [urfave/cli](https://github.com/urfave/cli) - 命令行界面框架
- [BurntSushi/toml](https://github.com/BurntSushi/toml) - TOML 解析器

---

**GitAI** —— Prompts 与版本控制的相遇之地。
