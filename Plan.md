# onering: TUI for Managing Code Agents

## Context

Build a lazygit-style TUI in Go that manages multiple code agents (Claude Code, Codex, Aider) with embedded side applications (nvim, lazygit). The user wants a sidebar + main panel layout, vim-style keybindings, and a polished UX with contextual help. The repo is empty — this is a greenfield project.

## Architecture

### Dual-Mode Agent Interaction

- **SDK Mode (Phase 1):** Agents run non-interactively with structured output. onering owns the UI completely, parsing events and rendering them natively. No VT emulation needed.
  - Claude Code: `claude --print --output-format stream-json --verbose --session-id <uuid>`
  - Codex: `codex exec --model <model> "<prompt>"`
  - Aider: `aider --message "<prompt>" --no-pretty --yes`
- **PTY Mode (Phase 3):** Full interactive embedding via `creack/pty` + VT emulator for nvim/lazygit/interactive agents.

### Bubbletea Model Hierarchy

```
AppModel (top-level: focus routing, input mode)
  ├── LayoutModel (sidebar + main panel geometry)
  │     ├── SidebarModel (session list with status indicators)
  │     └── MainPanelModel (active session content)
  │           ├── AgentViewModel (SDK mode: rendered agent output)
  │           └── PromptModel (text input for sending messages)
  ├── StatusBarModel (keybind hints, agent status, cost)
  └── HelpModel ('?' overlay)
```

### Input Modes

- **Navigation:** hjkl between panels, standard keybindings
- **Insert:** Keystrokes go to prompt textarea, Esc returns to navigation
- **Passthrough (Phase 3):** All keys forwarded to embedded PTY

### Default Keybindings

| Key | Action |
|-----|--------|
| `q` | Quit |
| `?` | Toggle help overlay |
| `h/l` | Focus sidebar / main panel |
| `j/k` | Navigate items / scroll |
| `i` | Enter insert mode (focus prompt) |
| `Esc` | Return to navigation mode |
| `Enter` | Send prompt (in insert mode) |
| `n` | New session |
| `d` | Delete session |
| `a` | Switch agent type |
| `e` | Open editor (nvim) |
| `g` | Open git (lazygit) |
| `f` | Zoom/fullscreen active panel |
| `Tab` | Cycle focus |
| `Ctrl+u/d` | Page up/down |

## Package Structure

```
onering/
  main.go
  go.mod
  cmd/
    root.go                     # Cobra root command, launches TUI
  internal/
    app/
      app.go                    # AppModel: top-level bubbletea model
      messages.go               # Custom tea.Msg types
      keybindings.go            # Binding definitions + help text
    config/
      config.go                 # YAML config loading + types
      defaults.go               # Default values
    agent/
      agent.go                  # Agent + Session interfaces
      types.go                  # AgentEvent, SessionState, ToolUse, EventMeta
      registry.go               # Agent registry (auto-detect available agents)
      claude.go                 # Claude Code implementation (stream-json)
      codex.go                  # Codex implementation
      aider.go                  # Aider implementation
    ui/
      layout.go                 # LayoutModel: sidebar + main geometry
      sidebar.go                # SidebarModel: session list
      mainpanel.go              # MainPanelModel: routes to active view
      agentview.go              # AgentViewModel: renders streaming output
      promptinput.go            # PromptModel: textarea input
      statusbar.go              # StatusBarModel: bottom bar
      help.go                   # HelpModel: '?' keybinding overlay
      styles.go                 # Lipgloss theme/style definitions
    terminal/                   # Phase 3: PTY embedding
      pty.go
      vt.go
      termview.go
    util/
      id.go                    # UUID generation
      log.go                   # File-based debug logging
```

## Key Interfaces

**Agent** — pluggable backend for each code agent CLI:
- `Name() string`
- `Available() (bool, error)` — checks if CLI is installed
- `StartSession(ctx, opts) (Session, error)`
- `ResumeSession(ctx, sessionID) (Session, error)`

**Session** — an active agent conversation:
- `ID() string`
- `Send(ctx, prompt) error`
- `Events() <-chan AgentEvent` — streaming events
- `Cancel()` / `Close() error`
- `State() SessionState` (Idle, Running, Error, Closed)

**AgentEvent** — streaming output from agent:
- Types: Init, Text, TextDone, ToolStart, ToolResult, Complete, Error
- Carries content, tool use details, and metadata (cost, tokens, duration)

## Configuration

YAML at `~/.config/onering/config.yaml`:
- `agents:` — per-agent enable/disable, command path, default model, extra args
- `side_apps:` — editor/git commands and keybindings
- `ui:` — sidebar width, show cost/tokens
- `keybindings:` — overrides for all keybindings

## Dependencies

- `charmbracelet/bubbletea` — TUI framework
- `charmbracelet/lipgloss` — styling
- `charmbracelet/bubbles` — viewport, list, textarea, spinner components
- `spf13/cobra` — CLI commands
- `gopkg.in/yaml.v3` — config parsing
- `google/uuid` — session IDs
- `creack/pty` (Phase 3) — PTY management
- `charmbracelet/x/vt` (Phase 3) — VT emulation

## Implementation Phases

### Phase 1: Working TUI with Claude Code SDK Mode

Build in this order:

1. **Scaffolding** — `go mod init`, main.go, cobra root command, directory structure
2. **Config** — YAML loading with defaults, XDG path resolution
3. **Agent interface + Claude implementation** — interfaces in `agent.go`, Claude stream-json parser in `claude.go`, test with real CLI
4. **Basic layout** — AppModel with sidebar + main panel split, lipgloss borders, focus indicators
5. **Agent output rendering** — AgentViewModel with viewport, styled text/tool-use blocks, streaming append
6. **Prompt input** — textarea at bottom of main panel, insert/navigation mode toggle
7. **Session management** — create (n), switch (j/k + Enter), delete (d), state indicators in sidebar
8. **Help overlay** — '?' shows context-sensitive keybinding reference
9. **Status bar** — mode indicator, agent name, cost/tokens, keybinding hints
10. **Polish** — resize handling, error display, spinner for running state, debug file logging

### Phase 2: Multi-Agent + Side App Zoom

- Codex and Aider agent implementations
- Agent selector (`a` key)
- Agent registry with auto-detection
- Side app zoom via `tea.Exec()` (e for nvim, g for lazygit)
- Session persistence to `~/.local/state/onering/sessions/`

### Phase 3: PTY Embedding

- `creack/pty` integration for spawning interactive processes
- VT emulator (`charmbracelet/x/vt`) to parse PTY output into cell grids
- Terminal view renderer converting cells to lipgloss strings
- Passthrough input mode for embedded interactive apps
- Split main panel: agent + side app simultaneously

### Phase 4: Advanced Features

- Configurable multi-panel layouts
- Diff viewer for agent file edits
- Session search/history
- Themes
- Agent orchestration (parallel agents, compare outputs)

## Verification

1. `go build ./...` compiles without errors
2. `go test ./...` passes (unit tests for JSON parsing, config loading)
3. Run `./onering` — sidebar appears on left, main panel on right with prompt
4. Press `n` — new Claude Code session created, appears in sidebar
5. Press `i`, type a prompt, press Enter — streaming response appears in main panel
6. Press `Esc`, `j/k` — navigate between sessions in sidebar
7. Press `?` — help overlay shows keybindings
8. Resize terminal — layout adjusts
