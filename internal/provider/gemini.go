package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var models = []string{
	"gemini-3-pro-preview", // primary
	"gemini-2.5-pro",       // fallback
}

type Gemini struct {
	client     *genai.Client
	maxRetries int
}

func NewGemini(apiKey string, maxRetries int) (*Gemini, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}
	return &Gemini{client: client, maxRetries: maxRetries}, nil
}

func (g *Gemini) Name() string {
	return "gemini"
}

func (g *Gemini) Generate(ctx context.Context, prompt string) (string, error) {
	var lastErr error

	for attempt := 0; attempt < g.maxRetries; attempt++ {
		modelName := models[attempt%len(models)]
		log.Printf("Generating with %s (Attempt %d/%d)", modelName, attempt+1, g.maxRetries)

		model := g.client.GenerativeModel(modelName)
		resp, err := model.GenerateContent(ctx, genai.Text(prompt))

		if err == nil && resp != nil {
			return extractText(resp), nil
		}

		lastErr = err
		errMsg := err.Error()
		log.Printf("Failed with %s: %s", modelName, errMsg)

		sleepTime := 15 * time.Duration(attempt+1) * time.Second
		if strings.Contains(errMsg, "503") || strings.Contains(strings.ToLower(errMsg), "overloaded") {
			log.Println("Server overloaded. Waiting 30s...")
			sleepTime = 30 * time.Second
		}

		if attempt < g.maxRetries-1 {
			time.Sleep(sleepTime)
		}
	}

	return "", fmt.Errorf("all retries failed: %w", lastErr)
}

func extractText(resp *genai.GenerateContentResponse) string {
	var result strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					result.WriteString(string(txt))
				}
			}
		}
	}
	return result.String()
}

func (g *Gemini) Close() error {
	return g.client.Close()
}
