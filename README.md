# GitAI - AI-Native Version Control System

> Traditional Review: Check code security and efficiency
> Future Review: Check prompt intent and token costs

**Git Compatible · Lightweight CLI · Pure Local · 7-Day Vibe Coding MVP**

---

## 🔥 One-Line Pitch

GitAI is an **AI-native version control system** that builds on top of traditional Git to directly manage:

- Prompt versions
- Token consumption
- Model calling costs
- New KPIs for AI engineers

Transform PR reviews from "code review" to "intent review, cost review, ROI review".

---

## ✨ Core Features

- ✅ Fully compatible with existing Git repositories - non-invasive, no migration needed
- ✅ Structured management and version comparison of `.prompt` files
- ✅ **Pure local token calculation** - no network, no API calls
- ✅ `gitai diff`: Compare prompt intent + token changes
- ✅ `gitai cost`: Statistics by file / branch / developer (USD/CNY)
- ✅ `gitai review`: Automatically check prompt clarity and redundancy
- ✅ Multi-model support: OpenAI / Qwen / Wenxin

---

## 🚀 Quick Start

```bash
# 1. Initialize in any Git repository
gitai init

# 2. Start managing your prompts
gitai track prompts/agent.prompt

# 3. View version and token changes
gitai diff
gitai log
gitai cost --user
gitai cost --branch
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

## 💡 Why GitAI?

### The Problem

In the AI era, software development has fundamentally changed:

| Traditional Era | AI Era |
|----------------|--------|
| Code = Logic | Code = Prompts |
| Review = Security/Performance | Review = Intent/Cost |
| KPI = Lines of Code | KPI = Token Efficiency |
| Version Control = Git | Version Control = ??? |

Git manages code changes perfectly, but **prompt files need specialized version control**:

- **Intent tracking**: What was the AI assistant asked to do?
- **Token accounting**: How much did this change cost?
- **Cost optimization**: Which prompts are most expensive?
- **ROI analysis**: Is prompt complexity worth the token cost?

### The Solution

GitAI lives **alongside Git** as a lightweight overlay:

```
your-project/
├── .git/              # Traditional Git (code)
├── .gitai/            # GitAI (prompts, tokens, costs)
├── src/
└── prompts/
    ├── agent.prompt
    └── reviewer.prompt
```

**No database. No web interface. Pure CLI. Zero network dependency.**

---

## 📖 Usage

### Initialize GitAI

```bash
$ gitai init
✓ Initialized .gitai/ in /path/to/repo
  Repository: /path/to/repo
  Branch: main
  Head: a1b2c3d4e5f6
```

### Track Prompt Files

Create a `.prompt` file with structured sections:

```markdown
# role: Senior Go Developer

You are an experienced Go developer who writes clean, efficient code.

## Task:
Help users with Go programming tasks including:
- Writing idiomatic Go code
- Debugging and optimization
- Code reviews and best practices

## Constraints:
- Always use Go 1.21+ features
- Follow Go best practices and idioms
- Include proper error handling

## Examples:
Example 1: Creating a REST API server
Example 2: Writing concurrent goroutines

## Input:
The user's Go programming question.

## Output:
Clear, well-explained Go code with comments.
```

Track it:

```bash
$ gitai track prompts/developer.prompt
✓ Tracked: prompts/developer.prompt
  Role: Senior Go Developer
  Total Tokens: 126
  Hash: 337f78512cf1
  Sections:
    task: 38 tokens
    constraints: 32 tokens
    examples: 28 tokens
    input: 12 tokens
    output: 16 tokens
```

### View Tracked Files

```bash
$ gitai status
File              Role                 Tokens  Hash
----              ----                 ------  ----
developer.prompt  Senior Go Developer  126     337f7851
coder.prompt      Full-Stack Developer 103     df636864

Total: 2 files, 229 tokens

Token breakdown:
  Task: 72 tokens
  Constraints: 55 tokens
  Examples: 55 tokens
```

### Show Detailed Information

```bash
$ gitai show developer.prompt
File: prompts/developer.prompt
  Role: Senior Go Developer
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
    Help users with Go programming tasks including:
    - Writing idiomatic Go code
    - Debugging and optimization
    ...
```

---

## 🗺️ Development Roadmap

### Phase 1: Foundation (Day 1-2) ✅
- [x] Project skeleton + Git integration
- [x] `.gitai/` directory storage
- [x] Track `.prompt` files
- [x] Local token estimation (Chinese/English mixed)

### Phase 2: Version Control (Day 3-4)
- [ ] `gitai diff`: Compare prompt versions (intent + tokens)
- [ ] `gitai log`: Show commit history with token deltas
- [ ] `gitai cost`: Calculate costs by file/branch/user
- [ ] Multi-currency support (USD/CNY)

### Phase 3: Configuration (Day 5-6)
- [ ] `.gitai.toml` configuration file
- [ ] Model switching (OpenAI/Qwen/Wenxin)
- [ ] Custom pricing per model
- [ ] Exchange rate configuration

### Phase 4: Optimization (Day 7)
- [ ] `gitai review`: Analyze prompt quality
- [ ] Token efficiency suggestions
- [ ] Redundancy detection
- [ ] Clarity scoring

---

## 🏗️ Architecture

```
gitai/
├── cmd/gitai/           # CLI entry point
│   └── main.go
├── internal/gitai/      # Core commands
│   ├── git.go           # Git integration (go-git)
│   ├── storage.go       # .gitai/ storage
│   ├── init.go          # init command
│   ├── track.go         # track command
│   ├── status.go        # status command
│   └── diff.go          # diff command (TODO)
├── pkg/token/           # Token calculation
│   ├── estimator.go     # Pure Go token estimation
│   └── parser.go        # Prompt file parser
└── go.mod
```

### Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.21+ |
| Git Integration | [go-git](https://github.com/go-git/go-git) |
| CLI Framework | [urfave/cli/v2](https://github.com/urfave/cli) |
| Storage | JSON (`.gitai/tracked.json`) |
| Configuration | TOML (`.gitai/config.toml`) |

### Key Design Decisions

1. **No database**: JSON file storage - simple, portable, git-trackedable
2. **No web UI**: Pure CLI - fast, scriptable, developer-friendly
3. **No network**: Pure local - private, fast, zero API costs
4. **Git compatible**: Parasitic design - works with any existing repo
5. **Standard library focused**: Minimal dependencies - easy to maintain

---

## 📊 Token Estimation

GitAI uses pure Go implementation for token estimation:

| Character Type | Token Ratio |
|----------------|-------------|
| Chinese (CJK) | ~0.7 tokens/char |
| English letters | ~0.25 tokens/char |
| Numbers | ~0.3 tokens/char |
| Punctuation | ~0.5 tokens/char |

**Why local estimation?**
- Zero network latency
- Zero API costs
- Works offline
- Privacy-preserving
- Sufficient for cost planning (within ~10% margin)

---

## 🤝 Contributing

GitAI is in active development (7-day Vibe Coding challenge).

Contributions welcome! Please feel free to submit issues and pull requests.

### Development Setup

```bash
git clone https://github.com/YuanyuanMa03/gitai.git
cd gitai
go mod tidy
go build -o gitai ./cmd/gitai
./gitai --help
```

### Running Tests

```bash
# Unit tests (TODO)
go test ./...

# Integration test
cd /tmp && mkdir test-repo && cd test-repo
git init
echo "test" > README.md && git add . && git commit -m "init"
/path/to/gitai init
```

---

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- [go-git](https://github.com/go-git/go-git) - Git pure Go implementation
- [urfave/cli](https://github.com/urfave/cli) - Command-line interface framework

---

## 📮 Contact

- GitHub: [@YuanyuanMa03](https://github.com/YuanyuanMa03)
- Issues: [GitHub Issues](https://github.com/YuanyuanMa03/gitai/issues)

---

**Built with ❤️ in 7 days via Vibe Coding**
