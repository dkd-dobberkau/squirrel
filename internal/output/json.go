package output

import (
	"encoding/json"

	"github.com/dkd-dobberkau/squirrel/internal/analyzer"
	"github.com/dkd-dobberkau/squirrel/internal/claude"
)

// RenderJSON returns the categorized projects as a JSON string.
func RenderJSON(data analyzer.CategorizedProjects) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ProjectDetail holds a single project with its recent prompts for detail output.
type ProjectDetail struct {
	Project       claude.ProjectInfo   `json:"project"`
	RecentPrompts []claude.HistoryEntry `json:"recentPrompts"`
}

// RenderProjectDetailJSON returns the project detail as a JSON string.
func RenderProjectDetailJSON(detail ProjectDetail) (string, error) {
	b, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
