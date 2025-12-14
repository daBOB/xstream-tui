# Playback Layer Quick Reference

**Package:** `internal/player/`
**Phase:** Phase 4
**Status:** Implementation Complete

## Architecture Overview

```
App (TUI Layer)
    │
    ├─→ Manager (Lifecycle & Interface)
    │   ├─→ MPVPlayer (Primary)
    │   │   └─→ IPCClient (JSON-RPC)
    │   │       └─→ Socket (Platform-specific)
    │   │
    │   └─→ VLCPlayer (Fallback)
```

## Key Components

### 1. Manager (`manager.go`)
**Responsibility:** Unified player interface and lifecycle management

**Public Interface:**
```go
type Manager struct {
    // Manages: current player, type, exit callback, playing state
}

// Core methods
func (m *Manager) Play(ctx context.Context, url, title string) error
func (m *Manager) Stop() error
func (m *Manager) TogglePause() error
func (m *Manager) Seek(seconds int) error
func (m *Manager) SetVolume(vol int) error

// State queries
func (m *Manager) IsPlaying() bool
func (m *Manager) Type() PlayerType // "mpv" or "vlc"

// Callback registration
func (m *Manager) OnExit(fn func(error))
```

**Key Pattern:** Thread-safe with `sync.Mutex` for all operations

### 2. mpv IPC (`mpv.go` + `ipc.go`)
**Responsibility:** Full-featured playback control via JSON-RPC socket

**MPVPlayer Methods:**
```go
// Playback control
func (p *MPVPlayer) Pause() error
func (p *MPVPlayer) Resume() error
func (p *MPVPlayer) TogglePause() error
func (p *MPVPlayer) Seek(seconds int) error
func (p *MPVPlayer) SetVolume(vol int) error

// State queries (IPC-based, network overhead)
func (p *MPVPlayer) GetPosition() (float64, error)
func (p *MPVPlayer) GetDuration() (float64, error)
func (p *MPVPlayer) IsPaused() (bool, error)

// Lifecycle
func (p *MPVPlayer) Wait() error  // Blocks until exit
func (p *MPVPlayer) Stop() error  // Graceful shutdown
```

**IPC Protocol:**
- Connection: 5s retry (100ms intervals) for socket creation
- Request: JSON object with "command" array and "request_id"
- Response: Async via goroutine, matched by request_id
- Timeout: 5s per command with cleanup

**Example Command Flow:**
```
Client sends:   { "command": ["seek", 10, "relative"], "request_id": 1 }
mpv processes:  Move position forward 10 seconds
mpv responds:   { "data": null, "error": "success", "request_id": 1 }
Client receives: Response via channel
```

### 3. Socket Management (`socket.go`)
**Responsibility:** Platform-specific IPC socket path handling

**Cross-Platform Support:**
- **Linux/macOS:** Unix socket in XDG_RUNTIME_DIR or /tmp
  - Example: `/run/user/1000/xstream-mpv-12345.sock`
- **Windows:** Named pipe (experimental)
  - Example: `\\.\pipe\xstream-mpv-12345`

**Socket Naming:** Uses process PID for isolation (no collisions)

### 4. VLC Fallback (`vlc.go`)
**Responsibility:** Lightweight fallback when mpv unavailable

**Limitations:**
- No IPC communication (subprocess only)
- Control methods return `nil` (no-ops)
- Play-and-exit mode (no resume capability)
- Platform-specific executable paths

**Platform Paths:**
- macOS: `/Applications/VLC.app/Contents/MacOS/VLC`
- Windows: `C:\Program Files\VideoLAN\VLC\vlc.exe`
- Linux: `vlc` (PATH lookup)

## Integration with TUI

### Message Flow
```
User selects stream (StreamSelectedMsg)
    │
    ├─→ App.Update() extracts URL/title
    │
    └─→ playerManager.Play(ctx, url, title)
        │
        ├─→ Success: onExit callback registered
        │           Manager returns nil
        │
        └─→ Failure: Manager.Play() returns error
                     App displays error to user
```

### Exit Callback
```go
// Register callback when creating player
manager.OnExit(func(err error) {
    if err != nil {
        // Player exited with error
        app.queue(PlayerStoppedMsg{Err: err})
    } else {
        // Player exited normally
        app.queue(PlayerStoppedMsg{Err: nil})
    }
})
```

**Thread Safety:** Callback invoked from monitor goroutine, safe to queue messages

## Error Handling

### Player Start Failures
**Scenario:** mpv not found or fails to start
**Flow:**
1. Manager.Play() tries `NewMPVPlayer()`
2. If error: falls back to `NewVLCPlayer()`
3. If both fail: returns error with both reasons
4. User sees: "No player available: mpv (...), vlc (...)"

### IPC Socket Timeout
**Scenario:** mpv starts but socket not created in time
**Flow:**
1. NewMPVPlayer() spawns mpv, tries to connect
2. Connection retries for 5 seconds (100ms intervals)
3. If no connection: falls back to VLC
4. User sees: "Player error, falling back to VLC"

### Command Timeout
**Scenario:** IPC command unresponded (5s timeout)
**Flow:**
1. IPCClient.Command() sends request
2. Waits 5 seconds for response
3. If timeout: returns error
4. App ignores error (non-critical, playback continues)
5. User may miss state display but sees playback

### Player Exit
**Scenario:** mpv/VLC process terminates
**Flow:**
1. Monitor goroutine detects exit via Wait()
2. Calls onExit callback with exit error (if any)
3. Callback sends PlayerStoppedMsg to TUI
4. App.Update() processes message
5. Returns to previous screen

## Configuration

No configuration needed. Manager auto-selects player:
1. Tries mpv first (full control)
2. Falls back to VLC if mpv unavailable
3. Errors if both unavailable

**Environment Variables:**
- `XDG_RUNTIME_DIR` - Controls socket location on Unix

## Testing Considerations

### Unit Testing
- Mock Player interface for Manager tests
- Mock IPCClient for mpv tests
- Test socket path generation per platform

### Integration Testing
- Real mpv with test URL (local file)
- VLC fallback when mpv disabled
- Exit monitoring with long/short videos

### Manual Testing
- Check socket cleanup on exit
- Verify IPC timeout doesn't hang
- Test fallback behavior
- Verify Windows named pipe (if testing on Windows)

## Performance Notes

### IPC Overhead
- Each property query = network round-trip (5s timeout)
- Use sparingly in tight loops
- Cache values when possible

### Socket Cleanup
- Both implementations clean up sockets
- PID-based naming prevents accumulation
- Automatic cleanup via cleanupSocket()

## Known Limitations

1. **VLC Control:** Limited to play/pause by design (no seek/volume via IPC)
2. **Windows:** Named pipe support experimental, untested
3. **IPC Overhead:** Property queries add latency (5s timeout)
4. **Timeout Fixed:** All timeouts hardcoded (5s socket, 5s command)

## Future Enhancements

- [ ] Configurable player preference (mpv/vlc toggle)
- [ ] Timeout configuration
- [ ] VLC socket-based IPC (if needed)
- [ ] Player health monitoring
- [ ] Playback metrics collection

---

**Last Updated:** 2025-12-14
**Reviewed Against:** Phase 4 Implementation
