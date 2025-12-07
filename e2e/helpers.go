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

// getNow returns the current time, or a mocked time if TEST_FIXED_DATE is set
func getNow() time.Time {
	if fixedDate := os.Getenv("TEST_FIXED_DATE"); fixedDate != "" {
		t, err := time.Parse("2006-01-02", fixedDate)
		if err == nil {
			return t
		}
	}
	return time.Now()
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
		DateCreated: getNow().Format("2006-01-02"),
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
	return getNow().Format("2006-01-02")
}

// Helper to get a future date string
func FutureDate(days int) string {
	return getNow().AddDate(0, 0, days).Format("2006-01-02")
}

// Helper to get the date of a specific day this week using ISO week (Monday-Sunday)
func DayThisWeek(targetWeekday time.Weekday) string {
	now := getNow()
	// Find Monday of this ISO week (Monday=0 offset, Sunday=6 offset from Monday)
	currentWeekday := now.Weekday()
	daysFromMonday := int(currentWeekday+6) % 7 // Sunday(0)->6, Monday(1)->0, Tue(2)->1, etc.
	monday := now.AddDate(0, 0, -daysFromMonday)
	// Calculate offset from Monday to target day (Monday=0, ..., Sunday=6)
	targetOffset := int(targetWeekday+6) % 7
	return monday.AddDate(0, 0, targetOffset).Format("2006-01-02")
}

// Helper to get Monday this week
func MondayThisWeek() string {
	return DayThisWeek(time.Monday)
}

// Helper to get Tuesday this week
func TuesdayThisWeek() string {
	return DayThisWeek(time.Tuesday)
}

// Helper to get Wednesday this week
func WednesdayThisWeek() string {
	return DayThisWeek(time.Wednesday)
}

// Helper to get Thursday this week
func ThursdayThisWeek() string {
	return DayThisWeek(time.Thursday)
}

// Helper to get Friday this week
func FridayThisWeek() string {
	return DayThisWeek(time.Friday)
}

// Helper to get Saturday this week
func SaturdayThisWeek() string {
	return DayThisWeek(time.Saturday)
}

// Helper to get Sunday this week
func SundayThisWeek() string {
	return DayThisWeek(time.Sunday)
}

// Helper to get next Monday (from today)
func NextMonday() string {
	now := getNow()
	daysUntilMonday := (7 - int(now.Weekday()) + int(time.Monday)) % 7
	if daysUntilMonday == 0 {
		// If today is Monday, next Monday is 7 days away
		daysUntilMonday = 7
	}
	return now.AddDate(0, 0, daysUntilMonday).Format("2006-01-02")
}

// CreateTodoWithTalkToTag creates a todo with to-talk-X tag in the content
func (h *TestHarness) CreateTodoWithTalkToTag(filename, title string, person string, tags []string, dueDate string, priority int) {
	content := fmt.Sprintf("- [ ] %s to-talk-%s", title, person)
	h.CreateTodoWithContent(filename, title, content, dueDate, priority)
}

// CreateTodoWithContent creates a todo with custom content (useful for subtasks and custom formatting)
func (h *TestHarness) CreateTodoWithContent(filename, title, content string, dueDate string, priority int) {
	fm := Frontmatter{
		Title:       title,
		DateCreated: getNow().Format("2006-01-02"),
		Tags:        []string{},
		DateDue:     dueDate,
		Done:        false,
		Priority:    priority,
	}
	h.createFileWithFrontmatter(filename, fm, content)
}

// ReadFileContent reads the full content of a file (including frontmatter)
func (h *TestHarness) ReadFileContent(filename string) string {
	path := filepath.Join(h.NotesDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		h.t.Fatalf("Failed to read file %s: %v", filename, err)
	}
	return string(content)
}

// VerifyTodoMarkedComplete verifies a specific line contains a completed todo [x]
func (h *TestHarness) VerifyTodoMarkedComplete(filename string, lineNumber int) {
	content := h.ReadFileContent(filename)
	lines := strings.Split(content, "\n")

	if lineNumber < 1 || lineNumber > len(lines) {
		h.t.Errorf("Line number %d out of range for file %s (has %d lines)", lineNumber, filename, len(lines))
		return
	}

	line := lines[lineNumber-1] // Convert to 0-indexed
	if !strings.Contains(line, "- [x]") {
		h.t.Errorf("Expected line %d in %s to be marked complete [x], got: %s", lineNumber, filename, line)
	}
}

// VerifyTodoInFile checks if a todo exists in file content (case-insensitive substring match)
func (h *TestHarness) VerifyTodoInFile(filename, expectedTodo string) {
	content := h.ReadFileContent(filename)
	if !strings.Contains(strings.ToLower(content), strings.ToLower(expectedTodo)) {
		h.t.Errorf("Expected to find %q in file %s, but it was not found.\nFile content:\n%s", expectedTodo, filename, content)
	}
}

// VerifyTodoNotInFile checks that a todo does NOT exist in file content
func (h *TestHarness) VerifyTodoNotInFile(filename, unexpectedTodo string) {
	content := h.ReadFileContent(filename)
	if strings.Contains(strings.ToLower(content), strings.ToLower(unexpectedTodo)) {
		h.t.Errorf("Expected NOT to find %q in file %s, but it was found.\nFile content:\n%s", unexpectedTodo, filename, content)
	}
}

// CountTodosInFile counts the number of todo items (both complete and incomplete) in a file
func (h *TestHarness) CountTodosInFile(filename string) int {
	content := h.ReadFileContent(filename)
	lines := strings.Split(content, "\n")

	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "- [x]") {
			count++
		}
	}
	return count
}

// VerifyFileContains checks if file contains a specific string
func (h *TestHarness) VerifyFileContains(filename, expectedContent string) {
	content := h.ReadFileContent(filename)
	if !strings.Contains(content, expectedContent) {
		h.t.Errorf("Expected file %s to contain %q, but it was not found.\nFile content:\n%s", filename, expectedContent, content)
	}
}

// VerifyFileNotContains checks that file does NOT contain a specific string
func (h *TestHarness) VerifyFileNotContains(filename, unexpectedContent string) {
	content := h.ReadFileContent(filename)
	if strings.Contains(content, unexpectedContent) {
		h.t.Errorf("Expected file %s NOT to contain %q, but it was found.\nFile content:\n%s", filename, unexpectedContent, content)
	}
}
