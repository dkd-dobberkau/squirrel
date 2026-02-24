package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
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
		count      int
		lastTS     int64
		firstTS    int64
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
			Path:            path,
			ShortName:       filepath.Base(path),
			PromptCount:     p.count,
			LastActivity:    time.UnixMilli(p.lastTS),
			FirstActivity:   time.UnixMilli(p.firstTS),
			LastPrompt:      p.lastPrompt,
			DaysSinceActive: int(time.Since(time.UnixMilli(p.lastTS)).Hours() / 24),
		})
	}

	return projects
}

// PromptsForProject filters history entries for a specific project path,
// sorted by timestamp descending, limited to max entries.
func PromptsForProject(entries []HistoryEntry, path string, max int) []HistoryEntry {
	var filtered []HistoryEntry
	for _, e := range entries {
		if e.Project == path {
			filtered = append(filtered, e)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp > filtered[j].Timestamp
	})

	if max > 0 && len(filtered) > max {
		filtered = filtered[:max]
	}

	return filtered
}
