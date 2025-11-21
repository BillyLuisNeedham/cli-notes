package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents the metadata in a note file
type Frontmatter struct {
	Title         string   `yaml:"title"`
	DateCreated   string   `yaml:"date-created"`
	Tags          []string `yaml:"tags"`
	DateDue       string   `yaml:"date-due,omitempty"`
	Done          bool     `yaml:"done,omitempty"`
	Priority      int      `yaml:"priority,omitempty"`
	ObjectiveID   string   `yaml:"objective-id,omitempty"`
	ObjectiveRole string   `yaml:"objective-role,omitempty"`
}

// CreateTestFile creates a file with the given name and content in the harness notes directory
func (h *TestHarness) CreateTestFile(filename string, content string) {
	path := filepath.Join(h.NotesDir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		h.t.Fatalf("Failed to create test file %s: %v", filename, err)
	}
}

// CreateTodo creates a todo file with frontmatter
func (h *TestHarness) CreateTodo(filename, title string, tags []string, due string, done bool, priority int) {
	fm := Frontmatter{
		Title:       title,
		DateCreated: time.Now().Format("2006-01-02"),
		Tags:        tags,
		DateDue:     due,
		Done:        done,
		Priority:    priority,
	}

	h.createFileWithFrontmatter(filename, fm, "Todo content")
}

func (h *TestHarness) createFileWithFrontmatter(filename string, fm Frontmatter, content string) {
	// Manually format frontmatter to avoid YAML marshaling issues
	// (yaml.Marshal adds quotes around strings, breaking date parsing in CLI)
	var frontmatter strings.Builder
	frontmatter.WriteString("---\n")
	frontmatter.WriteString(fmt.Sprintf("title: %s\n", fm.Title))
	frontmatter.WriteString(fmt.Sprintf("date-created: %s\n", fm.DateCreated))

	// Handle tags
	if len(fm.Tags) > 0 {
		frontmatter.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(fm.Tags, ", ")))
	}

	// Only add optional fields if present
	if fm.DateDue != "" {
		frontmatter.WriteString(fmt.Sprintf("date-due: %s\n", fm.DateDue))
	}

	// Always include done field for todos
	if fm.Done {
		frontmatter.WriteString("done: true\n")
	} else {
		frontmatter.WriteString("done: false\n")
	}

	if fm.Priority > 0 {
		frontmatter.WriteString(fmt.Sprintf("priority: %d\n", fm.Priority))
	}
	if fm.ObjectiveID != "" {
		frontmatter.WriteString(fmt.Sprintf("objective-id: %s\n", fm.ObjectiveID))
	}
	if fm.ObjectiveRole != "" {
		frontmatter.WriteString(fmt.Sprintf("objective-role: %s\n", fm.ObjectiveRole))
	}

	frontmatter.WriteString("---\n\n")

	fileContent := frontmatter.String() + content
	h.CreateTestFile(filename, fileContent)
}

// ParseFrontmatter reads a file and returns its frontmatter
func (h *TestHarness) ParseFrontmatter(filename string) Frontmatter {
	path := filepath.Join(h.NotesDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		h.t.Fatalf("Failed to read file %s: %v", filename, err)
	}

	parts := strings.Split(string(content), "---")
	if len(parts) < 3 {
		h.t.Fatalf("File %s does not have valid frontmatter format", filename)
	}

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		h.t.Fatalf("Failed to parse frontmatter for %s: %v", filename, err)
	}

	return fm
}

// AssertFrontmatterValue checks if a specific frontmatter field matches expected value
func (h *TestHarness) AssertFrontmatterValue(filename string, check func(Frontmatter) error) {
	fm := h.ParseFrontmatter(filename)
	if err := check(fm); err != nil {
		h.t.Errorf("Frontmatter assertion failed for %s: %v", filename, err)
	}
}

// Helper to get today's date string
func Today() string {
	return time.Now().Format("2006-01-02")
}

// Helper to get a future date string
func FutureDate(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}
