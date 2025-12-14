# xstream-tui: Project Overview & Product Development Requirements

## Executive Summary

xstream-tui is a modern, cross-platform terminal user interface (TUI) application for browsing and streaming IPTV content via Xtream Codes API. Built with Go and Bubble Tea framework, it provides a responsive, keyboard-driven experience for accessing 20,000+ streams across multiple categories.

## Project Vision

Enable efficient IPTV consumption in terminal environments without sacrificing UX or feature parity with GUI clients. Target power users, sys admins, and developers who work primarily in terminal interfaces.

## Functional Requirements

### FR-1: User Authentication & Credential Management
- Support Xtream Codes API credentials (username, password, server URL)
- Store credentials securely in `.env` file (with secure access patterns)
- Validate credentials on startup with error handling
- Support credential updates without restart
- **Acceptance Criteria:**
  - Credentials persist across sessions
  - Invalid credentials show clear error message
  - No credentials appear in logs or terminal history

### FR-2: Category Management
- Retrieve and display IPTV categories from XC API
- Cache categories locally for fast browsing
- Support category filtering and search
- Display category metadata (stream count, update time)
- **Acceptance Criteria:**
  - Categories load within 2 seconds (from cache)
  - Categories sync with API every 60 minutes
  - Handle empty category gracefully

### FR-3: Stream Browsing
- Display all streams in selected category with pagination/virtualization
- Support filtering by name, language, or metadata
- Search across all available streams
- Display stream metadata (title, type, language, EPG)
- Handle 20,000+ streams without performance degradation
- **Acceptance Criteria:**
  - Smooth scrolling through 10k+ streams
  - Search returns results in <500ms
  - No memory leaks with large datasets
  - Responsive UI during network operations

### FR-4: Playback Control
- Launch mpv player with stream URL
- Support VLC fallback if mpv unavailable
- Send commands to player via IPC socket
- Pause, resume, stop, seek, volume control
- Display now-playing metadata
- **Acceptance Criteria:**
  - Player launches within 1 second
  - All playback commands responsive (<100ms)
  - Graceful fallback if player fails
  - Support multiple player instances

### FR-5: User Interface Navigation
- Keyboard-driven navigation (vi-like bindings preferred)
- Multi-screen workflow: Login → Categories → Streams → Player
- State machine prevents invalid transitions
- Responsive to terminal resize events
- **Acceptance Criteria:**
  - All navigation responsive (<50ms)
  - UI adapts to terminal size changes instantly
  - No input lag on typical hardware
  - Clear visual feedback for current state

### FR-6: Configuration Management
- Support configuration via `.env` file
- Configurable: timeout, retry limits, cache paths, player preferences
- Environment variables override defaults
- Configuration validation on startup
- **Acceptance Criteria:**
  - All config keys documented
  - Invalid config shows helpful error
  - Changes take effect on restart
  - Defaults work without config file

## Non-Functional Requirements

### NFR-1: Performance
- Startup time: <2 seconds
- API calls: <3 seconds with retry
- UI render: <16ms (60 FPS target)
- Memory: <50MB baseline, <200MB with full cache
- Support 20,000+ streams without degradation

### NFR-2: Reliability
- Handle network timeouts gracefully (3-second timeout, 2 retries)
- Persist cache locally for offline browsing capability
- Automatic recovery from player crashes
- No unhandled panics in production
- Error messages guide user to resolution

### NFR-3: Security
- No credentials in logs, error messages, or terminal history
- HTTPS for API (if supported), HTTP fallback
- Input validation on all user inputs
- No arbitrary code execution vectors
- Secure credential file permissions (0600)

### NFR-4: Compatibility
- Go 1.21+ (slog support)
- Linux, macOS, Windows support
- Terminal: 80x24 minimum (responsive up to 4K)
- TUI framework: Bubble Tea 1.3+
- Tested with xterm, alacritty, tmux, screen

### NFR-5: Maintainability
- Code organized in three decoupled layers
- Single responsibility per file (<200 lines target)
- Comprehensive error context
- Clear package documentation
- >70% test coverage target

## Architecture Overview

### Three-Layer Pattern

```
Presentation Layer (TUI)
    ↓
Business Logic (Screens, State)
    ↓
Data Layer (XC Client) + Playback Layer (mpv)
```

### Key Design Principles

1. **Separation of Concerns:** Each layer has single responsibility
2. **Elm Architecture:** Unidirectional data flow (Model → Update → View)
3. **Immutability:** Messages drive state changes, not mutations
4. **Error Handling:** Explicit error propagation, no silent failures
5. **Testing:** Mock-friendly architecture with dependency injection

## Technical Stack

### Core Dependencies
- **Bubble Tea (v1.3.10+):** TUI framework implementing Elm architecture
- **Lipgloss (v1.1.0+):** Terminal styling and layout
- **Bubbles (v0.21.0+):** Pre-built UI components (textinput, list, etc.)
- **godotenv (v1.5.1+):** Environment file loading

### External Dependencies
- **mpv (system):** Primary media player with IPC socket support
- **VLC (optional):** Fallback player if mpv unavailable
- **XC API:** Xtream Codes compatible IPTV backend

## Implementation Phases

### Phase 1: Project Setup ✅ COMPLETE
- Initialize Go module with semantic versioning
- Install core dependencies (Bubble Tea, lipgloss, bubbles)
- Create directory structure following idiomatic Go patterns
- Establish Makefile with build, test, lint targets
- Create minimal working TUI entry point
- **Deliverables:** Working build, executable, test infrastructure

### Phase 2: Data Layer ✅ COMPLETE
- Implement XC API client with auth, categories, streams endpoints
- Handle JSON inconsistencies in API responses
- Implement local caching mechanism (SQLite or file-based)
- Error handling for network/API issues
- **Deliverables:** Testable XC client with 80%+ coverage

### Phase 3: Presentation Layer ✅ COMPLETE
- Implement login screen with credential input
- Build category screen with listing and filtering
- Build stream browser with virtualization for 20k+ items
- Implement search and filtering UI
- Create player screen with controls and metadata
- **Deliverables:** Complete TUI flow with navigation

### Phase 4: Playback Layer ✅ COMPLETE
- Implement mpv process manager and IPC communication
- Build VLC fallback support
- Implement command queue for playback control
- Handle player crashes and recovery
- **Deliverables:** Full playback control with both players

### Phase 5: Integration & Testing ✅ COMPLETE
- End-to-end testing across all layers
- Performance testing with large datasets
- Network resilience testing
- Cross-platform testing (Linux, macOS, Windows)
- User acceptance testing
- **Deliverables:** Integration test suite, performance baselines

### Phase 6: Polish & Release ✅ COMPLETE
- UI/UX refinement based on feedback
- Documentation and help system
- Optimization for memory and startup time
- Release automation and versioning
- **Deliverables:** v1.0 release with docs and examples

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Startup Time | <2 sec | Time from launch to TUI display |
| API Response | <3 sec | XC API call with retries |
| UI Responsiveness | <50ms | Keyboard input to visual feedback |
| Memory Usage | <50MB | Baseline with empty cache |
| Test Coverage | >70% | go tool cover report |
| Availability | 99% | Graceful error handling, no panics |

## Risk Management

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| XC API inconsistencies | High | Parsing failures | Robust JSON unmarshaling, extensive testing |
| Network timeouts | Medium | User frustration | Retry logic, caching, clear messaging |
| mpv unavailability | Low | No playback | VLC fallback, graceful degradation |
| Terminal compatibility | Low | Display issues | Test on major terminals, responsive design |
| Memory leaks | Low | Long-term stability | Profiling, GC monitoring, test loads |

## Dependencies & Constraints

### External Dependencies
- Go 1.21+ (slog, better generics)
- XC-compatible IPTV server with API access
- mpv or VLC media player on system
- POSIX-compatible terminal (Windows with ConPTY support)

### Technical Constraints
- TUI must work without X11 (SSH-friendly)
- No privileged operations required
- HTTP-only API communication (TLS if available)
- Limited to host's terminal size

## Documentation Structure

```
docs/
├── project-overview-pdr.md       # This file (requirements & overview)
├── code-standards.md              # Coding patterns and conventions
├── codebase-summary.md            # Current implementation status
├── system-architecture.md         # Detailed technical design
├── deployment-guide.md            # Installation and setup
└── project-roadmap.md             # Feature pipeline and timeline
```

## Roadmap

### v1.0 (MVP - COMPLETE 2025-12-14)
- [x] Project setup
- [x] Data layer (XC client)
- [x] Presentation layer (TUI)
- [x] Playback layer (mpv)
- [x] Integration & testing
- [x] Polish & Release

### v1.1 (Enhancement)
- [ ] Search optimization
- [ ] Playlist support
- [ ] Favorites/history
- [ ] Theme customization

### v1.2 (Advanced)
- [ ] Multiple account support
- [ ] Smart caching
- [ ] Recording support
- [ ] EPG integration

### v2.0 (Future)
- [ ] Web UI companion
- [ ] Mobile app
- [ ] Plugin system
- [ ] Commercial player support

## Glossary

| Term | Definition |
|------|-----------|
| XC API | Xtream Codes API - IPTV provider backend |
| TUI | Terminal User Interface |
| Bubble Tea | Go TUI framework implementing Elm architecture |
| Elm Architecture | Pattern: Model → Update → View (unidirectional flow) |
| IPC | Inter-Process Communication |
| VLC | VideoLAN Client - media player fallback |
| mpv | Modern media player with IPC socket support |
| Virtualization | Rendering only visible items (performance optimization) |

## Acceptance Criteria Summary

The project is considered complete when:

1. TUI launches in <2 seconds
2. User can authenticate with XC credentials
3. Categories load and display within 2 seconds
4. Streams display with pagination for 20,000+ items
5. Stream playback works via mpv with fallback to VLC
6. All user inputs responsive (<50ms)
7. No unhandled panics in normal usage
8. Code coverage >70%
9. Works on Linux, macOS, Windows
10. Documentation complete and accurate

---

**Document Version:** 1.0
**Last Updated:** 2025-12-14
**Status:** ✅ ALL PHASES COMPLETE - v1.0.0 Ready for Release
**Owner:** Development Team
**Next Review:** After v1.0.0 Public Release
