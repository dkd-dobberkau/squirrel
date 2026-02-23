package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCheckStatus_CleanRepo(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Run()
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	run("add", ".")
	run("commit", "-m", "init")

	status, err := CheckStatus(dir)
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if status.IsDirty {
		t.Error("expected clean repo, got dirty")
	}

	if status.Branch != "main" && status.Branch != "master" {
		t.Errorf("expected main or master branch, got %q", status.Branch)
	}

	if status.UncommittedFiles != 0 {
		t.Errorf("expected 0 uncommitted files, got %d", status.UncommittedFiles)
	}
}

func TestCheckStatus_DirtyRepo(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Run()
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	run("add", ".")
	run("commit", "-m", "init")

	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("changed"), 0644)
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0644)

	status, err := CheckStatus(dir)
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if !status.IsDirty {
		t.Error("expected dirty repo, got clean")
	}

	if status.UncommittedFiles < 1 {
		t.Errorf("expected at least 1 uncommitted file, got %d", status.UncommittedFiles)
	}
}

func TestCheckStatus_FeatureBranch(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Run()
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	run("add", ".")
	run("commit", "-m", "init")
	run("checkout", "-b", "feature/cool")

	status, err := CheckStatus(dir)
	if err != nil {
		t.Fatalf("CheckStatus failed: %v", err)
	}

	if status.Branch != "feature/cool" {
		t.Errorf("expected branch 'feature/cool', got %q", status.Branch)
	}

	if !status.IsFeatureBranch {
		t.Error("expected feature branch detection")
	}
}

func TestCheckStatus_NotARepo(t *testing.T) {
	dir := t.TempDir()

	status, err := CheckStatus(dir)
	if err != nil {
		t.Fatalf("CheckStatus should not error for non-repo, got: %v", err)
	}

	if status.IsRepo {
		t.Error("expected IsRepo=false for non-git directory")
	}
}
