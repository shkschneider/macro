# macro

A nano-like simple CLI text editor built with [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- Simple terminal-based text editor with Elm-inspired architecture
- Keyboard shortcuts similar to nano:
  - **Ctrl-S**: Save file
  - **Ctrl-Q**: Quit editor
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

Built using the Elm architecture pattern via Bubbletea:
- **Model**: Application state (textarea, filepath, status)
- **Update**: Handle messages and update state
- **View**: Render the UI based on current state