# macro

A nano-like simple CLI text editor built with [tview](https://github.com/rivo/tview) and [tcell](https://github.com/gdamore/tcell).

## Features

- Simple terminal-based text editor
- Keyboard shortcuts similar to nano:
  - **Ctrl-S**: Save file
  - **Ctrl-Q**: Quit editor

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

- Go 1.16 or later
- Terminal with support for ANSI/VT sequences