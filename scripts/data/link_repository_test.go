package data

import (
	"cli-notes/scripts"
	"os"
	"strings"
	"testing"
	"time"
)

// ============================================
// ParseLinks Tests
// ============================================

func TestParseLinks_ExtractsSingleLink(t *testing.T) {
	content := "Some text with [[My Note]] link"
	links := ParseLinks(content)

	if len(links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(links))
	}
	if len(links) > 0 && links[0] != "My Note" {
		t.Errorf("Expected 'My Note', got '%s'", links[0])
	}
}

func TestParseLinks_ExtractsMultipleLinks(t *testing.T) {
	content := "See [[Note A]] and also [[Note B]] for details"
	links := ParseLinks(content)

	if len(links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(links))
	}
	if len(links) >= 2 {
		if links[0] != "Note A" {
			t.Errorf("Expected first link 'Note A', got '%s'", links[0])
		}
		if links[1] != "Note B" {
			t.Errorf("Expected second link 'Note B', got '%s'", links[1])
		}
	}
}

func TestParseLinks_DeduplicatesLinks(t *testing.T) {
	content := "[[Same Note]] appears twice [[Same Note]]"
	links := ParseLinks(content)

	if len(links) != 1 {
		t.Errorf("Expected 1 unique link, got %d", len(links))
	}
}

func TestParseLinks_HandlesEmptyContent(t *testing.T) {
	links := ParseLinks("")

	if len(links) != 0 {
		t.Errorf("Expected 0 links for empty content, got %d", len(links))
	}
}

func TestParseLinks_NoLinksInContent(t *testing.T) {
	content := "Just regular text without any links"
	links := ParseLinks(content)

	if len(links) != 0 {
		t.Errorf("Expected 0 links, got %d", len(links))
	}
}

func TestParseLinks_IgnoresMalformedLinks(t *testing.T) {
	content := "This has [[ unclosed bracket and [single brackets]"
	links := ParseLinks(content)

	if len(links) != 0 {
		t.Errorf("Expected 0 valid links, got %d", len(links))
	}
}

func TestParseLinks_HandlesWhitespaceInLinks(t *testing.T) {
	content := "[[  Spaced Note  ]] should trim"
	links := ParseLinks(content)

	if len(links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(links))
	}
	if len(links) > 0 && links[0] != "Spaced Note" {
		t.Errorf("Expected 'Spaced Note', got '%s'", links[0])
	}
}

// ============================================
// ResolveLink Tests
// ============================================

func TestResolveLink_ByTitle(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "my-important-note.md",
		Title:     "My Important Note",
		CreatedAt: time.Now(),
		Content:   "# My Important Note\n\nContent here",
	})

	resolved, err := ResolveLink("My Important Note")
	if err != nil {
		t.Fatalf("ResolveLink failed: %v", err)
	}

	if resolved == nil {
		t.Fatal("Expected to resolve link, got nil")
	}

	if resolved.Name != "my-important-note.md" {
		t.Errorf("Expected 'my-important-note.md', got '%s'", resolved.Name)
	}
}

func TestResolveLink_ByFilename(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "project-notes.md",
		Title:     "Some Title",
		CreatedAt: time.Now(),
		Content:   "Content",
	})

	resolved, err := ResolveLink("project-notes")
	if err != nil {
		t.Fatalf("ResolveLink failed: %v", err)
	}

	if resolved == nil {
		t.Fatal("Expected to resolve link by filename, got nil")
	}

	if resolved.Name != "project-notes.md" {
		t.Errorf("Expected 'project-notes.md', got '%s'", resolved.Name)
	}
}

func TestResolveLink_CaseInsensitive(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "test.md",
		Title:     "My Test Note",
		CreatedAt: time.Now(),
		Content:   "Content",
	})

	// Test lowercase
	resolved, err := ResolveLink("my test note")
	if err != nil {
		t.Fatalf("ResolveLink failed: %v", err)
	}

	if resolved == nil {
		t.Error("Expected case-insensitive match, got nil")
	}
}

func TestResolveLink_NotFound(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	resolved, err := ResolveLink("Non Existent Note")
	if err != nil {
		t.Fatalf("ResolveLink failed with error: %v", err)
	}

	if resolved != nil {
		t.Error("Expected nil for non-existent note")
	}
}

// ============================================
// GetLinksFrom Tests
// ============================================

func TestGetLinksFrom_ReturnsLinkedFiles(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create target note
	createTestFile(t, scripts.File{
		Name:      "target.md",
		Title:     "Target Note",
		CreatedAt: time.Now(),
		Content:   "# Target Note",
	})

	// Create source note with link
	createTestFile(t, scripts.File{
		Name:      "source.md",
		Title:     "Source Note",
		CreatedAt: time.Now(),
		Content:   "# Source Note\n\nSee [[Target Note]] for details",
	})

	links, err := GetLinksFrom("source.md")
	if err != nil {
		t.Fatalf("GetLinksFrom failed: %v", err)
	}

	if len(links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(links))
	}

	if len(links) > 0 && links[0].Name != "target.md" {
		t.Errorf("Expected 'target.md', got '%s'", links[0].Name)
	}
}

func TestGetLinksFrom_SkipsUnresolvedLinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create source with dead link
	createTestFile(t, scripts.File{
		Name:      "source.md",
		Title:     "Source",
		CreatedAt: time.Now(),
		Content:   "[[Non Existent]] is dead link",
	})

	links, err := GetLinksFrom("source.md")
	if err != nil {
		t.Fatalf("GetLinksFrom failed: %v", err)
	}

	if len(links) != 0 {
		t.Errorf("Expected 0 links (dead links skipped), got %d", len(links))
	}
}

func TestGetLinksFrom_EmptyWhenNoLinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "no-links.md",
		Title:     "No Links",
		CreatedAt: time.Now(),
		Content:   "Just plain text",
	})

	links, err := GetLinksFrom("no-links.md")
	if err != nil {
		t.Fatalf("GetLinksFrom failed: %v", err)
	}

	if len(links) != 0 {
		t.Errorf("Expected 0 links, got %d", len(links))
	}
}

// ============================================
// GetBacklinks Tests
// ============================================

func TestGetBacklinks_FindsReferencingNotes(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create target note
	createTestFile(t, scripts.File{
		Name:      "target.md",
		Title:     "Target Note",
		CreatedAt: time.Now(),
		Content:   "# Target Note",
	})

	// Create source that links to target
	createTestFile(t, scripts.File{
		Name:      "source.md",
		Title:     "Source Note",
		CreatedAt: time.Now(),
		Content:   "See [[Target Note]]",
	})

	backlinks, err := GetBacklinks("target.md")
	if err != nil {
		t.Fatalf("GetBacklinks failed: %v", err)
	}

	if len(backlinks) != 1 {
		t.Errorf("Expected 1 backlink, got %d", len(backlinks))
	}

	if len(backlinks) > 0 && backlinks[0].Name != "source.md" {
		t.Errorf("Expected 'source.md', got '%s'", backlinks[0].Name)
	}
}

func TestGetBacklinks_MultipleBacklinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create target
	createTestFile(t, scripts.File{
		Name:      "target.md",
		Title:     "Target",
		CreatedAt: time.Now(),
		Content:   "# Target",
	})

	// Create multiple sources
	createTestFile(t, scripts.File{
		Name:      "source1.md",
		Title:     "Source 1",
		CreatedAt: time.Now(),
		Content:   "[[Target]]",
	})

	createTestFile(t, scripts.File{
		Name:      "source2.md",
		Title:     "Source 2",
		CreatedAt: time.Now(),
		Content:   "Also see [[Target]]",
	})

	backlinks, err := GetBacklinks("target.md")
	if err != nil {
		t.Fatalf("GetBacklinks failed: %v", err)
	}

	if len(backlinks) != 2 {
		t.Errorf("Expected 2 backlinks, got %d", len(backlinks))
	}
}

func TestGetBacklinks_NoBacklinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "lonely.md",
		Title:     "Lonely Note",
		CreatedAt: time.Now(),
		Content:   "No one links to me",
	})

	backlinks, err := GetBacklinks("lonely.md")
	if err != nil {
		t.Fatalf("GetBacklinks failed: %v", err)
	}

	if len(backlinks) != 0 {
		t.Errorf("Expected 0 backlinks, got %d", len(backlinks))
	}
}

// ============================================
// BuildLinkIndex Tests
// ============================================

func TestBuildLinkIndex_BuildsCompleteGraph(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create interconnected notes: A -> B -> C, A -> C
	createTestFile(t, scripts.File{
		Name:      "a.md",
		Title:     "A",
		CreatedAt: time.Now(),
		Content:   "[[B]] and [[C]]",
	})

	createTestFile(t, scripts.File{
		Name:      "b.md",
		Title:     "B",
		CreatedAt: time.Now(),
		Content:   "[[C]]",
	})

	createTestFile(t, scripts.File{
		Name:      "c.md",
		Title:     "C",
		CreatedAt: time.Now(),
		Content:   "End node",
	})

	index, err := BuildLinkIndex()
	if err != nil {
		t.Fatalf("BuildLinkIndex failed: %v", err)
	}

	// Check outlinks
	if len(index.OutLinks["a.md"]) != 2 {
		t.Errorf("Expected A to have 2 outlinks, got %d", len(index.OutLinks["a.md"]))
	}

	if len(index.OutLinks["b.md"]) != 1 {
		t.Errorf("Expected B to have 1 outlink, got %d", len(index.OutLinks["b.md"]))
	}

	// Check inlinks (backlinks)
	if len(index.InLinks["c.md"]) != 2 {
		t.Errorf("Expected C to have 2 backlinks, got %d", len(index.InLinks["c.md"]))
	}

	if len(index.InLinks["b.md"]) != 1 {
		t.Errorf("Expected B to have 1 backlink, got %d", len(index.InLinks["b.md"]))
	}
}

func TestBuildLinkIndex_HandlesCircularLinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create circular: A -> B -> A
	createTestFile(t, scripts.File{
		Name:      "a.md",
		Title:     "A",
		CreatedAt: time.Now(),
		Content:   "[[B]]",
	})

	createTestFile(t, scripts.File{
		Name:      "b.md",
		Title:     "B",
		CreatedAt: time.Now(),
		Content:   "[[A]]",
	})

	index, err := BuildLinkIndex()
	if err != nil {
		t.Fatalf("BuildLinkIndex failed: %v", err)
	}

	// Both should have 1 outlink and 1 inlink
	if len(index.OutLinks["a.md"]) != 1 {
		t.Errorf("Expected A to have 1 outlink, got %d", len(index.OutLinks["a.md"]))
	}

	if len(index.InLinks["a.md"]) != 1 {
		t.Errorf("Expected A to have 1 backlink, got %d", len(index.InLinks["a.md"]))
	}
}

// ============================================
// GetUnresolvedLinks Tests
// ============================================

func TestGetUnresolvedLinks_FindsDeadLinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	// Create note with both valid and dead links
	createTestFile(t, scripts.File{
		Name:      "exists.md",
		Title:     "Exists",
		CreatedAt: time.Now(),
		Content:   "I exist",
	})

	createTestFile(t, scripts.File{
		Name:      "source.md",
		Title:     "Source",
		CreatedAt: time.Now(),
		Content:   "[[Exists]] and [[Missing]] and [[Also Missing]]",
	})

	unresolved, err := GetUnresolvedLinks("source.md")
	if err != nil {
		t.Fatalf("GetUnresolvedLinks failed: %v", err)
	}

	if len(unresolved) != 2 {
		t.Errorf("Expected 2 unresolved links, got %d", len(unresolved))
	}
}

func TestGetUnresolvedLinks_EmptyWhenAllResolved(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "target.md",
		Title:     "Target",
		CreatedAt: time.Now(),
		Content:   "Target content",
	})

	createTestFile(t, scripts.File{
		Name:      "source.md",
		Title:     "Source",
		CreatedAt: time.Now(),
		Content:   "[[Target]]",
	})

	unresolved, err := GetUnresolvedLinks("source.md")
	if err != nil {
		t.Fatalf("GetUnresolvedLinks failed: %v", err)
	}

	if len(unresolved) != 0 {
		t.Errorf("Expected 0 unresolved, got %d", len(unresolved))
	}
}

// ============================================
// InsertLinkAtTop Tests
// ============================================

func TestInsertLinkAtTop_AddsNewLink(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note.md",
		Title:     "My Note",
		CreatedAt: time.Now(),
		Content:   "# My Note\n\nContent here",
	})

	err := InsertLinkAtTop("note.md", "Other Note")
	if err != nil {
		t.Fatalf("InsertLinkAtTop failed: %v", err)
	}

	// Reload and check
	file, err := LoadFileByName("note.md")
	if err != nil {
		t.Fatalf("LoadFileByName failed: %v", err)
	}

	if !strings.Contains(file.Content, "[[Other Note]]") {
		t.Error("Expected content to contain [[Other Note]]")
	}
}

func TestInsertLinkAtTop_AppendsToExistingLinks(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note.md",
		Title:     "My Note",
		CreatedAt: time.Now(),
		Content:   "[[First Link]]\n\n# My Note",
	})

	err := InsertLinkAtTop("note.md", "Second Link")
	if err != nil {
		t.Fatalf("InsertLinkAtTop failed: %v", err)
	}

	file, err := LoadFileByName("note.md")
	if err != nil {
		t.Fatalf("LoadFileByName failed: %v", err)
	}

	if !strings.Contains(file.Content, "[[First Link]]") {
		t.Error("Expected content to still contain [[First Link]]")
	}

	if !strings.Contains(file.Content, "[[Second Link]]") {
		t.Error("Expected content to contain [[Second Link]]")
	}
}

func TestInsertLinkAtTop_SkipsDuplicate(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	createTestFile(t, scripts.File{
		Name:      "note.md",
		Title:     "My Note",
		CreatedAt: time.Now(),
		Content:   "[[Existing Link]]\n\n# My Note",
	})

	err := InsertLinkAtTop("note.md", "Existing Link")
	if err != nil {
		t.Fatalf("InsertLinkAtTop failed: %v", err)
	}

	file, err := LoadFileByName("note.md")
	if err != nil {
		t.Fatalf("LoadFileByName failed: %v", err)
	}

	// Count occurrences - should only be 1
	count := strings.Count(file.Content, "[[Existing Link]]")
	if count != 1 {
		t.Errorf("Expected exactly 1 occurrence of link, got %d", count)
	}
}

// ============================================
// CreateNoteFromDeadLink Tests
// ============================================

func TestCreateNoteFromDeadLink_CreatesNewNote(t *testing.T) {
	th := setupTest(t)
	defer th.cleanup(t)

	newFile, err := CreateNoteFromDeadLink("My New Topic")
	if err != nil {
		t.Fatalf("CreateNoteFromDeadLink failed: %v", err)
	}

	if newFile == nil {
		t.Fatal("Expected new file, got nil")
	}

	if newFile.Title != "My New Topic" {
		t.Errorf("Expected title 'My New Topic', got '%s'", newFile.Title)
	}

	// Verify file exists on disk
	_, err = os.Stat("notes/" + newFile.Name)
	if os.IsNotExist(err) {
		t.Error("Expected file to exist on disk")
	}
}
