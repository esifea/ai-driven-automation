package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Provider
	Provider string // claude, gemini, ...
	APIKey   string

	// Task
	Mode     string // coder, reviewer
	TaskID   string
	PRNumber string

	// Q&A
	PRQuestion       string
	CommentPath      string
	CommentStartLine string
	CommentEndLine   string

	Feedback string

	BaseBranch   string
	ChangedFiles string
	MaxRetries   int
}

func Load() *Config {
	return &Config{
		Provider:         getEnv("AGENT_PROVIDER", "gemini"),
		APIKey:           getEnv("GEMINI_API_KEY", ""), // TODO: support multi agent
		Mode:             getEnv("MODE", "coder"),
		TaskID:           getEnv("TASK_ID", "01"),
		PRNumber:         getEnv("PR_NUMBER", ""),
		PRQuestion:       getEnv("PR_QUESTION", ""),
		CommentPath:      getEnv("COMMENT_PATH", ""),
		CommentStartLine: getEnv("COMMENT_START_LINE", ""),
		CommentEndLine:   getEnv("COMMENT_END_LINE", ""),
		Feedback:         getEnv("FEEDBACK", ""),
		BaseBranch:       getEnv("BASE_BRANCH", ""),
		ChangedFiles:     getEnv("CHANGED_FILES", ""),
		MaxRetries:       getEnvInt("MAX_RETRIES", 5),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return fallback
}
