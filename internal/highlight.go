// Package core provides core functionality for the macro editor.
package internal

import (
	"bytes"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Highlighter provides syntax highlighting capabilities using Chroma.
// It supports automatic language detection based on file extension and
// allows runtime registration of custom lexers.
type Highlighter struct {
	style     *chroma.Style
	formatter chroma.Formatter
	mu        sync.RWMutex
}

// DefaultHighlighter is the global highlighter instance.
var DefaultHighlighter = NewHighlighter()

// NewHighlighter creates a new Highlighter with the default terminal256 style.
func NewHighlighter() *Highlighter {
	return &Highlighter{
		style:     styles.Get("monokai"),
		formatter: formatters.Get("terminal256"),
	}
}

// SetStyle changes the highlighting style by name.
// Available styles can be listed with ListStyles().
// Returns true if the style was found and set, false otherwise.
func (h *Highlighter) SetStyle(styleName string) bool {
	// Check if style exists in available styles
	found := false
	for _, name := range styles.Names() {
		if name == styleName {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if s := styles.Get(styleName); s != nil {
		h.style = s
		return true
	}
	return false
}

// GetStyleName returns the current style name.
func (h *Highlighter) GetStyleName() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.style.Name
}

// ListStyles returns all available style names.
func ListStyles() []string {
	return styles.Names()
}

// ListLanguages returns all available language names.
func ListLanguages() []string {
	var names []string
	for _, l := range lexers.GlobalLexerRegistry.Lexers {
		config := l.Config()
		if config != nil {
			names = append(names, config.Name)
		}
	}
	return names
}

// DetectLanguage attempts to detect the language from a filename.
// Returns empty string if no language could be detected.
func DetectLanguage(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		return ""
	}
	config := lexer.Config()
	if config == nil {
		return ""
	}
	return config.Name
}

// GetLanguageByExtension returns the language name for a file extension.
// The extension should include the dot (e.g., ".go", ".py").
func GetLanguageByExtension(ext string) string {
	// Normalize extension
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	// Create a fake filename to use Chroma's detection
	lexer := lexers.Match("file" + ext)
	if lexer == nil {
		return ""
	}
	config := lexer.Config()
	if config == nil {
		return ""
	}
	return config.Name
}

// Highlight applies syntax highlighting to the given code.
// It returns the highlighted string with ANSI escape codes for terminal display.
// If language is empty, it will try to detect based on filename.
func (h *Highlighter) Highlight(code, filename, language string) string {
	h.mu.RLock()
	style := h.style
	formatter := h.formatter
	h.mu.RUnlock()

	// Determine lexer
	var lexer chroma.Lexer
	if language != "" {
		lexer = lexers.Get(language)
	}
	if lexer == nil && filename != "" {
		lexer = lexers.Match(filename)
	}
	if lexer == nil {
		// Fallback to plaintext
		lexer = lexers.Fallback
	}

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	// Format with ANSI codes
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code
	}

	return buf.String()
}

// HighlightLines highlights code and returns it as separate lines.
// This is useful for line-by-line rendering in the editor.
func (h *Highlighter) HighlightLines(code, filename, language string) []string {
	highlighted := h.Highlight(code, filename, language)
	return strings.Split(highlighted, "\n")
}

// RegisterLexer registers a custom lexer at runtime.
// This allows adding support for new languages without recompiling.
// The lexer must implement the chroma.Lexer interface.
func RegisterLexer(lexer chroma.Lexer) error {
	lexers.GlobalLexerRegistry.Register(lexer)
	return nil
}

// NewSimpleLexer creates a simple lexer from a LexerConfig.
// This provides a convenient way to define new languages at runtime.
func NewSimpleLexer(config *LexerConfig) (chroma.Lexer, error) {
	rules := make(chroma.Rules)

	// Convert our simplified rules to Chroma rules
	for state, patterns := range config.Rules {
		var chromaRules []chroma.Rule
		for _, p := range patterns {
			chromaRules = append(chromaRules, chroma.Rule{
				Pattern: p.Pattern,
				Type:    p.TokenType,
				Mutator: nil,
			})
		}
		rules[state] = chromaRules
	}

	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      config.Name,
			Aliases:   config.Aliases,
			Filenames: config.Filenames,
			MimeTypes: config.MimeTypes,
		},
		func() chroma.Rules {
			return rules
		},
	), nil
}

// LexerConfig defines configuration for creating a simple lexer at runtime.
type LexerConfig struct {
	// Name is the display name of the language
	Name string
	// Aliases are alternative names (e.g., "py" for "Python")
	Aliases []string
	// Filenames are glob patterns for matching files (e.g., "*.py")
	Filenames []string
	// MimeTypes are MIME types for this language
	MimeTypes []string
	// Rules define the tokenization rules by state
	Rules map[string][]PatternRule
}

// PatternRule defines a single tokenization rule.
type PatternRule struct {
	// Pattern is a regular expression pattern
	Pattern string
	// TokenType is the Chroma token type for matches
	TokenType chroma.TokenType
}

// GetFileExtension extracts the file extension from a path.
func GetFileExtension(path string) string {
	return filepath.Ext(path)
}

// HighlightCode is a convenience function using the default highlighter.
func HighlightCode(code, filename, language string) string {
	return DefaultHighlighter.Highlight(code, filename, language)
}

// HighlightCodeLines is a convenience function using the default highlighter.
func HighlightCodeLines(code, filename, language string) []string {
	return DefaultHighlighter.HighlightLines(code, filename, language)
}
