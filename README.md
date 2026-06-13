# lazycode

A lazygit-style TUI for managing multiple code AI agents in one unified dashboard. Embed Claude Code, OpenCode, Codex, and Aider side by side, with nvim and lazygit available as side applications.

## Features

- **Multi-Agent Dashboard** — Start, switch between, and manage agent sessions from a single terminal interface
- **PTY Embedding** — Agents run as embedded terminals, preserving their native TUI experience
- **Side Apps** — Launch nvim, lazygit (or any terminal app) alongside agent sessions
- **Vim-style Navigation** — hjkl movement, modal interaction (Normal / Insert / Passthrough)
- **Auto-Detection** — Automatically finds installed agents on your `$PATH`
- **Configurable** — YAML config at `~/.config/lazycode/config.yaml`

## Agents

| Agent | Status |
|---|---|
| Claude Code | ✅ PTY mode |
| OpenCode | ✅ PTY mode |
| Codex | ❌ Stub |
| Aider | ❌ Stub |

## Installation

```bash
go build -o lazycode .
./lazycode
```

Requires Go 1.26+.

## Configuration

`~/.config/lazycode/config.yaml`:

```yaml
agents:
  claude:
    enabled: true
    command: claude
  opencode:
    enabled: true
    command: opencode
  codex:
    enabled: false
    command: codex
  aider:
    enabled: false
    command: aider
side_apps:
  editor: nvim .
  git: lazygit
ui:
  sidebar_width: 30
  show_cost: true
  show_tokens: true
```

## Keybindings

| Key | Action |
|---|---|
| `h`/`l` or `Tab` | Focus sidebar / main panel |
| `j`/`k` | Navigate sessions list |
| `i` or `Enter` | Start / focus session |
| `n` | New session |
| `d` | Delete session |
| `e` | Launch editor |
| `g` | Launch lazygit |
| `Ctrl+Q` | Exit passthrough mode |
| `?` | Toggle help |
| `q` | Quit |

## Architecture

```
AppModel
├── LayoutModel
│   ├── SidebarModel      — session/app list
│   └── MainPanelModel    — terminal view or placeholder
├── StatusBarModel        — mode indicator + hints
├── HelpModel             — keybinding overlay
└── LabelModal            — new session dialog
```

Three input modes: **Navigation** (default, vim keys), **Insert** (prompt input), **Passthrough** (keys forwarded to PTY).

## Contributing

Contributions are welcome! Here's how to get started:

1. Fork the repo and create a branch from `main`.
2. Make your changes, keeping code style consistent with the existing codebase.
3. Run `go build ./...` to verify compilation.
4. Open a pull request describing what you changed and why.

For feature requests or bug reports, open an issue.

## License

MIT — see [LICENSE](LICENSE).
