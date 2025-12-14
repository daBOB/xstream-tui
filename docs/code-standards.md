# xstream-tui: Code Standards & Development Guidelines

## Overview

This document defines coding standards, architectural patterns, and development practices for the xstream-tui project. All contributors must adhere to these standards for consistency and maintainability.

## Core Principles

1. **Idiomatic Go:** Follow official Go conventions and best practices
2. **Single Responsibility:** Each package/function has one clear purpose
3. **Small Files:** Target <200 lines per file for readability
4. **Explicit Error Handling:** No silent failures; propagate context
5. **Testing First:** Write tests alongside implementation
6. **Clear Naming:** Self-documenting code with obvious intent

## Directory Structure & Packages

### Package Organization

```
cmd/xstream-tui/
    └── main.go              # Entry point only, thin wrapper

internal/
    ├── xc/                  # Data layer - XC API client
    │   ├── client.go        # HTTP client and auth
    │   ├── models.go        # API response models
    │   ├── categories.go    # Category endpoints
    │   ├── streams.go       # Stream endpoints
    │   └── *_test.go        # Unit tests
    │
    ├── tui/                 # Presentation layer - UI logic
    │   ├── app.go           # Main app model (tea.Model)
    │   ├── styles.go        # Lipgloss styles
    │   ├── screens/         # Screen components
    │   │   ├── login.go
    │   │   ├── categories.go
    │   │   ├── streams.go
    │   │   └── player.go
    │   ├── components/      # Reusable UI elements
    │   │   ├── header.go
    │   │   ├── statusbar.go
    │   │   └── list.go
    │   └── *_test.go
    │
    ├── player/              # Playback layer - media control
    │   ├── manager.go       # Process lifecycle
    │   ├── mpv.go           # mpv-specific logic
    │   ├── vlc.go           # VLC fallback
    │   ├── ipc.go           # IPC communication
    │   └── *_test.go
    │
    └── config/              # Configuration
        ├── config.go        # Config struct and loading
        └── *_test.go
```

### Package Responsibilities

| Package | Purpose | Exports |
|---------|---------|---------|
| `xc` | XC API client operations | Client, Categories, Streams, models |
| `tui` | Terminal UI framework | App (Model), Screens, Components, Styles |
| `player` | Media player management | Manager, PlaybackState |
| `config` | Application configuration | Config, Load, Validate |

## Naming Conventions

### Files
- **lowercase with underscores:** `client.go`, `auth_test.go`, `models.go`
- **Exceptions:** `main.go`, `DOC.md`
- **Tests:** `*_test.go` in same package

### Types (Exported)
- **PascalCase:** `type App struct`, `type XCClient struct`
- **Interfaces:** `type Writer interface`, `type Reader interface`
- **Constants:** `const MaxRetries = 3`, `const TimeoutSeconds = 30`

### Variables (Unexported)
- **camelCase:** `var retryCount int`, `defaultTimeout = 30 * time.Second`
- **Package-level unexported:** `var clients = make(map[string]*Client)`

### Functions
- **PascalCase (exported):** `func NewApp() App`, `func (c *Client) GetStreams() ([]Stream, error)`
- **camelCase (unexported):** `func (a *App) updateView()`, `func validateURL(s string) error`

### Receivers
- **Short names (1-2 chars):** `func (a *App)`, `func (c *Client)`, `func (m *Manager)`
- **Avoid:** `this`, `self`

## Coding Style

### Indentation & Formatting

```go
// Use gofmt (enforced via make lint)
// Tab indentation (standard Go)
// Line length: soft limit 100 chars, hard limit 120 chars

func (c *Client) GetCategories(ctx context.Context) (
    []Category,
    error,
) {
    // Line break for long signatures
}
```

### Imports

```go
package main

import (
    // Standard library (grouped, alphabetical)
    "context"
    "encoding/json"
    "fmt"

    // External dependencies (blank line separator, alphabetical)
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    // Internal packages (blank line separator, alphabetical)
    "github.com/altmueller/xstream-tui/internal/config"
    "github.com/altmueller/xstream-tui/internal/xc"
)
```

### Comments

```go
// Package xc provides Xtream Codes API client functionality.
// It handles authentication, category retrieval, and stream listing.
package xc

// NewClient creates a new XC API client with provided credentials.
// The server URL must include the protocol (http://).
func NewClient(serverURL, username, password string) *Client {
    // ...
}

// shortVariableNameNeeds explanation if not obvious
token := c.generateAccessToken() // comment on same line if brief

// Multi-line comment for complex logic
// explaining the approach and why we chose it
// over alternative implementations.
if err != nil {
    // ...
}
```

**Comment Rules:**
- Every exported type/func has a comment starting with name
- Unexported items comment if non-obvious
- Explain WHY, not WHAT (code shows what it does)
- Keep comments current with code changes

### Error Handling

```go
// Always check and handle errors
if err != nil {
    // Option 1: Wrap with context
    return fmt.Errorf("get categories: %w", err)

    // Option 2: Create custom error
    return &ValidationError{Field: "username", Msg: "empty"}

    // Option 3: Log and return
    log.Printf("retry %d failed: %v", attempt, err)
    return err
}

// Never ignore errors
_ = ioutil.WriteFile(...) // WRONG

// Unwrap and check specific errors
if errors.Is(err, context.DeadlineExceeded) {
    // Handle timeout specifically
}

// Type assertion with check
if perr, ok := err.(*ParseError); ok {
    // Handle parse error
}
```

### Zero Values & Initialization

```go
// Explicit initialization preferred over relying on zero values
client := &Client{
    timeout: 30 * time.Second,
    retries: 3,
    baseURL: serverURL,
}

// Use constructor functions for complex types
func NewApp() App {
    return App{
        width:  80,
        height: 24,
    }
}

// Avoid assuming zero values unless documented
var count int // OK: count defaults to 0
var m map[string]string // WRONG: map must be initialized
m = make(map[string]string) // OK
```

## Architecture Patterns

### The Elm Architecture (Tea.Model)

All TUI screens implement the Tea Model interface:

```go
import tea "github.com/charmbracelet/bubbletea"

type LoginScreen struct {
    username string
    password string
    focused  string // which field is focused
}

// Init initializes the model (called once at start)
func (m LoginScreen) Init() tea.Cmd {
    return nil // or return a command that fetches initial data
}

// Update handles messages and returns updated model + command
func (m LoginScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg), nil
    case tea.WindowSizeMsg:
        m.width = msg.Width
        return m, nil
    default:
        return m, nil
    }
}

// View renders the model as a string
func (m LoginScreen) View() string {
    return m.render()
}
```

**Key Rules:**
- Models are immutable (or appear immutable)
- Update returns new model instance
- View is pure function of state
- Commands handle side effects (API calls, file I/O)

### Error Types

```go
// Define specific error types for domains
type APIError struct {
    Code    int
    Message string
    URL     string
}

func (e *APIError) Error() string {
    return fmt.Sprintf("api error %d: %s (%s)", e.Code, e.Message, e.URL)
}

// Custom error with context
type ValidationError struct {
    Field string
    Value interface{}
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s = %v (%s)", e.Field, e.Value, e.Msg)
}
```

### Dependency Injection

```go
// Pass dependencies as constructor arguments
func NewManager(
    xcClient *xc.Client,
    playerMgr *player.Manager,
    config *config.Config,
) *Manager {
    return &Manager{
        xc:     xcClient,
        player: playerMgr,
        config: config,
    }
}

// Enable testing with mocks
type MockXCClient struct{}
func (m *MockXCClient) GetCategories(ctx context.Context) ([]Category, error) {
    // Return test data
}

func TestManager(t *testing.T) {
    mockXC := &MockXCClient{}
    mgr := NewManager(mockXC, nil, nil)
    // Test against mock
}
```

## Testing Standards

### Test File Organization

```go
// file: client_test.go
package xc

import (
    "context"
    "testing"
)

// Table-driven tests for multiple cases
func TestClientGetStreams(t *testing.T) {
    tests := []struct {
        name        string
        category    string
        wantCount   int
        wantErr     bool
        errContains string
    }{
        {
            name:      "valid category",
            category:  "HD Movies",
            wantCount: 100,
            wantErr:   false,
        },
        {
            name:        "invalid category",
            category:    "",
            wantErr:     true,
            errContains: "empty category",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := NewClient(testServerURL, testUser, testPass)
            got, err := client.GetStreams(context.Background(), tt.category)

            if (err != nil) != tt.wantErr {
                t.Errorf("GetStreams() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if err != nil && !strings.Contains(err.Error(), tt.errContains) {
                t.Errorf("GetStreams() error = %v, want error containing %q", err, tt.errContains)
            }

            if len(got) != tt.wantCount {
                t.Errorf("GetStreams() got %d items, want %d", len(got), tt.wantCount)
            }
        })
    }
}

// Subtests for complex scenarios
func TestClientAuth(t *testing.T) {
    t.Run("valid credentials", func(t *testing.T) {
        // test code
    })

    t.Run("invalid credentials", func(t *testing.T) {
        // test code
    })
}
```

### Testing Guidelines

```go
// Prefer table-driven tests
// Use descriptive test names: Test<Function><Scenario>
// Isolate tests (no shared state between tests)
// Use mocks for external dependencies
// Test error paths explicitly
// Check both positive and negative cases

// Bad test
func TestAuth(t *testing.T) {
    client := NewClient("http://example.com", "user", "pass")
    err := client.Authenticate()
    if err != nil {
        t.Error("auth failed")
    }
}

// Good test
func TestClientAuthenticate_WithValidCredentials_ReturnsNoError(t *testing.T) {
    client := NewClient(
        testServerURL,
        "validuser",
        "validpass",
    )
    err := client.Authenticate(context.Background())
    if err != nil {
        t.Fatalf("Authenticate() error = %v, want nil", err)
    }
}

func TestClientAuthenticate_WithInvalidCredentials_ReturnsError(t *testing.T) {
    client := NewClient(
        testServerURL,
        "invaliduser",
        "wrongpass",
    )
    err := client.Authenticate(context.Background())
    if err == nil {
        t.Error("Authenticate() error = nil, want error")
    }
}
```

## File Size Guidelines

### Target Sizes

| File Type | Target | Max |
|-----------|--------|-----|
| Models | 50-100 lines | 150 |
| Functions | 20-50 lines | 100 |
| Source file | 150-200 lines | 250 |
| Test file | 100-200 lines | 300 |

### When to Split Files

Split when:
- File exceeds 250 lines
- Multiple cohesive types defined
- Different functional areas grouped
- Improves readability/navigation

Example:
```go
// screens/streams.go - too large, split:
// screens/streams_list.go (rendering)
// screens/streams_filter.go (filtering logic)
// screens/streams_state.go (state management)
```

## Linting & Formatting

### Pre-commit Checks

```bash
# Format code
gofmt -w .

# Run linter
go vet ./...

# Run tests
go test -v ./...

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### CI/CD Requirements

Every commit must:
- [ ] Pass `gofmt` (auto-formatted)
- [ ] Pass `go vet` (no vet warnings)
- [ ] Pass `go test -v ./...` (all tests)
- [ ] Maintain >70% coverage on changed files
- [ ] Have meaningful commit messages

## Documentation Standards

### Package Documentation

```go
// Package xc implements the Xtream Codes API client.
//
// The client handles authentication, caching, and API communication.
// It provides methods for retrieving categories and streams.
//
// Example usage:
//
//  client := xc.NewClient("http://example.com", "user", "pass")
//  categories, err := client.GetCategories(ctx)
//  if err != nil {
//      log.Fatal(err)
//  }
package xc
```

### Function Documentation

```go
// GetStreams retrieves all streams in the specified category.
// It returns a slice of Stream structures and any error encountered.
// Calls are cached for 1 hour by default.
//
// Arguments:
//  ctx      - context for cancellation and timeouts
//  category - category ID to retrieve streams from
//
// Returns:
//  []Stream - slice of available streams
//  error    - non-nil if request fails (includes wrapped cause)
func (c *Client) GetStreams(ctx context.Context, category string) ([]Stream, error) {
    // ...
}
```

## Performance Considerations

### Memory

```go
// Reuse buffers for JSON unmarshaling
var v interface{}
decoder := json.NewDecoder(resp.Body)
err := decoder.Decode(&v) // preferred over ioutil.ReadAll

// Use streaming for large responses
for decoder.More() {
    var stream Stream
    decoder.Decode(&stream)
    // process stream
}

// Avoid unnecessary copies
slice := make([]Item, 10000) // pre-allocate if size known
```

### Concurrency

```go
// Use contexts for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Protect shared state
type Manager struct {
    mu     sync.RWMutex
    state  AppState
}

func (m *Manager) GetState() AppState {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.state
}

// Use channels for communication
result := make(chan StreamResult, 1)
go func() {
    result <- fetchStream(id)
}()
```

## Security Guidelines

### Credential Handling

```go
// Never log credentials
log.Printf("Authenticating user: %s", user.Password) // WRONG

// Clear sensitive data
defer func() {
    for i := range password {
        password[i] = 0
    }
}()

// Use secure defaults
const defaultTimeout = 30 * time.Second
const defaultRetries = 3

// Validate input
if username == "" {
    return fmt.Errorf("username required")
}
```

### API Communication

```go
// Always validate URLs
u, err := url.Parse(userInput)
if err != nil || u.Scheme == "" {
    return fmt.Errorf("invalid url: %w", err)
}

// Set timeouts
client := &http.Client{
    Timeout: 30 * time.Second,
}

// Validate responses
if resp.StatusCode >= 400 {
    return &APIError{Code: resp.StatusCode}
}
```

## Code Review Checklist

Before committing, ensure:

- [ ] Code follows naming conventions
- [ ] Functions are <100 lines
- [ ] Files are <250 lines
- [ ] All errors are handled
- [ ] Comments explain WHY
- [ ] Tests cover happy path and errors
- [ ] No hardcoded credentials
- [ ] No unused imports
- [ ] Code passes `gofmt` and `go vet`
- [ ] Coverage maintained >70%

## Git Commit Messages

```
Short summary (50 chars max)

Longer explanation if needed (wrap at 72 chars).
Explain what changed and why.

Fixes: #123
Relates to: #456
```

**Rules:**
- First line: imperative mood ("Add feature", not "Added feature")
- Reference issues and PRs
- One feature/fix per commit
- Keep commits small and atomic

---

**Document Version:** 1.0
**Last Updated:** 2025-12-14
**Enforced Since:** Phase 1 Complete
**Review Cycle:** Quarterly
