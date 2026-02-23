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
// the corresponding sessions-index.json files from claudeProjectsDir.
func EnrichWithSessions(projects []ProjectInfo, claudeProjectsDir string) {
	for i := range projects {
		dirName := projectPathToDir(projects[i].Path)
		idxPath := filepath.Join(claudeProjectsDir, dirName, "sessions-index.json")

		idx, err := ParseSessionsIndex(idxPath)
		if err != nil {
			continue
		}

		projects[i].Sessions = idx.Entries

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
