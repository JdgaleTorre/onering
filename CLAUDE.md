# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is onering

A lazygit-style TUI for managing multiple code agents (Claude Code, OpenCode, Codex, Aider) with embedded side applications (nvim, lazygit). Built in Go with Bubbletea.

## Build and Run

```bash
go build -o onering .    # build the binary
./onering                # run the TUI
go build ./...            # check compilation
go test ./...             # run all tests (no tests exist yet)
```

There is no linter or formatter configured. The binary `onering` is gitignored.

## Architecture

### Bubbletea Model Hierarchy

`AppModel` (internal/app/app.go) is the top-level Bubbletea model. It owns the input mode state machine, focus routing, session/app lifecycle, and sidebar cursor. It delegates rendering to:

- `LayoutModel` (internal/ui/layout.go) — splits sidebar + main panel, routes messages by focus
- `SidebarModel` — session list with project name/branch header, two sections: Sessions and Apps
- `MainPanelModel` — shows the active session's terminal view or a placeholder
- `StatusBarModel` — mode indicator and keybinding hints
- `HelpModel` — `?` overlay listing keybindings

### Input Modes

Three modes, tracked by `AppModel.mode`:
- **Navigation** — vim-style hjkl, session management keys
- **Insert** — keystrokes go to the prompt textarea (SDK-mode agents, not yet implemented)
- **Passthrough** — all keys forwarded to embedded PTY (Ctrl+Q to exit)

### Agent System

All agents implement `Agent` and `Session` interfaces (internal/agent/agent.go). Currently all working agents use PTY mode via `PTYSession` (internal/agent/opencode_session.go), which embeds the agent's TUI directly.

- `ClaudeAgent` and `OpenCodeAgent` — fully implemented via PTY embedding
- `CodexAgent` and `AiderAgent` — stubs (StartSession returns nil)
- `Registry` (internal/agent/registry.go) — auto-detects available agents via `exec.LookPath`

The original plan called for SDK mode (structured JSON streaming) as Phase 1 and PTY mode as Phase 3, but the implementation jumped to PTY mode. The SDK-mode event types (`AgentEvent`, `EventType`) in internal/agent/types.go are defined but unused.

### PTY / Terminal Embedding

`terminal.PTYHandle` wraps `creack/pty` for starting/resizing/closing child processes. `terminal.TermViewModel` uses `charmbracelet/x/vt` to emulate a terminal — it reads PTY output, feeds it to the VT emulator, and renders the cell grid via `emu.Render()`. Key presses are translated to VT key events via `keyMsgToUV` (internal/terminal/keys.go).

### Side Apps

Editor (nvim) and lazygit are embedded as PTY sessions, launched by `startSideApp` (internal/app/sideapps.go). They appear in the Apps section of the sidebar and run in passthrough mode.

### Configuration

YAML at `~/.config/onering/config.yaml` (XDG-aware). See `config.Config` for structure. Defaults in internal/config/defaults.go: claude and opencode enabled, codex and aider disabled, sidebar width 30.

### Debug Logging

`util.Debug()` writes to `$TMPDIR/onering-debug.log`. Logger must be initialized with `util.InitLogger()` first (currently not called from main).
