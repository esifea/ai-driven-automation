package aicontext

import (
	"testing"
)

func TestParseTaskMetadata_TargetFiles(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "basic target files",
			input: `# Task 01: Add Auth

TARGET FILES:
- internal/auth/handler.go
- internal/auth/middleware.go

## Objective
Add authentication.`,
			expected: []string{
				"internal/auth/handler.go",
				"internal/auth/middleware.go",
			},
		},
		{
			name: "target files with backticks",
			input: `TARGET FILES:
- ` + "`internal/auth/handler.go`" + `
- ` + "`internal/auth/middleware.go`" + `
`,
			expected: []string{
				"internal/auth/handler.go",
				"internal/auth/middleware.go",
			},
		},
		{
			name: "target files with comments",
			input: `TARGET FILES:
- internal/auth/handler.go # main handler
- internal/auth/middleware.go # auth middleware
`,
			expected: []string{
				"internal/auth/handler.go",
				"internal/auth/middleware.go",
			},
		},
		{
			name: "target files with leading ./",
			input: `TARGET FILES:
- ./internal/auth/handler.go
- ./internal/auth/middleware.go
`,
			expected: []string{
				"./internal/auth/handler.go",
				"./internal/auth/middleware.go",
			},
		},
		{
			name: "no target files section",
			input: `# Task 01: Add Auth

## Objective
Add authentication.`,
			expected: nil,
		},
		{
			name: "empty target files section",
			input: `TARGET FILES:

## Objective
Add authentication.`,
			expected: nil,
		},
		{
			name: "target files ends at next header",
			input: `TARGET FILES:
- file1.go
- file2.go
## Next Section
- not_a_file.go`,
			expected: []string{
				"file1.go",
				"file2.go",
			},
		},
		{
			name: "case insensitive TARGET FILES",
			input: `target files:
- file1.go
- file2.go
`,
			expected: []string{
				"file1.go",
				"file2.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := ParseTaskMetadata(tt.input)

			if len(meta.TargetFiles) != len(tt.expected) {
				t.Errorf("expected %d files, got %d", len(tt.expected), len(meta.TargetFiles))
				t.Errorf("expected: %v", tt.expected)
				t.Errorf("got: %v", meta.TargetFiles)
				return
			}

			for i, exp := range tt.expected {
				if meta.TargetFiles[i] != exp {
					t.Errorf("file %d: expected %q, got %q", i, exp, meta.TargetFiles[i])
				}
			}
		})
	}
}

func TestParseTaskMetadata_Language(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "language specified",
			input: `# Task 01
LANGUAGE: python
TARGET FILES:
- src/main.py
`,
			expected: "python",
		},
		{
			name: "language with extra spaces",
			input: `LANGUAGE:   go  
`,
			expected: "go",
		},
		{
			name: "language lowercase",
			input: `language: cpp
`,
			expected: "cpp",
		},
		{
			name: "no language specified",
			input: `# Task 01
TARGET FILES:
- src/main.go
`,
			expected: "",
		},
		{
			name: "language after target files",
			input: `TARGET FILES:
- src/main.py
LANGUAGE: python
`,
			expected: "python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := ParseTaskMetadata(tt.input)

			if meta.Language != tt.expected {
				t.Errorf("expected language %q, got %q", tt.expected, meta.Language)
			}
		})
	}
}

func TestParseTaskMetadata_Content(t *testing.T) {
	input := `# Task 01: Test
Some content here.`

	meta := ParseTaskMetadata(input)

	if meta.Content != input {
		t.Errorf("Content should be preserved exactly")
	}
}

func TestIsTargetFile(t *testing.T) {
	meta := &TaskMetadata{
		TargetFiles: []string{
			"internal/auth/handler.go",
			"./src/main.py",
		},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"internal/auth/handler.go", true},
		{"./internal/auth/handler.go", true},
		{"src/main.py", true},
		{"./src/main.py", true},
		{"internal/auth/other.go", false},
		{"handler.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := meta.IsTargetFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsTargetFile(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}
