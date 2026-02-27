# Agent T

A Go 1.22+ terminal workspace launcher for macOS using Bubble Tea TUI.

## Cursor Cloud specific instructions

- **Language/Runtime**: Go 1.22+. Dependencies managed via `go.mod`/`go.sum`.
- **Build**: `go build -o agent-t .`
- **Test**: `go test ./...` (18 tests across `internal/config`, `internal/launcher`, `internal/tui`)
- **Lint**: `go vet ./...` (no `.golangci.yml` configured; `go vet` is the available lint check)
- **Platform caveat**: The launcher (`internal/launcher`) generates AppleScript and invokes `osascript`, which only exists on macOS. On Linux, the TUI wizard runs fully but the final launch step fails with `osascript not found`. Tests pass on any platform since they validate script generation, not execution.
- **Running the TUI**: The binary scans the current working directory for subdirectories to present as projects. Run from a directory containing at least one subdirectory: `cd /some/parent && ./agent-t`
- **No external services** are required. This is a single standalone CLI binary with no database, API, or Docker dependencies.
