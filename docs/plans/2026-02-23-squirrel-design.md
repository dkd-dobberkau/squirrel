# Squirrel - Design Document

> "Wie ein Eichhoernchen, das seine Nuesse vergisst" - Ein Tool das dir hilft, deine vergessenen Claude Code Projekte wiederzufinden.

## Zusammenfassung

Squirrel ist ein Go CLI-Tool + Claude Code Skill, das die Claude Code History analysiert und dir zeigt:
- **Offene Baustellen** - Projekte mit uncommitted changes, Feature-Branches, unfertiger Arbeit
- **Aktivitaets-Timeline** - Was du in den letzten Tagen/Wochen gemacht hast
- **Smarte Empfehlungen** - Wo du am besten weiterarbeiten solltest

## Datenquellen

| Quelle | Level | Was wir daraus ziehen |
|--------|-------|----------------------|
| `~/.claude/history.jsonl` | Quick | Projekt-Aktivitaet, letzter Prompt, Haeufigkeit |
| `~/.claude/projects/*/sessions-index.json` | Quick | Session-Summaries, Git-Branch, Zeitstempel, Message-Count |
| Git-Status der Projekte | Medium | Uncommitted changes, offene Feature-Branches, Zeit seit letztem Commit |
| `~/.claude/projects/*/*.jsonl` (letzte Session) | Deep | TODOs, unfertiger Kontext, letzte User-Prompts |

## Heuristiken fuer "Offene Baustelle"

- Projekt wurde in den letzten 14 Tagen bearbeitet UND hat uncommitted changes
- Feature-Branch existiert (nicht main/master)
- Letzte Session-Summary enthaelt Hinweise auf unfertigen Stand
- Hohe Prompt-Frequenz in kurzer Zeit (= aktive Arbeit, evtl. unterbrochen)

## CLI-Befehle

```
squirrel                     # Alias fuer 'squirrel status --medium'
squirrel status              # Uebersicht aller Projekte (default: --medium)
  --quick                    # Nur History + Sessions
  --medium                   # + Git-Status (default)
  --deep                     # + Session-Kontext-Analyse
  --days 14                  # Zeitraum (default 14 Tage)
  --json                     # JSON-Output (fuer Skill-Integration)

squirrel stash               # Zeige nur offene Baustellen
squirrel timeline            # Chronologische Aktivitaets-Timeline
squirrel project <pfad>      # Detail-Ansicht fuer ein bestimmtes Projekt
```

## Beispiel-Ausgabe

```
Squirrel - Deine vergessenen Nuesse

Offene Baustellen (5)
  typo3-base          22.02. | 227 prompts | 3 uncommitted files | branch: feature/caching
  Steuer-2024         22.02. | 107 prompts | uncommitted changes
  GEO                 21.02. | 213 prompts | branch: feature/map-export
  eb2sw               23.02. |  31 prompts | uncommitted changes
  quasi               23.02. |  34 prompts | branch: feature/search

Letzte Aktivitaet (7 Tage)
  local-apple-llm      23.02. |   6 prompts | clean
  deuba-analyze         23.02. |  15 prompts | clean
  remote2claude         23.02. |  24 prompts | clean

Schlafende Projekte (kuerzlich aktiv, jetzt ruhig)
  dkd-umfrage          20.02. |  56 prompts | 3 Tage inaktiv
  claude-insights      19.02. | 243 prompts | 4 Tage inaktiv
```

## Claude Code Skill `/squirrel`

**Ablauf:**
1. User ruft `/squirrel` auf
2. Skill fuehrt `squirrel status --json --deep` aus
3. Claude analysiert das JSON-Ergebnis
4. Praesentiert: Offene Baustellen, Timeline, Top-3-Empfehlungen
5. Fragt: "An welchem Projekt willst du weiterarbeiten?"
6. User waehlt -> Claude zeigt letzten Kontext und schlaegt naechste Schritte vor

## Projektstruktur

```
squirrel/
├── cmd/
│   └── squirrel/
│       └── main.go              # CLI Entry Point (cobra)
├── internal/
│   ├── claude/
│   │   ├── history.go           # Parst history.jsonl
│   │   ├── sessions.go          # Parst sessions-index.json + Session-JSONLs
│   │   └── types.go             # Shared Types
│   ├── git/
│   │   └── status.go            # Git-Status-Checks (uncommitted, branches)
│   ├── analyzer/
│   │   ├── analyzer.go          # Kernlogik: kombiniert Datenquellen
│   │   ├── heuristics.go        # "Offene Baustelle"-Erkennung
│   │   └── scorer.go            # Scoring/Ranking der Projekte
│   └── output/
│       ├── terminal.go          # Huebsche Terminal-Ausgabe (lipgloss)
│       └── json.go              # JSON-Output fuer Skill-Integration
├── go.mod
├── go.sum
└── README.md
```

## Dependencies

- `github.com/spf13/cobra` - CLI-Framework
- `github.com/charmbracelet/lipgloss` - Terminal-Styling
- `github.com/go-git/go-git/v5` - Git-Status ohne externes Binary

## Technische Entscheidungen

- **Go** fuer schnelles Single-Binary, kein Runtime noetig
- **JSON als Zwischenformat** zwischen CLI und Skill
- **Konfigurierbare Analysetiefe** (quick/medium/deep) je nach Geduld
- **go-git** statt exec("git ...") fuer saubere Git-Integration
- **lipgloss** fuer huebsche aber nicht ueberladene Terminal-Ausgabe
