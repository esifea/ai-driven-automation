package aicontext

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
)

func ExtractSignatures(path, content string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".go":
		return extractGoSignatures(path, content)
	case ".py":
		return extractPythonSignatures(content)
	case ".cpp", ".cc", ".cxx", ".h", ".hpp":
		return extractCppSignatures(content)
	default:
		return truncateLines(content, 30) // unknown types
	}
}

func extractGoSignatures(path, content string) string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		// Fallback to regex if parsing fails
		return extractGoSignaturesRegex(content)
	}

	var b strings.Builder

	// Package declaration
	b.WriteString("package ")
	b.WriteString(file.Name.Name)
	b.WriteString("\n\n")

	// Imports
	if len(file.Imports) > 0 {
		b.WriteString("import (...)\n\n")
	}

	// Declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Type declarations (struct, interface)
					writeTypeSignature(&b, s)
				case *ast.ValueSpec:
					// Constants and variables
					if d.Tok == token.CONST || d.Tok == token.VAR {
						writeValueSignature(&b, d.Tok, s)
					}
				}
			}
		case *ast.FuncDecl:
			// Function/method declarations
			writeFuncSignature(&b, d)
		}
	}

	return b.String()
}

func extractGoSignaturesRegex(content string) string {
	var b strings.Builder

	// Package
	pkgRe := regexp.MustCompile(`(?m)^package\s+(\w+)`)
	if m := pkgRe.FindStringSubmatch(content); m != nil {
		b.WriteString("package ")
		b.WriteString(m[1])
		b.WriteString("\n\n")
	}

	// Functions
	funcRe := regexp.MustCompile(`(?m)^func\s+(\([^)]+\)\s+)?(\w+)\s*\([^)]*\)[^{]*`)
	for _, m := range funcRe.FindAllString(content, -1) {
		b.WriteString(strings.TrimSpace(m))
		b.WriteString("\n")
	}

	// Types
	typeRe := regexp.MustCompile(`(?m)^type\s+(\w+)\s+(struct|interface)\s*\{`)
	for _, m := range typeRe.FindAllStringSubmatch(content, -1) {
		b.WriteString("type ")
		b.WriteString(m[1])
		b.WriteString(" ")
		b.WriteString(m[2])
		b.WriteString(" {...}\n")
	}

	return b.String()
}

func extractPythonSignatures(content string) string {
	var b strings.Builder
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Import statements
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
			if i < 20 { // Only first 20 lines of imports
				b.WriteString(trimmed)
				b.WriteString("\n")
			}
			continue
		}

		// Class definitions
		if strings.HasPrefix(trimmed, "class ") {
			b.WriteString("\n")
			b.WriteString(trimmed)
			b.WriteString("\n")
			continue
		}

		// Function/method definitions
		if strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "async def ") {
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			b.WriteString(strings.Repeat("    ", indent/4))
			// Get just the signature line
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				b.WriteString(trimmed[:idx+1])
			} else {
				b.WriteString(trimmed)
			}
			b.WriteString(" ...\n")
			continue
		}

		// Decoration
		if strings.HasPrefix(trimmed, "@") {
			b.WriteString(trimmed)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func extractCppSignatures(content string) string {
	var b strings.Builder
	lines := strings.Split(content, "\n")

	inClass := false
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Include statements
		if strings.HasPrefix(trimmed, "#include") {
			b.WriteString(trimmed)
			b.WriteString("\n")
			continue
		}

		// Namespace
		if strings.HasPrefix(trimmed, "namespace ") {
			b.WriteString(trimmed)
			b.WriteString("\n")
			continue
		}

		// Class/struct declarations
		if strings.HasPrefix(trimmed, "class ") || strings.HasPrefix(trimmed, "struct ") {
			b.WriteString("\n")
			b.WriteString(trimmed)
			if !strings.HasSuffix(trimmed, ";") {
				inClass = true
				braceDepth = strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			}
			b.WriteString("\n")
			continue
		}

		// Inside class
		if inClass {
			openBraces := strings.Count(trimmed, "{")
			closeBraces := strings.Count(trimmed, "}")

			// Check closing brace
			if trimmed == "};" || trimmed == "}" {
				braceDepth -= closeBraces
				if braceDepth <= 0 {
					inClass = false
					b.WriteString("};\n")
				}
				continue
			}

			braceDepth += openBraces - closeBraces

			// Access specifiers
			if trimmed == "public:" || trimmed == "private:" || trimmed == "protected:" {
				b.WriteString(trimmed)
				b.WriteString("\n")
				continue
			}

			// Skip implementation
			if trimmed == "{" ||
				strings.HasPrefix(trimmed, "return ") ||
				strings.HasPrefix(trimmed, "if ") ||
				strings.HasPrefix(trimmed, "if(") ||
				strings.HasPrefix(trimmed, "for ") ||
				strings.HasPrefix(trimmed, "for(") ||
				strings.HasPrefix(trimmed, "while ") ||
				strings.HasPrefix(trimmed, "//") {
				continue
			}

			// Method/member declarations
			isDeclaration := false

			// Member variable: ends with ;
			if strings.HasSuffix(trimmed, ";") && !strings.Contains(trimmed, "(") {
				isDeclaration = true
			}

			// Method declaration: has ( and )
			if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
				isDeclaration = true
			}

			if isDeclaration {
				sig := trimmed
				// Remove inline implementation body
				if idx := strings.Index(sig, "{"); idx > 0 {
					sig = strings.TrimSpace(sig[:idx])
					if !strings.HasSuffix(sig, ";") {
						sig += ";"
					}
				}
				b.WriteString("    ")
				b.WriteString(sig)
				b.WriteString("\n")
			}
			continue
		}

		// Function declarations (outside class)
		if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
			// Skip function call
			if strings.HasPrefix(trimmed, "return ") || strings.HasPrefix(trimmed, "if ") {
				continue
			}

			sig := trimmed
			if idx := strings.Index(sig, "{"); idx > 0 {
				sig = strings.TrimSpace(sig[:idx])
				if !strings.HasSuffix(sig, ";") {
					sig += ";"
				}
			}
			b.WriteString(sig)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func writeTypeSignature(b *strings.Builder, spec *ast.TypeSpec) {
	b.WriteString("type ")
	b.WriteString(spec.Name.Name)

	switch t := spec.Type.(type) {
	case *ast.StructType:
		b.WriteString(" struct {")
		if t.Fields != nil && len(t.Fields.List) > 0 {
			b.WriteString(" /* ")
			b.WriteString(string(rune(len(t.Fields.List))))
			b.WriteString(" fields */ ")
		}
		b.WriteString("}\n")
	case *ast.InterfaceType:
		b.WriteString(" interface {")
		if t.Methods != nil && len(t.Methods.List) > 0 {
			b.WriteString(" /* ")
			b.WriteString(string(rune(len(t.Methods.List))))
			b.WriteString(" methods */ ")
		}
		b.WriteString("}\n")
	default:
		b.WriteString(" ...\n")
	}
}

func writeValueSignature(b *strings.Builder, tok token.Token, spec *ast.ValueSpec) {
	if tok == token.CONST {
		b.WriteString("const ")
	} else {
		b.WriteString("var ")
	}

	for i, name := range spec.Names {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(name.Name)
	}
	b.WriteString(" ...\n")
}

func writeFuncSignature(b *strings.Builder, decl *ast.FuncDecl) {
	b.WriteString("func ")

	// Receiver
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		b.WriteString("(")
		for _, field := range decl.Recv.List {
			if len(field.Names) > 0 {
				b.WriteString(field.Names[0].Name)
				b.WriteString(" ")
			}
			writeType(b, field.Type)
		}
		b.WriteString(") ")
	}

	// Function name
	b.WriteString(decl.Name.Name)

	// Parameters
	b.WriteString("(")
	writeFieldList(b, decl.Type.Params)
	b.WriteString(")")

	// Return types
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		b.WriteString(" ")
		if len(decl.Type.Results.List) > 1 {
			b.WriteString("(")
		}
		writeFieldList(b, decl.Type.Results)
		if len(decl.Type.Results.List) > 1 {
			b.WriteString(")")
		}
	}

	b.WriteString("\n")
}

func writeFieldList(b *strings.Builder, fields *ast.FieldList) {
	if fields == nil {
		return
	}

	for i, field := range fields.List {
		if i > 0 {
			b.WriteString(", ")
		}

		// Parameter names
		for j, name := range field.Names {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(name.Name)
		}

		if len(field.Names) > 0 {
			b.WriteString(" ")
		}

		writeType(b, field.Type)
	}
}

func writeType(b *strings.Builder, expr ast.Expr) {
	switch t := expr.(type) {
	case *ast.Ident:
		b.WriteString(t.Name)
	case *ast.StarExpr:
		b.WriteString("*")
		writeType(b, t.X)
	case *ast.ArrayType:
		b.WriteString("[]")
		writeType(b, t.Elt)
	case *ast.MapType:
		b.WriteString("map[")
		writeType(b, t.Key)
		b.WriteString("]")
		writeType(b, t.Value)
	case *ast.SelectorExpr:
		writeType(b, t.X)
		b.WriteString(".")
		b.WriteString(t.Sel.Name)
	case *ast.InterfaceType:
		b.WriteString("interface{}")
	case *ast.FuncType:
		b.WriteString("func(...)")
	default:
		b.WriteString("...")
	}
}

func truncateLines(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	return strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
}
