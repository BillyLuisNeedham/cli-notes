package presentation

import (
	"bytes"
	"cli-notes/scripts"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPrintAllFiles(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		files         []scripts.File
		expectedLines []string
	}{
		{
			name:          "empty files list",
			files:         []scripts.File{},
			expectedLines: []string{},
		},
		{
			name: "single file",
			files: []scripts.File{
				{
					Name:     "test.md",
					DueAt:    time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
					Priority: scripts.P2,
				},
			},
			expectedLines: []string{
				"test.md  due: 2023-05-15  P2",
			},
		},
		{
			name: "multiple files same priority",
			files: []scripts.File{
				{
					Name:     "task1.md",
					DueAt:    time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
					Priority: scripts.P1,
				},
				{
					Name:     "task2.md",
					DueAt:    time.Date(2023, 6, 20, 0, 0, 0, 0, time.UTC),
					Priority: scripts.P1,
				},
			},
			expectedLines: []string{
				"task1.md  due: 2023-05-15  P1",
				"task2.md  due: 2023-06-20  P1",
			},
		},
		{
			name: "multiple files different priorities",
			files: []scripts.File{
				{
					Name:     "task1.md",
					DueAt:    time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
					Priority: scripts.P1,
				},
				{
					Name:     "task2.md",
					DueAt:    time.Date(2023, 6, 20, 0, 0, 0, 0, time.UTC),
					Priority: scripts.P2,
				},
				{
					Name:     "meeting.md",
					DueAt:    time.Date(2023, 7, 30, 0, 0, 0, 0, time.UTC),
					Priority: scripts.P3,
				},
			},
			expectedLines: []string{
				"task1.md  due: 2023-05-15  P1",
				"",
				"task2.md  due: 2023-06-20  P2",
				"",
				"meeting.md  due: 2023-07-30  P3",
			},
		},
		{
			name: "files with zero time",
			files: []scripts.File{
				{
					Name:  "no-due-date.md",
					DueAt: time.Time{},
				},
			},
			expectedLines: []string{
				"no-due-date.md  due: 0001-01-01",
			},
		},
		{
			name: "files with no priority",
			files: []scripts.File{
				{
					Name:  "no-priority.md",
					DueAt: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedLines: []string{
				"no-priority.md  due: 2023-05-15",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the function
			PrintAllFiles(tc.files)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify output
			outputLines := strings.Split(strings.TrimRight(output, "\n"), "\n")
			if len(outputLines) == 1 && outputLines[0] == "" {
				outputLines = []string{}
			}

			if len(outputLines) != len(tc.expectedLines) {
				t.Errorf("Expected %d lines, got %d lines: %v", len(tc.expectedLines), len(outputLines), outputLines)
				return
			}

			for i, expectedLine := range tc.expectedLines {
				if i < len(outputLines) && outputLines[i] != expectedLine {
					t.Errorf("Line %d: expected '%s', got '%s'", i, expectedLine, outputLines[i])
				}
			}
		})
	}
}
