package parser

import (
	"testing"
)

func TestParseFiles_Basic(t *testing.T) {
	input := `Here is the implementation:

### File: internal/auth/handler.go
` + "```go" + `
package auth

func Login() error {
    return nil
}
` + "```" + `

### File: internal/auth/middleware.go
` + "```go" + `
package auth

func Middleware() {}
` + "```" + `
`

	files := ParseFiles(input)

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	if _, ok := files["internal/auth/handler.go"]; !ok {
		t.Error("should contain handler.go")
	}

	if _, ok := files["internal/auth/middleware.go"]; !ok {
		t.Error("should contain middleware.go")
	}

	// Check content
	handler := files["internal/auth/handler.go"]
	if !containsString(handler, "func Login()") {
		t.Error("handler.go should contain Login function")
	}
}

func TestParseFiles_NoCodeBlock(t *testing.T) {
	input := `### File: main.go
package main

func main() {
    println("hello")
}

### File: util.go
package main

func helper() {}
`

	files := ParseFiles(input)

	t.Logf("Parsed %d files", len(files))
}

func TestParseFiles_DifferentLanguages(t *testing.T) {
	input := `### File: main.py
` + "```python" + `
def main():
    print("hello")
` + "```" + `

### File: handler.cpp
` + "```cpp" + `
#include <iostream>
void handler() {}
` + "```" + `

### File: config.go
` + "```go" + `
package config
var Debug = true
` + "```" + `
`

	files := ParseFiles(input)

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	if _, ok := files["main.py"]; !ok {
		t.Error("should contain main.py")
	}
	if _, ok := files["handler.cpp"]; !ok {
		t.Error("should contain handler.cpp")
	}
	if _, ok := files["config.go"]; !ok {
		t.Error("should contain config.go")
	}
}

func TestParseFiles_EmptyInput(t *testing.T) {
	files := ParseFiles("")
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestParseFiles_NoFiles(t *testing.T) {
	input := `I cannot help with that request.`
	files := ParseFiles(input)
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestParseFiles_TrimsContent(t *testing.T) {
	input := `### File: test.go
` + "```go" + `

package main

func test() {}

` + "```" + `
`

	files := ParseFiles(input)
	content := files["test.go"]

	if content[0] == '\n' {
		t.Error("should trim leading newlines")
	}
	if content[len(content)-1] == '\n' {
		t.Error("should trim trailing newlines")
	}
}

func TestParseFiles_PathWithSpaces(t *testing.T) {
	input := `### File:   internal/auth/handler.go  
` + "```go" + `
package auth
` + "```" + `
`

	files := ParseFiles(input)

	if _, ok := files["internal/auth/handler.go"]; !ok {
		t.Error("should handle paths with surrounding spaces")
		t.Logf("Got files: %v", files)
	}
}

func TestParseFiles_MultipleCodeBlocksPerFile(t *testing.T) {
	input := `### File: test.go

Some explanation here.

` + "```go" + `
package main

func main() {}
` + "```" + `

More explanation after.
`

	files := ParseFiles(input)
	content := files["test.go"]

	if containsString(content, "explanation") {
		t.Error("should not include text outside code blocks")
	}
	if !containsString(content, "func main()") {
		t.Error("should include code inside block")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
