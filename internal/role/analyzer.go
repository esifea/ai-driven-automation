package role

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/esifea/ai-driven-automation/internal/provider"
)

type AnalysisMode string

const (
	AnalysisModeCoder AnalysisMode = "coder"
	AnalysisModeQA    AnalysisMode = "qa"
)

type AnalysisRequest struct {
	Mode        AnalysisMode
	Instruction string // Task instruction (coder) or Question (qa)
	Context     string // Signatures + target files or diff
	Overview    string // Global rules
}

type AnalysisResult struct {
	FilesToModify []FileAction `json:"files_to_modify"`
	FilesToCreate []FileAction `json:"files_to_create"`
	FilesToRead   []FileAction `json:"files_to_read"` // For Q&A mode
}

type FileAction struct {
	Path     string   `json:"path"`
	Sections []string `json:"sections,omitempty"`
	Reason   string   `json:"reason"`
}

func RunAnalysis(ctx context.Context, provider provider.Provider, req *AnalysisRequest) (*AnalysisResult, error) {
	prompt := buildAnalysisPrompt(req)

	response, err := provider.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return parseAnalysisResult(response)
}

func buildAnalysisPrompt(req *AnalysisRequest) string {
	var b strings.Builder

	b.WriteString("You are a Senior Engineer analyzing a codebase.\n\n")

	b.WriteString("GLOBAL PROJECT RULES:\n")
	b.WriteString(req.Overview)
	b.WriteString("\n\n")

	switch req.Mode {
	case AnalysisModeCoder:
		b.WriteString("TASK INSTRUCTIONS:\n")
		b.WriteString(req.Instruction)
		b.WriteString("\n\n")

		b.WriteString("CODEBASE CONTEXT:\n")
		b.WriteString(req.Context)
		b.WriteString("\n\n")

		b.WriteString(`INSTRUCTIONS:
1. Analyze the task requirements
2. Review the TARGET FILES (full content) and OTHER FILES (signatures)
3. Identify which additional files need full content to complete the task

OUTPUT FORMAT (JSON only, no markdown):
{
  "files_to_modify": [
    {"path": "path/to/file.go", "sections": ["FunctionName"], "reason": "why"}
  ],
  "files_to_create": [
    {"path": "path/to/new_file.go", "reason": "why"}
  ]
}

RULES:
- Do NOT include files already shown with full content
- Only request files whose signatures suggest they need modification
- Be conservative - only request files you truly need
- Output valid JSON only`)

	case AnalysisModeQA:
		b.WriteString("USER QUESTION:\n")
		b.WriteString(req.Instruction)
		b.WriteString("\n\n")

		b.WriteString("CONTEXT (Diff + Signatures):\n")
		b.WriteString(req.Context)
		b.WriteString("\n\n")

		b.WriteString(`INSTRUCTIONS:
1. Analyze what the user is asking
2. Review the DIFF (changed files) and SIGNATURES (other files)
3. Identify which additional files need full content to answer accurately

OUTPUT FORMAT (JSON only, no markdown):
{
  "files_to_read": [
    {"path": "path/to/file.go", "reason": "why this file helps answer the question"}
  ]
}

RULES:
- Do NOT include files already shown in the diff with full content
- Only request files that are necessary to understand the context
- Consider files that: implement related logic, define types used, show patterns
- Be conservative - only request files you truly need
- Output valid JSON only`)
	}

	return b.String()
}

func parseAnalysisResult(response string) (*AnalysisResult, error) {
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Try to extract JSON object
	jsonRe := regexp.MustCompile(`\{[\s\S]*\}`)
	match := jsonRe.FindString(response)
	if match == "" {
		return &AnalysisResult{}, nil
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(match), &result); err != nil {
		return &AnalysisResult{}, nil
	}

	return &result, nil
}

func (a *AnalysisResult) GetAdditionalFilePaths() []string {
	seen := make(map[string]bool)
	var paths []string

	for _, f := range a.FilesToModify {
		if !seen[f.Path] {
			paths = append(paths, f.Path)
			seen[f.Path] = true
		}
	}

	for _, f := range a.FilesToRead {
		if !seen[f.Path] {
			paths = append(paths, f.Path)
			seen[f.Path] = true
		}
	}

	return paths
}
