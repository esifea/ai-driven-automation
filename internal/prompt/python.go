package prompt

func getPythonPrompt() *LanguagePrompt {
	return &LanguagePrompt{
		Language: LangPython,
		Standards: `Python Coding Standards (PEP 8 + Modern Python):

1. FORMATTING
   - Use 2 spaces for indentation (no tabs)
   - Maximum line length: 88 characters (Black default)
   - Use blank lines to separate functions and classes
   - Imports at top: stdlib, third-party, local (separated by blank lines)

2. NAMING
   - Variables/functions: snake_case
   - Classes: PascalCase
   - Constants: UPPER_SNAKE_CASE
   - Private: _single_leading_underscore
   - Name mangling: __double_leading_underscore (rare)

3. TYPE HINTS (Required for Python 3.9+)
   - All function parameters and return types must have hints
   - Use built-in types: list, dict, set (not typing.List)
   - Use | for unions: str | None (not Optional[str])
   - Use typing.TypeAlias for complex types

4. ERROR HANDLING
   - Be specific with exceptions: except ValueError, not except Exception
   - Use context managers (with statement) for resources
   - Raise exceptions with meaningful messages
   - Use custom exception classes for domain errors

5. DOCUMENTATION
   - All public modules, classes, and functions need docstrings
   - Use Google-style or NumPy-style docstrings consistently
   - Include Args, Returns, Raises sections

6. STRUCTURE
   - One class per file for large classes
   - Use __all__ to define public API
   - Use if __name__ == "__main__": for scripts`,

		Patterns: `- Use dataclasses or Pydantic for data structures
- Use pathlib.Path instead of os.path
- Use f-strings for string formatting
- Use list/dict/set comprehensions when readable
- Use generators for large sequences
- Use @property for computed attributes
- Use enum.Enum for fixed choices
- Use functools.lru_cache for memoization`,

		AntiPatterns: `- Don't use mutable default arguments: def f(x=[])
- Don't use bare except: clauses
- Don't use wildcard imports: from x import *
- Don't use single-letter names except for loops/lambdas
- Don't use global keyword
- Don't use type() for type checking, use isinstance()
- Don't use string concatenation in loops, use join()
- Don't ignore type checker errors with # type: ignore without reason`,
	}
}
