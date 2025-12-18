package prompt

import (
	"os"
	"path/filepath"
	"strings"
)

type Language string

const (
	LangGo      Language = "go"
	LangPython  Language = "python"
	LangCpp     Language = "cpp"
	LangUnknown Language = "unknown"
)

type LanguagePrompt struct {
	Language     Language
	Standards    string
	Patterns     string
	AntiPatterns string
}

func GetLanguagePrompt(lang Language) *LanguagePrompt {
	// Load from docs/tasks/languages/
	if custom := loadCustomLanguageDoc(lang); custom != "" {
		return &LanguagePrompt{
			Language:  lang,
			Standards: custom,
		}
	}

	// Fall back to built-in prompts
	switch lang {
	case LangGo:
		return getGoPrompt()
	case LangPython:
		return getPythonPrompt()
	case LangCpp:
		return getCppPrompt()
	default:
		return &LanguagePrompt{Language: LangUnknown}
	}
}

func loadCustomLanguageDoc(lang Language) string {
	path := filepath.Join("docs", "tasks", "languages", string(lang)+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func DetectLanguage(envVar string, taskContent string, targetFiles []string, changedFiles []string) Language {
	// Priority 1: Environment variable
	if envVar != "" {
		if lang := parseLanguage(envVar); lang != LangUnknown {
			return lang
		}
	}

	// Priority 2: LANGUAGE: in task file
	if lang := parseLanguageFromTask(taskContent); lang != LangUnknown {
		return lang
	}

	// Priority 3: TARGET FILES extensions
	if lang := detectFromFiles(targetFiles); lang != LangUnknown {
		return lang
	}

	// Priority 4: Changed files (Q&A mode)
	if lang := detectFromFiles(changedFiles); lang != LangUnknown {
		return lang
	}

	// Default: Go
	return LangGo
}

func parseLanguage(s string) Language {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "go", "golang":
		return LangGo
	case "python", "py":
		return LangPython
	case "cpp", "c++", "cxx", "cc":
		return LangCpp
	default:
		return LangUnknown
	}
}

func parseLanguageFromTask(content string) Language {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(trimmed), "LANGUAGE:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "LANGUAGE:"))
			value = strings.TrimSpace(strings.TrimPrefix(value, "language:"))
			return parseLanguage(value)
		}
	}
	return LangUnknown
}

func detectFromFiles(files []string) Language {
	counts := make(map[Language]int)

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		switch ext {
		case ".go":
			counts[LangGo]++
		case ".py":
			counts[LangPython]++
		case ".cpp", ".cc", ".cxx", ".h", ".hpp":
			counts[LangCpp]++
		}
	}

	// Return the most common language
	var maxLang Language = LangUnknown
	var maxCount int
	for lang, count := range counts {
		if count > maxCount {
			maxLang = lang
			maxCount = count
		}
	}

	return maxLang
}

func (p *LanguagePrompt) FormatForPrompt() string {
	if p.Language == LangUnknown {
		return ""
	}

	var b strings.Builder
	b.WriteString("=== LANGUAGE-SPECIFIC STANDARDS (")
	b.WriteString(strings.ToUpper(string(p.Language)))
	b.WriteString(") ===\n\n")

	if p.Standards != "" {
		b.WriteString(p.Standards)
		b.WriteString("\n\n")
	}

	if p.Patterns != "" {
		b.WriteString("RECOMMENDED PATTERNS:\n")
		b.WriteString(p.Patterns)
		b.WriteString("\n\n")
	}

	if p.AntiPatterns != "" {
		b.WriteString("AVOID:\n")
		b.WriteString(p.AntiPatterns)
		b.WriteString("\n")
	}

	return b.String()
}
