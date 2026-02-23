# Squirrel Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI tool + Claude Code skill that analyzes Claude Code history to find forgotten projects, show activity timelines, and recommend where to continue working.

**Architecture:** Go binary reads `~/.claude/history.jsonl` and `~/.claude/projects/*/sessions-index.json` for quick analysis, adds git status checks for medium depth, and parses session JSONL files for deep analysis. Output is either styled terminal text (lipgloss) or JSON for skill integration. A Claude Code skill wraps the binary with AI-powered recommendations.

**Tech Stack:** Go 1.26, cobra (CLI), lipgloss (terminal styling), go-git (git status), encoding/json (parsing)

---

## Task 1: Go Module + Cobra Skeleton

**Files:**
- Create: `go.mod`
- Create: `cmd/squirrel/main.go`

**Step 1: Initialize Go module**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go mod init github.com/oliverthiele/squirrel`
Expected: `go.mod` created

**Step 2: Install cobra**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go get github.com/spf13/cobra@latest`
Expected: `go.sum` created, cobra in go.mod

**Step 3: Write main.go with root + status command**

Create `cmd/squirrel/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	depth   string
	days    int
	jsonOut bool
)

var rootCmd = &cobra.Command{
	Use:   "squirrel",
	Short: "Find your forgotten Claude Code projects",
	Long:  "Squirrel helps you find projects you started in Claude Code but forgot about.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusCmd.RunE(cmd, args)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project overview with open work, activity, and sleeping projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("squirrel status - not yet implemented")
		return nil
	},
}

var stashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Show only projects with uncommitted changes or feature branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("squirrel stash - not yet implemented")
		return nil
	},
}

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Show chronological activity timeline",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("squirrel timeline - not yet implemented")
		return nil
	},
}

var projectCmd = &cobra.Command{
	Use:   "project [path]",
	Short: "Show detailed view for a specific project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("squirrel project %s - not yet implemented\n", args[0])
		return nil
	},
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&depth, "depth", "medium", "Analysis depth: quick, medium, or deep")
	pf.BoolVar(&jsonOut, "json", false, "Output as JSON (for skill integration)")
	pf.IntVar(&days, "days", 14, "Number of days to look back")

	// Shortcut flags on status (also inherited by root)
	statusCmd.Flags().Bool("quick", false, "Shortcut for --depth=quick")
	statusCmd.Flags().Bool("medium", false, "Shortcut for --depth=medium")
	statusCmd.Flags().Bool("deep", false, "Shortcut for --depth=deep")

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stashCmd)
	rootCmd.AddCommand(timelineCmd)
	rootCmd.AddCommand(projectCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 4: Build and verify**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go build -o squirrel ./cmd/squirrel && ./squirrel --help`
Expected: Help text showing `squirrel`, `status`, `stash`, `timeline`, `project` commands

**Step 5: Commit**

```bash
git add go.mod go.sum cmd/
git commit -m "feat: cobra CLI skeleton with status, stash, timeline, project commands"
```

---

## Task 2: Types + History Parser

**Files:**
- Create: `internal/claude/types.go`
- Create: `internal/claude/history.go`
- Create: `internal/claude/history_test.go`

**Step 1: Write types.go**

Create `internal/claude/types.go`:

```go
package claude

import "time"

// HistoryEntry is a single line from ~/.claude/history.jsonl
type HistoryEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"`
	Project   string `json:"project"`
}

// SessionEntry is one entry from sessions-index.json
type SessionEntry struct {
	SessionID   string `json:"sessionId"`
	FullPath    string `json:"fullPath"`
	FirstPrompt string `json:"firstPrompt"`
	Summary     string `json:"summary"`
	MsgCount    int    `json:"messageCount"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
	GitBranch   string `json:"gitBranch"`
	ProjectPath string `json:"projectPath"`
	IsSidechain bool   `json:"isSidechain"`
}

// SessionsIndex is the top-level structure of sessions-index.json
type SessionsIndex struct {
	Version int            `json:"version"`
	Entries []SessionEntry `json:"entries"`
}

// SessionMessage is a single line from a session JSONL file (for deep mode)
type SessionMessage struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	CWD       string `json:"cwd"`
	GitBranch string `json:"gitBranch"`
}

// ProjectInfo aggregates all data we know about a project
type ProjectInfo struct {
	Path            string         `json:"path"`
	ShortName       string         `json:"shortName"`
	PromptCount     int            `json:"promptCount"`
	LastActivity    time.Time      `json:"lastActivity"`
	FirstActivity   time.Time      `json:"firstActivity"`
	LastPrompt      string         `json:"lastPrompt"`
	Sessions        []SessionEntry `json:"sessions,omitempty"`
	LatestSummary   string         `json:"latestSummary,omitempty"`
	LatestBranch    string         `json:"latestBranch,omitempty"`
	// Populated by medium/deep analysis
	GitDirty        bool     `json:"gitDirty"`
	GitBranch       string   `json:"gitBranch"`
	UncommittedFiles int     `json:"uncommittedFiles"`
	DaysSinceActive int      `json:"daysSinceActive"`
	IsOpenWork      bool     `json:"isOpenWork"`
	Score           float64  `json:"score"`
}
```

**Step 2: Write the failing test for history parsing**

Create `internal/claude/history_test.go`:

```go
package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseHistory(t *testing.T) {
	// Create a temp history.jsonl
	dir := t.TempDir()
	histFile := filepath.Join(dir, "history.jsonl")

	content := `{"display":"hello world","timestamp":1759336699341,"project":"/Users/test/project-a"}
{"display":"fix bug","timestamp":1759336700000,"project":"/Users/test/project-a"}
{"display":"add feature","timestamp":1759336800000,"project":"/Users/test/project-b"}
`
	if err := os.WriteFile(histFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := ParseHistory(histFile)
	if err != nil {
		t.Fatalf("ParseHistory failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	if entries[0].Display != "hello world" {
		t.Errorf("expected 'hello world', got %q", entries[0].Display)
	}

	if entries[0].Project != "/Users/test/project-a" {
		t.Errorf("expected project-a, got %q", entries[0].Project)
	}
}

func TestAggregateByProject(t *testing.T) {
	entries := []HistoryEntry{
		{Display: "first", Timestamp: 1759336699341, Project: "/Users/test/project-a"},
		{Display: "second", Timestamp: 1759336700000, Project: "/Users/test/project-a"},
		{Display: "third", Timestamp: 1759336800000, Project: "/Users/test/project-b"},
	}

	projects := AggregateByProject(entries, 30)

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

	// Find project-a
	var projA *ProjectInfo
	for i := range projects {
		if projects[i].Path == "/Users/test/project-a" {
			projA = &projects[i]
			break
		}
	}

	if projA == nil {
		t.Fatal("project-a not found")
	}

	if projA.PromptCount != 2 {
		t.Errorf("expected 2 prompts for project-a, got %d", projA.PromptCount)
	}

	if projA.LastPrompt != "second" {
		t.Errorf("expected last prompt 'second', got %q", projA.LastPrompt)
	}

	if projA.ShortName != "project-a" {
		t.Errorf("expected short name 'project-a', got %q", projA.ShortName)
	}
}
```

**Step 3: Run test to verify it fails**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/claude/ -v`
Expected: FAIL - `ParseHistory` and `AggregateByProject` not defined

**Step 4: Write history.go**

Create `internal/claude/history.go`:

```go
package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ParseHistory reads ~/.claude/history.jsonl and returns all entries.
func ParseHistory(path string) ([]HistoryEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for long lines
	for scanner.Scan() {
		var e HistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue // skip malformed lines
		}
		entries = append(entries, e)
	}
	return entries, scanner.Err()
}

// AggregateByProject groups history entries by project and computes per-project stats.
// Only includes projects with activity in the last `days` days.
func AggregateByProject(entries []HistoryEntry, days int) []ProjectInfo {
	cutoff := time.Now().AddDate(0, 0, -days)

	type projectAcc struct {
		count     int
		lastTS    int64
		firstTS   int64
		lastPrompt string
	}

	acc := make(map[string]*projectAcc)

	for _, e := range entries {
		t := time.UnixMilli(e.Timestamp)
		if t.Before(cutoff) {
			continue
		}

		p, ok := acc[e.Project]
		if !ok {
			p = &projectAcc{firstTS: e.Timestamp}
			acc[e.Project] = p
		}
		p.count++
		if e.Timestamp > p.lastTS {
			p.lastTS = e.Timestamp
			p.lastPrompt = e.Display
		}
		if e.Timestamp < p.firstTS {
			p.firstTS = e.Timestamp
		}
	}

	projects := make([]ProjectInfo, 0, len(acc))
	for path, p := range acc {
		projects = append(projects, ProjectInfo{
			Path:          path,
			ShortName:     filepath.Base(path),
			PromptCount:   p.count,
			LastActivity:  time.UnixMilli(p.lastTS),
			FirstActivity: time.UnixMilli(p.firstTS),
			LastPrompt:    p.lastPrompt,
			DaysSinceActive: int(time.Since(time.UnixMilli(p.lastTS)).Hours() / 24),
		})
	}

	return projects
}
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/claude/ -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/claude/
git commit -m "feat: types and history.jsonl parser with aggregation"
```

---

## Task 3: Sessions Parser

**Files:**
- Create: `internal/claude/sessions.go`
- Create: `internal/claude/sessions_test.go`

**Step 1: Write failing test**

Create `internal/claude/sessions_test.go`:

```go
package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSessionsIndex(t *testing.T) {
	dir := t.TempDir()
	idx := SessionsIndex{
		Version: 1,
		Entries: []SessionEntry{
			{
				SessionID:   "abc-123",
				FirstPrompt: "hello",
				Summary:     "Built something cool",
				MsgCount:    42,
				Created:     "2026-02-20T10:00:00.000Z",
				Modified:    "2026-02-20T12:00:00.000Z",
				GitBranch:   "feature/cool",
				ProjectPath: "/Users/test/project-a",
			},
			{
				SessionID:   "def-456",
				FirstPrompt: "fix bug",
				Summary:     "Fixed the thing",
				MsgCount:    10,
				Created:     "2026-02-19T08:00:00.000Z",
				Modified:    "2026-02-19T09:00:00.000Z",
				GitBranch:   "main",
				ProjectPath: "/Users/test/project-a",
			},
		},
	}

	data, _ := json.Marshal(idx)
	idxFile := filepath.Join(dir, "sessions-index.json")
	os.WriteFile(idxFile, data, 0644)

	result, err := ParseSessionsIndex(idxFile)
	if err != nil {
		t.Fatalf("ParseSessionsIndex failed: %v", err)
	}

	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}

	if result.Entries[0].Summary != "Built something cool" {
		t.Errorf("unexpected summary: %q", result.Entries[0].Summary)
	}
}

func TestEnrichWithSessions(t *testing.T) {
	claudeDir := t.TempDir()

	// Create a project dir with sessions-index.json
	projDir := filepath.Join(claudeDir, "-Users-test-project-a")
	os.MkdirAll(projDir, 0755)

	idx := SessionsIndex{
		Version: 1,
		Entries: []SessionEntry{
			{
				SessionID:   "abc-123",
				Summary:     "Built feature X",
				MsgCount:    42,
				Modified:    "2026-02-20T12:00:00.000Z",
				GitBranch:   "feature/x",
				ProjectPath: "/Users/test/project-a",
			},
		},
	}
	data, _ := json.Marshal(idx)
	os.WriteFile(filepath.Join(projDir, "sessions-index.json"), data, 0644)

	projects := []ProjectInfo{
		{Path: "/Users/test/project-a", ShortName: "project-a"},
	}

	EnrichWithSessions(projects, claudeDir)

	if len(projects[0].Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(projects[0].Sessions))
	}

	if projects[0].LatestSummary != "Built feature X" {
		t.Errorf("expected summary 'Built feature X', got %q", projects[0].LatestSummary)
	}

	if projects[0].LatestBranch != "feature/x" {
		t.Errorf("expected branch 'feature/x', got %q", projects[0].LatestBranch)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/claude/ -v -run TestParseSessions`
Expected: FAIL - functions not defined

**Step 3: Write sessions.go**

Create `internal/claude/sessions.go`:

```go
package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ParseSessionsIndex reads a sessions-index.json file.
func ParseSessionsIndex(path string) (*SessionsIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx SessionsIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// projectPathToDir converts a project path like "/Users/olivier/Versioncontrol/local/foo"
// to the Claude projects directory name "-Users-olivier-Versioncontrol-local-foo".
func projectPathToDir(projectPath string) string {
	return strings.ReplaceAll(projectPath, "/", "-")
}

// EnrichWithSessions adds session data to ProjectInfo entries by reading
// the corresponding sessions-index.json files from claudeDir/projects/.
func EnrichWithSessions(projects []ProjectInfo, claudeProjectsDir string) {
	for i := range projects {
		dirName := projectPathToDir(projects[i].Path)
		idxPath := filepath.Join(claudeProjectsDir, dirName, "sessions-index.json")

		idx, err := ParseSessionsIndex(idxPath)
		if err != nil {
			continue
		}

		projects[i].Sessions = idx.Entries

		// Find latest session by Modified timestamp
		var latestModified string
		for _, s := range idx.Entries {
			if s.Modified > latestModified {
				latestModified = s.Modified
				projects[i].LatestSummary = s.Summary
				projects[i].LatestBranch = s.GitBranch
			}
		}
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/claude/ -v`
Expected: PASS (all tests)

**Step 5: Commit**

```bash
git add internal/claude/sessions.go internal/claude/sessions_test.go
git commit -m "feat: sessions-index.json parser with project enrichment"
```

---

## Task 4: Git Status Checker

**Files:**
- Create: `internal/git/status.go`
- Create: `internal/git/status_test.go`

**Step 1: Install go-git**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go get github.com/go-git/go-git/v5@latest`

**Step 2: Write failing test**

Create `internal/git/status_test.go`:

```go
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCheckStatus_CleanRepo(t *testing.T) {
	dir := t.TempDir()

	// Init a git repo with one commit
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Run()
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	run("add", ".")
	run("commit", "-m", "init")

	status, err := CheckStatus(dir)
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if status.IsDirty {
		t.Error("expected clean repo, got dirty")
	}

	if status.Branch != "main" && status.Branch != "master" {
		t.Errorf("expected main or master branch, got %q", status.Branch)
	}

	if status.UncommittedFiles != 0 {
		t.Errorf("expected 0 uncommitted files, got %d", status.UncommittedFiles)
	}
}

func TestCheckStatus_DirtyRepo(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Run()
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	run("add", ".")
	run("commit", "-m", "init")

	// Make it dirty
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("changed"), 0644)
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644)

	status, err := CheckStatus(dir)
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if !status.IsDirty {
		t.Error("expected dirty repo, got clean")
	}

	if status.UncommittedFiles < 1 {
		t.Errorf("expected at least 1 uncommitted file, got %d", status.UncommittedFiles)
	}
}

func TestCheckStatus_FeatureBranch(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Run()
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	run("add", ".")
	run("commit", "-m", "init")
	run("checkout", "-b", "feature/cool")

	status, err := CheckStatus(dir)
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if status.Branch != "feature/cool" {
		t.Errorf("expected branch 'feature/cool', got %q", status.Branch)
	}

	if !status.IsFeatureBranch {
		t.Error("expected feature branch detection")
	}
}

func TestCheckStatus_NotARepo(t *testing.T) {
	dir := t.TempDir()

	status, err := CheckStatus(dir)
	if err != nil {
		t.Fatalf("CheckStatus should not error for non-repo, got: %v", err)
	}

	if status.IsRepo {
		t.Error("expected IsRepo=false for non-git directory")
	}
}
```

**Step 3: Run test to verify it fails**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/git/ -v`
Expected: FAIL - `CheckStatus` not defined

**Step 4: Write status.go**

Create `internal/git/status.go`:

```go
package git

import (
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

// RepoStatus holds the git status of a project directory.
type RepoStatus struct {
	IsRepo           bool   `json:"isRepo"`
	IsDirty          bool   `json:"isDirty"`
	Branch           string `json:"branch"`
	IsFeatureBranch  bool   `json:"isFeatureBranch"`
	UncommittedFiles int    `json:"uncommittedFiles"`
}

// CheckStatus checks the git status of a directory.
// Returns a zero-value RepoStatus with IsRepo=false if the directory is not a git repo.
func CheckStatus(path string) (RepoStatus, error) {
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		return RepoStatus{IsRepo: false}, nil
	}

	status := RepoStatus{IsRepo: true}

	// Get current branch
	head, err := repo.Head()
	if err == nil {
		ref := head.Name().Short()
		status.Branch = ref
		status.IsFeatureBranch = isFeatureBranch(ref)
	}

	// Get worktree status
	wt, err := repo.Worktree()
	if err != nil {
		return status, nil
	}

	ws, err := wt.Status()
	if err != nil {
		return status, nil
	}

	for _, s := range ws {
		if s.Worktree != gogit.Unmodified || s.Staging != gogit.Unmodified {
			status.UncommittedFiles++
		}
	}
	status.IsDirty = status.UncommittedFiles > 0

	return status, nil
}

func isFeatureBranch(name string) bool {
	lower := strings.ToLower(name)
	if lower == "main" || lower == "master" || lower == "develop" || lower == "dev" {
		return false
	}
	return true
}
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/git/ -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/git/
git commit -m "feat: git status checker with dirty/branch/feature detection"
```

---

## Task 5: Analyzer - Core Logic + Heuristics + Scoring

**Files:**
- Create: `internal/analyzer/analyzer.go`
- Create: `internal/analyzer/analyzer_test.go`

**Step 1: Write failing test**

Create `internal/analyzer/analyzer_test.go`:

```go
package analyzer

import (
	"testing"
	"time"

	"github.com/oliverthiele/squirrel/internal/claude"
)

func TestCategorize(t *testing.T) {
	now := time.Now()

	projects := []claude.ProjectInfo{
		// Open work: dirty + recent
		{ShortName: "dirty-recent", GitDirty: true, LastActivity: now, DaysSinceActive: 0, PromptCount: 50},
		// Open work: feature branch + recent
		{ShortName: "feature-recent", GitBranch: "feature/x", LastActivity: now.Add(-24 * time.Hour), DaysSinceActive: 1, PromptCount: 30},
		// Active and clean
		{ShortName: "clean-recent", GitDirty: false, GitBranch: "main", LastActivity: now, DaysSinceActive: 0, PromptCount: 10},
		// Sleeping: was active, now quiet
		{ShortName: "sleeping", GitDirty: false, GitBranch: "main", LastActivity: now.Add(-5 * 24 * time.Hour), DaysSinceActive: 5, PromptCount: 100},
	}

	result := Categorize(projects)

	if len(result.OpenWork) != 2 {
		t.Errorf("expected 2 open work items, got %d", len(result.OpenWork))
	}

	if len(result.RecentActivity) != 1 {
		t.Errorf("expected 1 recent activity, got %d", len(result.RecentActivity))
	}

	if len(result.Sleeping) != 1 {
		t.Errorf("expected 1 sleeping project, got %d", len(result.Sleeping))
	}
}

func TestScore(t *testing.T) {
	now := time.Now()

	p := claude.ProjectInfo{
		GitDirty:        true,
		GitBranch:       "feature/x",
		PromptCount:     100,
		DaysSinceActive: 1,
		LastActivity:    now.Add(-24 * time.Hour),
	}

	score := Score(p)
	if score <= 0 {
		t.Errorf("expected positive score for active dirty project, got %f", score)
	}

	clean := claude.ProjectInfo{
		GitDirty:        false,
		GitBranch:       "main",
		PromptCount:     5,
		DaysSinceActive: 10,
		LastActivity:    now.Add(-10 * 24 * time.Hour),
	}

	cleanScore := Score(clean)
	if cleanScore >= score {
		t.Errorf("clean old project should score lower than dirty recent one: %f >= %f", cleanScore, score)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/analyzer/ -v`
Expected: FAIL

**Step 3: Write analyzer.go**

Create `internal/analyzer/analyzer.go`:

```go
package analyzer

import (
	"math"
	"sort"

	"github.com/oliverthiele/squirrel/internal/claude"
	gitpkg "github.com/oliverthiele/squirrel/internal/git"
)

// CategorizedProjects holds projects sorted into categories.
type CategorizedProjects struct {
	OpenWork       []claude.ProjectInfo `json:"openWork"`
	RecentActivity []claude.ProjectInfo `json:"recentActivity"`
	Sleeping       []claude.ProjectInfo `json:"sleeping"`
}

// Categorize sorts projects into open work, recent activity, and sleeping.
func Categorize(projects []claude.ProjectInfo) CategorizedProjects {
	var result CategorizedProjects

	for _, p := range projects {
		p.Score = Score(p)
		p.IsOpenWork = isOpenWork(p)

		switch {
		case p.IsOpenWork:
			result.OpenWork = append(result.OpenWork, p)
		case p.DaysSinceActive <= 3:
			result.RecentActivity = append(result.RecentActivity, p)
		default:
			result.Sleeping = append(result.Sleeping, p)
		}
	}

	// Sort each category by score descending
	sortByScore := func(s []claude.ProjectInfo) {
		sort.Slice(s, func(i, j int) bool {
			return s[i].Score > s[j].Score
		})
	}
	sortByScore(result.OpenWork)
	sortByScore(result.RecentActivity)
	sortByScore(result.Sleeping)

	return result
}

func isOpenWork(p claude.ProjectInfo) bool {
	if p.GitDirty {
		return true
	}
	branch := p.GitBranch
	if branch == "" {
		branch = p.LatestBranch
	}
	if branch != "" && branch != "main" && branch != "master" && branch != "develop" && branch != "dev" {
		return true
	}
	return false
}

// Score computes a priority score for a project. Higher = more likely the user should work on it.
func Score(p claude.ProjectInfo) float64 {
	score := 0.0

	// Recency: exponential decay, halves every 3 days
	recency := math.Exp(-float64(p.DaysSinceActive) * math.Ln2 / 3.0)
	score += recency * 50

	// Activity: log scale of prompt count
	if p.PromptCount > 0 {
		score += math.Log2(float64(p.PromptCount)) * 5
	}

	// Dirty bonus
	if p.GitDirty {
		score += 30
	}

	// Feature branch bonus
	branch := p.GitBranch
	if branch == "" {
		branch = p.LatestBranch
	}
	if branch != "" && branch != "main" && branch != "master" && branch != "develop" && branch != "dev" {
		score += 20
	}

	return score
}

// EnrichWithGit adds git status data to projects (medium depth).
func EnrichWithGit(projects []claude.ProjectInfo) {
	for i := range projects {
		status, err := gitpkg.CheckStatus(projects[i].Path)
		if err != nil || !status.IsRepo {
			continue
		}
		projects[i].GitDirty = status.IsDirty
		projects[i].GitBranch = status.Branch
		projects[i].UncommittedFiles = status.UncommittedFiles
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/analyzer/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/analyzer/
git commit -m "feat: project categorization, scoring, and git enrichment"
```

---

## Task 6: JSON Output

**Files:**
- Create: `internal/output/json.go`
- Create: `internal/output/json_test.go`

**Step 1: Write failing test**

Create `internal/output/json_test.go`:

```go
package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/oliverthiele/squirrel/internal/analyzer"
	"github.com/oliverthiele/squirrel/internal/claude"
)

func TestRenderJSON(t *testing.T) {
	data := analyzer.CategorizedProjects{
		OpenWork: []claude.ProjectInfo{
			{ShortName: "project-a", PromptCount: 50, LastActivity: time.Now(), GitDirty: true},
		},
		RecentActivity: []claude.ProjectInfo{
			{ShortName: "project-b", PromptCount: 10, LastActivity: time.Now()},
		},
	}

	result, err := RenderJSON(data)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if _, ok := parsed["openWork"]; !ok {
		t.Error("expected 'openWork' key in JSON output")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/output/ -v`
Expected: FAIL

**Step 3: Write json.go**

Create `internal/output/json.go`:

```go
package output

import (
	"encoding/json"

	"github.com/oliverthiele/squirrel/internal/analyzer"
)

// RenderJSON returns the categorized projects as a JSON string.
func RenderJSON(data analyzer.CategorizedProjects) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./internal/output/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/output/json.go internal/output/json_test.go
git commit -m "feat: JSON output renderer for skill integration"
```

---

## Task 7: Terminal Output (lipgloss)

**Files:**
- Create: `internal/output/terminal.go`

**Step 1: Install lipgloss**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go get github.com/charmbracelet/lipgloss@latest`

**Step 2: Write terminal.go**

Create `internal/output/terminal.go`:

```go
package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oliverthiele/squirrel/internal/analyzer"
	"github.com/oliverthiele/squirrel/internal/claude"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF8C00")).
		MarginBottom(1)

	sectionStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#87CEEB")).
		MarginTop(1)

	warnStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700"))

	okStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#98FB98"))

	sleepStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080"))

	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))
)

// RenderTerminal prints the categorized projects to stdout.
func RenderTerminal(data analyzer.CategorizedProjects) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Squirrel - Deine vergessenen Nuesse"))
	b.WriteString("\n\n")

	// Open work
	if len(data.OpenWork) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("Offene Baustellen (%d)", len(data.OpenWork))))
		b.WriteString("\n")
		for _, p := range data.OpenWork {
			b.WriteString(warnStyle.Render("  ! "))
			b.WriteString(formatProject(p))
			b.WriteString("\n")
		}
	}

	// Recent activity
	if len(data.RecentActivity) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("\nLetzte Aktivitaet (%d)", len(data.RecentActivity))))
		b.WriteString("\n")
		for _, p := range data.RecentActivity {
			b.WriteString(okStyle.Render("  + "))
			b.WriteString(formatProject(p))
			b.WriteString("\n")
		}
	}

	// Sleeping
	if len(data.Sleeping) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("\nSchlafende Projekte (%d)", len(data.Sleeping))))
		b.WriteString("\n")
		for _, p := range data.Sleeping {
			b.WriteString(sleepStyle.Render("  ~ "))
			b.WriteString(formatProject(p))
			b.WriteString("\n")
		}
	}

	if len(data.OpenWork) == 0 && len(data.RecentActivity) == 0 && len(data.Sleeping) == 0 {
		b.WriteString(dimStyle.Render("  Keine Projekte im gewaehlten Zeitraum gefunden."))
		b.WriteString("\n")
	}

	return b.String()
}

func formatProject(p claude.ProjectInfo) string {
	date := p.LastActivity.Format("02.01.")
	name := fmt.Sprintf("%-22s", truncate(p.ShortName, 22))
	prompts := fmt.Sprintf("%4d prompts", p.PromptCount)

	details := []string{name, date, prompts}

	if p.UncommittedFiles > 0 {
		details = append(details, warnStyle.Render(fmt.Sprintf("%d uncommitted", p.UncommittedFiles)))
	}

	branch := p.GitBranch
	if branch == "" {
		branch = p.LatestBranch
	}
	if branch != "" && branch != "main" && branch != "master" {
		details = append(details, dimStyle.Render("branch: "+branch))
	}

	if p.DaysSinceActive > 0 {
		details = append(details, dimStyle.Render(fmt.Sprintf("%d Tage inaktiv", p.DaysSinceActive)))
	}

	return strings.Join(details, " | ")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "~"
}
```

**Step 3: Build to verify it compiles**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go build ./internal/output/`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/output/terminal.go
git commit -m "feat: lipgloss terminal output with color-coded categories"
```

---

## Task 8: Wire Everything Together in CLI

**Files:**
- Modify: `cmd/squirrel/main.go`

**Step 1: Update main.go to wire all components**

Replace `cmd/squirrel/main.go` with the fully wired version:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/oliverthiele/squirrel/internal/analyzer"
	"github.com/oliverthiele/squirrel/internal/claude"
	"github.com/oliverthiele/squirrel/internal/output"
)

var (
	depth   string
	days    int
	jsonOut bool
)

func claudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

func runAnalysis() (analyzer.CategorizedProjects, error) {
	cDir := claudeDir()
	histPath := filepath.Join(cDir, "history.jsonl")

	// Parse history
	entries, err := claude.ParseHistory(histPath)
	if err != nil {
		return analyzer.CategorizedProjects{}, fmt.Errorf("reading history: %w", err)
	}

	// Aggregate by project
	projects := claude.AggregateByProject(entries, days)

	// Quick: add session data
	claude.EnrichWithSessions(projects, filepath.Join(cDir, "projects"))

	// Medium: add git status
	if depth == "medium" || depth == "deep" {
		analyzer.EnrichWithGit(projects)
	}

	// Categorize and score
	return analyzer.Categorize(projects), nil
}

func renderOutput(data analyzer.CategorizedProjects) error {
	if jsonOut {
		s, err := output.RenderJSON(data)
		if err != nil {
			return err
		}
		fmt.Println(s)
	} else {
		fmt.Print(output.RenderTerminal(data))
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "squirrel",
	Short: "Find your forgotten Claude Code projects",
	Long:  "Squirrel helps you find projects you started in Claude Code but forgot about.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusCmd.RunE(cmd, args)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project overview with open work, activity, and sleeping projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveDepthShortcuts(cmd)
		data, err := runAnalysis()
		if err != nil {
			return err
		}
		return renderOutput(data)
	},
}

var stashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Show only projects with uncommitted changes or feature branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		if depth == "quick" {
			depth = "medium" // stash needs git info
		}
		data, err := runAnalysis()
		if err != nil {
			return err
		}
		// Only show open work
		data.RecentActivity = nil
		data.Sleeping = nil
		return renderOutput(data)
	},
}

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Show chronological activity timeline",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := runAnalysis()
		if err != nil {
			return err
		}
		// Merge all categories into one list sorted by date
		all := append(data.OpenWork, data.RecentActivity...)
		all = append(all, data.Sleeping...)
		data.OpenWork = nil
		data.RecentActivity = all
		data.Sleeping = nil
		return renderOutput(data)
	},
}

var projectCmd = &cobra.Command{
	Use:   "project [path]",
	Short: "Show detailed view for a specific project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("squirrel project %s - not yet implemented\n", args[0])
		return nil
	},
}

func resolveDepthShortcuts(cmd *cobra.Command) {
	if q, _ := cmd.Flags().GetBool("quick"); q {
		depth = "quick"
	}
	if m, _ := cmd.Flags().GetBool("medium"); m {
		depth = "medium"
	}
	if d, _ := cmd.Flags().GetBool("deep"); d {
		depth = "deep"
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&depth, "depth", "medium", "Analysis depth: quick, medium, or deep")
	pf.BoolVar(&jsonOut, "json", false, "Output as JSON (for skill integration)")
	pf.IntVar(&days, "days", 14, "Number of days to look back")

	statusCmd.Flags().Bool("quick", false, "Shortcut for --depth=quick")
	statusCmd.Flags().Bool("medium", false, "Shortcut for --depth=medium")
	statusCmd.Flags().Bool("deep", false, "Shortcut for --depth=deep")

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stashCmd)
	rootCmd.AddCommand(timelineCmd)
	rootCmd.AddCommand(projectCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 2: Build and test end-to-end**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go build -o squirrel ./cmd/squirrel && ./squirrel`
Expected: Styled output showing real project data from `~/.claude/`

**Step 3: Test JSON output**

Run: `./squirrel --json | python3 -m json.tool | head -30`
Expected: Valid formatted JSON with openWork/recentActivity/sleeping

**Step 4: Test subcommands**

Run: `./squirrel stash && echo "---" && ./squirrel timeline --days 7`
Expected: Stash shows only open work, timeline shows all projects from last 7 days

**Step 5: Run all tests**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./... -v`
Expected: All tests PASS

**Step 6: Commit**

```bash
git add cmd/squirrel/main.go
git commit -m "feat: wire all components into CLI with status, stash, timeline commands"
```

---

## Task 9: Claude Code Skill

**Files:**
- Create: `~/.claude/skills/squirrel/SKILL.md`

**Step 1: Build and install the binary**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go build -o squirrel ./cmd/squirrel && cp squirrel /usr/local/bin/squirrel`
Expected: `squirrel` available in PATH

**Step 2: Create the skill file**

Create `~/.claude/skills/squirrel/SKILL.md`:

```markdown
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
```

**Step 3: Test the skill**

In a Claude Code session, run `/squirrel` and verify it works.

**Step 4: Commit the skill to the squirrel repo**

```bash
# Also keep a copy in the repo for reference
mkdir -p /Users/olivier/Versioncontrol/local/squirrel/skill
cp ~/.claude/skills/squirrel/SKILL.md /Users/olivier/Versioncontrol/local/squirrel/skill/SKILL.md
cd /Users/olivier/Versioncontrol/local/squirrel
git add skill/
git commit -m "feat: add Claude Code /squirrel skill"
```

---

## Task 10: Final Polish + README

**Files:**
- Create: `README.md`

**Step 1: Write README.md**

Create `README.md`:

```markdown
# Squirrel

> Like a squirrel that forgot where it buried its nuts - find your forgotten Claude Code projects.

Squirrel analyzes your Claude Code history to find projects you started but forgot about.

## Install

```bash
go install github.com/oliverthiele/squirrel/cmd/squirrel@latest
```

Or build from source:

```bash
git clone https://github.com/oliverthiele/squirrel.git
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
```

**Step 2: Run full test suite one last time**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go test ./... -v`
Expected: All PASS

**Step 3: Build final binary**

Run: `cd /Users/olivier/Versioncontrol/local/squirrel && go build -o squirrel ./cmd/squirrel && ./squirrel`
Expected: Clean output with real data

**Step 4: Commit**

```bash
git add README.md
git commit -m "docs: add README with install and usage instructions"
```
