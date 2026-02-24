package claude

import "strings"

// FindProject finds the best matching project for a query string.
// Matching priority: exact path > ShortName > path suffix > substring.
func FindProject(projects []ProjectInfo, query string) (ProjectInfo, bool) {
	// Exact path match
	for _, p := range projects {
		if p.Path == query {
			return p, true
		}
	}

	// ShortName match (case-insensitive)
	q := strings.ToLower(query)
	for _, p := range projects {
		if strings.ToLower(p.ShortName) == q {
			return p, true
		}
	}

	// Path suffix match (e.g. "local/squirrel" matches "/Users/olivier/Versioncontrol/local/squirrel")
	for _, p := range projects {
		if strings.HasSuffix(strings.ToLower(p.Path), "/"+q) {
			return p, true
		}
	}

	// Substring match on path (case-insensitive)
	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Path), q) {
			return p, true
		}
	}

	return ProjectInfo{}, false
}
