package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestHarness manages the test environment and CLI execution
type TestHarness struct {
	t        *testing.T
	TempDir  string
	NotesDir string
	CLIPath  string
	Env      []string
}

// NewTestHarness creates a new test harness with a temporary directory
func NewTestHarness(t *testing.T) *TestHarness {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cli-notes-e2e-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create notes directory within temp dir
	notesDir := filepath.Join(tempDir, "notes")
	if err := os.Mkdir(notesDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// We are likely in .../cli-tests/e2e when running tests
	// We need the root .../cli-tests
	repoRoot := filepath.Dir(wd)

	// Verify main.go exists there
	mainPath := filepath.Join(repoRoot, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		// Maybe we are already in root?
		repoRoot = wd
		mainPath = filepath.Join(repoRoot, "main.go")
		if _, err := os.Stat(mainPath); os.IsNotExist(err) {
			t.Fatalf("Could not find main.go at %s or %s", filepath.Join(filepath.Dir(wd), "main.go"), filepath.Join(wd, "main.go"))
		}
	}

	// Build the CLI binary
	// We build it to a temporary location
	binPath := filepath.Join(tempDir, "cli-notes")
	// Build the package in the current directory (repoRoot)
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = repoRoot // Build from repo root so it finds go.mod and all files
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build CLI: %v\nOutput: %s", err, out)
	}

	// Ensure binary is executable
	if err := os.Chmod(binPath, 0755); err != nil {
		t.Fatalf("Failed to chmod binary: %v", err)
	}

	// Setup environment
	env := os.Environ()
	env = append(env, "CLI_NOTES_TEST_MODE=true")
	env = append(env, "EDITOR=echo")
	// Filter out any existing CLI_NOTES_* env vars if needed, or just append

	h := &TestHarness{
		t:        t,
		TempDir:  tempDir,
		NotesDir: notesDir,
		CLIPath:  binPath,
		Env:      env,
	}

	// Register cleanup
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return h
}

// RunCommand runs the CLI with the given arguments and input
func (h *TestHarness) RunCommand(input string) (string, string, error) {
	// Get absolute path to main.go
	// We assume the test is run from the repo root or we can find it relative to the test file.
	// But since we are in a temp dir, we need the absolute path.
	// We stored the repo root in the harness? No, we need to.

	// Let's capture the repo root when creating the harness.
	// But wait, NewTestHarness is called from the test.
	// We can get the wd there.

	// DEBUG: Log execution details
	h.t.Logf("=== RunCommand Debug ===")
	h.t.Logf("CLI Path: %s", h.CLIPath)
	h.t.Logf("Temp Dir: %s", h.TempDir)
	h.t.Logf("Notes Dir: %s", h.NotesDir)
	h.t.Logf("Input: %q", input)

	ctx := context.Background() // Define ctx for use with CommandContext
	cmd := exec.CommandContext(ctx, h.CLIPath)
	cmd.Dir = h.TempDir
	cmd.Env = h.Env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("failed to start command: %w", err)
	}

	// Write input to stdin
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()

	// Wait for command to finish
	// We use a timeout to prevent hanging tests
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			h.t.Logf("failed to kill process: %v", err)
		}
		h.t.Logf("STDOUT:\n%s", stdout.String())
		h.t.Logf("STDERR:\n%s", stderr.String())
		return stdout.String(), stderr.String(), fmt.Errorf("command timed out after 10s\nStdout:\n%s\nStderr:\n%s", stdout.String(), stderr.String())
	case err := <-done:
		// DEBUG: Always log output
		h.t.Logf("STDOUT:\n%s", stdout.String())
		h.t.Logf("STDERR:\n%s", stderr.String())
		
		if err != nil {
			return stdout.String(), stderr.String(), fmt.Errorf("command failed: %v\nStdout:\n%s\nStderr:\n%s", err, stdout.String(), stderr.String())
		}
		return stdout.String(), stderr.String(), nil
	}
}

// RunCommandWithArgs runs the CLI with specific arguments (if the CLI supported flags, which it doesn't seem to much, but good for future)
// Since this CLI is interactive-first, we mostly use RunCommand with input.
// However, if we want to test specific start-up flags, we can use this.
func (h *TestHarness) RunCommandWithArgs(args []string, input string) (string, string, error) {
	cmd := exec.Command(h.CLIPath, args...)
	cmd.Dir = h.TempDir
	cmd.Env = h.Env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("failed to start command: %w", err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			h.t.Logf("failed to kill process: %v", err)
		}
		return stdout.String(), stderr.String(), fmt.Errorf("command timed out")
	case err := <-done:
		return stdout.String(), stderr.String(), err
	}
}

// AssertFileExists checks if a file exists in the notes directory
func (h *TestHarness) AssertFileExists(filename string) {
	path := filepath.Join(h.NotesDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		h.t.Errorf("File %s does not exist", filename)
	}
}

// AssertFileContent checks if a file contains specific content
func (h *TestHarness) AssertFileContent(filename string, expectedContent string) {
	path := filepath.Join(h.NotesDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		h.t.Errorf("Failed to read file %s: %v", filename, err)
		return
	}

	if !strings.Contains(string(content), expectedContent) {
		h.t.Errorf("File %s does not contain expected content.\nExpected: %s\nGot: %s", filename, expectedContent, string(content))
	}
}

// AssertFileNotExists checks if a file does not exist
func (h *TestHarness) AssertFileNotExists(filename string) {
	path := filepath.Join(h.NotesDir, filename)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		h.t.Errorf("File %s exists but should not", filename)
	}
}
