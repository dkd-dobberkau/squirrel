package analyzer

import (
	"math"
	"sort"

	"github.com/dkd-dobberkau/squirrel/internal/claude"
	gitpkg "github.com/dkd-dobberkau/squirrel/internal/git"
)

// CategorizedProjects holds projects sorted into categories.
type CategorizedProjects struct {
	OpenWork       []claude.ProjectInfo `json:"openWork"`
	RecentActivity []claude.ProjectInfo `json:"recentActivity"`
	Sleeping       []claude.ProjectInfo `json:"sleeping"`
	Acknowledged   []claude.ProjectInfo `json:"acknowledged"`
}

// Categorize sorts projects into open work, recent activity, sleeping, and acknowledged.
func Categorize(projects []claude.ProjectInfo, ackedPaths map[string]bool) CategorizedProjects {
	var result CategorizedProjects

	for _, p := range projects {
		p.Score = Score(p)
		p.IsOpenWork = isOpenWork(p)

		if ackedPaths[p.Path] {
			result.Acknowledged = append(result.Acknowledged, p)
			continue
		}

		switch {
		case p.IsOpenWork:
			result.OpenWork = append(result.OpenWork, p)
		case p.DaysSinceActive <= 3:
			result.RecentActivity = append(result.RecentActivity, p)
		default:
			result.Sleeping = append(result.Sleeping, p)
		}
	}

	sortByScore := func(s []claude.ProjectInfo) {
		sort.Slice(s, func(i, j int) bool {
			return s[i].Score > s[j].Score
		})
	}
	sortByScore(result.OpenWork)
	sortByScore(result.RecentActivity)
	sortByScore(result.Sleeping)
	sortByScore(result.Acknowledged)

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

// Score computes a priority score for a project.
func Score(p claude.ProjectInfo) float64 {
	score := 0.0

	recency := math.Exp(-float64(p.DaysSinceActive) * math.Ln2 / 3.0)
	score += recency * 50

	if p.PromptCount > 0 {
		score += math.Log2(float64(p.PromptCount)) * 5
	}

	if p.GitDirty {
		score += 30
	}

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
