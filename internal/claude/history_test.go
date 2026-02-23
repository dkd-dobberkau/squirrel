package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseHistory(t *testing.T) {
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

func TestParseHistorySkipsMalformed(t *testing.T) {
	dir := t.TempDir()
	histFile := filepath.Join(dir, "history.jsonl")

	content := `{"display":"valid","timestamp":1759336699341,"project":"/Users/test/proj"}
not valid json at all
{"display":"also valid","timestamp":1759336700000,"project":"/Users/test/proj"}
`
	if err := os.WriteFile(histFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := ParseHistory(histFile)
	if err != nil {
		t.Fatalf("ParseHistory failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (skipping malformed), got %d", len(entries))
	}
}

func TestParseHistoryFileNotFound(t *testing.T) {
	_, err := ParseHistory("/nonexistent/path/history.jsonl")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestAggregateByProject(t *testing.T) {
	// Use recent timestamps so the 30-day filter includes them
	now := time.Now()
	ts1 := now.Add(-2 * time.Hour).UnixMilli()
	ts2 := now.Add(-1 * time.Hour).UnixMilli()
	ts3 := now.Add(-30 * time.Minute).UnixMilli()

	entries := []HistoryEntry{
		{Display: "first", Timestamp: ts1, Project: "/Users/test/project-a"},
		{Display: "second", Timestamp: ts2, Project: "/Users/test/project-a"},
		{Display: "third", Timestamp: ts3, Project: "/Users/test/project-b"},
	}

	projects := AggregateByProject(entries, 30)

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

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

func TestAggregateByProjectFiltersOldEntries(t *testing.T) {
	now := time.Now()
	recentTS := now.Add(-1 * time.Hour).UnixMilli()
	oldTS := now.AddDate(0, 0, -60).UnixMilli() // 60 days ago

	entries := []HistoryEntry{
		{Display: "recent", Timestamp: recentTS, Project: "/Users/test/active"},
		{Display: "old", Timestamp: oldTS, Project: "/Users/test/stale"},
	}

	projects := AggregateByProject(entries, 30)

	if len(projects) != 1 {
		t.Fatalf("expected 1 project (old one filtered), got %d", len(projects))
	}

	if projects[0].Path != "/Users/test/active" {
		t.Errorf("expected active project, got %q", projects[0].Path)
	}
}

func TestParseHistoryRoundTrip(t *testing.T) {
	// Test that ParseHistory + AggregateByProject work together end-to-end
	dir := t.TempDir()
	histFile := filepath.Join(dir, "history.jsonl")

	now := time.Now()
	ts1 := now.Add(-1 * time.Hour).UnixMilli()
	ts2 := now.Add(-30 * time.Minute).UnixMilli()

	content := fmt.Sprintf(
		`{"display":"init project","timestamp":%d,"project":"/Users/test/myapp"}
{"display":"add tests","timestamp":%d,"project":"/Users/test/myapp"}
`, ts1, ts2)

	if err := os.WriteFile(histFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := ParseHistory(histFile)
	if err != nil {
		t.Fatalf("ParseHistory failed: %v", err)
	}

	projects := AggregateByProject(entries, 30)

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	if projects[0].PromptCount != 2 {
		t.Errorf("expected 2 prompts, got %d", projects[0].PromptCount)
	}

	if projects[0].LastPrompt != "add tests" {
		t.Errorf("expected last prompt 'add tests', got %q", projects[0].LastPrompt)
	}

	if projects[0].ShortName != "myapp" {
		t.Errorf("expected short name 'myapp', got %q", projects[0].ShortName)
	}
}
