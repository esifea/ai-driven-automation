package prompt

func getGoPrompt() *LanguagePrompt {
	return &LanguagePrompt{
		Language: LangGo,
		Standards: `Go Coding Standards:

1. FORMATTING
   - Use gofmt/goimports for all code
   - No unused imports or variables
   - Line length: prefer under 100 characters

2. NAMING
   - Use MixedCaps or mixedCaps, not underscores
   - Acronyms should be consistent case: URL, HTTP, ID (not Url, Http, Id)
   - Interface names: Reader, Writer, Formatter (not IReader)
   - Single-method interfaces: name by method + "er" suffix

3. ERROR HANDLING
   - Always handle errors explicitly, never ignore with _
   - Wrap errors with context: fmt.Errorf("operation failed: %w", err)
   - Return early on errors, keep happy path unindented
   - Use errors.Is() and errors.As() for error checking

4. DOCUMENTATION
   - All exported types, functions, and packages need doc comments
   - Doc comments start with the name: "Handler manages HTTP requests"
   - Use complete sentences with proper punctuation

5. PACKAGES
   - Package names: short, lowercase, no underscores
   - Avoid package-level state when possible
   - One package per directory`,

		Patterns: `- Use defer for cleanup (files, locks, connections)
- Use context.Context for cancellation and timeouts
- Use table-driven tests
- Use functional options for complex constructors
- Use sync.Once for lazy initialization
- Return concrete types, accept interfaces
- Keep interfaces small (1-3 methods)`,

		AntiPatterns: `- Don't use init() unless absolutely necessary
- Don't use panic for normal error handling
- Don't use global variables for state
- Don't return interfaces from functions
- Don't use naked returns in long functions
- Don't ignore error returns
- Don't use dot imports except in tests`,
	}
}
