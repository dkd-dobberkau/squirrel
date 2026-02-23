package analyzer

import (
	"testing"
	"time"

	"github.com/dkd-dobberkau/squirrel/internal/claude"
)

func TestCategorize(t *testing.T) {
	now := time.Now()

	projects := []claude.ProjectInfo{
		{ShortName: "dirty-recent", GitDirty: true, LastActivity: now, DaysSinceActive: 0, PromptCount: 50},
		{ShortName: "feature-recent", GitBranch: "feature/x", LastActivity: now.Add(-24 * time.Hour), DaysSinceActive: 1, PromptCount: 30},
		{ShortName: "clean-recent", GitDirty: false, GitBranch: "main", LastActivity: now, DaysSinceActive: 0, PromptCount: 10},
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
	p := claude.ProjectInfo{
		GitDirty:        true,
		GitBranch:       "feature/x",
		PromptCount:     100,
		DaysSinceActive: 1,
		LastActivity:    time.Now().Add(-24 * time.Hour),
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
		LastActivity:    time.Now().Add(-10 * 24 * time.Hour),
	}

	cleanScore := Score(clean)
	if cleanScore >= score {
		t.Errorf("clean old project should score lower than dirty recent one: %f >= %f", cleanScore, score)
	}
}
