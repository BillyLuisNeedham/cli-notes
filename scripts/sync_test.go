package scripts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSyncCopiesMdFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source .md files
	os.WriteFile(filepath.Join(srcDir, "note1.md"), []byte("# Note 1"), 0644)
	os.WriteFile(filepath.Join(srcDir, "note2.md"), []byte("# Note 2"), 0644)

	copied, err := SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if copied != 2 {
		t.Errorf("expected 2 files copied, got %d", copied)
	}

	// Verify files exist in destination
	for _, name := range []string{"note1.md", "note2.md"} {
		content, err := os.ReadFile(filepath.Join(dstDir, name))
		if err != nil {
			t.Errorf("expected %s to exist in destination: %v", name, err)
		}
		if name == "note1.md" && string(content) != "# Note 1" {
			t.Errorf("expected '# Note 1', got %q", string(content))
		}
	}
}

func TestSyncSkipsUnchangedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "note.md")
	os.WriteFile(srcPath, []byte("content"), 0644)

	// First sync
	_, err := SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second sync should skip the file (same size, mod time >= source)
	copied, err := SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if copied != 0 {
		t.Errorf("expected 0 files copied on second sync, got %d", copied)
	}
}

func TestSyncCopiesModifiedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "note.md")
	os.WriteFile(srcPath, []byte("original"), 0644)

	// First sync
	_, err := SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify source file (different size triggers copy)
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(srcPath, []byte("modified content"), 0644)

	// Second sync should copy the modified file
	copied, err := SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if copied != 1 {
		t.Errorf("expected 1 file copied, got %d", copied)
	}

	content, _ := os.ReadFile(filepath.Join(dstDir, "note.md"))
	if string(content) != "modified content" {
		t.Errorf("expected 'modified content', got %q", string(content))
	}
}

func TestSyncRemovesStaleFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source and sync
	os.WriteFile(filepath.Join(srcDir, "keep.md"), []byte("keep"), 0644)
	os.WriteFile(filepath.Join(srcDir, "remove.md"), []byte("remove"), 0644)

	_, err := SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete source file
	os.Remove(filepath.Join(srcDir, "remove.md"))

	// Sync again - should remove stale file from destination
	_, err = SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dstDir, "remove.md")); !os.IsNotExist(err) {
		t.Error("expected remove.md to be deleted from destination")
	}

	if _, err := os.Stat(filepath.Join(dstDir, "keep.md")); err != nil {
		t.Error("expected keep.md to still exist in destination")
	}
}

func TestSyncCreatesDestinationDirectory(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "nested", "backup")

	os.WriteFile(filepath.Join(srcDir, "note.md"), []byte("content"), 0644)

	copied, err := SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if copied != 1 {
		t.Errorf("expected 1 file copied, got %d", copied)
	}

	if _, err := os.Stat(filepath.Join(dstDir, "note.md")); err != nil {
		t.Error("expected note.md to exist in nested destination")
	}
}

func TestSyncIgnoresNonMdFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	os.WriteFile(filepath.Join(srcDir, "note.md"), []byte("note"), 0644)
	os.WriteFile(filepath.Join(srcDir, "image.png"), []byte("png"), 0644)
	os.WriteFile(filepath.Join(srcDir, "data.json"), []byte("{}"), 0644)

	copied, err := SyncNotesToBackup(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if copied != 1 {
		t.Errorf("expected 1 file copied, got %d", copied)
	}

	if _, err := os.Stat(filepath.Join(dstDir, "image.png")); !os.IsNotExist(err) {
		t.Error("expected image.png to not be copied")
	}

	if _, err := os.Stat(filepath.Join(dstDir, "data.json")); !os.IsNotExist(err) {
		t.Error("expected data.json to not be copied")
	}
}
