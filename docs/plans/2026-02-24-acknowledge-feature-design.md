# Design: Acknowledge (ack) Feature

## Problem

Squirrel shows all discovered projects in its output. Users need a way to acknowledge projects they're aware of but don't want prominently displayed — reducing noise without losing visibility entirely.

## Decision Summary

- **Command:** `squirrel ack <project>` / `squirrel unack <project>`
- **Storage:** `~/.config/squirrel/config.json`
- **Display:** Acknowledged projects shown in a separate "Acknowledged" category (greyed out, at bottom)
- **Duration:** Permanent by default, optional `--for <duration>` for time-limited ack

## Config File

Location: `~/.config/squirrel/config.json`

```json
{
  "acknowledged": [
    {
      "path": "/Users/olivier/Versioncontrol/typo3/typo3",
      "ackedAt": "2026-02-24T14:30:00Z",
      "expiresAt": "2026-03-26T14:30:00Z"
    },
    {
      "path": "/Users/olivier/Versioncontrol/local/oldproject",
      "ackedAt": "2026-02-24T14:30:00Z",
      "expiresAt": null
    }
  ]
}
```

## CLI Commands

```
squirrel ack <project>            # Acknowledge permanently
squirrel ack <project> --for 30d  # Acknowledge for 30 days
squirrel unack <project>          # Remove acknowledgement
```

Project matching uses existing `FindProject()` logic (exact path > ShortName > path suffix > substring).

## Architecture Changes

### New package: `internal/config`

Responsible for loading/saving `~/.config/squirrel/config.json`. Provides:
- `Load() (*Config, error)` — load config, create if missing
- `Save(*Config) error` — write config to disk
- `(*Config) IsAcknowledged(path string) bool` — check if project is acked (and not expired)
- `(*Config) Ack(path string, expiresAt *time.Time)` — add ack entry
- `(*Config) Unack(path string) bool` — remove ack entry

### Modified: `analyzer.CategorizedProjects`

Add `Acknowledged []claude.ProjectInfo` field.

### Modified: `analyzer.Categorize()`

Accept config parameter. After scoring, check if project is acknowledged. If yes, move to `Acknowledged` category instead of its normal category.

### Modified: `output/terminal.go`

Add "Acknowledged" section at the end with greyed-out styling and checkmark prefix.

### Modified: `output/json.go`

Include `acknowledged` field in JSON output.

### New commands in `cmd/squirrel/main.go`

- `ackCmd` — loads config, resolves project via analysis + FindProject, adds ack entry
- `unackCmd` — loads config, removes ack entry by path matching

## Duration Parsing

The `--for` flag accepts formats like `7d`, `30d`, `2w`, `3m`. Parsed in `internal/config` with a simple duration parser (no external dependency needed).

## Terminal Output

```
Acknowledged (2)
  ✓ typo3 typo3              | 24.02. |   13 prompts | expires 26.03.
  ✓ oldproject               | 15.01. |    5 prompts | permanent
```

Uses existing `dimStyle`/`sleepStyle` for greyed-out appearance.
