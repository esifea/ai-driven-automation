package aicontext

import (
	"strings"
	"testing"
)

func TestExtractGoSignatures(t *testing.T) {
	input := `package auth

import (
	"context"
	"fmt"
)

// Handler manages authentication
type Handler struct {
	db    *sql.DB
	cache *redis.Client
}

// Config holds configuration
type Config struct {
	Secret string
	TTL    int
}

// Authenticator defines auth interface
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (bool, error)
	Refresh(token string) (string, error)
}

// Login authenticates a user
func (h *Handler) Login(ctx context.Context, req LoginRequest) (*Token, error) {
	// Implementation details...
	user, err := h.db.GetUser(req.Username)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &Token{Value: "token"}, nil
}

// Logout invalidates the token
func (h *Handler) Logout(ctx context.Context, token string) error {
	return h.cache.Delete(token)
}

// NewHandler creates a new handler
func NewHandler(db *sql.DB, cache *redis.Client) *Handler {
	return &Handler{db: db, cache: cache}
}

var defaultTimeout = 30
const maxRetries = 3
`

	result := ExtractSignatures("test.go", input)

	// Should contain package
	if !strings.Contains(result, "package auth") {
		t.Error("should contain package declaration")
	}

	// Should contain imports (collapsed)
	if !strings.Contains(result, "import (...)") {
		t.Error("should contain collapsed imports")
	}

	// Should contain type declarations
	if !strings.Contains(result, "type Handler struct") {
		t.Error("should contain Handler struct")
	}
	if !strings.Contains(result, "type Config struct") {
		t.Error("should contain Config struct")
	}
	if !strings.Contains(result, "type Authenticator interface") {
		t.Error("should contain Authenticator interface")
	}

	// Should contain function signatures
	if !strings.Contains(result, "func (h *Handler) Login(") {
		t.Error("should contain Login method signature")
	}
	if !strings.Contains(result, "func (h *Handler) Logout(") {
		t.Error("should contain Logout method signature")
	}
	if !strings.Contains(result, "func NewHandler(") {
		t.Error("should contain NewHandler function signature")
	}

	// Should NOT contain implementation details
	if strings.Contains(result, "GetUser") {
		t.Error("should not contain implementation details")
	}
	if strings.Contains(result, "fmt.Errorf") {
		t.Error("should not contain implementation code")
	}

	t.Logf("Extracted signatures:\n%s", result)
}

func TestExtractPythonSignatures(t *testing.T) {
	input := `import os
from typing import Optional
from dataclasses import dataclass

@dataclass
class Config:
    secret: str
    ttl: int = 30

class AuthHandler:
    """Handles authentication."""
    
    def __init__(self, db: Database, cache: Cache) -> None:
        self.db = db
        self.cache = cache
    
    async def login(self, username: str, password: str) -> Optional[Token]:
        """Authenticate user and return token."""
        user = await self.db.get_user(username)
        if not user:
            return None
        if not verify_password(password, user.password_hash):
            return None
        return Token(user_id=user.id)
    
    def logout(self, token: str) -> bool:
        """Invalidate the given token."""
        return self.cache.delete(token)

def create_handler(db: Database) -> AuthHandler:
    """Factory function for AuthHandler."""
    cache = Cache()
    return AuthHandler(db, cache)
`

	result := ExtractSignatures("test.py", input)

	// Should contain imports
	if !strings.Contains(result, "import os") {
		t.Error("should contain import")
	}
	if !strings.Contains(result, "from typing import Optional") {
		t.Error("should contain from import")
	}

	// Should contain class definitions
	if !strings.Contains(result, "class Config:") {
		t.Error("should contain Config class")
	}
	if !strings.Contains(result, "class AuthHandler:") {
		t.Error("should contain AuthHandler class")
	}

	// Should contain decorators
	if !strings.Contains(result, "@dataclass") {
		t.Error("should contain dataclass decorator")
	}

	// Should contain function/method signatures
	if !strings.Contains(result, "def __init__(self") {
		t.Error("should contain __init__ signature")
	}
	if !strings.Contains(result, "async def login(self") {
		t.Error("should contain login signature")
	}
	if !strings.Contains(result, "def logout(self") {
		t.Error("should contain logout signature")
	}
	if !strings.Contains(result, "def create_handler(") {
		t.Error("should contain create_handler signature")
	}

	// Should NOT contain implementation details
	if strings.Contains(result, "await self.db.get_user") {
		t.Error("should not contain implementation details")
	}

	t.Logf("Extracted signatures:\n%s", result)
}

func TestExtractCppSignatures(t *testing.T) {
	input := `#include <string>
#include <memory>

namespace auth {

class Config {
public:
    std::string secret;
    int ttl = 30;
};

class Handler {
public:
    Handler(Database* db, Cache* cache);
    ~Handler();
    
    Token* login(const std::string& username, const std::string& password);
    bool logout(const std::string& token);

private:
    Database* db_;
    Cache* cache_;
    
    bool validatePassword(const std::string& password, const std::string& hash) {
        // implementation
        return password == hash;
    }
};

Handler* createHandler(Database* db);

}  // namespace auth
`

	result := ExtractSignatures("test.hpp", input)

	// Should contain includes
	if !strings.Contains(result, "#include <string>") {
		t.Error("should contain #include")
	}

	// Should contain namespace
	if !strings.Contains(result, "namespace auth") {
		t.Error("should contain namespace")
	}

	// Should contain class declarations
	if !strings.Contains(result, "class Config") {
		t.Error("should contain Config class")
	}
	if !strings.Contains(result, "class Handler") {
		t.Error("should contain Handler class")
	}

	// Should contain public/private markers
	if !strings.Contains(result, "public:") {
		t.Error("should contain public:")
	}
	if !strings.Contains(result, "private:") {
		t.Error("should contain private:")
	}

	// Should contain method declarations
	if !strings.Contains(result, "login(") {
		t.Error("should contain login declaration")
	}
	if !strings.Contains(result, "logout(") {
		t.Error("should contain logout declaration")
	}

	// Should not contain implementation
	if strings.Contains(result, "return password == hash") {
		t.Error("should not contain implementation details")
	}

	t.Logf("Extracted signatures:\n%s", result)
}

func TestExtractSignatures_UnknownExtension(t *testing.T) {
	input := `line 1
line 2
line 3
` + strings.Repeat("line N\n", 50)

	result := ExtractSignatures("test.xyz", input)

	// Should truncate to ~30 lines
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) > 35 {
		t.Errorf("should truncate unknown files, got %d lines", len(lines))
	}

	if !strings.Contains(result, "truncated") {
		t.Error("should indicate truncation")
	}
}

func TestExtractGoSignatures_InvalidSyntax(t *testing.T) {
	// Fallback to regex-based extraction
	input := `package main

func broken( {
	// missing closing paren
}

func valid() string {
	return "ok"
}
`

	result := ExtractSignatures("test.go", input)

	if !strings.Contains(result, "package main") {
		t.Error("should contain package even with invalid syntax")
	}

	t.Logf("Fallback extraction result:\n%s", result)
}
