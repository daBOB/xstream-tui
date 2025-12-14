# xstream-tui: System Architecture

## Architecture Overview

xstream-tui implements a four-layer architecture with clean separation of concerns:

```
┌────────────────────────────────────────────────────────────────┐
│                 Presentation Layer (TUI)                       │
│           Bubble Tea Model-View-Update Pattern                 │
│                                                                │
│  ┌───────────┐  ┌──────────┐  ┌────────┐  ┌────────────────┐ │
│  │  Login    │→ │Categories│→ │Streams │→ │  Player Screen │ │
│  │  Screen   │  │  Screen  │  │ Screen │  │  (Playback)    │ │
│  └───────────┘  └──────────┘  └────────┘  └────────────────┘ │
│                                                                │
│  Download Queue Overlay (toggled with 'D')                    │
│  ┌──────────────────────────────────────────┐                │
│  │ Downloads: Progress, Cancel, Remove      │                │
│  └──────────────────────────────────────────┘                │
│                                                                │
│            State Machine (App struct) at center               │
└────────────────┬──────────────┬────────────────┬──────────────┘
                 │              │                │
         ┌───────▼─────┐        │                │
         │  Data Layer │        │                │
         │   (XC API)  │        │                │
         └─────────────┘        │                │
                        ┌───────▼──────────────┐ │
                        │  Playback Layer      │ │
                        │  (mpv/VLC)           │ │
                        └──────────────────────┘ │
                        ┌───────▼──────────────┐ │
                        │ Download Layer       │ │
                        │ (Queue Manager)      │ │
                        └──────────────────────┘ │
                                        ┌────────▼──────┐
                                        │ Styling &     │
                                        │ Layout        │
                                        └───────────────┘
```

## Layer Descriptions

### 1. Presentation Layer (TUI)

**Location:** `internal/tui/`

**Responsibility:** User interaction, event handling, and visual rendering using Bubble Tea framework.

**Components:**

#### Application Model (`internal/tui/app.go`)
- Implements `tea.Model` interface (Init, Update, View)
- Central state holder for entire application
- Coordinates between screens
- Responds to terminal resize events

```go
type App struct {
    width  int
    height int
    // Additional state: current screen, user data, etc.
}

// Init called once at startup
// Update called for each message
// View returns rendered string
```

#### Screen Components (`internal/tui/screens/`)
Each screen is a separate model:
- **LoginScreen:** Credential input and authentication
- **CategoriesScreen:** Category listing and selection
- **StreamsScreen:** Stream browsing with virtualization
- **PlayerScreen:** Playback control and metadata display

Pattern: Each screen implements `tea.Model` independently
Composition: App delegates to active screen

#### Reusable Components (`internal/tui/components/`)
- **Header:** Application title and status
- **StatusBar:** Current action/mode display
- **List:** Virtualized list rendering (handles 20k+ items)
- **Input:** Text input with validation

#### Styling (`internal/tui/styles.go`)
- Lipgloss style definitions
- Color scheme and theme constants
- Layout dimensions and padding
- Responsive design helpers

### 2. Data Layer

**Location:** `internal/xc/`

**Responsibility:** Xtream Codes API communication, authentication, and data transformation.

**Components:**

#### XC Client (`internal/xc/client.go`)
- HTTP client with retry logic (3 retries, exponential backoff)
- Request signing and authentication
- Error handling and recovery
- Context support for cancellation

```go
type Client struct {
    baseURL   string
    username  string
    password  string
    httpClient *http.Client
    cache     *Cache // local caching
}

func (c *Client) Authenticate(ctx context.Context) error
func (c *Client) GetCategories(ctx context.Context) ([]Category, error)
func (c *Client) GetStreams(ctx context.Context, categoryID string) ([]Stream, error)
```

#### Data Models (`internal/xc/models.go`)
- `Category`: Category metadata and stream count
- `Stream`: Individual stream information (title, type, language, EPG)
- `User`: Authentication and account info

API Response handling:
- Graceful JSON unmarshaling (XC API inconsistencies)
- Type conversions (string → int, null handling)
- Validation of required fields

#### Caching (`internal/xc/cache.go`)
- Local storage of categories and streams
- TTL-based invalidation (categories: 60 min, streams: 24 hours)
- File-based storage for offline access
- Graceful fallback if cache unavailable

#### Endpoints (`internal/xc/endpoints.go`)
- URL construction for XC API endpoints
- Parameter validation
- Response parsing

### 3. Playback Layer

**Location:** `internal/player/`

**Responsibility:** Media player lifecycle management and IPC communication for playback control.

**Components:**

#### Player Manager (`internal/player/manager.go`)
Central coordinator for player lifecycle:
- Unified player interface for mpv/VLC
- Auto-fallback: tries mpv first, falls back to VLC if unavailable
- Lifecycle tracking (spawn → wait → exit)
- Graceful shutdown with exit monitoring
- Thread-safe operations (sync.Mutex protection)
- Exit callback mechanism for TUI integration

```go
type Manager struct {
    current    Player      // Active player (MPVPlayer or VLCPlayer)
    playerType PlayerType  // "mpv" or "vlc"
    onExit     func(error) // Callback when player exits
    playing    bool
}

func (m *Manager) Play(ctx context.Context, url, title string) error
func (m *Manager) Stop() error
func (m *Manager) TogglePause() error
func (m *Manager) Seek(seconds int) error
func (m *Manager) SetVolume(vol int) error
func (m *Manager) IsPlaying() bool
```

#### mpv IPC Integration (`internal/player/mpv.go`)
Full-featured playback control via IPC socket:
- Process spawning with IPC socket path
- Real-time property queries (position, duration, pause state)
- Command execution (pause, resume, seek, volume)
- Graceful shutdown with quit command

```go
type MPVPlayer struct {
    cmd        *exec.Cmd
    ipc        *IPCClient
    socketPath string
}

// Query/Control methods via IPC
func (p *MPVPlayer) GetPosition() (float64, error)
func (p *MPVPlayer) GetDuration() (float64, error)
func (p *MPVPlayer) IsPaused() (bool, error)
func (p *MPVPlayer) SetVolume(vol int) error
func (p *MPVPlayer) Seek(seconds int) error
```

#### IPC Communication (`internal/player/ipc.go`)
JSON-RPC bidirectional communication with mpv:
- Socket connection management with retry logic (5s timeout)
- Request buffering with request ID tracking
- Asynchronous response handling via goroutine
- Response timeout protection (5s per command)

```go
type IPCClient struct {
    conn      net.Conn
    pending   map[int]chan Response  // Track in-flight requests
    reader    *bufio.Reader
}

func NewIPCClient(socketPath string) (*IPCClient, error)
func (c *IPCClient) Command(args ...any) (Response, error)
```

#### Socket Management (`internal/player/socket.go`)
Platform-specific socket path handling:
- **Linux/macOS:** Unix sockets in XDG_RUNTIME_DIR or temp
- **Windows:** Named pipes (experimental)
- PID-based socket naming for isolation
- Cleanup on player termination

#### VLC Fallback (`internal/player/vlc.go`)
Lightweight fallback when mpv unavailable:
- Direct subprocess execution (no IPC)
- Basic launch parameters only
- No runtime control (pause/seek/volume no-ops)
- Platform-specific executable paths

### 4. Download Layer

**Location:** `internal/download/`

**Responsibility:** Queue-based file downloads with progress tracking and cancellation.

**Components:**

#### Download Manager (`internal/download/manager.go`)
- Single-threaded queue processor (one concurrent download)
- FIFO queue with status tracking
- Item struct: ID, Name, URL, FilePath, Status, Progress, Size, Downloaded, Error

```go
type Manager struct {
    queue       []*Item
    downloadDir string
    httpClient  *http.Client
    onProgress  func(ProgressUpdate)
    nextID      int
    isProcessing bool  // Prevents concurrent queue processing
}

// Key methods
func (m *Manager) Add(name, urlStr string) string       // Returns ID
func (m *Manager) Cancel(id string) bool                // Cancels active download
func (m *Manager) Remove(id string) bool                // Removes completed item
func (m *Manager) Queue() []Item                        // Returns current queue
func (m *Manager) ActiveDownload() *Item                // Current download
func (m *Manager) SetProgressCallback(cb func(...))     // Register progress handler
```

**Features:**
- 32KB buffered file writes for efficiency
- Temp file handling (.tmp suffix) with atomic rename on completion
- HTTP context cancellation for graceful stopping
- Progress updates on every chunk received
- Status lifecycle: Queued → Downloading → Completed/Failed/Cancelled

**Download Directory:**
- Default: ~/Downloads/xstream-tui/
- Override via XSTREAM_DOWNLOAD_DIR environment variable
- Auto-creates directory with 0755 permissions

#### Queue Component Integration (`internal/tui/components/download_queue.go`)
- Overlay panel displaying active downloads
- Keyboard navigation: `j`/`k` or arrow keys to select
- Action keys: `d` (cancel), `x` (remove), `Esc` (close)
- Visual feedback: Progress bars, status icons, percentage display
- Auto-hides when queue is empty

### 5. Configuration Layer

**Location:** `internal/config/`

**Responsibility:** Configuration loading, validation, and management.

**Components:**

#### Configuration (`internal/config/config.go`)
- Load from `.env` file via `godotenv`
- Environment variable overrides
- Validation of required fields
- Default values for optional settings

```go
type Config struct {
    // XC API
    XCServerURL string // http://iptv.example.com
    XCUsername  string
    XCPassword  string

    // Player
    PlayerTimeout  time.Duration
    PreferredPlayer string // "mpv" or "vlc"

    // Cache
    CachePath     string
    CacheTTL      time.Duration

    // TUI
    TerminalWidth  int
    TerminalHeight int
}
```

## Data Flow

### Authentication Flow

```
User Input (LoginScreen)
    │
    ▼
App.Update(KeyMsg)
    │
    ├─→ LoginScreen.Update(KeyMsg)
    │   └─→ Update input fields
    │
    └─→ User presses Enter
        │
        ▼
        App.Update(LoginSubmitted)
        │
        ├─→ xcClient.Authenticate()
        │   │
        │   ├─→ HTTP POST /api/user
        │   ├─→ Validate response
        │   └─→ Return error or token
        │
        ├─→ On success: Switch to CategoriesScreen
        └─→ On error: Display error message
```

### Stream Listing Flow

```
User selects Category
    │
    ▼
App.Update(CategorySelected)
    │
    ├─→ CategoriesScreen.Update()
    │
    ├─→ App switches to StreamsScreen
    │
    ├─→ StreamsScreen.Init()
    │   │
    │   ├─→ Check cache
    │   │   └─→ If valid: Load from cache
    │   │
    │   └─→ If not cached or expired:
    │       │
    │       ├─→ xcClient.GetStreams(categoryID)
    │       ├─→ Store in cache
    │       └─→ Return streams
    │
    └─→ StreamsScreen.View()
        │
        ├─→ Render only visible items (virtualization)
        ├─→ Apply filters/search
        └─→ Display pagination info
```

### Playback Flow

```
User presses Enter on Stream
    │
    ▼
App.Update(StreamSelectedMsg)
    │
    ├─→ Extract URL and title
    │
    ├─→ playerManager.Play(ctx, url, title)
    │   │
    │   ├─→ Try NewMPVPlayer()
    │   │   ├─→ Spawn mpv process with IPC socket
    │   │   ├─→ Connect to IPC socket (5s retry)
    │   │   └─→ Success: Return MPVPlayer
    │   │
    │   └─→ If mpv fails:
    │       ├─→ Try NewVLCPlayer()
    │       │   ├─→ Spawn VLC process
    │       │   └─→ Success: Return VLCPlayer
    │       │
    │       └─→ If VLC fails: Return error
    │
    ├─→ monitorExitLocked() spawns goroutine
    │   └─→ Waits for player process exit
    │       └─→ Calls onExit callback
    │
    ├─→ onExit callback sends PlayerStoppedMsg
    │   └─→ Triggers App.Update(PlayerStoppedMsg)
    │
    └─→ App updates UI and returns to StreamsScreen
```

**Key Details:**
- mpv connection retries for 5 seconds (100ms intervals) to allow socket creation
- Each player gets unique socket name via PID
- VLC used only when mpv unavailable or fails
- Player monitor goroutine handles async exit detection
- Exit callback ensures non-blocking player cleanup

## Message Types (Bubble Tea)

```go
// Terminal events (from tea framework)
tea.KeyMsg          // User pressed key
tea.WindowSizeMsg   // Terminal resized

// Application events
type LoginSubmitted struct {
    username string
    password string
}

type CategorySelected struct {
    categoryID string
}

type StreamSelected struct {
    streamID  string
    streamURL string
}

type PlaybackStateChanged struct {
    state   PlaybackState
    message string
}

type APIError struct {
    endpoint string
    err      error
}

// Playback layer messages
type PlayerStartedMsg struct {
    PlayerType string // "mpv" or "vlc"
}

type PlayerStoppedMsg struct {
    Err error // nil if ended normally
}

// Download layer messages
type DownloadRequestMsg struct {
    Name string  // Display name
    URL  string  // Download URL
}

type DownloadProgressMsg struct {
    ID         string
    Progress   float64
    Downloaded int64
    Size       int64
    Status     download.Status
    Error      error
}

type DownloadQueueToggleMsg struct{} // Toggle queue visibility with 'D'

type DownloadCancelMsg struct {
    ID string  // Download ID to cancel
}

type DownloadRemoveMsg struct {
    ID string  // Download ID to remove
}
```

## State Machine

```
Entry Point
    │
    ▼
[LOGIN]
 │ Success
 ├─→ AuthToken acquired
 │
 ▼
[CATEGORIES]
 │ Category selected
 │
 ▼
[STREAMS]
 │ Stream selected
 │
 ▼
[PLAYER]
 │ User presses q
 │
 ▼
[STREAMS] (or back to CATEGORIES)
 │ User presses Esc
 │
 ▼
[CATEGORIES]
 │ User presses Esc
 │
 ▼
[EXIT]

Backpressure:
- All states can quit with 'q'
- Escape goes back to parent state
- Network errors can transition to error state
```

## Concurrency Model

### Single-Threaded Main Loop (Bubble Tea)

Bubble Tea's event loop is single-threaded:

```go
for {
    // 1. Get next message
    msg := getNextMessage()

    // 2. Call Update (blocking)
    model, cmd := model.Update(msg)

    // 3. Execute command if returned
    if cmd != nil {
        result := executeCommand(cmd)
        // Result becomes next message
    }

    // 4. Render with View()
    screen := model.View()
    display(screen)
}
```

### Commands for Async Operations

```go
// Command returns a function that takes a message channel
type Cmd func() Msg

// Example: Fetch categories async
func (a *App) fetchCategoriesCmd() tea.Cmd {
    return func() tea.Msg {
        categories, err := a.xcClient.GetCategories(a.ctx)
        if err != nil {
            return CategoriesError{err: err}
        }
        return CategoriesLoaded{items: categories}
    }
}

// Update handles the result
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case CategoriesLoaded:
        a.categories = msg.items
        return a, nil
    }
}
```

## Error Handling Strategy

### Error Types

1. **Network Errors:** Timeout, connection refused
   - Action: Retry with exponential backoff
   - User: "Connecting... (Retry 2/3)"

2. **API Errors:** Auth failed, invalid category
   - Action: Display error, return to previous screen
   - User: "Authentication failed: Invalid credentials"

3. **Player Errors:** Span multiple layers
   - **Process Start Error:** mpv/VLC executable not found
     - Action: Try fallback player (mpv → VLC)
     - User: "No player available: mpv (error), vlc (error)"
   - **IPC Socket Error:** Socket creation timeout or connection refused
     - Action: Fallback to VLC, both implementations handle gracefully
     - User: "Player error, falling back to VLC"
   - **Command Timeout:** IPC command unresponded (5s timeout)
     - Action: Continue playback (read-only failure)
     - User: May miss state display but playback continues
   - **Player Exit:** Abnormal process termination
     - Action: Monitor callback sends PlayerStoppedMsg
     - User: "Playback ended" (with optional error details)

4. **Validation Errors:** Empty username, invalid URL
   - Action: Highlight field, show error
   - User: Display error below input field

### Propagation

```
Low-level error (socket, file I/O)
    │
    ▼
Wrapped with context
    fmt.Errorf("get categories: %w", err)
    │
    ▼
Handled at appropriate level
    - Network layer: Retry or fail
    - Business logic: Map to user-friendly message
    - UI layer: Display or log
```

## Performance Considerations

### Virtualization
- Only render visible items (terminal height)
- Keyboard navigation scrolls viewport
- Memory: O(viewport) not O(total items)
- Handles 20,000+ streams smoothly

### Caching
- Categories cached 60 minutes
- Streams cached 24 hours
- File-based cache survives restarts
- Graceful degradation if cache corrupted

### API Optimization
- Batch requests where possible
- Progressive loading (show categories while fetching streams)
- Request timeouts prevent hanging
- Exponential backoff for retries

### Memory Management
- Streaming JSON parsing for large responses
- Pre-allocated slices when size known
- Proper cleanup on screen transitions
- Context cancellation for pending requests

## Security Considerations

### Credential Management
- Credentials stored in `.env` (not version controlled)
- Memory cleared after use (overwrite with zeros)
- No logging of sensitive data
- No credential transmission in logs

### Input Validation
- All user inputs validated before use
- URL parsing and validation
- Integer bounds checking
- Whitelist characters for filters

### API Communication
- HTTPS when available (graceful fallback to HTTP)
- Request timeout prevents indefinite hangs
- Response validation (expected fields)
- No arbitrary code execution

## Deployment Architecture

### Single Binary
- Go statically compiles to single executable
- No runtime dependencies (except mpv/VLC)
- Can run from any directory
- No installation required

### System Dependencies
- **Linux/macOS:** mpv or VLC (optional, playback only)
- **Windows:** ConPTY support (Windows 10+)
- **Terminal:** 80x24 minimum (responsive design)

### Configuration
- `.env` file in working directory
- Environment variables override `.env`
- Defaults work without configuration
- Validation on startup

## Scalability

### Horizontal
- Application is single-user, single-instance
- No distributed considerations
- Each user runs their own instance

### Vertical
- Handles 20,000+ streams efficiently
- Memory bounded (cache + viewport)
- CPU light (event-driven, not polling)
- Network-bound (API latency dominates)

## Testing Architecture

### Unit Tests
- Mock XC client for TUI layer
- Test business logic in isolation
- Snapshot tests for UI rendering

### Integration Tests
- Real API communication (against test server)
- Player mock or local test player
- End-to-end screen navigation

### Performance Tests
- Large stream count (20k+) rendering
- Cache efficiency under load
- Memory profiling

## Integration Patterns (Phase 5)

### Complete Message Flow

Integration wires three layers (Presentation → Auth → Data → Playback) through tea.Cmd async pattern and message-driven architecture.

#### Login → Auth → XC Client

```
User Input (LoginScreen)
    │
    ├─→ TextInput fields (host, port, user, pass)
    │   └─→ Form validation (non-empty fields)
    │
    ├─→ Enter key on password field
    │   └─→ submit() returns tea.Cmd
    │
    ├─→ Cmd executes async:
    │   │
    │   ├─→ xc.NewClient(url, user, pass)
    │   │   └─→ Client initialized
    │   │
    │   ├─→ client.Authenticate(ctx) via HTTP
    │   │   └─→ GET /api/user?username=...&password=...
    │   │
    │   └─→ Return AuthSuccessMsg or AuthErrorMsg
    │
    ├─→ App.Update(AuthSuccessMsg)
    │   ├─→ Store client in app.client
    │   ├─→ Clear password from memory (security)
    │   ├─→ Store UserInfo (username, expiration, etc)
    │   └─→ navigateTo(ContentTypeScreen)
    │
    └─→ On AuthErrorMsg:
        ├─→ Display error in login form
        └─→ User can retry
```

**Key Patterns:**
- Password cleared immediately after successful auth (line 144 in app.go)
- Form input cleared on submit for security (login.go:152)
- Async command prevents UI freeze during HTTP request
- Error handling delegates back to login screen for retry

#### ContentType → Categories → API Loading

```
User selects content type (Live/VOD/Series)
    │
    ├─→ ContentTypeScreen.Update() returns tea.Cmd
    │   └─→ ContentTypeSelectedMsg
    │
    ├─→ App.Update(ContentTypeSelectedMsg)
    │   ├─→ Store currentType in app state
    │   ├─→ Set loading flag
    │   └─→ navigateTo(CategoriesScreen)
    │
    ├─→ navigateTo() initializes next screen:
    │   ├─→ SetClient(app.client) - inject shared client
    │   ├─→ SetContentType(currentType) - triggers load
    │   └─→ Returns tea.Cmd from screen init
    │
    ├─→ SetContentType() → loadCategories() cmd:
    │   │
    │   ├─→ Check cache first
    │   │   └─→ Cache key: content_type
    │   │
    │   ├─→ If cached & valid TTL:
    │   │   └─→ Return CategoriesLoadedMsg immediately
    │   │
    │   └─→ If expired or missing:
    │       ├─→ client.GetCategories(ctx, contentType)
    │       ├─→ Await API response
    │       └─→ Return CategoriesLoadedMsg with data
    │
    ├─→ App.Update(CategoriesLoadedMsg)
    │   ├─→ Set loading = false
    │   ├─→ Pass to CategoriesScreen for rendering
    │   └─→ Return spinner stop cmd
    │
    └─→ CategoriesScreen renders categories list
```

**Key Patterns:**
- Dependency injection through SetClient/SetContentType
- SetContentType() returns tea.Cmd for async loading
- Navigation triggers initial data load
- Loading spinner controlled by App.loading flag

#### Streams → Player Launch

```
User selects category → StreamsScreen populated
    │
    ├─→ CategorySelectedMsg triggers:
    │   ├─→ Store currentCat
    │   ├─→ Set loading flag
    │   └─→ navigateTo(StreamsScreen)
    │
    ├─→ navigateTo(StreamsScreen):
    │   ├─→ SetClient(client)
    │   ├─→ SetCategory(cat, contentType) → load streams async
    │   └─→ Spinner tick begins
    │
    ├─→ StreamsLoadedMsg:
    │   ├─→ Populate list with streams
    │   └─→ Stop spinner
    │
    ├─→ User selects stream → StreamSelectedMsg
    │   ├─→ Extract URL and stream name
    │   └─→ Return tea.Cmd
    │
    ├─→ App.Update(StreamSelectedMsg):
    │   ├─→ startPlayback(url, name) → tea.Cmd
    │   └─→ Returns cmd that will execute in loop
    │
    ├─→ startPlayback() cmd executes:
    │   │
    │   ├─→ Cancel previous playerStop context
    │   ├─→ Create new context for current playback
    │   ├─→ player.Play(ctx, url, title)
    │   │   ├─→ Try mpv (full featured)
    │   │   └─→ Fallback to VLC if mpv fails
    │   │
    │   └─→ Return PlayerStartedMsg or ErrorMsg
    │
    ├─→ Player runs in background (its own process)
    │   └─→ TUI remains responsive during playback
    │
    ├─→ player.Stop() context cancel closes player
    │   └─→ Exit callback → PlayerStoppedMsg
    │
    └─→ User presses 'q' on StreamsScreen
        ├─→ App detects key
        ├─→ Call app.StopPlayback()
        └─→ PlayerStoppedMsg
```

**Key Patterns:**
- Navigation chain: ContentType → Categories → Streams → Player
- Each screen receives client + type info via setters
- Async loading commands prevent UI blocking
- Player isolated in separate process; TUI stays responsive

### Navigation Stack Management

```go
// Stack-based navigation (app.go)
app.navStack []Screen  // History of screens visited

// Forward navigation (navigateTo)
navigateTo(screen) {
    navStack.append(current)  // Save current to stack
    current = screen          // Switch to new screen
    // Initialize new screen with data
}

// Backward navigation (navigateBack)
navigateBack() {
    current = navStack.pop()  // Restore previous
    // Keep state intact
}

// Global key handling (app.go:125-135)
Ctrl+C → Quit           // Always available
Esc → navigateBack()    // Except on LoginScreen
Q → Quit               // Only on Login/ContentType

// State preserved across navigation
app.client              // Persists through all screens
app.userInfo            // Persists (username, account info)
app.currentType         // Persists (Live/VOD/Series selection)
app.currentCat          // Persists (category context)
```

### Async Pattern with tea.Cmd

Tea.Cmd is a function that returns a message (blocking operation → non-blocking):

```go
type Cmd func() Msg

// Example: LoadCategories cmd
func (m *CategoriesModel) loadCategories() tea.Cmd {
    return func() tea.Msg {
        // This runs in background thread
        categories, err := client.GetCategories(ctx)
        if err != nil {
            return tui.ErrorMsg{Err: err}
        }
        return tui.CategoriesLoadedMsg{Categories: categories}
    }
}

// Usage in Update
return a.navigateTo(CategoriesScreen)
// Returns (app, cmd)
// Cmd executes in event loop → produces message → Update again

// Result
App.Update(msg)
    ├─→ Cmd returns → executes async
    │   └─→ Blocking API call
    │
    ├─→ Result becomes message
    │   └─→ CategoriesLoadedMsg
    │
    └─→ App.Update(CategoriesLoadedMsg)
        └─→ Render loaded data
```

**Command Batching:**

```go
// Single command
return a, cmd

// Multiple parallel commands
return a, tea.Batch(cmd1, cmd2, cmd3)

// Example: Navigation + Spinner
if a.loading {
    return a, tea.Batch(cmd, a.spinnerTick())
}
return a, cmd
```

### Message Flow Architecture

All messages route through central `App.Update()`:

```
Bubble Tea Event Loop
    │
    ├─→ Receive: tea.KeyMsg, tea.WindowSizeMsg, Custom messages
    │
    ├─→ App.Update(msg)
    │   ├─→ Handle global keys (Ctrl+C, Esc, Q)
    │   │
    │   ├─→ Handle navigation msgs (AuthSuccessMsg, etc)
    │   │
    │   ├─→ Update app state (loading, errorMsg, currentType, etc)
    │   │
    │   └─→ Forward to screen: updateScreen(msg)
    │       └─→ Screen.Update(msg) → returns tea.Cmd
    │
    ├─→ Execute returned Cmd
    │   ├─→ Async operation (API call, file I/O)
    │   └─→ Result becomes new message
    │
    └─→ Loop back to Update

// Screen message types (messages.go)
AuthSuccessMsg           // Login success → navigation trigger
AuthErrorMsg             // Login fail → error display
ContentTypeSelectedMsg   // Navigation: ContentType → Categories
CategorySelectedMsg      // Navigation: Categories → Streams
StreamSelectedMsg        // Navigation: Streams → Player
CategoriesLoadedMsg      // Data arrival
StreamsLoadedMsg         // Data arrival
PlayerStartedMsg         // Playback event
PlayerStoppedMsg         // Playback end
ErrorMsg                 // Generic error
LoadingMsg               // Loading state
SpinnerTickMsg           // Animation tick
```

### Error Handling Integration

Errors propagate through message chain:

```
API Error (network, auth, etc)
    │
    ├─→ Wrapped: fmt.Errorf("operation: %w", err)
    │
    ├─→ Cmd returns ErrorMsg{Err: err}
    │   └─→ App.Update(ErrorMsg)
    │       ├─→ Set loading = false
    │       ├─→ Set errorMsg string
    │       └─→ Render error overlay (app.go:349)
    │
    ├─→ Error persists until:
    │   ├─→ User presses Esc (clears overlay)
    │   ├─→ Navigation (clears error)
    │   └─→ ClearErrorMsg
    │
    └─→ View renders error bar at bottom
        └─→ renderWithError(content) overlays red bar
```

---

**Document Version:** 1.3
**Last Updated:** 2025-12-14
**Phases Complete:** Phases 1-7 (Project Setup, Data Layer, Presentation, Playback, Integration, Polish, Download Queue)
**Status:** Four-layer architecture with auth, navigation, playback, and download queue features
**Latest Feature:** Download manager with single-threaded queue, progress tracking, and UI integration
