package player

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"
)

// mockConn simulates a net.Conn for testing IPC
type mockConn struct {
	net.Conn
	readData  []byte
	readPos   int
	writeData []byte
	closed    bool
	mu        sync.Mutex
}

func newMockConn() *mockConn {
	return &mockConn{}
}

func (m *mockConn) Read(b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readPos >= len(m.readData) {
		// Simulate blocking read
		time.Sleep(10 * time.Millisecond)
		return 0, nil
	}

	n := copy(b, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockConn) Write(b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockConn) SetResponse(resp Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, _ := json.Marshal(resp)
	m.readData = append(data, '\n')
	m.readPos = 0
}

func TestIPCClient_Close(t *testing.T) {
	conn := newMockConn()
	client := &IPCClient{
		conn:    conn,
		pending: make(map[int]chan Response),
		reader:  bufio.NewReader(conn),
	}

	// Add pending request
	ch := make(chan Response, 1)
	client.pending[1] = ch

	// Close client
	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !client.closed {
		t.Error("client should be marked as closed")
	}

	// Verify pending channel was closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("pending channel should be closed")
		}
	default:
		t.Error("pending channel should be closed")
	}

	// Verify double close doesn't error
	err = client.Close()
	if err != nil {
		t.Errorf("second Close() error = %v", err)
	}
}

func TestIPCClient_Command_Closed(t *testing.T) {
	conn := newMockConn()
	client := &IPCClient{
		conn:    conn,
		pending: make(map[int]chan Response),
		reader:  bufio.NewReader(conn),
		closed:  true,
	}

	_, err := client.Command("get_property", "pause")
	if err == nil {
		t.Error("Command() should return error when client is closed")
	}
}

func TestResponse_Fields(t *testing.T) {
	resp := Response{
		Data:      "test",
		Error:     "success",
		RequestID: 1,
	}

	if resp.Data != "test" {
		t.Errorf("Data = %v, want 'test'", resp.Data)
	}
	if resp.Error != "success" {
		t.Errorf("Error = %v, want 'success'", resp.Error)
	}
	if resp.RequestID != 1 {
		t.Errorf("RequestID = %d, want 1", resp.RequestID)
	}
}

func TestIPCClient_ReadLoop_Cleanup(t *testing.T) {
	// Create a pipe to simulate connection
	server, client := net.Pipe()
	defer server.Close()

	ipc := &IPCClient{
		conn:    client,
		pending: make(map[int]chan Response),
		reader:  bufio.NewReader(client),
	}

	// Add pending request
	ch := make(chan Response, 1)
	ipc.pending[1] = ch

	// Start readLoop
	go ipc.readLoop()

	// Close server side to trigger error in readLoop
	server.Close()

	// Wait for cleanup
	time.Sleep(50 * time.Millisecond)

	ipc.mu.Lock()
	if !ipc.closed {
		t.Error("readLoop should mark client as closed on error")
	}
	if len(ipc.pending) != 0 {
		t.Errorf("pending should be empty after cleanup, got %d", len(ipc.pending))
	}
	ipc.mu.Unlock()
}
