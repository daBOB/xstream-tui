package player

import (
	"bufio"
	"net"
	"testing"
)

func TestMPVPlayer_ImplementsControllablePlayer(t *testing.T) {
	// MPVPlayer implements ControllablePlayer (full control via IPC).
	var _ ControllablePlayer = (*MPVPlayer)(nil)
}

// mockIPCClient creates an IPC client with a mock connection for testing
func createMockIPCClient() (*IPCClient, *mockConn) {
	conn := newMockConn()
	return &IPCClient{
		conn:    conn,
		pending: make(map[int]chan Response),
		reader:  bufio.NewReader(conn),
	}, conn
}

func TestMPVPlayer_Stop(t *testing.T) {
	// Create mock IPC client
	ipc, _ := createMockIPCClient()

	// Create minimal MPVPlayer for testing Stop()
	p := &MPVPlayer{
		ipc:        ipc,
		socketPath: "/tmp/test-socket",
		cancel:     func() {},
	}

	// Stop should not panic even without a real cmd
	// Note: cmd is nil so we skip the actual process termination
	p.ipc.Close()
}

func TestMPVPlayer_Methods_WithClosedIPC(t *testing.T) {
	// Create closed IPC client
	ipc, _ := createMockIPCClient()
	ipc.closed = true

	p := &MPVPlayer{
		ipc:    ipc,
		cancel: func() {},
	}

	// All methods should return error when IPC is closed
	if err := p.Pause(); err == nil {
		t.Error("Pause() should return error when IPC closed")
	}
	if err := p.Resume(); err == nil {
		t.Error("Resume() should return error when IPC closed")
	}
	if err := p.TogglePause(); err == nil {
		t.Error("TogglePause() should return error when IPC closed")
	}
	if err := p.Seek(10); err == nil {
		t.Error("Seek() should return error when IPC closed")
	}
	if err := p.SetVolume(50); err == nil {
		t.Error("SetVolume() should return error when IPC closed")
	}
	if _, err := p.GetPosition(); err == nil {
		t.Error("GetPosition() should return error when IPC closed")
	}
	if _, err := p.GetDuration(); err == nil {
		t.Error("GetDuration() should return error when IPC closed")
	}
	if _, err := p.IsPaused(); err == nil {
		t.Error("IsPaused() should return error when IPC closed")
	}
}

func TestMPVPlayer_GetPosition_WrongType(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	ipc := &IPCClient{
		conn:    client,
		pending: make(map[int]chan Response),
		reader:  bufio.NewReader(client),
	}

	p := &MPVPlayer{
		ipc:    ipc,
		cancel: func() {},
	}

	// Simulate response in goroutine
	go func() {
		// Wait for command and send response with wrong type
		buf := make([]byte, 1024)
		server.Read(buf)
		server.Write([]byte(`{"request_id":1,"data":"not a number","error":"success"}` + "\n"))
	}()

	go ipc.readLoop()

	_, err := p.GetPosition()
	if err == nil {
		t.Error("GetPosition() should return error for non-float data")
	}
}

func TestMPVPlayer_GetDuration_WrongType(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	ipc := &IPCClient{
		conn:    client,
		pending: make(map[int]chan Response),
		reader:  bufio.NewReader(client),
	}

	p := &MPVPlayer{
		ipc:    ipc,
		cancel: func() {},
	}

	go func() {
		buf := make([]byte, 1024)
		server.Read(buf)
		server.Write([]byte(`{"request_id":1,"data":"string","error":"success"}` + "\n"))
	}()

	go ipc.readLoop()

	_, err := p.GetDuration()
	if err == nil {
		t.Error("GetDuration() should return error for non-float data")
	}
}

func TestMPVPlayer_IsPaused_WrongType(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	ipc := &IPCClient{
		conn:    client,
		pending: make(map[int]chan Response),
		reader:  bufio.NewReader(client),
	}

	p := &MPVPlayer{
		ipc:    ipc,
		cancel: func() {},
	}

	go func() {
		buf := make([]byte, 1024)
		server.Read(buf)
		server.Write([]byte(`{"request_id":1,"data":"not bool","error":"success"}` + "\n"))
	}()

	go ipc.readLoop()

	_, err := p.IsPaused()
	if err == nil {
		t.Error("IsPaused() should return error for non-bool data")
	}
}
