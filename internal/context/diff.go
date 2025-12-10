package aicontext

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Before/after context for Q&A
type DiffContext struct {
	BaseBranch    string
	CurrentBranch string
	Diff          string            // Full diff
	ChangedFiles  []string
	FilesBefore   map[string]string
	FilesAfter    map[string]string
}

func GetDiffContext(baseBranch string) (*DiffContext, error) {
	if baseBranch == "" {
		baseBranch = getBaseBranch()
	}

	ctx := &DiffContext{
		BaseBranch:  baseBranch,
		FilesBefore: make(map[string]string),
		FilesAfter:  make(map[string]string),
	}

	// Get current branch name
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		ctx.CurrentBranch = strings.TrimSpace(string(out))
	}

	// Get diff
	diffOut, err := exec.Command("git", "diff", baseBranch+"...HEAD").Output()
	if err != nil {
		diffOut, err = exec.Command("git", "diff", baseBranch, "HEAD").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get diff: %w", err)
		}
	}
	ctx.Diff = string(diffOut)

	// Get list of changed files
	filesOut, err := exec.Command("git", "diff", "--name-only", baseBranch+"...HEAD").Output()
	if err != nil {
		filesOut, _ = exec.Command("git", "diff", "--name-only", baseBranch, "HEAD").Output()
	}
	if len(filesOut) > 0 {
		ctx.ChangedFiles = strings.Split(strings.TrimSpace(string(filesOut)), "\n")
	}

  // Get file contents
	for _, file := range ctx.ChangedFiles {
		// Content from base branch
		beforeOut, err := exec.Command("git", "show", baseBranch+":"+file).Output()
		if err == nil {
			ctx.FilesBefore[file] = string(beforeOut)
		}

		// Content from current branch (HEAD)
		afterOut, err := exec.Command("git", "show", "HEAD:"+file).Output()
		if err == nil {
			ctx.FilesAfter[file] = string(afterOut)
		}
	}

	return ctx, nil
}

func (d *DiffContext) GetContextForQA(targetFile string, targetLines string) string {
	var b strings.Builder

  b.WriteString(fmt.Sprintf("BRANCH: %s (base branch: %s)\n\n", d.CurrentBranch, d.BaseBranch))

	if targetFile != "" {
    // Asking about a specific file
		b.WriteString(fmt.Sprintf("=== TARGET FILE: %s ===\n\n", targetFile))

		if before, ok := d.FilesBefore[targetFile]; ok {
			b.WriteString("--- BEFORE (original) ---\n")
			b.WriteString(before)
			b.WriteString("\n\n")
		}

		if after, ok := d.FilesAfter[targetFile]; ok {
			b.WriteString("--- AFTER (current) ---\n")
			b.WriteString(after)
			b.WriteString("\n\n")
		}

		// File-specific diff
		b.WriteString("--- DIFF ---\n")
		b.WriteString(extractFileDiff(d.Diff, targetFile))
		b.WriteString("\n\n")
	} else {
		// Full PR context
		b.WriteString("=== CHANGED FILES ===\n")
		for _, f := range d.ChangedFiles {
			b.WriteString(fmt.Sprintf("- %s\n", f))
		}
		b.WriteString("\n")

		b.WriteString("=== FULL DIFF ===\n")
		b.WriteString(d.Diff)
		b.WriteString("\n")
	}

	return b.String()
}

func (d *DiffContext) GetTargetFileContent(path string) string {
	if content, ok := d.FilesAfter[path]; ok {
		return content
	}
	return ""
}

func (d *DiffContext) LoadAdditionalFiles(paths []string) map[string]string {
	additional := make(map[string]string)

	for _, path := range paths {
		// Skip if already exists
		if _, exists := d.FilesAfter[path]; exists {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		additional[path] = string(content)
	}

	return additional
}

func (d *DiffContext) GetContextForQAWithAdditional(targetFile, targetLines string, additionalFiles map[string]string, signatures string) string {
	var b strings.Builder

  b.WriteString(fmt.Sprintf("BRANCH: %s (base branch: %s)\n\n", d.CurrentBranch, d.BaseBranch))

	b.WriteString("=== CHANGED FILES IN THIS PR ===\n\n")

	if targetFile != "" {
    // Asking about a specific file
		b.WriteString(fmt.Sprintf("=== TARGET FILE: %s ===\n\n", targetFile))

		if before, ok := d.FilesBefore[targetFile]; ok {
			b.WriteString("--- BEFORE (original) ---\n")
			b.WriteString(before)
			b.WriteString("\n\n")
		}

		if after, ok := d.FilesAfter[targetFile]; ok {
			b.WriteString("--- AFTER (current) ---\n")
			b.WriteString(after)
			b.WriteString("\n\n")
		}

		// File-specific diff
		b.WriteString("--- DIFF ---\n")
		b.WriteString(extractFileDiff(d.Diff, targetFile))
		b.WriteString("\n\n")

		// Additional files
		for file, after := range d.FilesAfter {
			if file == targetFile {
				continue
			}
			b.WriteString(fmt.Sprintf("=== Changed: %s ===\n", file))
			if before, ok := d.FilesBefore[file]; ok {
        b.WriteString("--- BEFORE (original) ---\n")
				b.WriteString(before)
				b.WriteString("\n\n")
			}
			b.WriteString("--- AFTER (current) ---\n")
			b.WriteString(after)
			b.WriteString("\n\n")
		}
	} else {
    b.WriteString("=== CHANGED FILES ===\n")

		// All changed files
		for file, after := range d.FilesAfter {
			b.WriteString(fmt.Sprintf("--- %s ---\n", file))
			if before, ok := d.FilesBefore[file]; ok {
				b.WriteString("BEFORE:\n")
				b.WriteString(before)
				b.WriteString("\n\n")
			}
			b.WriteString("AFTER:\n")
			b.WriteString(after)
			b.WriteString("\n\n")
		}

		b.WriteString("FULL DIFF:\n")
		b.WriteString(d.Diff)
		b.WriteString("\n\n")
	}

	// Additional files from analysis
	if len(additionalFiles) > 0 {
		b.WriteString("=== RELATED FILES (for context) ===\n\n")
		for path, content := range additionalFiles {
			b.WriteString(fmt.Sprintf("--- %s ---\n", path))
			b.WriteString(content)
			b.WriteString("\n\n")
		}
	}

	// Signatures for reference
	if signatures != "" {
		b.WriteString("=== OTHER FILES (signatures only) ===\n\n")
		b.WriteString(signatures)
	}

	return b.String()
}

func extractFileDiff(fullDiff, targetFile string) string {
	lines := strings.Split(fullDiff, "\n")
	var result []string
	inTargetFile := false

	for _, line := range lines {
		// Detect file header
		if strings.HasPrefix(line, "diff --git") {
			inTargetFile = strings.Contains(line, targetFile)
		}

		if inTargetFile {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func getBaseBranch() string {
	// Common base branches
	candidates := []string{"main", "dev"}

	for _, branch := range candidates {
		err := exec.Command("git", "rev-parse", "--verify", branch).Run()
		if err == nil {
			return branch
		}
	}

	return "main" // default
}
