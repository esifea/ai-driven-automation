package prompt

func getCppPrompt() *LanguagePrompt {
	return &LanguagePrompt{
		Language: LangCpp,
		Standards: `C++ Coding Standards (C++17/20 Modern Style):

1. MEMORY MANAGEMENT
   - Use smart pointers: std::unique_ptr, std::shared_ptr
   - NEVER use raw new/delete for ownership
   - Use std::make_unique and std::make_shared
   - Follow RAII for all resource management

2. NAMING
   - Classes/Structs: PascalCase
   - Functions/Methods: snake_case (be consistent)
   - Variables: camelCase
   - Constants: kConstantName or UPPER_SNAKE_CASE
   - Private members: trailing underscore (member_)

3. CONST CORRECTNESS
   - Use const everywhere possible
   - Mark methods const if they don't modify state
   - Use constexpr for compile-time constants
   - Prefer const references for parameters: const std::string&

4. MODERN C++ FEATURES
   - Use auto for complex types, explicit for simple ones
   - Use range-based for loops
   - Use nullptr instead of NULL or 0
   - Use enum class instead of enum
   - Use [[nodiscard]] for functions with important return values
   - Use std::optional for optional values
   - Use std::variant instead of unions
   - Use structured bindings: auto [x, y] = pair;

5. ERROR HANDLING
   - Use exceptions for exceptional cases
   - Use std::expected (C++23) or return codes for expected failures
   - Use noexcept where appropriate
   - RAII ensures cleanup even with exceptions

6. HEADERS
   - Use #pragma once or include guards
   - Forward declare when possible
   - Include what you use (IWYU)
   - Order: related header, C system, C++ stdlib, other libs, project`,

		Patterns: `- Use std::string_view for read-only string parameters
- Use std::span for array parameters (C++20)
- Use std::array for fixed-size arrays
- Use std::vector for dynamic arrays
- Use std::unordered_map/set for O(1) lookup
- Use lambdas for callbacks and algorithms
- Use CRTP for static polymorphism
- Use std::move for transferring ownership`,

		AntiPatterns: `- Don't use raw pointers for ownership
- Don't use C-style casts, use static_cast/dynamic_cast
- Don't use C arrays, use std::array or std::vector
- Don't use #define for constants, use constexpr
- Don't use using namespace std; in headers
- Don't ignore compiler warnings
- Don't use std::endl (use '\n' for performance)
- Don't throw in destructors
- Don't slice objects (assign derived to base by value)`,
	}
}
