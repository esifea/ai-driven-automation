package role

import (
	"context"
	"fmt"
	"strings"

	"github.com/esifea/ai-driven-automation/internal/config"
	"github.com/esifea/ai-driven-automation/internal/provider"
)

func RunCoder(ctx context.Context, provider provider.Provider, cfg *config.Config, instruction, contextStr, overview string) (string, error) {
	feedback := cfg.Feedback
	if feedback == "" {
		feedback = "None."
	}

	prompt := fmt.Sprintf(`You are a Senior Engineer. Implement the following task.

GLOBAL PROJECT RULES (MUST FOLLOW):
%s

CONTEXT:
%s

TASK INSTRUCTIONS:
%s

REQUIREMENTS:
1. Output the FULL content of any file you create or modify.
2. Format:
   ### File: path/to/file.ext
   `+"```"+`
   // content
   `+"```"+`
3. FIX issues from the FEEDBACK below (if any).
4. Only modify files shown in CONTEXT - do not invent new paths.

PREVIOUS REVIEWER FEEDBACK:
%s`, overview, contextStr, instruction, feedback)

	return provider.Generate(ctx, prompt)
}

func RunQA(ctx context.Context, provider provider.Provider, cfg *config.Config, contextStr, overview string) (string, error) {
	userRequest := strings.Replace(cfg.PRQuestion, "/ask", "", 1)
	userRequest = strings.TrimSpace(userRequest)

	targetInfo := ""
	if cfg.CommentPath != "" && cfg.CommentEndLine != "" {
		lineInfo := fmt.Sprintf("TARGET LINE: %s", cfg.CommentEndLine)

		if cfg.CommentStartLine != "" && cfg.CommentStartLine != "None" && cfg.CommentStartLine != cfg.CommentEndLine {
			lineInfo = fmt.Sprintf("TARGET LINES: %s-%s", cfg.CommentStartLine, cfg.CommentEndLine)
		}

		targetInfo = fmt.Sprintf("\nTARGET FILE: %s\n%s\n", cfg.CommentPath, lineInfo)
	}

	prompt := fmt.Sprintf(`You are a Helpful Senior Engineer Assistant reviewing a Pull Request.

GLOBAL PROJECT RULES:
%s

CONTEXT:
%s
%s
USER QUESTION:
%s

INSTRUCTIONS:
- You can see BEFORE (original) and AFTER (current) versions of changed files
- You also have full content of related files for deeper understanding
- Use this context to give accurate, specific answers
- Reference actual code when explaining
- If suggesting code changes, output the FULL file content using format:
  ### File: path/to/file.ext
`, overview, contextStr, targetInfo, userRequest)

	return provider.Generate(ctx, prompt)
}

func RunReviewer(ctx context.Context, provider provider.Provider, instruction, contextStr string) (string, error) {
	prompt := fmt.Sprintf(`You are a Strict Code Reviewer (Principal Engineer).

Verify the code below against instructions:
%s

GENERATED CODE:
%s

OUTPUT FORMAT:
First line: STATUS: [PASS or FAIL]
Subsequent lines: Bullet points of critique.`, instruction, contextStr)

	return provider.Generate(ctx, prompt)
}
