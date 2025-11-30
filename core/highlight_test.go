package core

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
)

func TestHighlighter_Highlight(t *testing.T) {
	h := NewHighlighter()

	tests := []struct {
		name     string
		code     string
		filename string
		language string
		contains string // substring that should be present (ANSI codes)
	}{
		{
			name:     "Go code",
			code:     "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
			filename: "main.go",
			language: "",
			contains: "package", // should contain the keyword
		},
		{
			name:     "Python code",
			code:     "def hello():\n    print('Hello')",
			filename: "hello.py",
			language: "",
			contains: "def",
		},
		{
			name:     "Explicit language",
			code:     "SELECT * FROM users",
			filename: "",
			language: "sql",
			contains: "SELECT",
		},
		{
			name:     "Plaintext fallback",
			code:     "Just some plain text",
			filename: "unknown.xyz",
			language: "",
			contains: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Highlight(tt.code, tt.filename, tt.language)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Highlight() result should contain %q, got %q", tt.contains, result)
			}
			// For known languages, result should contain ANSI escape codes
			if tt.language != "" || (tt.filename != "" && !strings.HasSuffix(tt.filename, ".xyz")) {
				if !strings.Contains(result, "\x1b[") {
					t.Errorf("Highlight() should contain ANSI codes for %s", tt.name)
				}
			}
		})
	}
}

func TestHighlighter_HighlightLines(t *testing.T) {
	h := NewHighlighter()

	code := "package main\n\nfunc main() {\n}"
	lines := h.HighlightLines(code, "main.go", "")

	if len(lines) != 4 {
		t.Errorf("HighlightLines() returned %d lines, want 4", len(lines))
	}
}

func TestHighlighter_SetStyle(t *testing.T) {
	h := NewHighlighter()

	// Default style should be monokai
	if h.GetStyleName() != "monokai" {
		t.Errorf("Default style should be monokai, got %s", h.GetStyleName())
	}

	// Change style
	if !h.SetStyle("dracula") {
		t.Error("SetStyle(dracula) should return true")
	}
	if h.GetStyleName() != "dracula" {
		t.Errorf("Style should be dracula after SetStyle, got %s", h.GetStyleName())
	}

	// Invalid style should not change current style and return false
	if h.SetStyle("nonexistent-style") {
		t.Error("SetStyle(nonexistent-style) should return false")
	}
	if h.GetStyleName() != "dracula" {
		t.Errorf("Style should remain dracula after invalid SetStyle, got %s", h.GetStyleName())
	}
}

func TestListStyles(t *testing.T) {
	styleNames := ListStyles()

	if len(styleNames) == 0 {
		t.Error("ListStyles() should return at least one style")
	}

	// Check for some known styles
	foundMonokai := false
	foundDracula := false
	for _, name := range styleNames {
		if name == "monokai" {
			foundMonokai = true
		}
		if name == "dracula" {
			foundDracula = true
		}
	}

	if !foundMonokai {
		t.Error("ListStyles() should include 'monokai'")
	}
	if !foundDracula {
		t.Error("ListStyles() should include 'dracula'")
	}
}

func TestListLanguages(t *testing.T) {
	languages := ListLanguages()

	if len(languages) == 0 {
		t.Error("ListLanguages() should return at least one language")
	}

	// Check for some common languages
	found := make(map[string]bool)
	for _, lang := range languages {
		found[strings.ToLower(lang)] = true
	}

	commonLangs := []string{"go", "python", "javascript", "java", "c"}
	for _, lang := range commonLangs {
		if !found[lang] {
			t.Errorf("ListLanguages() should include %q", lang)
		}
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"main.go", "Go"},
		{"script.py", "Python"},
		{"app.js", "JavaScript"},
		{"style.css", "CSS"},
		{"data.json", "JSON"},
		{"unknown.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := DetectLanguage(tt.filename)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetLanguageByExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".go", "Go"},
		{".py", "Python"},
		{".js", "JavaScript"},
		{"go", "Go"},  // without dot
		{"py", "Python"},
		{".unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := GetLanguageByExtension(tt.ext)
			if result != tt.expected {
				t.Errorf("GetLanguageByExtension(%q) = %q, want %q", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/file.go", ".go"},
		{"file.py", ".py"},
		{"noext", ""},
		{"/path/to/file.tar.gz", ".gz"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := GetFileExtension(tt.path)
			if result != tt.expected {
				t.Errorf("GetFileExtension(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	code := "func main() {}"

	// Test HighlightCode
	result := HighlightCode(code, "main.go", "")
	if !strings.Contains(result, "func") {
		t.Error("HighlightCode() should contain 'func'")
	}

	// Test HighlightCodeLines
	lines := HighlightCodeLines(code, "main.go", "")
	if len(lines) == 0 {
		t.Error("HighlightCodeLines() should return at least one line")
	}
}

func TestNewSimpleLexer(t *testing.T) {
	config := &LexerConfig{
		Name:      "TestLang",
		Aliases:   []string{"test", "tl"},
		Filenames: []string{"*.tl", "*.test"},
		MimeTypes: []string{"text/x-testlang"},
		Rules: map[string][]PatternRule{
			"root": {
				{Pattern: `\b(keyword1|keyword2)\b`, TokenType: chroma.Keyword},
				{Pattern: `"[^"]*"`, TokenType: chroma.String},
				{Pattern: `//.*`, TokenType: chroma.Comment},
				{Pattern: `\d+`, TokenType: chroma.Number},
			},
		},
	}

	lexer, err := NewSimpleLexer(config)
	if err != nil {
		t.Fatalf("NewSimpleLexer() error = %v", err)
	}

	if lexer == nil {
		t.Fatal("NewSimpleLexer() returned nil lexer")
	}

	// Register and test
	err = RegisterLexer(lexer)
	if err != nil {
		t.Fatalf("RegisterLexer() error = %v", err)
	}

	// Test highlighting with the custom lexer
	h := NewHighlighter()
	result := h.Highlight(`keyword1 "hello" 123 // comment`, "", "TestLang")

	// Should contain ANSI codes
	if !strings.Contains(result, "\x1b[") {
		t.Error("Custom lexer should produce ANSI-highlighted output")
	}
}
