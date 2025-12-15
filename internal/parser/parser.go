package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ParseFiles(response string) map[string]string {
	files := make(map[string]string)
	lines := strings.Split(response, "\n")

	var currentFile string
	var codeLines []string
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Format: ### File: path/to/file
		if strings.HasPrefix(trimmed, "### File:") {
			// Save previous file if exists
			if currentFile != "" && len(codeLines) > 0 {
				files[currentFile] = strings.TrimSpace(strings.Join(codeLines, "\n"))
			}
			currentFile = strings.TrimSpace(strings.TrimPrefix(trimmed, "### File:"))
			codeLines = nil
			inBlock = false
			continue
		}

		// Code block
		if strings.HasPrefix(trimmed, "```") {
			inBlock = !inBlock
			continue
		}

		if inBlock && currentFile != "" {
			codeLines = append(codeLines, line)
		}
	}

	if currentFile != "" && len(codeLines) > 0 {
		files[currentFile] = strings.TrimSpace(strings.Join(codeLines, "\n"))
	}

	return files
}

func WriteFiles(files map[string]string) (int, error) {
	count := 0
	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return count, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return count, fmt.Errorf("failed to write file %s: %w", path, err)
		}
		fmt.Printf("Wrote: %s\n", path)
		count++
	}
	return count, nil
}
