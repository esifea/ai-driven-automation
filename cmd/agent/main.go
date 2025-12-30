package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/esifea/ai-driven-automation/internal/config"
	ctx "github.com/esifea/ai-driven-automation/internal/context"
	"github.com/esifea/ai-driven-automation/internal/parser"
	"github.com/esifea/ai-driven-automation/internal/provider"
	"github.com/esifea/ai-driven-automation/internal/role"
)

func main() {
	// Parse flags
	mode := flag.String("mode", "", "Agent mode: coder or reviewer")
	taskID := flag.String("task", "", "Task ID")
	providerName := flag.String("provider", "", "LLM provider: claude, gemini")
	flag.Parse()

	// Init config
	cfg := config.Load()

	// Override configs with flags if provided
	if *mode != "" {
		cfg.Mode = *mode
	}
	if *taskID != "" {
		cfg.TaskID = *taskID
	}
	if *providerName != "" {
		cfg.Provider = *providerName
	}

	llm, err := provider.NewProvider(cfg.Provider, cfg.APIKey, cfg.MaxRetries)
	if err != nil {
		log.Fatalf("Failed to initialize provider: %v", err)
	}

	// Handle Q&A mode
	if cfg.PRQuestion != "" {
		runQAMode(llm, cfg)
		return
	}

	// Handle modes
	switch cfg.Mode {
	case "reviewer":
		runReviewerMode(llm, cfg)
	default:
		runCoderMode(llm, cfg)
	}
}

func runQAMode(llm provider.Provider, cfg *config.Config) {
	log.Println("--- Q&A MODE ---")
	log.Println("Question:")
	log.Printf("%s", cfg.PRQuestion)
	log.Println("------------------")

	overview := ctx.GetOverviewDoc()

	// Get diff context (before/after changes)
	diffCtx, err := ctx.GetDiffContext(cfg.BaseBranch)
	if err != nil {
		log.Printf("Warning: Could not get diff context: %v", err)
		log.Println("Falling back to current codebase context only")
		runQAFallback(llm, cfg, overview)
		return
	}

	log.Printf("Diff context: %d files changed (base: %s)", len(diffCtx.ChangedFiles), diffCtx.BaseBranch)

	taskMetadata := &ctx.TaskMetadata{}
	codebaseCtx := ctx.GetCodebaseContext(taskMetadata)

	// Build analysis context: diff + signatures
	analysisContext := diffCtx.GetContextForQA(cfg.CommentPath, cfg.CommentEndLine)
	analysisContext += "\n\n=== OTHER FILES (signatures) ===\n"
	for path, sig := range codebaseCtx.SignatureFiles {
		// Skip files already in diff
		if _, inDiff := diffCtx.FilesAfter[path]; inDiff {
			continue
		}
		analysisContext += fmt.Sprintf("--- %s ---\n%s\n\n", path, sig)
	}

	// === Pass 1: Analyze what files are needed ===
	log.Println("=== Q&A Pass 1: Analyzing question ===")

	question := cfg.PRQuestion
	if cfg.CommentPath != "" {
		question = fmt.Sprintf("[File: %s, Line: %s] %s", cfg.CommentPath, cfg.CommentEndLine, question)
	}

	analysis, err := role.RunAnalysis(
		context.Background(), llm,
		&role.AnalysisRequest{
			Mode:        role.AnalysisModeQA,
			Instruction: question,
			Context:     analysisContext,
			Overview:    overview,
		},
	)

	var additionalFiles map[string]string
	if err != nil {
		log.Printf("Warning: Q&A analysis failed: %v", err)
	} else {
		additionalPaths := analysis.GetAdditionalFilePaths()
		if len(additionalPaths) > 0 {
			log.Printf("Pass 1 identified files needed: %v", additionalPaths)
			additionalFiles = diffCtx.LoadAdditionalFiles(additionalPaths)
		}
	}

	// Build signatures string (excluding diff and additional files)
	var signaturesStr string
	for path, sig := range codebaseCtx.SignatureFiles {
		if _, inDiff := diffCtx.FilesAfter[path]; inDiff {
			continue
		}
		if _, inAdditional := additionalFiles[path]; inAdditional {
			continue
		}
		signaturesStr += fmt.Sprintf("--- %s ---\n%s\n\n", path, sig)
	}

	// === Pass 2: Answer with full context ===
	log.Println("=== Q&A Pass 2: Generating answer ===")

	fullContext := diffCtx.GetContextForQAWithAdditional(
		cfg.CommentPath, cfg.CommentEndLine,
		additionalFiles, signaturesStr,
	)

	answer, err := role.RunQA(
		context.Background(), llm, cfg,
		fullContext, overview,
	)
	if err != nil {
		log.Fatalf("Q&A failed: %v", err)
	}

	writeAnswer(answer)
}

// runQAFallback handles Q&A when diff context is not available
func runQAFallback(llm provider.Provider, cfg *config.Config, overview string) {
	taskMetadata := &ctx.TaskMetadata{}
	codebaseCtx := ctx.GetCodebaseContext(taskMetadata)

	answer, err := role.RunQA(
		context.Background(), llm, cfg,
		codebaseCtx.GetContextForAnalysis(), overview,
	)
	if err != nil {
		log.Fatalf("Q&A failed: %v", err)
	}
	writeAnswer(answer)
}

func writeAnswer(answer string) {
	if err := os.WriteFile("answer.md", []byte(answer), 0644); err != nil {
		log.Fatalf("Failed to write answer: %v", err)
	}
	log.Println("Answer written to answer.md")
}

func runCoderMode(llm provider.Provider, cfg *config.Config) {
	log.Printf("--- CODER AGENT STARTED (Task: %s) ---", cfg.TaskID)

	if cfg.Feedback != "" {
		log.Printf("Feedback received: %s", cfg.Feedback)
	}

	// Load context
	overview := ctx.GetOverviewDoc()
	instruction, err := ctx.GetInstructionDoc(cfg.TaskID)
	if err != nil {
		log.Fatalf("Failed to load task instructions: %v", err)
	}

	// Parse task metadata (TARGET FILES, DEPENDS_ON)
	taskMetadata := ctx.ParseTaskMetadata(instruction)
	log.Printf("Target files from task: %v", taskMetadata.TargetFiles)

	var dependentContext string

	if len(taskMetadata.DependsOn) > 0 {
		log.Printf("Task depends on: %v", taskMetadata.DependsOn)
		dependentContext = ctx.GetDependentContext(taskMetadata.DependsOn)
	}

	// Build initial context (targets full, others signatures)
	codebaseCtx := ctx.GetCodebaseContext(taskMetadata)
	log.Printf("Loaded %d target files, %d signature files",
		len(codebaseCtx.TargetFiles), len(codebaseCtx.SignatureFiles))

	// === Pass 1: Analysis ===
	log.Println("=== Pass 1: Analyzing task and identifying files ===")

	analysisContext := codebaseCtx.GetContextForAnalysis()
	if dependentContext != "" {
		analysisContext = dependentContext + "\n\n" + analysisContext
	}

	analysis, err := role.RunAnalysis(
		context.Background(), llm,
		&role.AnalysisRequest{
			Mode:        role.AnalysisModeCoder,
			Instruction: instruction,
			Context:     analysisContext,
			Overview:    overview,
		},
	)
	if err != nil {
		log.Printf("Warning: Analysis failed, proceeding with target files only: %v", err)
	} else {
		additionalPaths := analysis.GetAdditionalFilePaths()
		if len(additionalPaths) > 0 {
			log.Printf("Pass 1 identified additional files: %v", additionalPaths)
			codebaseCtx.ReloadFiles(additionalPaths)
		}
	}

	// === Pass 2: Implementation ===
	log.Println("=== Pass 2: Generating implementation ===")
	generated, err := role.RunCoder(
		context.Background(), llm, cfg,
		instruction, codebaseCtx.GetContextForImplementation(), overview,
	)
	if err != nil {
		log.Fatalf("Code generation failed: %v", err)
	}

	// Parse and write files
	files := parser.ParseFiles(generated)
	if len(files) == 0 {
		log.Println("Warning: Coder generated no file output.")
		return
	}

	count, err := parser.WriteFiles(files)
	if err != nil {
		log.Fatalf("Failed to write files: %v", err)
	}
	log.Printf("Coder wrote %d files to disk.", count)
}

func runReviewerMode(llm provider.Provider, cfg *config.Config) {
	log.Println("--- REVIEWER AGENT STARTED ---")

	if cfg.PRNumber == "" {
		log.Fatal("PR_NUMBER is required for reviewer mode")
	}

	// Load context
	instruction, err := ctx.GetInstructionDoc(cfg.TaskID)
	if err != nil {
		log.Fatalf("Failed to load task instructions: %v", err)
	}

	taskMetadata := ctx.ParseTaskMetadata(instruction)
	codebaseCtx := ctx.GetCodebaseContext(taskMetadata)

	// Generate review
	review, err := role.RunReviewer(
		context.Background(), llm,
		instruction, codebaseCtx.GetContextForImplementation(),
	)
	if err != nil {
		log.Fatalf("Review generation failed: %v", err)
	}

	fmt.Println("=== REVIEW CONTENT ===")
	fmt.Println(review)
	fmt.Println("=== END REVIEW ===")

	log.Println("Review generated. Submitting to GitHub...")

	// Determine status
	eventType := "REQUEST_CHANGES"
	body := fmt.Sprintf("## AI Review: CHANGES REQUESTED ❌\n\n%s", review)

	if strings.Contains(review, "STATUS: PASS") {
		eventType = "APPROVE"
		body = fmt.Sprintf("## AI Review: PASS ✅\n\n%s", review)
	}

	// Submit via gh CLI
	ghFlag := "--" + strings.ToLower(strings.ReplaceAll(eventType, "_", "-"))
	cmd := exec.Command("gh", "pr", "review", cfg.PRNumber, ghFlag, "--body", body)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
    // Check self-review error
    log.Printf("Warning: Failed to submit review: %v", err)
    log.Println("Reviewing your own PR not allowed by Github")

    return
	}
	log.Printf("Submitted review: %s", eventType)
}
