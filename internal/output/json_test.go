package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dkd-dobberkau/squirrel/internal/analyzer"
	"github.com/dkd-dobberkau/squirrel/internal/claude"
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

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if _, ok := parsed["openWork"]; !ok {
		t.Error("expected 'openWork' key in JSON output")
	}
}
