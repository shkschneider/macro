# macro

A nano-like simple CLI text editor built with [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- Simple terminal-based text editor with Elm-inspired architecture
- Keyboard shortcuts similar to nano:
  - **Ctrl-S**: Save file
  - **Ctrl-Space**: Open file switcher dialog with instant fuzzy search
  - **Ctrl-Q**: Quit editor
- File switcher dialog with centered overlay
  - Custom fuzzy search with text input at bottom
  - Just start typing to filter files instantly (no need to press `/`)
  - Lists all files in the current file's directory
  - Use arrow keys to navigate, Enter to open
  - Esc or Ctrl-Space to close
- **Syntax highlighting** powered by [Chroma](https://github.com/alecthomas/chroma)
  - Automatic language detection based on file extension
  - 300+ supported languages out of the box
  - Multiple color themes available (default: Monokai)
  - Runtime registration of custom lexers for new languages
- Clean, modern UI with syntax highlighting for status messages

## Installation

```bash
go build -o macro
```

## Usage

```bash
# Create or edit a new file
./macro filename.txt

# The editor will open in your terminal
# Use Ctrl-S to save changes
# Use Ctrl-Q to quit
```

## Requirements

- Go 1.23 or later
- Terminal with support for ANSI/VT sequences

## Architecture

### Overview

The application follows the **Elm Architecture** pattern (Model-View-Update) via Bubbletea, providing a functional, reactive approach to building terminal UIs.

```
┌─────────────────────────────────────────────────┐
│                   main()                        │
│  - Parse CLI args (filename)                    │
│  - Initialize Bubbletea program                 │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│            initialModel()                       │
│  - Create textarea component (Bubbles)          │
│  - Load file content if exists                  │
│  - Set initial state                            │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│         Tea.Program Runtime                     │
│                                                 │
│  ┌───────────────────────────────────────────┐ │
│  │            Model (State)                   │ │
│  │  - textarea: Bubbles TextArea component   │ │
│  │  - filePath: Current file path            │ │
│  │  - status: Status bar message             │ │
│  │  - err: Error state flag                  │ │
│  └───────────────────────────────────────────┘ │
│                                                 │
│  ┌───────────────────────────────────────────┐ │
│  │         Update (Message Handler)          │ │
│  │  - KeyMsg: Handle keyboard input          │ │
│  │    • Ctrl-Q → Quit program                │ │
│  │    • Ctrl-S → Save file to disk           │ │
│  │    • Other keys → Pass to textarea        │ │
│  │  - WindowSizeMsg: Adjust textarea size    │ │
│  └───────────────────────────────────────────┘ │
│                                                 │
│  ┌───────────────────────────────────────────┐ │
│  │          View (Renderer)                  │ │
│  │  1. Title (styled with Lipgloss)          │ │
│  │  2. TextArea content                      │ │
│  │  3. Status bar (color-coded):             │ │
│  │     - Yellow: Default shortcuts           │ │
│  │     - Green: Success messages             │ │
│  │     - Red: Error messages                 │ │
│  └───────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

### Component Breakdown

**1. Model (`type model struct`)**
- Holds all application state
- Immutable updates via Update function
- Contains:
  - `textarea`: Bubbles TextArea for editing
  - `filePath`: Path to file being edited
  - `status`: Current status message
  - `err`: Error tracking

**2. Update (`func (m model) Update`)**
- Pure function: `(Model, Message) → (Model, Command)`
- Handles three message types:
  - **KeyMsg**: Keyboard events (Ctrl-S, Ctrl-Q, typing)
  - **WindowSizeMsg**: Terminal resize events
  - Delegates to textarea for text editing
- Returns new model state and optional commands

**3. View (`func (m model) View`)**
- Pure function: `Model → String`
- Renders UI from current state
- Uses Lipgloss for styling:
  - Title in bold blue
  - Status bar with conditional colors
- Layout: Title + TextArea + Status

**4. Styling (Lipgloss)**
- `titleStyle`: Bold blue for title bar
- `statusStyle`: Yellow for default status
- `successStyle`: Green for save confirmations
- `errorStyle`: Red for errors

### Data Flow

```
User Input (Keyboard)
    ↓
Message (tea.KeyMsg)
    ↓
Update Function
    ↓
New Model State
    ↓
View Function
    ↓
Rendered UI (Terminal)
```

This architecture ensures:
- **Predictable state management**: All state changes go through Update
- **Testability**: Pure functions are easy to test
- **Maintainability**: Clear separation of concerns
- **Extensibility**: Easy to add new features via new message types

## Syntax Highlighting

Macro uses [Chroma](https://github.com/alecthomas/chroma), a general purpose syntax highlighter in pure Go, to provide syntax highlighting capabilities.

### Supported Languages

Over 300 programming and markup languages are supported out of the box, including:
- Go, Python, JavaScript, TypeScript, Java, C, C++, Rust
- HTML, CSS, JSON, YAML, XML, Markdown
- SQL, Shell/Bash, PowerShell
- And many more...

Language detection is automatic based on file extension.

### Color Themes

Multiple color themes are available. The default theme is Monokai. Other popular themes include:
- Dracula
- GitHub
- Solarized (Dark/Light)
- Nord
- Gruvbox

### Adding Custom Languages at Runtime

You can register custom lexers for new languages at runtime using the `core.RegisterLexer` function. Here's an example:

```go
import (
    "github.com/alecthomas/chroma/v2"
    "github.com/shkschneider/macro/core"
)

// Define a custom lexer configuration
config := &core.LexerConfig{
    Name:      "MyLanguage",
    Aliases:   []string{"mylang", "ml"},
    Filenames: []string{"*.ml", "*.mylang"},
    MimeTypes: []string{"text/x-mylang"},
    Rules: map[string][]core.PatternRule{
        "root": {
            {Pattern: `\b(func|var|return)\b`, TokenType: chroma.Keyword},
            {Pattern: `"[^"]*"`, TokenType: chroma.String},
            {Pattern: `//.*`, TokenType: chroma.Comment},
            {Pattern: `\d+`, TokenType: chroma.Number},
        },
    },
}

// Create and register the lexer
lexer, err := core.NewSimpleLexer(config)
if err != nil {
    log.Fatal(err)
}
core.RegisterLexer(lexer)
```

After registration, files matching the specified patterns will be highlighted using your custom rules.

### API Reference

The `core` package provides the following syntax highlighting functions:

- `HighlightCode(code, filename, language string) string` - Highlight code with ANSI colors
- `HighlightCodeLines(code, filename, language string) []string` - Highlight and return as lines
- `DetectLanguage(filename string) string` - Detect language from filename
- `GetLanguageByExtension(ext string) string` - Get language for extension
- `ListLanguages() []string` - List all available languages
- `ListStyles() []string` - List all available color themes
- `RegisterLexer(lexer chroma.Lexer) error` - Register a custom lexer
- `NewSimpleLexer(config *LexerConfig) (chroma.Lexer, error)` - Create a simple lexer from config
