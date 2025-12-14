# xstream-tui Codebase Summary

## Project Overview

xstream-tui is a Go-based terminal user interface (TUI) for IPTV streaming via Xtream Codes API with mpv playback integration. Implements The Elm Architecture pattern using Bubble Tea framework.

**Repository:** github.com/altmueller/xstream-tui
**Go Version:** 1.25.5
**Status:** Phase 2 (Data Layer) - Complete

## Architecture

Three decoupled layers:

```
┌─────────────────────────────────────────────────────────────┐
│                      main.go (Entry)                        │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   TUI Layer (Bubble Tea)                     │
│  ┌─────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐           │
│  │ Login   │→│Categories│→│Streams │→│ Player   │           │
│  └─────────┘ └──────────┘ └────────┘ └──────────┘           │
│                     State Machine                            │
└────────────┬────────────────────────────────────┬───────────┘
             │                                    │
┌────────────▼────────────┐        ┌──────────────▼───────────┐
│     Data Layer          │        │    Playback Layer        │
│  ┌───────────────┐      │        │  ┌─────────────────┐     │
│  │ XC Client     │      │        │  │ Process Manager │     │
│  │ - Auth        │      │        │  │ - mpv spawn     │     │
│  │ - Categories  │      │        │  │ - IPC socket    │     │
│  │ - Streams     │      │        │  │ - VLC fallback  │     │
│  └───────────────┘      │        │  └─────────────────┘     │
└─────────────────────────┘        └──────────────────────────┘
```

## Directory Structure

```
xstream-tui/
├── cmd/
│   └── xstream-tui/
│       └── main.go                # Entry point
├── internal/
│   ├── xc/                        # Data layer (Xtream Codes)
│   │   ├── client.go              # HTTP client with validation
│   │   ├── models.go              # Data models & FlexibleID type
│   │   ├── endpoints.go           # 12 API endpoints
│   │   ├── models_test.go         # Model unit tests
│   │   └── client_test.go         # Client + integration tests
│   ├── tui/                       # Presentation layer
│   │   ├── app.go                 # Main application model (TEA)
│   │   ├── screens/               # Screen components (planned)
│   │   ├── components/            # Reusable UI elements (planned)
│   │   └── styles.go              # Styling utilities (planned)
│   ├── player/                    # Playback layer
│   │   ├── manager.go             # Process manager (planned)
│   │   ├── mpv.go                 # mpv integration (planned)
│   │   └── ipc.go                 # IPC communication (planned)
│   └── config/                    # Configuration
│       └── config.go              # Config management (planned)
├── go.mod                         # Module definition
├── go.sum                         # Dependency checksums
├── Makefile                       # Build targets
├── .gitignore                     # Git exclusions
└── docs/                          # Documentation
```

## Core Components

### Entry Point (`cmd/xstream-tui/main.go`)

- Initializes Bubble Tea program with alternate screen buffer
- Creates App instance and runs TUI event loop
- Handles runtime errors with stderr output

**Key Code:**
```go
p := tea.NewProgram(tui.NewApp(), tea.WithAltScreen())
if _, err := p.Run(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

### Data Layer (`internal/xc/`)

**Client** (`client.go`)
- HTTP client for Xtream Codes API with validation
- Credentials injected via URL path/query (API requirement)
- Configurable timeout and custom HTTP client support
- Error handling with size limits (maxErrorBodySize = 4096)
- Security documentation: credentials in URLs, use HTTPS when available, store securely

**Models** (`models.go`)
- `FlexibleID`: Handles IDs as int or string (provider inconsistency)
  - Methods: `String()`, `Int()`, `IsZero()`, `NewFlexibleID()`, `NewFlexibleIDFromString()`
  - Implements `json.Marshaler` and `json.Unmarshaler`
- `UserInfo`: Account info (username, status, active connections, expiry)
- `ServerInfo`: XC server metadata (URL, ports, protocol, timezone)
- `AuthResponse`: Authentication result (user + server info)
- `Category`: Content category (ID, name, parent ID)
- `LiveStream`: Live TV channel (ID, name, icon, category, EPG channel ID, archive)
- `VODStream`: Video-on-demand item (ID, name, rating, container, direct source)
- `Series`: TV series (ID, name, cover, added date)
- `SeriesInfo`: Detailed series with episodes
- `Episode`: Individual season/episode
- `EPGShort`: Current + next program
- `EPGEntry`: EPG listing with start time and duration

**Endpoints** (`endpoints.go`) - 12 methods
- `Authenticate(ctx)`: Verify credentials, validate account status
- `GetLiveCategories(ctx)`, `GetVODCategories(ctx)`, `GetSeriesCategories(ctx)`: Category listings
- `GetLiveStreams(ctx, categoryID)`, `GetVODStreams(ctx, categoryID)`, `GetSeries(ctx, categoryID)`: Stream listings
- `GetSeriesInfo(ctx, seriesID)`: Series details with episodes
- `GetVODInfo(ctx, vodID)`: VOD item details
- `GetShortEPG(ctx, streamID)`: Current/next program
- `GetSimpleDataTable(ctx, streamID)`: EPG entries for date range
- `GetAllLiveStreams(ctx)`, `GetAllVODStreams(ctx)`, `GetAllSeries(ctx)`: Unfiltered listings

**Tests** (`models_test.go`, `client_test.go`)
- Unit tests for FlexibleID marshaling/unmarshaling
- Client initialization and validation tests
- Integration tests using httptest mock server

### TUI Model (`internal/tui/app.go`)

Implements Bubble Tea's `tea.Model` interface using The Elm Architecture:

**Type:** `App` struct
- `width`: Terminal width
- `height`: Terminal height

**Methods:**
- `NewApp()`: Creates application instance
- `Init()`: Initialization hook (returns no commands)
- `Update(msg tea.Msg)`: Handles messages and state updates
  - Quit on 'q' or Ctrl+C
  - Tracks terminal dimensions on WindowSizeMsg
- `View()`: Renders centered placeholder with branding

**Styling:** Uses lipgloss for terminal styling (color 86 for title, gray for instructions)

### Dependencies

**Core TUI Stack:**
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI framework
- `github.com/charmbracelet/bubbles` v0.21.0 - UI components (optional)
- `github.com/charmbracelet/lipgloss` v1.1.0 - Styling library
- `github.com/charmbracelet/colorprofile` v0.2.3 - Color support

**Utilities:**
- `github.com/joho/godotenv` v1.5.1 - Environment file loading
- Supporting libraries for terminal/ANSI handling

See `go.mod` for complete dependency tree with versions.

## Build & Development

### Makefile Targets

```makefile
make build           # Build binary to bin/xstream-tui
make run            # Run application directly
make test           # Run tests with verbose output
make test-coverage  # Generate coverage report
make clean          # Remove build artifacts
make lint           # Run go vet
make tidy           # Tidy dependencies
```

### Build Output
- Binary location: `./bin/xstream-tui`
- Platform: Linux/macOS/Windows (Go standard cross-compilation supported)

## Development Patterns

### Code Organization
- **Package per layer:** Clear separation of concerns
- **Small files:** Target <200 lines per file (development guideline)
- **Import paths:** Use absolute imports (github.com/altmueller/xstream-tui/...)
- **Naming:** Follow Go conventions (CamelCase for exports, lowercase for internal)

### Error Handling
- Stderr for error output
- Exit code 1 on fatal errors
- Bubble Tea framework handles I/O errors internally

### Testing
- Unit tests in `*_test.go` files
- Coverage reports in HTML format
- Run with `make test` or `make test-coverage`

## Phase Status

**Phase 1: Project Setup** ✅ COMPLETE
- Go module initialized
- Dependencies installed
- Directory structure created
- Minimal TUI working
- Build targets operational
- Code review passed

**Phase 2: Data Layer** ✅ COMPLETE
- FlexibleID type for int/string JSON handling
- Complete data models (UserInfo, ServerInfo, Category, LiveStream, VODStream, Series, SeriesInfo, Episode, EPG)
- HTTP client with validation and error handling
- 12 API endpoints implemented (Authenticate, GetLiveCategories, GetVODCategories, GetSeriesCategories, GetLiveStreams, GetVODStreams, GetSeries, GetSeriesInfo, GetVODInfo, GetShortEPG, GetSimpleDataTable, GetAllStreams)
- Unit tests for models (FlexibleID marshaling)
- Integration tests using httptest mock server

**Next Phases:**
- Phase 3: Presentation Layer (Screens & components)
- Phase 4: Playback Layer (mpv integration)
- Phase 5: Integration & testing
- Phase 6: Polish & optimization

## Key Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Go Version | 1.25.5 (supports 1.21+) | slog availability, generics support |
| TUI Framework | Bubble Tea | Cross-platform, active maintenance, Elm architecture |
| Credential Storage | .env file | Simple, secure via .gitignore |
| API Protocol | HTTP only | XC API standard (no HTTPS verification) |
| Cache | Persistent | Session across restarts for UX |

## Security Considerations

- `.gitignore` prevents accidental credential commits
- No hardcoded credentials in code
- Environment variables via .env loading
- Terminal runs in alt screen (separate from shell history)

## Links

- [Project Overview & PDR](./project-overview-pdr.md)
- [Code Standards](./code-standards.md)
- [System Architecture](./system-architecture.md)
- [Phase 1 Implementation Plan](../plans/251213-1633-iptv-tui-implementation/phase-01-project-setup.md)

---

**Last Updated:** 2025-12-14
**Status:** Phase 2 Complete
**Author:** Docs Manager
