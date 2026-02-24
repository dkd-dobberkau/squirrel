package claude

import "testing"

func TestFindProjectExactPath(t *testing.T) {
	projects := []ProjectInfo{
		{Path: "/Users/test/project-a", ShortName: "project-a"},
		{Path: "/Users/test/project-b", ShortName: "project-b"},
	}

	p, ok := FindProject(projects, "/Users/test/project-a")
	if !ok {
		t.Fatal("expected to find project by exact path")
	}
	if p.ShortName != "project-a" {
		t.Errorf("expected project-a, got %q", p.ShortName)
	}
}

func TestFindProjectByShortName(t *testing.T) {
	projects := []ProjectInfo{
		{Path: "/Users/test/project-a", ShortName: "project-a"},
		{Path: "/Users/test/squirrel", ShortName: "squirrel"},
	}

	p, ok := FindProject(projects, "squirrel")
	if !ok {
		t.Fatal("expected to find project by short name")
	}
	if p.Path != "/Users/test/squirrel" {
		t.Errorf("expected squirrel path, got %q", p.Path)
	}
}

func TestFindProjectByShortNameCaseInsensitive(t *testing.T) {
	projects := []ProjectInfo{
		{Path: "/Users/test/MyApp", ShortName: "MyApp"},
	}

	p, ok := FindProject(projects, "myapp")
	if !ok {
		t.Fatal("expected case-insensitive short name match")
	}
	if p.ShortName != "MyApp" {
		t.Errorf("expected MyApp, got %q", p.ShortName)
	}
}

func TestFindProjectByPathSuffix(t *testing.T) {
	projects := []ProjectInfo{
		{Path: "/Users/olivier/Versioncontrol/local/squirrel", ShortName: "squirrel"},
	}

	p, ok := FindProject(projects, "local/squirrel")
	if !ok {
		t.Fatal("expected to find project by path suffix")
	}
	if p.ShortName != "squirrel" {
		t.Errorf("expected squirrel, got %q", p.ShortName)
	}
}

func TestFindProjectBySubstring(t *testing.T) {
	projects := []ProjectInfo{
		{Path: "/Users/test/my-cool-project", ShortName: "my-cool-project"},
	}

	p, ok := FindProject(projects, "cool")
	if !ok {
		t.Fatal("expected to find project by substring")
	}
	if p.ShortName != "my-cool-project" {
		t.Errorf("expected my-cool-project, got %q", p.ShortName)
	}
}

func TestFindProjectNotFound(t *testing.T) {
	projects := []ProjectInfo{
		{Path: "/Users/test/project-a", ShortName: "project-a"},
	}

	_, ok := FindProject(projects, "nonexistent")
	if ok {
		t.Fatal("expected no match for nonexistent query")
	}
}

func TestFindProjectPriority(t *testing.T) {
	// ShortName match should take priority over substring match
	projects := []ProjectInfo{
		{Path: "/Users/test/foo-squirrel-bar", ShortName: "foo-squirrel-bar"},
		{Path: "/Users/test/squirrel", ShortName: "squirrel"},
	}

	p, ok := FindProject(projects, "squirrel")
	if !ok {
		t.Fatal("expected to find project")
	}
	// ShortName exact match should win over substring
	if p.Path != "/Users/test/squirrel" {
		t.Errorf("expected ShortName match to win, got path %q", p.Path)
	}
}
