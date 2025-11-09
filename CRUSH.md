# Macro Codebase Guidelines for AI Agents

This document outlines essential information for AI agents working in the `macro` codebase, a TUI (Terminal User Interface) text editor built with Go and Charmbracelet libraries.

## 1. Project Overview

`macro` is a TUI text editor. Its core functionality involves loading and editing text files, with various features (file switching, buffer switching, help, save, quit) implemented as modular components.

## 2. Essential Commands

*   **Build**: Compile the application executable.
    ```bash
    go build -o macro
    ```
*   **Run**: Run the application.
    ```bash
    ./macro [optional_filepath]
    ```
*   **Test**: Execute all unit tests.
    ```bash
    go test ./...
    ```
*   **Lint (Basic)**: Run Go's built-in vet tool for basic static analysis.
    ```bash
    go vet ./...
    ```
*   **Integration Tests**: The `integration_test.sh` script sets up various test files and demonstrates how to manually run the editor with them. It does not run automated checks itself.
    ```bash
    ./integration_test.sh
    ```

## 3. Code Organization and Structure

The codebase follows a modular structure:

*   **`main.go`**: The application's entry point. It initializes the Bubble Tea program and registers all available features as commands.
*   **`app.go`**: Contains the core `tea.Model` implementation for the TUI application. It manages the global state (buffers, active dialogs, messages), handles `tea.Msg` updates, and orchestrates the rendering of different UI components.
*   **`buffer.go`**: (Inferred) Likely defines the `Buffer` struct and related methods for managing individual open files and their content.
*   **`command.go`**: (Inferred) Likely defines the `Command` struct used for registering features and their execution logic.
*   **`feature/`**: This package contains implementations of distinct editor functionalities. Each feature typically defines:
    *   A `CommandDef` for registration.
    *   Its own `tea.Model` or a component that implements the `core.Dialog` interface for interactive features (e.g., `file_switcher.go`, `buffer_switcher.go`, `help.go`, `save.go`, `quit.go`).
    *   Custom `tea.Msg` types for communication back to the main `app.go` model.
*   **`core/`**: This package provides shared UI components, styling, and utility functions used across the application.
    *   `dialog.go`: Defines the `Dialog` interface for interactive overlays.
    *   `overlay.go`: Contains logic for rendering dialogs centered on top of the main view, handling `lipgloss` styling and ANSI escape sequences.
    *   `styles.go`: Defines `lipgloss.Style` objects for consistent visual themes.

## 4. Naming Conventions and Style Patterns

*   **Go Standard**: Adheres to standard Go naming conventions (CamelCase for public identifiers, lowercase for private).
*   **Package Names**: Singular and lowercase (e.g., `feature`, `core`).
*   **Command Definitions**: Features register themselves using `macro.CommandDef` structs.
*   **Message Types**: Custom messages for inter-component communication typically end with `Msg` (e.g., `FileSelectedMsg`).
*   **Imports**: Generally grouped: standard library, then external dependencies (e.g., `charmbracelet` libraries), then internal project packages (`github.com/shkschneider/macro/...`).

## 5. Testing Approach

*   **Unit Tests**: Standard Go unit tests (`go test`) are used for individual functions and components. (e.g., `core/overlay_test.go`).
*   **Integration Tests**: The `integration_test.sh` script is provided for setting up various test files to manually verify editor behavior. Automated integration tests are not explicitly implemented using `go test`.

## 6. Important Gotchas and Non-Obvious Patterns

*   **Bubble Tea (TUI Framework)**: The application is built using `github.com/charmbracelet/bubbletea`. Understanding its `tea.Model`, `Init()`, `Update(msg tea.Msg)`, and `View()` methods, as well as `tea.Cmd` for side effects, is fundamental.
*   **Charmbracelet Ecosystem**: Extensive use of `github.com/charmbracelet/bubbles/*` components (textarea, filepicker, viewport) and `github.com/charmbracelet/lipgloss` for styling.
*   **Dialogs and Overlays**: Interactive features are often implemented as dialogs conforming to the `core.Dialog` interface. These dialogs are overlaid on the main view using `core.OverlayDialog`, which requires careful handling of terminal dimensions and `lipgloss` styling to ensure correct rendering.
*   **Command Palette/Keybindings**: Features are activated via `CommandDef` registrations. Keybindings (e.g., `Ctrl-Space` for command palette, `Ctrl-P` for file switcher) are handled in `app.go`'s `Update` method by looking up and executing registered commands.
*   **Read-Only Files**: The editor explicitly checks and displays read-only status for files, preventing modifications in such cases.
*   **Fuzzy Matching**: The `feature/file_switcher.go` and `feature/buffer_switcher.go` utilize `github.com/sahilm/fuzzy` for filtering lists of items.
