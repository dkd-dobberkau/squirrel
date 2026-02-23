---
name: squirrel
description: Find forgotten Claude Code projects - shows open work, activity timeline, and recommendations
user_invocable: true
---

# Squirrel - Find Your Forgotten Projects

Run the squirrel CLI to analyze Claude Code history and present results.

## Steps

1. Run the squirrel command with JSON output and deep analysis:

   ```bash
   squirrel status --json --deep --days 14
   ```

2. Parse the JSON output and present the results in a structured way:

   **For each category (openWork, recentActivity, sleeping):**
   - Show the project name, last activity date, prompt count
   - For open work: highlight uncommitted files and feature branches
   - For sleeping projects: show days since last activity

3. After presenting the overview, provide:
   - **Top 3 Empfehlungen:** Which projects the user should focus on (highest score)
   - **Quick Summary:** "Du hast X offene Baustellen, Y aktive Projekte und Z schlafende Projekte"

4. Ask the user: "An welchem Projekt moechtest du weiterarbeiten?"

5. When the user picks a project:
   - Show the last session summary for that project
   - Show the last few prompts from history
   - Suggest: "Soll ich in das Projektverzeichnis wechseln?"

## Notes

- If `squirrel` is not in PATH, build it first: `cd ~/Versioncontrol/local/squirrel && go build -o squirrel ./cmd/squirrel && cp squirrel /usr/local/bin/squirrel`
- The `--deep` flag takes longer but provides richer context
- Use `--days 30` for a broader view
