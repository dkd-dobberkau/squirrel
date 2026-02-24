# Acknowledge (ack) Feature — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `squirrel ack/unack` commands so users can acknowledge projects, moving them to a greyed-out "Acknowledged" section instead of cluttering the main categories.

**Architecture:** New `internal/config` package handles a JSON config file at `~/.config/squirrel/config.json`. The analyzer's `Categorize()` function accepts acknowledged paths and filters them into a new `Acknowledged` category. Two new cobra commands (`ack`, `unack`) manage the config.

**Tech Stack:** Go stdlib (encoding/json, os, time, regexp), cobra (existing), lipgloss (existing)

---

### Task 1: Config package — types and duration parsing

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the failing tests for duration parsing**

```go
package config

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"7d", 7 * 24 * time.Hour, false},
		{"30d", 30 * 24 * time.Hour, false},
		{"2w", 14 * 24 * time.Hour, false},
		{"3m", 90 * 24 * time.Hour, false},
		{"1m", 30 * 24 * time.Hour, false},
		{"", 0, true},
		{"abc", 0, true},
		{"0d", 0, true},
		{"-5d", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestParseDuration -v`
Expected: FAIL — package doesn't exist yet

**Step 3: Write types and ParseDuration implementation**

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// AckEntry represents a single acknowledged project.
type AckEntry struct {
	Path      string     `json:"path"`
	AckedAt   time.Time  `json:"ackedAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// Config is the top-level squirrel configuration.
type Config struct {
	Acknowledged []AckEntry `json:"acknowledged"`
}

var durationRe = regexp.MustCompile(`^(\d+)([dwm])$`)

// ParseDuration parses a human-friendly duration like "7d", "2w", "3m".
func ParseDuration(s string) (time.Duration, error) {
	matches := durationRe.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration %q (use e.g. 7d, 2w, 3m)", s)
	}

	n, _ := strconv.Atoi(matches[1])
	if n <= 0 {
		return 0, fmt.Errorf("duration must be positive: %q", s)
	}

	var days int
	switch matches[2] {
	case "d":
		days = n
	case "w":
		days = n * 7
	case "m":
		days = n * 30
	}

	return time.Duration(days) * 24 * time.Hour, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config/ -run TestParseDuration -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add config types and duration parsing"
```

---

### Task 2: Config package — Load, Save, IsAcknowledged, Ack, Unack

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Step 1: Write the failing tests**

Add to `internal/config/config_test.go`:

```go
func TestLoadSaveRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{}
	expires := time.Now().Add(30 * 24 * time.Hour)
	cfg.Ack("/projects/foo", nil)
	cfg.Ack("/projects/bar", &expires)

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Acknowledged) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded.Acknowledged))
	}

	if loaded.Acknowledged[0].Path != "/projects/foo" {
		t.Errorf("expected path /projects/foo, got %s", loaded.Acknowledged[0].Path)
	}
}

func TestLoadNonexistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load of nonexistent file should not error: %v", err)
	}
	if len(cfg.Acknowledged) != 0 {
		t.Errorf("expected empty config, got %d entries", len(cfg.Acknowledged))
	}
}

func TestIsAcknowledged(t *testing.T) {
	cfg := &Config{}
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	cfg.Ack("/projects/permanent", nil)
	cfg.Ack("/projects/valid", &future)
	cfg.Ack("/projects/expired", &past)

	if !cfg.IsAcknowledged("/projects/permanent") {
		t.Error("permanent ack should be acknowledged")
	}
	if !cfg.IsAcknowledged("/projects/valid") {
		t.Error("non-expired ack should be acknowledged")
	}
	if cfg.IsAcknowledged("/projects/expired") {
		t.Error("expired ack should NOT be acknowledged")
	}
	if cfg.IsAcknowledged("/projects/unknown") {
		t.Error("unknown project should NOT be acknowledged")
	}
}

func TestAckIdempotent(t *testing.T) {
	cfg := &Config{}
	cfg.Ack("/projects/foo", nil)
	cfg.Ack("/projects/foo", nil)

	if len(cfg.Acknowledged) != 1 {
		t.Errorf("double ack should not create duplicates, got %d", len(cfg.Acknowledged))
	}
}

func TestUnack(t *testing.T) {
	cfg := &Config{}
	cfg.Ack("/projects/foo", nil)
	cfg.Ack("/projects/bar", nil)

	if !cfg.Unack("/projects/foo") {
		t.Error("Unack should return true for existing entry")
	}
	if len(cfg.Acknowledged) != 1 {
		t.Errorf("expected 1 entry after unack, got %d", len(cfg.Acknowledged))
	}
	if cfg.Unack("/projects/unknown") {
		t.Error("Unack should return false for nonexistent entry")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/config/ -v`
Expected: FAIL — Load, Save, Ack, Unack, IsAcknowledged not defined

**Step 3: Implement Load, Save, IsAcknowledged, Ack, Unack**

Add to `internal/config/config.go`:

```go
// DefaultPath returns the default config file path.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "squirrel", "config.json")
}

// Load reads the config from path. Returns empty config if file doesn't exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to path, creating parent directories as needed.
func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// IsAcknowledged checks if a project path is acknowledged and not expired.
func (c *Config) IsAcknowledged(path string) bool {
	for _, e := range c.Acknowledged {
		if e.Path == path {
			if e.ExpiresAt != nil && e.ExpiresAt.Before(time.Now()) {
				return false
			}
			return true
		}
	}
	return false
}

// Ack adds or updates an acknowledgement for a project path.
func (c *Config) Ack(path string, expiresAt *time.Time) {
	for i, e := range c.Acknowledged {
		if e.Path == path {
			c.Acknowledged[i].AckedAt = time.Now()
			c.Acknowledged[i].ExpiresAt = expiresAt
			return
		}
	}
	c.Acknowledged = append(c.Acknowledged, AckEntry{
		Path:      path,
		AckedAt:   time.Now(),
		ExpiresAt: expiresAt,
	})
}

// Unack removes the acknowledgement for a project path. Returns true if found.
func (c *Config) Unack(path string) bool {
	for i, e := range c.Acknowledged {
		if e.Path == path {
			c.Acknowledged = append(c.Acknowledged[:i], c.Acknowledged[i+1:]...)
			return true
		}
	}
	return false
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add Load, Save, Ack, Unack, IsAcknowledged"
```

---

### Task 3: Analyzer — add Acknowledged category

**Files:**
- Modify: `internal/analyzer/analyzer.go`
- Modify: `internal/analyzer/analyzer_test.go`

**Step 1: Write the failing test**

Add to `internal/analyzer/analyzer_test.go`:

```go
func TestCategorizeWithAcknowledged(t *testing.T) {
	now := time.Now()

	projects := []claude.ProjectInfo{
		{Path: "/p/dirty", ShortName: "dirty", GitDirty: true, LastActivity: now, DaysSinceActive: 0, PromptCount: 50},
		{Path: "/p/acked", ShortName: "acked", GitDirty: true, LastActivity: now, DaysSinceActive: 0, PromptCount: 30},
		{Path: "/p/clean", ShortName: "clean", GitBranch: "main", LastActivity: now, DaysSinceActive: 0, PromptCount: 10},
	}

	ackedPaths := map[string]bool{"/p/acked": true}

	result := Categorize(projects, ackedPaths)

	if len(result.OpenWork) != 1 {
		t.Errorf("expected 1 open work, got %d", len(result.OpenWork))
	}
	if len(result.Acknowledged) != 1 {
		t.Errorf("expected 1 acknowledged, got %d", len(result.Acknowledged))
	}
	if result.Acknowledged[0].Path != "/p/acked" {
		t.Errorf("expected acked project, got %s", result.Acknowledged[0].Path)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/analyzer/ -run TestCategorizeWithAcknowledged -v`
Expected: FAIL — Categorize signature doesn't accept ackedPaths

**Step 3: Update CategorizedProjects and Categorize**

In `internal/analyzer/analyzer.go`:

1. Add `Acknowledged` field to `CategorizedProjects`:
```go
type CategorizedProjects struct {
	OpenWork       []claude.ProjectInfo `json:"openWork"`
	RecentActivity []claude.ProjectInfo `json:"recentActivity"`
	Sleeping       []claude.ProjectInfo `json:"sleeping"`
	Acknowledged   []claude.ProjectInfo `json:"acknowledged"`
}
```

2. Change `Categorize` signature to accept `ackedPaths map[string]bool`:
```go
func Categorize(projects []claude.ProjectInfo, ackedPaths map[string]bool) CategorizedProjects {
```

3. At the start of the loop, before the switch, add:
```go
if ackedPaths[p.Path] {
    result.Acknowledged = append(result.Acknowledged, p)
    continue
}
```

4. Add sort for Acknowledged at the end (same `sortByScore` call).

**Step 4: Fix existing test**

Update `TestCategorize` in `analyzer_test.go` — pass `nil` as second arg:
```go
result := Categorize(projects, nil)
```

**Step 5: Run all analyzer tests**

Run: `go test ./internal/analyzer/ -v`
Expected: PASS

**Step 6: Fix callers of Categorize in main.go**

In `cmd/squirrel/main.go`, update `runAnalysis()` line 47:
```go
categorized := analyzer.Categorize(projects, nil)
```

This is a temporary `nil` — Task 6 will wire in the real config.

**Step 7: Run full test suite to verify nothing broke**

Run: `go test ./...`
Expected: PASS

**Step 8: Commit**

```bash
git add internal/analyzer/analyzer.go internal/analyzer/analyzer_test.go cmd/squirrel/main.go
git commit -m "feat(analyzer): add Acknowledged category to Categorize"
```

---

### Task 4: Terminal output — render Acknowledged section

**Files:**
- Modify: `internal/output/terminal.go`

**Step 1: Add the Acknowledged section rendering**

In `internal/output/terminal.go`, in `RenderTerminal()`, add before the "no projects" check (before line 74):

```go
if len(data.Acknowledged) > 0 {
    b.WriteString(sectionStyle.Render(fmt.Sprintf("\nAcknowledged (%d)", len(data.Acknowledged))))
    b.WriteString("\n")
    for _, p := range data.Acknowledged {
        b.WriteString(dimStyle.Render("  ✓ "))
        b.WriteString(dimStyle.Render(formatProjectAck(p)))
        b.WriteString("\n")
    }
}
```

**Step 2: Add formatProjectAck helper**

Add to `internal/output/terminal.go`:

```go
func formatProjectAck(p claude.ProjectInfo) string {
	date := p.LastActivity.Format("02.01.")
	name := fmt.Sprintf("%-22s", truncate(p.ShortName, 22))
	prompts := fmt.Sprintf("%4d prompts", p.PromptCount)
	return strings.Join([]string{name, date, prompts}, " | ")
}
```

**Step 3: Update the empty-check condition**

Change line 74 from:
```go
if len(data.OpenWork) == 0 && len(data.RecentActivity) == 0 && len(data.Sleeping) == 0 {
```
to:
```go
if len(data.OpenWork) == 0 && len(data.RecentActivity) == 0 && len(data.Sleeping) == 0 && len(data.Acknowledged) == 0 {
```

**Step 4: Build and verify**

Run: `go build ./cmd/squirrel/`
Expected: builds successfully

**Step 5: Commit**

```bash
git add internal/output/terminal.go
git commit -m "feat(output): render Acknowledged section in terminal output"
```

---

### Task 5: JSON output — verify acknowledged field

**Files:**
- Modify: `internal/output/json_test.go`

**Step 1: Add test for acknowledged in JSON**

Add to `internal/output/json_test.go`:

```go
func TestRenderJSONWithAcknowledged(t *testing.T) {
	data := analyzer.CategorizedProjects{
		Acknowledged: []claude.ProjectInfo{
			{ShortName: "acked-project", PromptCount: 13, LastActivity: time.Now()},
		},
	}

	result, err := RenderJSON(data)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	acked, ok := parsed["acknowledged"]
	if !ok {
		t.Fatal("expected 'acknowledged' key in JSON output")
	}
	arr, ok := acked.([]interface{})
	if !ok || len(arr) != 1 {
		t.Errorf("expected 1 acknowledged entry, got %v", acked)
	}
}
```

**Step 2: Run test**

Run: `go test ./internal/output/ -run TestRenderJSONWithAcknowledged -v`
Expected: PASS (the struct field already has the json tag from Task 3)

**Step 3: Commit**

```bash
git add internal/output/json_test.go
git commit -m "test(output): verify acknowledged field in JSON output"
```

---

### Task 6: CLI commands — ack and unack

**Files:**
- Modify: `cmd/squirrel/main.go`

**Step 1: Add ack command**

Add to `cmd/squirrel/main.go`:

```go
var forDuration string

var ackCmd = &cobra.Command{
	Use:   "ack [project]",
	Short: "Acknowledge a project (moves it to the Acknowledged section)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := config.DefaultPath()
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		// Resolve project
		cDir := claudeDir()
		histPath := filepath.Join(cDir, "history.jsonl")
		entries, err := claude.ParseHistory(histPath)
		if err != nil {
			return fmt.Errorf("reading history: %w", err)
		}
		projects := claude.AggregateByProject(entries, 365)
		project, ok := claude.FindProject(projects, args[0])
		if !ok {
			return fmt.Errorf("project %q not found", args[0])
		}

		var expiresAt *time.Time
		if forDuration != "" {
			d, err := config.ParseDuration(forDuration)
			if err != nil {
				return err
			}
			t := time.Now().Add(d)
			expiresAt = &t
		}

		cfg.Ack(project.Path, expiresAt)
		if err := config.Save(cfg, cfgPath); err != nil {
			return err
		}

		if expiresAt != nil {
			fmt.Printf("Acknowledged %s (expires %s)\n", project.ShortName, expiresAt.Format("02.01.2006"))
		} else {
			fmt.Printf("Acknowledged %s (permanent)\n", project.ShortName)
		}
		return nil
	},
}

var unackCmd = &cobra.Command{
	Use:   "unack [project]",
	Short: "Remove acknowledgement from a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := config.DefaultPath()
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		// Resolve project
		cDir := claudeDir()
		histPath := filepath.Join(cDir, "history.jsonl")
		entries, err := claude.ParseHistory(histPath)
		if err != nil {
			return fmt.Errorf("reading history: %w", err)
		}
		projects := claude.AggregateByProject(entries, 365)
		project, ok := claude.FindProject(projects, args[0])
		if !ok {
			return fmt.Errorf("project %q not found", args[0])
		}

		if cfg.Unack(project.Path) {
			if err := config.Save(cfg, cfgPath); err != nil {
				return err
			}
			fmt.Printf("Removed acknowledgement for %s\n", project.ShortName)
		} else {
			fmt.Printf("%s was not acknowledged\n", project.ShortName)
		}
		return nil
	},
}
```

**Step 2: Wire ack into init()**

In `init()`, add:
```go
ackCmd.Flags().StringVar(&forDuration, "for", "", "Duration (e.g. 7d, 2w, 3m)")
rootCmd.AddCommand(ackCmd)
rootCmd.AddCommand(unackCmd)
```

**Step 3: Wire config into runAnalysis()**

Update `runAnalysis()` to load config and pass acked paths to `Categorize()`:

```go
func runAnalysis() (analyzer.CategorizedProjects, error) {
	cDir := claudeDir()
	histPath := filepath.Join(cDir, "history.jsonl")

	entries, err := claude.ParseHistory(histPath)
	if err != nil {
		return analyzer.CategorizedProjects{}, fmt.Errorf("reading history: %w", err)
	}

	projects := claude.AggregateByProject(entries, days)
	claude.EnrichWithSessions(projects, filepath.Join(cDir, "projects"))

	if depth == "medium" || depth == "deep" {
		analyzer.EnrichWithGit(projects)
	}

	// Load config for acknowledged projects
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		cfg = &config.Config{} // gracefully continue without config
	}
	ackedPaths := make(map[string]bool)
	for _, p := range projects {
		if cfg.IsAcknowledged(p.Path) {
			ackedPaths[p.Path] = true
		}
	}

	categorized := analyzer.Categorize(projects, ackedPaths)
	// ... rest of deep enrichment stays the same
```

Don't forget to add `"github.com/dkd-dobberkau/squirrel/internal/config"` to the import block.

**Step 4: Build and manual test**

Run: `go build ./cmd/squirrel/ && ./squirrel ack --help`
Expected: shows ack usage with --for flag

**Step 5: Commit**

```bash
git add cmd/squirrel/main.go
git commit -m "feat: add squirrel ack/unack commands"
```

---

### Task 7: Full integration test

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: all tests PASS

**Step 2: Manual smoke test**

Run: `go build -o squirrel ./cmd/squirrel/ && ./squirrel status`
Expected: output shows all categories (Acknowledged empty until you ack something)

**Step 3: Commit any remaining fixes**

If any fixes were needed, commit them.
