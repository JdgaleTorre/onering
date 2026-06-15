# onering

A lazygit-style TUI for managing multiple code AI agents in one unified dashboard. Embed Claude Code, OpenCode, Codex, and Aider side by side, with nvim and lazygit available as side applications.

## Features

- **Multi-Agent Dashboard** — Start, switch between, and manage agent sessions from a single terminal interface
- **PTY Embedding** — Agents run as embedded terminals, preserving their native TUI experience
- **Side Apps** — Launch nvim, lazygit (or any terminal app) alongside agent sessions
- **Vim-style Navigation** — hjkl movement, modal interaction (Normal / Insert / Passthrough)
- **Auto-Detection** — Automatically finds installed agents on your `$PATH`
- **Configurable** — YAML config at `~/.config/onering/config.yaml`

## Agents

| Agent | Status |
|---|---|
| Claude Code | ✅ PTY mode |
| OpenCode | ✅ PTY mode |
| Codex | ❌ Stub |
| Aider | ❌ Stub |

## Installation

```bash
go build -o onering .
./onering
```

Requires Go 1.26+.

## Configuration

`~/.config/onering/config.yaml`:

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
  docker: lazydocker
  extra:
    - name: monitor
      command: btop
  enable:
    docker: false
ui:
  sidebar_width: 30
  show_cost: true
  show_tokens: true
```

**Side apps** — three built-in: `editor` (nvim), `git` (lazygit), `docker` (lazydocker).  
Apps with `enable: false` are hidden.  
Uninstalled apps show `!` in the sidebar; pressing Enter shows install instructions.

### Extra Apps

You can add any terminal application as a side app under `side_apps.extra`. Each entry needs a `name` and a `command`:

```yaml
side_apps:
  extra:
    - name: monitor
      command: btop
    - name: logs
      command: tail -f /var/log/syslog
    - name: k9s
      command: k9s
```

Extra apps appear in the **Apps** section of the sidebar alongside the built-in apps. You can launch, kill, and interact with them the same way — select with `Enter` to start, `d` to kill, and `Ctrl+Q` to exit passthrough mode.

To disable an extra app without removing it from the config, add it to the `enable` map:

```yaml
side_apps:
  enable:
    monitor: false
```

## Tasks

onering automatically detects runnable tasks from your project and lists them in the sidebar under the **Tasks** section (jump there with `3`).

### Supported Sources

| Source | Detected via | Tasks |
|---|---|---|
| npm / pnpm / yarn / bun | `package.json` + lock file | `install` + all `scripts` entries |
| Make | `Makefile` | All targets |
| Go | `go.mod` | `build`, `test`, `vet`, `fmt` |

The package manager is auto-detected from lock files (`pnpm-lock.yaml`, `yarn.lock`, `bun.lock`). To override, set `tasks.package_manager` in your config:

```yaml
tasks:
  package_manager: pnpm  # force pnpm instead of auto-detecting
```

### Running Tasks

- **`Enter`** — runs the task and captures output in the main panel (piped mode)
- **`p`** — runs the task in a full interactive terminal (PTY mode), useful for dev servers or watch commands

Kill a running task with `d`. Refresh the task list with `r` to pick up changes to your project files.

### Favorites

Press `f` to mark a task as a favorite — it gets pinned to the top of the list with a ★ icon. Favorites are saved per project and persist across sessions.

### Status Indicators

| Icon | Meaning |
|---|---|
| ● | Running |
| ✓ | Completed (exit code 0) |
| ✗ | Failed (non-zero exit code) |

## Keybindings

Context-sensitive hints are shown in the status bar as you navigate between sections.

### Navigation

| Key | Action |
|---|---|
| `h`/`l` or `Tab` | Focus sidebar / main panel |
| `j`/`k` | Navigate sidebar items |
| `0`/`1`/`2`/`3` | Jump to project info / sessions / apps / tasks |
| `?` | Toggle help |
| `q` | Quit |

### Sessions

| Key | Action |
|---|---|
| `n` | New session |
| `d` | Delete session |
| `i` or `Enter` | Enter / activate session |

### Apps

| Key | Action |
|---|---|
| `Enter` | Launch app |
| `d` | Kill app |
| `Ctrl+E` | Launch editor (nvim) |
| `Ctrl+G` | Launch lazygit |
| `Ctrl+D` | Launch lazydocker |

### Tasks

| Key | Action |
|---|---|
| `Enter` | Run task |
| `p` | Run task in PTY |
| `r` | Refresh task list |
| `f` | Toggle favorite |

### Modes

| Key | Action |
|---|---|
| `Ctrl+Q` | Exit passthrough mode |
| `Esc` | Back to navigation mode |

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
