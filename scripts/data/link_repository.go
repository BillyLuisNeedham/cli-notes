package data

import (
	"cli-notes/scripts"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// linkPattern matches [[link text]] syntax
var linkPattern = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

// ParseLinks extracts all [[...]] link references from content
func ParseLinks(content string) []string {
	matches := linkPattern.FindAllStringSubmatch(content, -1)
	links := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			linkText := strings.TrimSpace(match[1])
			if linkText != "" && !seen[linkText] {
				links = append(links, linkText)
				seen[linkText] = true
			}
		}
	}

	return links
}

// ResolveLink finds a file matching the link text (by title or filename)
func ResolveLink(linkText string) (*scripts.File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	notesPath := filepath.Join(currentDir, DirectoryPath)
	linkTextLower := strings.ToLower(linkText)

	var matchedFile *scripts.File

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		// Match by filename (without extension)
		if strings.ToLower(fileNameWithoutExt) == linkTextLower {
			file, err := LoadFileByName(fileName)
			if err != nil {
				return err
			}
			matchedFile = &file
			return filepath.SkipAll
		}

		// Match by title
		file, err := LoadFileByName(fileName)
		if err != nil {
			return nil // Skip files that can't be loaded
		}

		if strings.ToLower(file.Title) == linkTextLower {
			matchedFile = &file
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, err
	}

	return matchedFile, nil
}

// GetLinksFrom returns all files that the given file links to
func GetLinksFrom(fileName string) ([]scripts.File, error) {
	file, err := LoadFileByName(fileName)
	if err != nil {
		return nil, err
	}

	// Parse links from content (includes the area after frontmatter)
	links := ParseLinks(file.Content)

	linkedFiles := make([]scripts.File, 0, len(links))
	for _, linkText := range links {
		linkedFile, err := ResolveLink(linkText)
		if err != nil {
			continue // Skip unresolvable links
		}
		if linkedFile != nil {
			linkedFiles = append(linkedFiles, *linkedFile)
		}
	}

	return linkedFiles, nil
}

// GetBacklinks returns all files that link TO the given file
func GetBacklinks(fileName string) ([]scripts.File, error) {
	targetFile, err := LoadFileByName(fileName)
	if err != nil {
		return nil, err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	notesPath := filepath.Join(currentDir, DirectoryPath)
	targetNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	targetTitleLower := strings.ToLower(targetFile.Title)
	targetNameLower := strings.ToLower(targetNameWithoutExt)

	backlinks := make([]scripts.File, 0)

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() == fileName {
			return nil
		}

		file, err := LoadFileByName(info.Name())
		if err != nil {
			return nil // Skip files that can't be loaded
		}

		// Check if this file links to our target
		links := ParseLinks(file.Content)
		for _, linkText := range links {
			linkTextLower := strings.ToLower(linkText)
			if linkTextLower == targetTitleLower || linkTextLower == targetNameLower {
				backlinks = append(backlinks, file)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return backlinks, nil
}

// LinkIndex represents the full graph of note connections
type LinkIndex struct {
	// OutLinks maps filename to list of filenames it links to
	OutLinks map[string][]string
	// InLinks maps filename to list of filenames that link to it
	InLinks map[string][]string
	// FilesByName maps filename to File struct
	FilesByName map[string]scripts.File
	// FilesByTitle maps lowercase title to filename for resolution
	FilesByTitle map[string]string
}

// BuildLinkIndex builds a complete index of all links in the notes directory
func BuildLinkIndex() (*LinkIndex, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	notesPath := filepath.Join(currentDir, DirectoryPath)

	index := &LinkIndex{
		OutLinks:     make(map[string][]string),
		InLinks:      make(map[string][]string),
		FilesByName:  make(map[string]scripts.File),
		FilesByTitle: make(map[string]string),
	}

	// First pass: load all files and build title index
	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := LoadFileByName(info.Name())
		if err != nil {
			return nil // Skip files that can't be loaded
		}

		index.FilesByName[info.Name()] = file
		if file.Title != "" {
			index.FilesByTitle[strings.ToLower(file.Title)] = info.Name()
		}
		// Also index by filename without extension
		nameWithoutExt := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		index.FilesByTitle[strings.ToLower(nameWithoutExt)] = info.Name()

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Second pass: build link graph
	for fileName, file := range index.FilesByName {
		links := ParseLinks(file.Content)

		for _, linkText := range links {
			// Resolve the link to a filename
			linkTextLower := strings.ToLower(linkText)
			targetFileName, exists := index.FilesByTitle[linkTextLower]
			if !exists {
				continue // Skip unresolvable links
			}

			// Add outgoing link
			index.OutLinks[fileName] = append(index.OutLinks[fileName], targetFileName)

			// Add incoming link (backlink)
			index.InLinks[targetFileName] = append(index.InLinks[targetFileName], fileName)
		}
	}

	return index, nil
}

// GetUnresolvedLinks finds all [[...]] links in a file that don't resolve to existing notes
func GetUnresolvedLinks(fileName string) ([]string, error) {
	file, err := LoadFileByName(fileName)
	if err != nil {
		return nil, err
	}

	links := ParseLinks(file.Content)
	unresolved := make([]string, 0)

	for _, linkText := range links {
		resolved, err := ResolveLink(linkText)
		if err != nil || resolved == nil {
			unresolved = append(unresolved, linkText)
		}
	}

	return unresolved, nil
}

// CreateNoteFromDeadLink creates a new note from an unresolved link text
func CreateNoteFromDeadLink(linkText string) (*scripts.File, error) {
	// Create a new todo note with the link text as title
	now := time.Now()

	// Generate safe filename
	safeTitle := strings.ToLower(linkText)
	safeTitle = strings.ReplaceAll(safeTitle, " ", "-")

	fileName := fmt.Sprintf("%s-%s.md", safeTitle, now.Format("2006-01-02"))

	newFile := scripts.File{
		Name:      fileName,
		Title:     linkText,
		Tags:      []string{"todo"},
		CreatedAt: now,
		DueAt:     now,
		Done:      false,
		Priority:  scripts.P2,
		Content:   "# " + linkText + "\n\n",
	}

	err := WriteFile(newFile)
	if err != nil {
		return nil, err
	}

	return &newFile, nil
}

// InsertLinkAtTop adds a [[link]] reference at the top of a note (after frontmatter, before content)
func InsertLinkAtTop(fileName string, linkText string) error {
	file, err := LoadFileByName(fileName)
	if err != nil {
		return err
	}

	linkRef := "[[" + linkText + "]]"

	// Check if link already exists
	if strings.Contains(file.Content, linkRef) {
		return nil // Link already exists, nothing to do
	}

	// Insert link at the beginning of content
	// If content starts with links already, append to that line
	// Otherwise, add a new line with the link
	content := file.Content
	trimmedContent := strings.TrimLeft(content, "\n")

	if strings.HasPrefix(trimmedContent, "[[") {
		// Find the first line that starts with [[
		lines := strings.SplitN(trimmedContent, "\n", 2)
		if len(lines) > 0 {
			// Append to existing links line
			lines[0] = lines[0] + " " + linkRef
			if len(lines) > 1 {
				file.Content = lines[0] + "\n" + lines[1]
			} else {
				file.Content = lines[0]
			}
		}
	} else {
		// Add link as new first line
		file.Content = linkRef + "\n\n" + trimmedContent
	}

	return WriteFile(file)
}
