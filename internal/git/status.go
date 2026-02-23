package git

import (
	"os/exec"
	"strings"
)

// RepoStatus holds the git status of a project directory.
type RepoStatus struct {
	IsRepo           bool   `json:"isRepo"`
	IsDirty          bool   `json:"isDirty"`
	Branch           string `json:"branch"`
	IsFeatureBranch  bool   `json:"isFeatureBranch"`
	UncommittedFiles int    `json:"uncommittedFiles"`
}

// CheckStatus checks the git status of a directory using native git commands.
// This respects .gitignore, .git/info/exclude, and the global gitignore.
// Returns a zero-value RepoStatus with IsRepo=false if the directory is not a git repo.
func CheckStatus(path string) (RepoStatus, error) {
	if !isGitRepo(path) {
		return RepoStatus{IsRepo: false}, nil
	}

	status := RepoStatus{IsRepo: true}

	if branch, err := gitCommand(path, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		status.Branch = strings.TrimSpace(branch)
		status.IsFeatureBranch = isFeatureBranch(status.Branch)
	}

	if porcelain, err := gitCommand(path, "status", "--porcelain"); err == nil {
		lines := strings.Split(strings.TrimSpace(porcelain), "\n")
		for _, line := range lines {
			if line != "" {
				status.UncommittedFiles++
			}
		}
		status.IsDirty = status.UncommittedFiles > 0
	}

	return status, nil
}

func isGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func gitCommand(path string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", path}, args...)
	out, err := exec.Command("git", fullArgs...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func isFeatureBranch(name string) bool {
	lower := strings.ToLower(name)
	if lower == "main" || lower == "master" || lower == "develop" || lower == "dev" {
		return false
	}
	return true
}
