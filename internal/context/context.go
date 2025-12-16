package aicontext

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	excludeDirs = map[string]bool{
		".git":         true,
		".github":      true,
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"bin":          true,
	}

	skipFiles = map[string]bool{
		"go.sum":            true,
		"go.mod":            true,
		"package-lock.json": true,
	}

	codeExtensions = map[string]bool{
		".go":  true,
		".md":  true,
		".tf":  true,
		".py":  true,
		".h":   true,
		".hpp": true,
		".c":   true,
		".cpp": true,
	}
)

type ContextType struct {
	TargetFiles     map[string]string // Full content
	SignatureFiles  map[string]string // Partial content (signature)
	AdditionalFiles map[string]string // Full content
}

//--- Load contexts from docs ---//

// docs/tasks/00_overview.md
func GetOverviewDoc() string {
	path := "docs/tasks/00_overview.md"

	content, err := loadFile(path)
	if err != nil {
		return ""
	}

	return content
}

// docs/tasks/{TASK_ID}_*.md (exclude *_completed.md)
func GetInstructionDoc(taskID string) (string, error) {
	pattern := fmt.Sprintf("docs/tasks/%s_*.md", taskID)

	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		return "", fmt.Errorf("no instruction file found for Task %s", taskID)
	}

	for _, f := range files {
		if !strings.Contains(f, "_completed") {
			return loadFile(f)
		}
	}

	// XXX: Support multiple instruction files
	return loadFile(files[0])
}

// docs/tasks/{TASK_ID}_*_completed.md
func GetCompletedDoc(taskID string) (string, error) {
	pattern := fmt.Sprintf("docs/tasks/%s_*_completed.md", taskID)

	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		return "", fmt.Errorf("no completed doc found for Task %s", taskID)
	}

	return loadFile(files[0])
}

func GetDependentContext(dependsOn []string) string {
	if len(dependsOn) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("=== DEPENDENT TASKS CONTEXT ===\n\n")

	for _, taskID := range dependsOn {
		// Load completed summary
		if completed, err := GetCompletedDoc(taskID); err == nil {
			b.WriteString(fmt.Sprintf("--- Task %s (Completed) ---\n", taskID))
			b.WriteString(completed)
			b.WriteString("\n\n")
			continue
		}

		// Fallback to original instructions
		if instruction, err := GetInstructionDoc(taskID); err == nil {
			b.WriteString(fmt.Sprintf("--- Task %s (Instructions Only) ---\n", taskID))
			b.WriteString(instruction)
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

//--- Load contexts from sources ---//

func GetCodebaseContext(taskMetadata *TaskMetadata) *ContextType {
	result := &ContextType{
		TargetFiles:     make(map[string]string),
		SignatureFiles:  make(map[string]string),
		AdditionalFiles: make(map[string]string),
	}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip pre-defined files
		if info.IsDir() {
			if excludeDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if skipFiles[info.Name()] {
			return nil
		}

		if strings.Contains(path, "00_overview.md") {
			return nil
		}

		// Filter code files
		ext := filepath.Ext(path)
		if !codeExtensions[ext] {
			return nil
		}

		content, err := loadFile(path)
		if err != nil {
			return nil
		}

		path = strings.TrimPrefix(path, "./")

		if taskMetadata != nil && taskMetadata.IsTargetFile(path) { // get full context of target files
			result.TargetFiles[path] = content
		} else {
			result.SignatureFiles[path] = ExtractSignatures(path, content) // get signatures for token saving
		}

		return nil
	})

	return result
}

func (c *ContextType) GetContextForAnalysis() string {
	var b strings.Builder

	// Target files
	if len(c.TargetFiles) > 0 {
		b.WriteString("=== TARGET FILES (Full content) ===\n\n")
		for path, content := range c.TargetFiles {
			b.WriteString(fmt.Sprintf("--- File: %s ---\n", path))
			b.WriteString(content)
			b.WriteString("\n\n")
		}
	}

	// Signature files
	if len(c.SignatureFiles) > 0 {
		b.WriteString("=== OTHER FILES (Signatures only) ===\n\n")
		for path, sig := range c.SignatureFiles {
			b.WriteString(fmt.Sprintf("--- File: %s ---\n", path))
			b.WriteString(sig)
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func (c *ContextType) GetContextForImplementation() string {
	var b strings.Builder

	// Target files
	if len(c.TargetFiles) > 0 {
		b.WriteString("=== TARGET FILES (Full content) ===\n\n")
		for path, content := range c.TargetFiles {
			b.WriteString(fmt.Sprintf("--- File: %s ---\n", path))
			b.WriteString(content)
			b.WriteString("\n\n")
		}
	}

	// Additional files (added after analysis)
	if len(c.AdditionalFiles) > 0 {
		b.WriteString("=== ADDITIONAL FILES (Full content) ===\n\n")
		for path, content := range c.AdditionalFiles {
			b.WriteString(fmt.Sprintf("--- File: %s ---\n", path))
			b.WriteString(content)
			b.WriteString("\n\n")
		}
	}

	// Signature files (for reference)
	if len(c.SignatureFiles) > 0 {
		b.WriteString("=== OTHER FILES (Signatures only) ===\n\n")
		for path, sig := range c.SignatureFiles {
			b.WriteString(fmt.Sprintf("--- File: %s ---\n", path))
			b.WriteString(sig)
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func (c *ContextType) ReloadFiles(paths []string) {
	for _, path := range paths {
		// Skip if already loaded or not exist
		_, ok1 := c.TargetFiles[path]
		_, ok2 := c.AdditionalFiles[path]
		_, ok3 := c.SignatureFiles[path]
		if ok1 || ok2 || !ok3 {
			continue
		}

		content, err := loadFile(path)
		if err != nil {
			continue
		}

		delete(c.SignatureFiles, path)
		c.AdditionalFiles[path] = content
	}
}

func loadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
