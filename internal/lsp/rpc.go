package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// rpcMessage represents a generic JSON-RPC 2.0 message used by LSP.
type rpcMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  *json.RawMessage `json:"params,omitempty"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *rpcError        `json:"error,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// transport implements simple Content-Length framed JSON-RPC over an
// io.ReadWriteCloser. It is intentionally minimal for sprint 2.
// TODO: add deadlines/cancellation hooks and metrics.
type transport struct {
	rw     io.ReadWriteCloser
	mu     sync.Mutex // protects writes to rw
	reader *bufio.Reader
	rmu    sync.Mutex // protects reader initialization
}

// newTransport creates a transport wrapping rw.
func newTransport(rw io.ReadWriteCloser) *transport {
	return &transport{
		rw:     rw,
		reader: bufio.NewReader(rw),
	}
}

// Send serializes the rpcMessage and writes it with a Content-Length header.
// This is safe for concurrent callers: writes are serialized by t.mu.
func (t *transport) Send(ctx context.Context, msg *rpcMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal rpc message: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))

	t.mu.Lock()
	defer t.mu.Unlock()

	// Write header then body. Using simple writes is sufficient for now.
	if _, err := t.rw.Write([]byte(header)); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if _, err := t.rw.Write(body); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	return nil
}

// Read reads the next framed JSON-RPC message from the transport and
// returns the decoded rpcMessage.
func (t *transport) Read() (*rpcMessage, error) {
	// Use the buffered reader for efficient reading of headers and body.
	t.rmu.Lock()
	r := t.reader
	if r == nil {
		// Shouldn't happen if constructed with newTransport, but guard anyway.
		t.reader = bufio.NewReader(t.rw)
		r = t.reader
	}
	t.rmu.Unlock()

	// Read headers until empty line.
	var contentLen int
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header line: %w", err)
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			// blank line => end of headers
			break
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			// ignore malformed header lines for robustness
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if strings.EqualFold(key, "Content-Length") {
			n, err := strconv.Atoi(val)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid Content-Length: %q", val)
			}
			contentLen = n
		}
	}

	if contentLen <= 0 {
		return nil, fmt.Errorf("missing or invalid Content-Length")
	}

	buf := make([]byte, contentLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var msg rpcMessage
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal rpc message: %w", err)
	}
	return &msg, nil
}
