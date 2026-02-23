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
