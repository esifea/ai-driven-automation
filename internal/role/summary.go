package role

import (
	"context"
	"fmt"

	"github.com/esifea/ai-driven-automation/internal/provider"
)

type SummaryRequest struct {
	TaskID       string
	Instruction  string
	FilesChanged map[string]string
	PRNumber     string
}

func GenerateCompletionSummary(ctx context.Context, provider provider.Provider, req *SummaryRequest) (string, error) {
	fileList := ""
	for path, content := range req.FilesChanged {
		// FIXME: Include first 50 lines of each file for context (heuristic)
		lines := truncateContent(content, 50)
		fileList += fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, lines)
	}

	prompt := fmt.Sprintf(`You are a technical documentation writer.

A task has been completed and merged. Generate a completion summary for future reference.

TASK ID: %s
PR NUMBER: %s

ORIGINAL TASK INSTRUCTIONS:
%s

FILES IMPLEMENTED:
%s

Generate a completion summary in this EXACT format:

# Task %s - Completed

MERGED: [current date]
PR: #%s

## Final Implementation
[List each file with a brief description of what it contains]

## Key Decisions
[List important implementation decisions, patterns used, or deviations from original instructions]

## API/Interface Summary
[List public functions, types, or endpoints created - this helps dependent tasks]

## Patterns Established
[List any conventions or patterns that future tasks should follow]

## Notes for Dependent Tasks
[Any important context for tasks that build on this work]

Keep it concise but informative. Focus on what future tasks need to know.`,
		req.TaskID, req.PRNumber, req.Instruction, fileList, req.TaskID, req.PRNumber)

	return provider.Generate(ctx, prompt)
}

func truncateContent(content string, maxLines int) string {
	lines := splitLines(content)
	if len(lines) <= maxLines {
		return content
	}
	result := joinLines(lines[:maxLines])
	result += "\n// ... (truncated)"
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}
	return result
}
