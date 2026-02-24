# 🐿️ Squirrel

> Like a squirrel that forgot where it buried its nuts — find your forgotten Claude Code projects.

Squirrel analyzes your Claude Code history to find projects you started but forgot about.

## 🔧 Install

```bash
go install github.com/dkd-dobberkau/squirrel/cmd/squirrel@latest
```

Via Homebrew (macOS):

```bash
brew install dkd-dobberkau/tap/squirrel
```

Or build from source:

```bash
git clone https://github.com/dkd-dobberkau/squirrel.git
cd squirrel
go build -o squirrel ./cmd/squirrel
cp squirrel /usr/local/bin/
```

## 🚀 Usage

```bash
squirrel                       # Show everything (default: medium depth, 14 days)
squirrel status                # Same as above
squirrel stash                 # Only show open work (uncommitted changes, feature branches)
squirrel timeline              # Chronological activity view
squirrel project <query>       # Detail view for a single project

# Project lookup supports flexible matching:
squirrel project myapp         # Match by short name
squirrel project local/myapp   # Match by path suffix
squirrel project /full/path    # Match by exact path

# Options
squirrel --quick               # Fast: only history + sessions
squirrel --depth=deep          # Deep: includes TODO extraction from session data
squirrel --days 30             # Look back 30 days
squirrel --json                # JSON output for scripting

# Combine options
squirrel project myapp --deep  # Detail view with extracted TODOs
squirrel status --deep --json  # Full analysis with TODOs as JSON
```

## 🤖 Claude Code Skill

Install the `/squirrel` skill for Claude Code:

```bash
squirrel install-skill
```

Then use `/squirrel` in any Claude Code session for AI-powered project recommendations.

## 🧠 How It Works

Squirrel reads:
- `~/.claude/history.jsonl` — your prompt history across all projects
- `~/.claude/projects/*/sessions-index.json` — session summaries per project
- Git status of project directories (medium/deep mode)
- `~/.claude/projects/*/*.jsonl` — session JSONL files (deep mode only)

It categorizes projects into:
- 🚧 **Open Work** — uncommitted changes, feature branches
- ✅ **Recent Activity** — clean projects you worked on recently
- 😴 **Sleeping** — projects that went quiet

### Analysis Depths

| Depth | What it does |
|-------|-------------|
| `--quick` | History + sessions only (fastest) |
| `--medium` | + Git status (default) |
| `--deep` | + TODO/FIXME/HACK extraction from session JSONL files |

## 📄 License

MIT — see [LICENSE](LICENSE)
