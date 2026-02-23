# Squirrel

> Like a squirrel that forgot where it buried its nuts - find your forgotten Claude Code projects.

Squirrel analyzes your Claude Code history to find projects you started but forgot about.

## Install

```bash
go install github.com/dkd-dobberkau/squirrel/cmd/squirrel@latest
```

Or build from source:

```bash
git clone https://github.com/dkd-dobberkau/squirrel.git
cd squirrel
go build -o squirrel ./cmd/squirrel
cp squirrel /usr/local/bin/
```

## Usage

```bash
squirrel                # Show everything (default: medium depth, 14 days)
squirrel status         # Same as above
squirrel stash          # Only show open work (uncommitted changes, feature branches)
squirrel timeline       # Chronological activity view

# Options
squirrel --quick        # Fast: only history + sessions
squirrel --depth=deep   # Deep: includes session context analysis
squirrel --days 30      # Look back 30 days
squirrel --json         # JSON output for scripting
```

## Claude Code Skill

Copy `skill/SKILL.md` to `~/.claude/skills/squirrel/SKILL.md`, then use `/squirrel` in any Claude Code session.

## How It Works

Squirrel reads:
- `~/.claude/history.jsonl` - your prompt history across all projects
- `~/.claude/projects/*/sessions-index.json` - session summaries per project
- Git status of project directories (medium/deep mode)

It categorizes projects into:
- **Open Work** - uncommitted changes, feature branches
- **Recent Activity** - clean projects you worked on recently
- **Sleeping** - projects that went quiet
