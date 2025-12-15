package aicontext

import (
	"bufio"
	"strings"
)

type TaskMetadata struct {
	TargetFiles []string // Files to include with full content
	Content     string   // Full task content
	Language    string   // Language if specified
}

func ParseTaskMetadata(taskContent string) *TaskMetadata {
	meta := &TaskMetadata{
		Content: taskContent,
	}

	scanner := bufio.NewScanner(strings.NewReader(taskContent))
	inTargetFiles := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect language
		if strings.HasPrefix(strings.ToUpper(trimmed), "LANGUAGE:") {
			meta.Language = strings.TrimSpace(trimmed[9:]) // len("LANGUAGE:") = 9
			continue
		}

		// Retrieve target files
		if strings.HasPrefix(strings.ToUpper(trimmed), "TARGET FILES:") {
			inTargetFiles = true
			continue
		}

		if inTargetFiles {
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "##") {
				inTargetFiles = false
				continue
			}

			// Parse file path from "- path/to/filename" format
			if strings.HasPrefix(trimmed, "-") {
				path := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
				// Remove comments
				if idx := strings.Index(path, "#"); idx > 0 {
					path = strings.TrimSpace(path[:idx])
				}
				// Remove backticks
				path = strings.Trim(path, "`")
				if path != "" {
					meta.TargetFiles = append(meta.TargetFiles, path)
				}
			}

			if trimmed == "" && len(meta.TargetFiles) > 0 {
				inTargetFiles = false
			}
		}
	}

	return meta
}

func (m *TaskMetadata) IsTargetFile(path string) bool {
	// Normalize path (remove leading ./)
	path = strings.TrimPrefix(path, "./")

	for _, target := range m.TargetFiles {
		target = strings.TrimPrefix(target, "./")
		if path == target {
			return true
		}
	}
	return false
}
