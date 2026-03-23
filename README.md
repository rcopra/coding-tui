# coding-tui

A terminal client for [Exercism](https://exercism.org) built with [BubbleTea v2](https://github.com/charmbracelet/bubbletea). Browse tracks, read exercises, test and submit solutions — all without leaving the terminal.

```
┌─────────────────────────┬──────────────────────────────┐
│  coding-tui              │  nvim                        │
│                          │                              │
│  [Track: Go]             │  // two-fer.go               │
│                          │  package twofer              │
│  ## Two Fer              │                              │
│                          │  func ShareWith(name string) │
│  Create a sentence of    │    string {                  │
│  the form "One for X,    │    if name == "" {           │
│  one for me."            │      name = "you"            │
│                          │    }                         │
│  ──────────────────────  │    return fmt.Sprintf(       │
│  t test  s submit        │      "One for %s, one for   │
│  d download  h hints     │      me.", name)             │
│  c community  o browser  │  }                           │
└─────────────────────────┴──────────────────────────────┘
```

Designed to run in a tmux pane alongside your editor. The filesystem is the bridge — exercises download to your exercism workspace and you edit them however you want.

## Install

```sh
go install github.com/rcopra/coding-tui@latest
```

Or build from source:

```sh
git clone https://github.com/rcopra/coding-tui.git
cd coding-tui
go build -o coding-tui .
```

## Prerequisites

- [Exercism CLI](https://exercism.org/docs/using/solving-exercises/working-locally) configured with your API token:
  ```sh
  exercism configure --token=YOUR_TOKEN
  ```
- Go 1.21+ (for building)
- Language-specific test runners for tracks you use (e.g. `ruby`, `python3`, `node`)

## Usage

```sh
coding-tui
```

That's it. You'll see a list of all Exercism tracks. Navigate with vim keybindings.

## Keybindings

| Key | Action |
|-----|--------|
| `j/k` | Navigate / scroll |
| `/` | Filter list |
| `gg` / `G` | Jump to top / bottom |
| `enter` | Select |
| `q` / `esc` | Back (quit at root) |
| `ctrl+c` | Force quit |
| `d` | Download exercise |
| `t` | Run tests locally |
| `s` | Submit to Exercism |
| `h` | Toggle hints |
| `c` | Community solutions |
| `o` | Open in browser |

## Workflow

1. Launch `coding-tui` in a tmux pane
2. Browse to a track and pick an exercise
3. Read the instructions (rendered markdown with syntax highlighting)
4. Press `d` to download — files land in your exercism workspace
5. Switch to your editor pane, open the file, write your solution
6. Switch back, press `t` to run tests
7. Press `s` to submit when tests pass

## Supported Tracks (local testing)

Local test execution is supported for 20+ tracks including Go, Rust, Python, JavaScript, TypeScript, Ruby, Elixir, Java, C, C++, C#, Swift, Kotlin, Haskell, Clojure, Scala, Lua, Bash, Zig, and PHP. Other tracks can still be browsed, downloaded, and submitted — just run tests manually.

## Architecture

```
internal/
├── api/          Exercism API client (v1 + v2 endpoints)
├── cache/        File-based response cache (10min TTL)
├── config/       Reads ~/.config/exercism/user.json
├── ui/           BubbleTea screens (tracks, exercises, detail, tests, community)
└── workspace/    File operations and local test runner
```

Built with the [Charm](https://charm.sh) stack:
- [BubbleTea v2](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss v2](https://github.com/charmbracelet/lipgloss) — Styling
- [Bubbles v2](https://github.com/charmbracelet/bubbles) — Components (list, viewport, help)
- [Glamour v2](https://github.com/charmbracelet/glamour) — Markdown rendering

## License

MIT
