package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Response represents an mpv IPC response.
type Response struct {
	Data      any    `json:"data"`
	Error     string `json:"error"`
	RequestID int    `json:"request_id"`
}

// IPCClient handles JSON-based IPC communication with mpv.
type IPCClient struct {
	conn      net.Conn
	requestID int
	pending   map[int]chan Response
	mu        sync.Mutex
	reader    *bufio.Reader
	closed    bool
}

// NewIPCClient connects to mpv IPC socket with retry logic.
// mpv needs time to create the socket after startup.
func NewIPCClient(socketPath string) (*IPCClient, error) {
	var conn net.Conn
	var err error

	// Retry connection up to 5 seconds (50 * 100ms)
	for i := 0; i < 50; i++ {
		conn, err = dialSocket(socketPath)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("connect to mpv socket: %w", err)
	}

	c := &IPCClient{
		conn:    conn,
		pending: make(map[int]chan Response),
		reader:  bufio.NewReader(conn),
	}
	go c.readLoop()
	return c, nil
}

// readLoop continuously reads responses from mpv.
func (c *IPCClient) readLoop() {
	for {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			return // Socket closed
		}

		var resp Response
		if json.Unmarshal(line, &resp) == nil && resp.RequestID > 0 {
			c.mu.Lock()
			if ch, ok := c.pending[resp.RequestID]; ok {
				ch <- resp
				delete(c.pending, resp.RequestID)
			}
			c.mu.Unlock()
		}
	}
}

// Command sends a command to mpv and waits for response.
func (c *IPCClient) Command(args ...any) (Response, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return Response{}, fmt.Errorf("ipc client closed")
	}

	c.requestID++
	id := c.requestID
	ch := make(chan Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	cmd := map[string]any{
		"command":    args,
		"request_id": id,
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return Response{}, fmt.Errorf("marshal command: %w", err)
	}
	data = append(data, '\n')

	if _, err := c.conn.Write(data); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return Response{}, fmt.Errorf("write command: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != "" && resp.Error != "success" {
			return resp, fmt.Errorf("mpv error: %s", resp.Error)
		}
		return resp, nil
	case <-time.After(5 * time.Second):
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return Response{}, fmt.Errorf("timeout waiting for mpv response")
	}
}

// Close closes the IPC connection.
func (c *IPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	// Close all pending channels
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}

	return c.conn.Close()
}
