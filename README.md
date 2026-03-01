# GitAI - AI-Native Version Control System

> **Traditional code review**: Check code security and efficiency
> **AI-era code review**: Check prompt intent, token costs, and ROI

**Git Compatible · Lightweight CLI · Pure Local · 100% Go**

---

## 🎯 Philosophy

### The Shift to AI-Native Development

The AI era has fundamentally changed software development:

| Aspect | Traditional Era | AI Era |
|--------|---------------|---------|
| **What we write** | Code | Prompts |
| **What we review** | Security, Performance | Intent, Cost, Effectiveness |
| **What we measure** | Lines of Code | Token Efficiency, ROI |
| **Version control** | Git manages code | What manages prompts? |

**GitAI** fills this gap - it's an AI-native version control system that treats **prompts as first-class citizens** alongside your code.

### Core Principles

1. **Intent as First-Class Data**
   - Every commit tracks: *why* it was made, *what* prompts were used, *how much* it cost
   - Transform commits from opaque code changes into traceable intent chains

2. **Prompts Are Code, But Different**
   - Prompts have versions, dependencies, and quality metrics
   - Token consumption is a real cost that needs tracking
   - Prompt changes can have rippling effects across codebases

3. **Bidirectional Traceability**
   - Forward: Intent → Prompts → Code → Behavior
   - Backward: Code → Prompts → Intents → Decisions
   - Nothing gets lost in the AI development loop

4. **Continuous Learning**
   - Track what works and what doesn't
   - Build institutional knowledge from feedback
   - Optimize prompts based on real outcomes

---

## ✨ Key Features

### 📋 Prompt Version Control
- Track `.prompt` files alongside code
- View token changes between versions (`gitai diff`)
- See commit history with token deltas (`gitai log`)
- Structured prompt format: role, task, constraints, examples, input, output

### 💰 Cost Intelligence
- Calculate costs per file, section, role, user, or branch
- Multi-model support: OpenAI, Qwen, Wenxin, and custom
- Currency conversion: USD/CNY with configurable exchange rates
- Visualize spending trends over time (`gitai trend`)

### 🔍 Impact Analysis
- See what changes if you modify a prompt (`gitai impact <prompt>`)
- Auto-detect prompt→code dependencies (`gitai scan-deps`)
- Trace which prompts affect which files
- Find prompts that share affected files

### 🧪 Quality Assurance
- Automated prompt quality scoring (clarity, conciseness, efficiency)
- Pre-commit hooks to catch low-quality prompts
- Review suggestions for improvement
- Quality dashboard (`gitai review`, `gitai quality`)

### 🎯 Intent Tracking
- Every commit records developer intent
- Automatic category detection (fix/feature/refactor/ai-gen/experimental)
- Link commits to the prompts that generated them
- Full intent history (`gitai intent-log`)

### 📚 Organization
- **Templates**: Create reusable prompt templates (`gitai template`)
- **Tags**: Organize prompts with custom tags (`gitai tag`)
- **Search**: Full-text search across all prompts
- **Export/Import**: Backup and restore prompt library

### 🧠 Feedback Loop
- Record outcomes: good/bad/neutral feedback (`gitai feedback`)
- Track satisfaction metrics over time
- Learn from results to optimize prompts
- AI-generated insights from your data (`gitai insights`)

### 🔗 Full Traceability
- Trace intent → prompts → affected files (`gitai trace <intent>`)
- Trace file → prompts → intents (`gitai trace <file>`)
- See the complete chain from idea to implementation
- Understand downstream effects of changes

### ⚙️ Automation
- Git hooks for automatic quality checks
- Pre-commit: validate prompt quality before committing
- Post-commit: automatically record intent
- No manual overhead - it just works

---

## 🚀 Quick Start

```bash
# 1. Initialize in your Git repository
git init
gitai init

# 2. Track your first prompt
gitai track prompts/agent.prompt

# 3. See what changed
gitai diff                    # Token changes
gitai log                     # Commit history
gitai cost --user             # Cost by developer
gitai impact prompts/agent.prompt  # Impact analysis

# 4. Learn from your data
gitai insights                # AI-generated insights
gitai trace <intent-id>       # Full trace chain
```

---

## 📖 Command Reference

### Getting Started

| Command | Description |
|---------|-------------|
| `gitai init` | Initialize `.gitai/` in current repository |
| `gitai track <file>` | Track a `.prompt` file |
| `gitai status` | Show tracked files status |
| `gitai show <file>` | Show detailed file information |

### Version Control

| Command | Description |
|---------|-------------|
| `gitai diff <file>` | Compare prompt versions with token changes |
| `gitai log` | Show commit history with token deltas |
| `gitai commit` | Create commit with intent tracking |

### Cost Analysis

| Command | Description |
|---------|-------------|
| `gitai cost` | Show cost breakdown (by file/section/role) |
| `gitai cost --user` | Cost breakdown by developer |
| `gitai cost --branch` | Cost breakdown by branch |
| `gitai price` | List model pricing |
| `gitai stats` | Comprehensive statistics dashboard |
| `gitai trend` | Token growth over time |

### Quality & Review

| Command | Description |
|---------|-------------|
| `gitai review` | Quality analysis with improvement suggestions |
| `gitai quality` | Quick quality overview |
| `gitai hooks install` | Install automatic quality checks |

### Organization

| Command | Description |
|---------|-------------|
| `gitai template create` | Create prompt template |
| `gitai template use` | Use template to create file |
| `gitai tag add <file> <tag>` | Add tags to file |
| `gitai search <query>` | Search across prompts |
| `gitai export` | Export to JSON/CSV |
| `gitai import` | Import from export |

### AI-Era Features

| Command | Description |
|---------|-------------|
| `gitai impact <prompt>` | Analyze impact of changing a prompt |
| `gitai scan-deps` | Auto-detect prompt dependencies |
| `gitai show-deps` | Show dependency graph |
| `gitai feedback [good\|bad]` | Record feedback on commits |
| `gitai insights` | Show AI-generated insights |
| `gitai trace <id\|file>` | Trace intent↔code relationships |

---

## 💡 Use Cases

### 1. Understand Your AI Spending

```bash
$ gitai cost --user
Cost breakdown by user (model: gpt-4o)

User          Commits  Files  Tokens    USD      CNY
----          -------  -----  ------    ---      ---
alice         15       23     45,230    $0.12    ¥0.86
bob           8        12     12,400    $0.03    ¥0.22

TOTAL         23       35     57,630    $0.15    ¥1.08
```

### 2. Trace an Idea to Implementation

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

### 3. Analyze Impact Before Changes

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

### 4. Learn From Your Data

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

## 📦 Installation

### From Source

```bash
git clone https://github.com/YuanyuanMa03/gitai.git
cd gitai
go build -o gitai ./cmd/gitai
sudo mv gitai /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/YuanyuanMa03/gitai/cmd/gitai@latest
```

---

## 🏗️ Architecture

```
gitai/
├── cmd/gitai/              # CLI entry point
├── internal/gitai/         # Core commands
│   ├── git.go              # Git integration
│   ├── storage.go          # Storage layer
│   ├── intent.go           # Intent tracking
│   ├── feedback.go          # Feedback system
│   ├── dependency.go       # Dependency graph
│   ├── hooks.go            # Git hooks
│   ├── insights.go         # Analytics & trace
│   └── ...                 # Other commands
├── pkg/token/              # Token calculation
│   ├── estimator.go        # Pure Go estimation
│   └── parser.go           # Prompt parser
└── go.mod
```

### Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.21+ |
| Git Integration | [go-git](https://github.com/go-git/go-git) |
| CLI Framework | [urfave/cli/v2](https://github.com/urfave/cli) |
| Storage | JSON (`.gitai/*.json`) |
| Configuration | TOML (`.gitai/config.toml`) |

### Design Principles

1. **No Database** - JSON file storage, simple and portable
2. **No Web UI** - Pure CLI, fast and scriptable
3. **No Network** - 100% local, private and fast
4. **Git Compatible** - Works with any existing Git repo
5. **Minimal Dependencies** - Easy to maintain and audit

---

## 📊 Token Estimation

GitAI uses pure Go implementation for accurate token estimation:

| Character Type | Token Ratio |
|----------------|-------------|
| Chinese (CJK) | ~0.7 tokens/char |
| English letters | ~0.25 tokens/char |
| Numbers | ~0.3 tokens/char |
| Punctuation | ~0.5 tokens/char |

**Why local estimation?**
- ✅ Zero network latency
- ✅ Zero API costs
- ✅ Works offline
- ✅ Privacy-preserving
- ✅ Sufficient for cost planning (~10% margin)

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/gitai -run TestIntent
```

GitAI is built with **Test-Driven Development (TDD)** - all features have comprehensive test coverage.

---

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🤝 Contributing

Contributions welcome! Please feel free to submit issues and pull requests.

### Development Setup

```bash
git clone https://github.com/YuanyuanMa03/gitai.git
cd gitai
go mod tidy
go build -o gitai ./cmd/gitai
./gitai --help
```

---

## 📮 Contact

- GitHub: [@YuanyuanMa03](https://github.com/YuanyuanMa03)
- Issues: [GitHub Issues](https://github.com/YuanyuanMa03/gitai/issues)

---

## 🙏 Acknowledgments

- [go-git](https://github.com/go-git/go-git) - Git pure Go implementation
- [urfave/cli](https://github.com/urfave/cli) - Command-line interface framework
- [BurntSushi/toml](https://github.com/BurntSushi/toml) - TOML parser for Go

---

**GitAI** - Where prompts meet version control.
