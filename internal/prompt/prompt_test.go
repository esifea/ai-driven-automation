package prompt

import (
	"testing"
)

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected Language
	}{
		{"go", LangGo},
		{"Go", LangGo},
		{"GO", LangGo},
		{"golang", LangGo},
		{"python", LangPython},
		{"Python", LangPython},
		{"py", LangPython},
		{"cpp", LangCpp},
		{"c++", LangCpp},
		{"C++", LangCpp},
		{"cxx", LangCpp},
		{"cc", LangCpp},
		{"", LangUnknown},
		{"java", LangUnknown},
		{"rust", LangUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLanguage(tt.input)
			if result != tt.expected {
				t.Errorf("parseLanguage(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseLanguageFromTask(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Language
	}{
		{
			name:     "language at start",
			input:    "LANGUAGE: python\n\n# Task",
			expected: LangPython,
		},
		{
			name:     "language in middle",
			input:    "# Task\n\nLANGUAGE: go\n\n## Objective",
			expected: LangGo,
		},
		{
			name:     "lowercase language",
			input:    "language: cpp",
			expected: LangCpp,
		},
		{
			name:     "no language",
			input:    "# Task\n\n## Objective",
			expected: LangUnknown,
		},
		{
			name:     "language with extra spaces",
			input:    "LANGUAGE:   python  ",
			expected: LangPython,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLanguageFromTask(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetectFromFiles(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected Language
	}{
		{
			name:     "all go files",
			files:    []string{"main.go", "handler.go", "util.go"},
			expected: LangGo,
		},
		{
			name:     "all python files",
			files:    []string{"main.py", "handler.py"},
			expected: LangPython,
		},
		{
			name:     "all cpp files",
			files:    []string{"main.cpp", "handler.hpp", "util.h"},
			expected: LangCpp,
		},
		{
			name:     "mixed - go majority",
			files:    []string{"main.go", "handler.go", "script.py"},
			expected: LangGo,
		},
		{
			name:     "mixed - python majority",
			files:    []string{"main.py", "handler.py", "util.py", "build.go"},
			expected: LangPython,
		},
		{
			name:     "no recognized files",
			files:    []string{"main.java", "handler.rs"},
			expected: LangUnknown,
		},
		{
			name:     "empty list",
			files:    []string{},
			expected: LangUnknown,
		},
		{
			name:     "cpp header files",
			files:    []string{"handler.h", "util.hpp"},
			expected: LangCpp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectFromFiles(tt.files)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDetectLanguage_Priority(t *testing.T) {
	// Priority: env > task > target files > changed files > default

	t.Run("env var takes priority", func(t *testing.T) {
		result := DetectLanguage(
			"python",               // env var
			"LANGUAGE: go",         // task content
			[]string{"main.cpp"},   // target files
			[]string{"handler.go"}, // changed files
		)
		if result != LangPython {
			t.Errorf("env var should take priority, got %v", result)
		}
	})

	t.Run("task content second priority", func(t *testing.T) {
		result := DetectLanguage(
			"",                     // no env var
			"LANGUAGE: go",         // task content
			[]string{"main.cpp"},   // target files
			[]string{"handler.py"}, // changed files
		)
		if result != LangGo {
			t.Errorf("task content should be second priority, got %v", result)
		}
	})

	t.Run("target files third priority", func(t *testing.T) {
		result := DetectLanguage(
			"",                        // no env var
			"# No language specified", // no language in task
			[]string{"main.cpp"},      // target files
			[]string{"handler.py"},    // changed files
		)
		if result != LangCpp {
			t.Errorf("target files should be third priority, got %v", result)
		}
	})

	t.Run("changed files fourth priority", func(t *testing.T) {
		result := DetectLanguage(
			"",                        // no env var
			"# No language specified", // no language in task
			[]string{},                // no target files
			[]string{"handler.py"},    // changed files
		)
		if result != LangPython {
			t.Errorf("changed files should be fourth priority, got %v", result)
		}
	})

	t.Run("default to go", func(t *testing.T) {
		result := DetectLanguage(
			"",                        // no env var
			"# No language specified", // no language in task
			[]string{},                // no target files
			[]string{},                // no changed files
		)
		if result != LangGo {
			t.Errorf("should default to Go, got %v", result)
		}
	})
}

func TestGetLanguagePrompt(t *testing.T) {
	t.Run("go prompt", func(t *testing.T) {
		prompt := GetLanguagePrompt(LangGo)
		if prompt.Language != LangGo {
			t.Errorf("expected LangGo, got %v", prompt.Language)
		}
		if prompt.Standards == "" {
			t.Error("Go prompt should have standards")
		}
		if prompt.Patterns == "" {
			t.Error("Go prompt should have patterns")
		}
		if prompt.AntiPatterns == "" {
			t.Error("Go prompt should have anti-patterns")
		}
	})

	t.Run("python prompt", func(t *testing.T) {
		prompt := GetLanguagePrompt(LangPython)
		if prompt.Language != LangPython {
			t.Errorf("expected LangPython, got %v", prompt.Language)
		}
		if prompt.Standards == "" {
			t.Error("Python prompt should have standards")
		}
	})

	t.Run("cpp prompt", func(t *testing.T) {
		prompt := GetLanguagePrompt(LangCpp)
		if prompt.Language != LangCpp {
			t.Errorf("expected LangCpp, got %v", prompt.Language)
		}
		if prompt.Standards == "" {
			t.Error("C++ prompt should have standards")
		}
	})

	t.Run("unknown returns empty", func(t *testing.T) {
		prompt := GetLanguagePrompt(LangUnknown)
		if prompt.Language != LangUnknown {
			t.Errorf("expected LangUnknown, got %v", prompt.Language)
		}
	})
}

func TestFormatForPrompt(t *testing.T) {
	t.Run("formats language prompt", func(t *testing.T) {
		prompt := GetLanguagePrompt(LangGo)
		formatted := prompt.FormatForPrompt()

		if formatted == "" {
			t.Error("formatted prompt should not be empty")
		}
		if !contains(formatted, "GO") {
			t.Error("should contain language name")
		}
		if !contains(formatted, "LANGUAGE-SPECIFIC STANDARDS") {
			t.Error("should contain header")
		}
	})

	t.Run("unknown returns empty string", func(t *testing.T) {
		prompt := GetLanguagePrompt(LangUnknown)
		formatted := prompt.FormatForPrompt()

		if formatted != "" {
			t.Error("unknown language should return empty string")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
