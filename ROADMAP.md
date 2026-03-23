# Exercism TUI — Roadmap

> A BubbleTea v2 terminal client for Exercism. Browse tracks, read exercises, test and submit solutions — all without leaving the terminal.

## Architecture Overview

```
coding-tui/
├── main.go                  # Entry point, tea.NewProgram
├── internal/
│   ├── api/                 # Exercism API client
│   │   ├── client.go        # HTTP client, auth, request helpers
│   │   ├── tracks.go        # Track endpoints
│   │   ├── exercises.go     # Exercise endpoints
│   │   ├── solutions.go     # Solution/submission endpoints
│   │   └── types.go         # Response structs
│   ├── config/              # Config loading
│   │   └── config.go        # Read ~/.config/exercism/user.json
│   ├── cache/               # Local response cache
│   │   └── cache.go         # File-based JSON cache with TTL
│   ├── ui/                  # BubbleTea models (screens)
│   │   ├── root.go          # Root model, screen router
│   │   ├── tracks.go        # Track browser (list)
│   │   ├── exercises.go     # Exercise list per track
│   │   ├── detail.go        # Exercise detail (markdown instructions)
│   │   ├── testrun.go       # Test results display
│   │   ├── community.go     # Community solutions browser
│   │   └── common.go        # Shared styles, keybindings, helpers
│   └── workspace/           # File system operations
│       └── workspace.go     # Download, detect existing, path helpers
└── go.mod
```

**Two API surfaces:**
- `https://exercism.org/api/v2/` — website API, no auth needed for browsing tracks/exercises
- `https://api.exercism.org/v1/` — CLI API, token required, used for solutions/submissions

**Config:** Reuses `~/.config/exercism/user.json` (token + workspace path). No separate config needed.

**Navigation model:** Stack-based. Push screens on enter, pop on `q`/`esc`. Root model holds the stack and routes messages to the top screen.

---

## Phase 1 — Foundation + Track Browser

**Goal:** Launch the app, authenticate, browse all Exercism tracks in a filterable list.

### Tasks
- [ ] `go mod init` with BubbleTea v2 dependencies
- [ ] Config loader: read `~/.config/exercism/user.json`, extract token + workspace path
- [ ] API client: base HTTP client with Bearer auth, JSON decoding, error handling
- [ ] API: `GET /api/v2/tracks` — fetch and parse track list
- [ ] Root model: screen stack, window size tracking, global keybindings (`q` quit, `?` help)
- [ ] Track browser screen: bubbles/list with track name, language, exercise count, progress
- [ ] Vim keybindings: `j/k` scroll, `/` filter, `gg/G` top/bottom, `enter` select
- [ ] Loading spinner while fetching tracks
- [ ] Help bar at bottom showing available keys

### Done when
You can launch the app, see all Exercism tracks, filter them, and select one (which does nothing yet).

---

## Phase 2 — Exercise List

**Goal:** After selecting a track, see its exercises with status indicators.

### Tasks
- [ ] API: `GET /api/v2/tracks/:slug/exercises` — fetch exercises for a track
- [ ] Exercise list screen: bubbles/list with exercise name, difficulty, type (concept/practice), status
- [ ] Status indicators: not started, in progress, completed (using unicode symbols)
- [ ] Difficulty display: easy/medium/hard with color coding
- [ ] `q`/`esc` to go back to track list
- [ ] Exercise blurb shown in list item description

### Done when
You can browse into a track and see all its exercises with their status. Back navigation works.

---

## Phase 3 — Exercise Detail View

**Goal:** View exercise instructions rendered as markdown in the terminal. Download exercise files.

### Tasks
- [ ] API: fetch exercise instructions (from solution initial files or GitHub .docs/instructions.md)
- [ ] Detail screen: glamour-rendered markdown in a viewport with scrolling
- [ ] Workspace manager: download exercise files to `~/exercism/<track>/<exercise>/`
- [ ] `d` to download/start exercise — files land in workspace, compatible with exercism CLI layout
- [ ] Show file path after download so user knows where to find it in nvim
- [ ] `h` to toggle hints (if available)
- [ ] Detect if exercise is already downloaded, show indicator

### Done when
You can read exercise instructions fully rendered in the terminal, download the exercise, and switch to nvim to find the files.

---

## Phase 4 — Test & Submit

**Goal:** Run tests and submit solutions from the TUI.

### Tasks
- [ ] `t` to run tests: execute the track's test command locally (e.g., `go test`, `python -m pytest`)
- [ ] Test run screen: show stdout/stderr with syntax highlighting, pass/fail summary
- [ ] API: `POST /api/v2/solutions/:uuid/submissions` — submit solution files
- [ ] API: `POST /api/v2/solutions/:uuid/iterations` — create iteration
- [ ] `s` to submit: upload solution files from workspace
- [ ] Show submission status (tests running on server → pass/fail)
- [ ] API: `GET /api/iterations/:uuid/automated_feedback` — show analyzer feedback
- [ ] `c` to mark exercise complete

### Done when
Full exercise lifecycle works: read instructions → download → edit in nvim → test in TUI → submit → see results → complete.

---

## Phase 5 — Community Solutions + Polish

**Goal:** Browse community solutions and polish the experience.

### Tasks
- [ ] API: `GET /api/v2/tracks/:slug/exercises/:slug/community_solutions` — browse others' solutions
- [ ] Community solutions screen: list of solutions, select to view code with syntax highlighting
- [ ] File-based cache: cache track/exercise lists with TTL to reduce API calls
- [ ] Progress indicators: show track completion percentage in track list
- [ ] Error handling polish: network errors, auth failures, missing config — all shown gracefully
- [ ] `o` to open exercise in browser (fallback escape hatch)
- [ ] Responsive layout: adapt to terminal width/height
- [ ] Color theme: respect terminal colors, dark-mode friendly

### Done when
The app is a complete replacement for the Exercism browser experience for the core learning workflow.

---

## Future (Not in v1)

- Mentoring (request + view discussions)
- Notifications
- Multiple exercise tabs / session management
- Neovim plugin wrapping the Go core
- Offline mode with full exercise cache
- Custom keybinding configuration

---

## Tech Stack

| Component | Package | Version |
|-----------|---------|---------|
| TUI framework | `charm.land/bubbletea/v2` | v2.0.2 |
| Styling | `charm.land/lipgloss/v2` | v2.0.2 |
| Components | `charm.land/bubbles/v2` | v2.0.0 |
| Markdown | `charm.land/glamour/v2` | v2.0.0 |
| Forms | `charm.land/huh/v2` | v2.0.3 |

## API Reference

| Action | Method | URL | Auth |
|--------|--------|-----|------|
| List tracks | GET | `exercism.org/api/v2/tracks` | No |
| Track detail | GET | `exercism.org/api/v2/tracks/:slug` | No |
| List exercises | GET | `exercism.org/api/v2/tracks/:slug/exercises` | No |
| Start exercise | PATCH | `exercism.org/api/v2/tracks/:slug/exercises/:slug/start` | Yes |
| Get solution | GET | `exercism.org/api/v2/solutions/:uuid` | Yes |
| Submit files | POST | `exercism.org/api/v2/solutions/:uuid/submissions` | Yes |
| Create iteration | POST | `exercism.org/api/v2/solutions/:uuid/iterations` | Yes |
| Test run results | GET | `exercism.org/api/v2/solutions/:uuid/submissions/:uuid/test_run` | Yes |
| Automated feedback | GET | `exercism.org/api/v2/iterations/:uuid/automated_feedback` | Yes |
| Complete exercise | PATCH | `exercism.org/api/v2/solutions/:uuid/complete` | Yes |
| Community solutions | GET | `exercism.org/api/v2/tracks/:slug/exercises/:slug/community_solutions` | No |
| Parse markdown | POST | `exercism.org/api/v2/markdown/parse` | No |
