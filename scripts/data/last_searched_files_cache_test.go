package data

import (
	"cli-notes/scripts"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func createTestFiles(count int) []scripts.File {
	files := make([]scripts.File, count)
	for i := 0; i < count; i++ {
		files[i] = scripts.File{
			Name:      fmt.Sprintf("test-file-%d.md", i),
			Title:     fmt.Sprintf("Test File %d", i),
			Tags:      []string{"test"},
			CreatedAt: time.Now(),
		}
	}
	return files
}

func TestNewSearchedFilesStore(t *testing.T) {
	store := NewSearchedFilesStore()
	
	if store == nil {
		t.Fatal("NewSearchedFilesStore returned nil")
	}
	
	if len(store.filesThatHaveBeenSearched) != 0 {
		t.Errorf("Expected empty files slice, got %d items", len(store.filesThatHaveBeenSearched))
	}
	
	if store.selectedIndex != -1 {
		t.Errorf("Expected selectedIndex to be -1, got %d", store.selectedIndex)
	}
}

func TestSetAndGetFilesSearched(t *testing.T) {
	store := NewSearchedFilesStore()
	testFiles := createTestFiles(3)
	
	store.SetFilesSearched(testFiles)
	
	if !reflect.DeepEqual(store.GetFilesSearched(), testFiles) {
		t.Error("GetFilesSearched did not return the same files that were set")
	}
	
	if store.selectedIndex != -1 {
		t.Errorf("Expected selectedIndex to be reset to -1 after SetFilesSearched, got %d", store.selectedIndex)
	}
}

func TestGetNextFileEmptyList(t *testing.T) {
	store := NewSearchedFilesStore()
	
	file := store.GetNextFile()
	
	if file != nil {
		t.Errorf("Expected nil when getting next file from empty list, got %v", file)
	}
	
	if store.selectedIndex != -1 {
		t.Errorf("Expected selectedIndex to remain -1, got %d", store.selectedIndex)
	}
}

func TestGetPreviousFileEmptyList(t *testing.T) {
	store := NewSearchedFilesStore()
	
	file := store.GetPreviousFile()
	
	if file != nil {
		t.Errorf("Expected nil when getting previous file from empty list, got %v", file)
	}
	
	if store.selectedIndex != -1 {
		t.Errorf("Expected selectedIndex to remain -1, got %d", store.selectedIndex)
	}
}

func TestGetNextFileSingleFile(t *testing.T) {
	store := NewSearchedFilesStore()
	testFiles := createTestFiles(1)
	store.SetFilesSearched(testFiles)
	
	// First call should return the single file
	file := store.GetNextFile()
	
	if file == nil {
		t.Fatal("Expected file, got nil")
	}
	
	if file.Name != testFiles[0].Name {
		t.Errorf("Expected file name %s, got %s", testFiles[0].Name, file.Name)
	}
	
	if store.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0, got %d", store.selectedIndex)
	}
	
	// Second call should return the same file (no further navigation)
	file = store.GetNextFile()
	
	if file == nil {
		t.Fatal("Expected file on second call, got nil")
	}
	
	if file.Name != testFiles[0].Name {
		t.Errorf("Expected file name %s on second call, got %s", testFiles[0].Name, file.Name)
	}
	
	if store.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to remain 0, got %d", store.selectedIndex)
	}
}

func TestGetPreviousFileSingleFile(t *testing.T) {
	store := NewSearchedFilesStore()
	testFiles := createTestFiles(1)
	store.SetFilesSearched(testFiles)
	
	// First call should return the single file
	file := store.GetPreviousFile()
	
	if file == nil {
		t.Fatal("Expected file, got nil")
	}
	
	if file.Name != testFiles[0].Name {
		t.Errorf("Expected file name %s, got %s", testFiles[0].Name, file.Name)
	}
	
	if store.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0, got %d", store.selectedIndex)
	}
	
	// Second call should return the same file (no further navigation)
	file = store.GetPreviousFile()
	
	if file == nil {
		t.Fatal("Expected file on second call, got nil")
	}
	
	if file.Name != testFiles[0].Name {
		t.Errorf("Expected file name %s on second call, got %s", testFiles[0].Name, file.Name)
	}
	
	if store.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to remain 0, got %d", store.selectedIndex)
	}
}

func TestGetNextFileMultipleFiles(t *testing.T) {
	store := NewSearchedFilesStore()
	testFiles := createTestFiles(3)
	store.SetFilesSearched(testFiles)
	
	// First call should return the last file
	file := store.GetNextFile()
	
	if file == nil {
		t.Fatal("Expected file, got nil")
	}
	
	if file.Name != testFiles[2].Name {
		t.Errorf("Expected last file name %s, got %s", testFiles[2].Name, file.Name)
	}
	
	if store.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex to be 2, got %d", store.selectedIndex)
	}
	
	// Second call should return the middle file
	file = store.GetNextFile()
	
	if file == nil {
		t.Fatal("Expected file on second call, got nil")
	}
	
	if file.Name != testFiles[1].Name {
		t.Errorf("Expected middle file name %s, got %s", testFiles[1].Name, file.Name)
	}
	
	if store.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1, got %d", store.selectedIndex)
	}
	
	// Third call should return the first file
	file = store.GetNextFile()
	
	if file == nil {
		t.Fatal("Expected file on third call, got nil")
	}
	
	if file.Name != testFiles[0].Name {
		t.Errorf("Expected first file name %s, got %s", testFiles[0].Name, file.Name)
	}
	
	if store.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0, got %d", store.selectedIndex)
	}
	
	// Fourth call should return the first file again (boundary case)
	file = store.GetNextFile()
	
	if file == nil {
		t.Fatal("Expected file on fourth call, got nil")
	}
	
	if file.Name != testFiles[0].Name {
		t.Errorf("Expected first file name again %s, got %s", testFiles[0].Name, file.Name)
	}
	
	if store.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to remain 0, got %d", store.selectedIndex)
	}
}

func TestGetPreviousFileMultipleFiles(t *testing.T) {
	store := NewSearchedFilesStore()
	testFiles := createTestFiles(3)
	store.SetFilesSearched(testFiles)
	
	// First call should return the first file
	file := store.GetPreviousFile()
	
	if file == nil {
		t.Fatal("Expected file, got nil")
	}
	
	if file.Name != testFiles[0].Name {
		t.Errorf("Expected first file name %s, got %s", testFiles[0].Name, file.Name)
	}
	
	if store.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0, got %d", store.selectedIndex)
	}
	
	// Second call should return the middle file
	file = store.GetPreviousFile()
	
	if file == nil {
		t.Fatal("Expected file on second call, got nil")
	}
	
	if file.Name != testFiles[1].Name {
		t.Errorf("Expected middle file name %s, got %s", testFiles[1].Name, file.Name)
	}
	
	if store.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1, got %d", store.selectedIndex)
	}
	
	// Third call should return the last file
	file = store.GetPreviousFile()
	
	if file == nil {
		t.Fatal("Expected file on third call, got nil")
	}
	
	if file.Name != testFiles[2].Name {
		t.Errorf("Expected last file name %s, got %s", testFiles[2].Name, file.Name)
	}
	
	if store.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex to be 2, got %d", store.selectedIndex)
	}
	
	// Fourth call should return the last file again (boundary case)
	file = store.GetPreviousFile()
	
	if file == nil {
		t.Fatal("Expected file on fourth call, got nil")
	}
	
	if file.Name != testFiles[2].Name {
		t.Errorf("Expected last file name again %s, got %s", testFiles[2].Name, file.Name)
	}
	
	if store.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex to remain 2, got %d", store.selectedIndex)
	}
}

func TestNavigationCombo(t *testing.T) {
	store := NewSearchedFilesStore()
	testFiles := createTestFiles(3)
	store.SetFilesSearched(testFiles)
	
	// Start with GetNextFile
	file := store.GetNextFile() // Gets last file (index 2)
	if file.Name != testFiles[2].Name {
		t.Errorf("Expected file name %s, got %s", testFiles[2].Name, file.Name)
	}
	
	// Then call GetPreviousFile
	file = store.GetPreviousFile() // Should move to index 3, but boundary is 2
	if file.Name != testFiles[2].Name {
		t.Errorf("Expected last file again %s, got %s", testFiles[2].Name, file.Name)
	}
	
	// Now call GetNextFile twice
	file = store.GetNextFile() // Move to index 1
	file = store.GetNextFile() // Move to index 0
	if file.Name != testFiles[0].Name {
		t.Errorf("Expected first file %s, got %s", testFiles[0].Name, file.Name)
	}
	
	// Call GetPreviousFile
	file = store.GetPreviousFile() // Move to index 1
	if file.Name != testFiles[1].Name {
		t.Errorf("Expected middle file %s, got %s", testFiles[1].Name, file.Name)
	}
}

func TestSetFilesSearchedResetsIndex(t *testing.T) {
	store := NewSearchedFilesStore()
	
	// Set initial files and navigate
	testFiles1 := createTestFiles(3)
	store.SetFilesSearched(testFiles1)
	store.GetNextFile() // Move to index 2
	
	if store.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex to be 2, got %d", store.selectedIndex)
	}
	
	// Set new files and verify index resets
	testFiles2 := createTestFiles(2)
	store.SetFilesSearched(testFiles2)
	
	if store.selectedIndex != -1 {
		t.Errorf("Expected selectedIndex to reset to -1, got %d", store.selectedIndex)
	}
	
	// Verify the new files were set
	if !reflect.DeepEqual(store.GetFilesSearched(), testFiles2) {
		t.Error("Files were not updated correctly")
	}
} 