package git

import (
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

// RepoStatus holds the git status of a project directory.
type RepoStatus struct {
	IsRepo           bool   `json:"isRepo"`
	IsDirty          bool   `json:"isDirty"`
	Branch           string `json:"branch"`
	IsFeatureBranch  bool   `json:"isFeatureBranch"`
	UncommittedFiles int    `json:"uncommittedFiles"`
}

// CheckStatus checks the git status of a directory.
// Returns a zero-value RepoStatus with IsRepo=false if the directory is not a git repo.
func CheckStatus(path string) (RepoStatus, error) {
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		return RepoStatus{IsRepo: false}, nil
	}

	status := RepoStatus{IsRepo: true}

	head, err := repo.Head()
	if err == nil {
		ref := head.Name().Short()
		status.Branch = ref
		status.IsFeatureBranch = isFeatureBranch(ref)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return status, nil
	}

	ws, err := wt.Status()
	if err != nil {
		return status, nil
	}

	for _, s := range ws {
		if s.Worktree != gogit.Unmodified || s.Staging != gogit.Unmodified {
			status.UncommittedFiles++
		}
	}
	status.IsDirty = status.UncommittedFiles > 0

	return status, nil
}

func isFeatureBranch(name string) bool {
	lower := strings.ToLower(name)
	if lower == "main" || lower == "master" || lower == "develop" || lower == "dev" {
		return false
	}
	return true
}
