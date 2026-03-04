package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInitGitRepo_CreatesRepoAndGitignore(t *testing.T) {
	dir := t.TempDir()

	err := InitGitRepo(dir)
	if err != nil {
		t.Fatalf("InitGitRepo failed: %v", err)
	}

	// .git directory exists
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		t.Fatal(".git directory was not created")
	}

	// .gitignore exists with correct content
	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}
	expected := "*\n!*.md\n!.gitignore\n"
	if string(content) != expected {
		t.Fatalf(".gitignore content = %q, want %q", string(content), expected)
	}

	// Initial commit exists
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %v\n%s", err, output)
	}
	if len(output) == 0 {
		t.Fatal("No commits found after InitGitRepo")
	}
}

func TestInitGitRepo_Idempotent(t *testing.T) {
	dir := t.TempDir()

	err := InitGitRepo(dir)
	if err != nil {
		t.Fatalf("First InitGitRepo failed: %v", err)
	}

	// Create a file and commit to verify second call doesn't reset
	os.WriteFile(filepath.Join(dir, "note.md"), []byte("hello"), 0644)
	CommitChanges(dir)

	err = InitGitRepo(dir)
	if err != nil {
		t.Fatalf("Second InitGitRepo failed: %v", err)
	}

	// Should have 2 commits (initial + the note commit)
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %v\n%s", err, output)
	}
	lines := 0
	for _, b := range output {
		if b == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Fatalf("Expected 2 commits, got %d:\n%s", lines, output)
	}
}

func TestCommitChanges_NewFiles(t *testing.T) {
	dir := t.TempDir()
	InitGitRepo(dir)

	os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test"), 0644)

	err := CommitChanges(dir)
	if err != nil {
		t.Fatalf("CommitChanges failed: %v", err)
	}

	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %v\n%s", err, output)
	}
	lines := 0
	for _, b := range output {
		if b == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Fatalf("Expected 2 commits (initial + new file), got %d:\n%s", lines, output)
	}
}

func TestCommitChanges_NoChanges(t *testing.T) {
	dir := t.TempDir()
	InitGitRepo(dir)

	err := CommitChanges(dir)
	if err != nil {
		t.Fatalf("CommitChanges with no changes should return nil, got: %v", err)
	}
}

func TestCommitChanges_ModifiedFiles(t *testing.T) {
	dir := t.TempDir()
	InitGitRepo(dir)

	notePath := filepath.Join(dir, "note.md")
	os.WriteFile(notePath, []byte("# Original"), 0644)
	CommitChanges(dir)

	os.WriteFile(notePath, []byte("# Modified"), 0644)

	err := CommitChanges(dir)
	if err != nil {
		t.Fatalf("CommitChanges on modified file failed: %v", err)
	}

	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %v\n%s", err, output)
	}
	lines := 0
	for _, b := range output {
		if b == '\n' {
			lines++
		}
	}
	if lines != 3 {
		t.Fatalf("Expected 3 commits, got %d:\n%s", lines, output)
	}
}

func TestCommitChanges_DeletedFiles(t *testing.T) {
	dir := t.TempDir()
	InitGitRepo(dir)

	notePath := filepath.Join(dir, "to-delete.md")
	os.WriteFile(notePath, []byte("# Delete me"), 0644)
	CommitChanges(dir)

	os.Remove(notePath)

	err := CommitChanges(dir)
	if err != nil {
		t.Fatalf("CommitChanges on deleted file failed: %v", err)
	}

	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log failed: %v\n%s", err, output)
	}
	lines := 0
	for _, b := range output {
		if b == '\n' {
			lines++
		}
	}
	if lines != 3 {
		t.Fatalf("Expected 3 commits, got %d:\n%s", lines, output)
	}
}

func TestGitignore_ExcludesNonMdFiles(t *testing.T) {
	dir := t.TempDir()
	InitGitRepo(dir)

	// Create tracked and untracked files
	os.WriteFile(filepath.Join(dir, "note.md"), []byte("# Note"), 0644)
	os.WriteFile(filepath.Join(dir, ".DS_Store"), []byte("junk"), 0644)
	os.WriteFile(filepath.Join(dir, "drawing.excalidraw"), []byte("{}"), 0644)

	CommitChanges(dir)

	// Check that only .md and .gitignore are tracked
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git ls-files failed: %v\n%s", err, output)
	}

	tracked := string(output)
	if !contains(tracked, "note.md") {
		t.Fatal("note.md should be tracked")
	}
	if !contains(tracked, ".gitignore") {
		t.Fatal(".gitignore should be tracked")
	}
	if contains(tracked, ".DS_Store") {
		t.Fatal(".DS_Store should NOT be tracked")
	}
	if contains(tracked, "drawing.excalidraw") {
		t.Fatal(".excalidraw files should NOT be tracked")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
