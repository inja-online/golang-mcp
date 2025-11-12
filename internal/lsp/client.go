package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Client provides a lightweight JSON-RPC over LSP transport client.
// This file implements request/notify semantics and pending-request routing.
// TODO: add metrics, cancellation propagation to Send, and graceful shutdown.
type Client struct {
	rw        io.ReadWriteCloser
	opts      ClientOptions
	transport *transport

	mu       sync.Mutex // protects handlers
	handlers map[string]func(params json.RawMessage)

	pendingMu sync.Mutex
	pending   map[string]chan *rpcMessage // keyed by id as JSON bytes (e.g. "1")

	nextID int64
	closed int32
}

// ClientOptions configures client behavior.
type ClientOptions struct {
	RequestTimeout time.Duration
	Logger         *log.Logger
}

// NewClient creates a new Client. rw must be a ReadWriteCloser connected to the LSP process.
func NewClient(rw io.ReadWriteCloser, opts ClientOptions) *Client {
	if opts.RequestTimeout == 0 {
		opts.RequestTimeout = 10 * time.Second
	}
	return &Client{
		rw:       rw,
		opts:     opts,
		handlers: make(map[string]func(params json.RawMessage)),
		pending:  make(map[string]chan *rpcMessage),
	}
}

// Start initializes transport and launches the receive loop.
func (c *Client) Start(ctx context.Context) error {
	if atomic.LoadInt32(&c.closed) == 1 {
		return fmt.Errorf("client already closed")
	}
	c.transport = newTransport(c.rw)
	// start receive loop
	go c.receiveLoop()
	return nil
}

// Shutdown attempts to mark client closed and close underlying rw.
func (c *Client) Shutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return nil
	}
	// Best-effort close of underlying connection.
	if c.rw != nil {
		_ = c.rw.Close()
	}
	// notify pending requests of shutdown
	c.pendingMu.Lock()
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	c.pendingMu.Unlock()
	return nil
}

// Request sends a JSON-RPC request and waits for a response. result may be nil.
// Uses context for cancellation; falls back to ClientOptions.RequestTimeout when no deadline is set.
func (c *Client) Request(ctx context.Context, method string, params interface{}, result interface{}) error {
	if atomic.LoadInt32(&c.closed) == 1 {
		return fmt.Errorf("client is closed")
	}
	// ensure transport exists
	if c.transport == nil {
		return fmt.Errorf("transport not started")
	}

	// prepare id
	id := atomic.AddInt64(&c.nextID, 1)
	idStr := strconv.FormatInt(id, 10)
	rawID := json.RawMessage([]byte(idStr))

	// prepare params
	var paramsRaw *json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		tmp := json.RawMessage(b)
		paramsRaw = &tmp
	}

	msg := &rpcMessage{
		JSONRPC: "2.0",
		ID:      &rawID,
		Method:  method,
		Params:  paramsRaw,
	}

	// create pending channel
	respCh := make(chan *rpcMessage, 1)
	c.pendingMu.Lock()
	c.pending[idStr] = respCh
	c.pendingMu.Unlock()

	// send
	if err := c.transport.Send(ctx, msg); err != nil {
		// cleanup pending
		c.pendingMu.Lock()
		delete(c.pending, idStr)
		c.pendingMu.Unlock()
		return fmt.Errorf("send request: %w", err)
	}

	// prepare wait context
	waitCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, c.opts.RequestTimeout)
		defer cancel()
	}

	select {
	case <-waitCtx.Done():
		// remove pending
		c.pendingMu.Lock()
		delete(c.pending, idStr)
		c.pendingMu.Unlock()
		return waitCtx.Err()
	case resp, ok := <-respCh:
		if !ok || resp == nil {
			return fmt.Errorf("response channel closed")
		}
		if resp.Error != nil {
			return fmt.Errorf("rpc error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
		}
		if result != nil && resp.Result != nil {
			if err := json.Unmarshal(*resp.Result, result); err != nil {
				return fmt.Errorf("unmarshal result: %w", err)
			}
		}
		return nil
	}
}

// Notify sends a JSON-RPC notification (no response expected).
func (c *Client) Notify(ctx context.Context, method string, params interface{}) error {
	if atomic.LoadInt32(&c.closed) == 1 {
		return fmt.Errorf("client is closed")
	}
	if c.transport == nil {
		return fmt.Errorf("transport not started")
	}

	var paramsRaw *json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		tmp := json.RawMessage(b)
		paramsRaw = &tmp
	}

	msg := &rpcMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsRaw,
	}
	if err := c.transport.Send(ctx, msg); err != nil {
		return fmt.Errorf("send notify: %w", err)
	}
	return nil
}

// RegisterNotificationHandler registers a handler for server notifications (e.g., textDocument/publishDiagnostics).
func (c *Client) RegisterNotificationHandler(method string, handler func(params json.RawMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[method] = handler
}

// receiveLoop continuously reads messages from transport and routes them.
func (c *Client) receiveLoop() {
	for {
		if atomic.LoadInt32(&c.closed) == 1 {
			return
		}
		msg, err := c.transport.Read()
		if err != nil {
			// log and exit loop; transport.Read will return errors when connection closed.
			if c.opts.Logger != nil {
				c.opts.Logger.Printf("lsp transport read error: %v", err)
			}
			return
		}
		// responses (have ID)
		if msg.ID != nil {
			idKey := string(*msg.ID)
			c.pendingMu.Lock()
			ch, ok := c.pending[idKey]
			if ok {
				// deliver and remove pending
				select {
				case ch <- msg:
				default:
				}
				close(ch)
				delete(c.pending, idKey)
			}
			c.pendingMu.Unlock()
			continue
		}
		// notifications
		if msg.Method != "" {
			c.mu.Lock()
			h := c.handlers[msg.Method]
			c.mu.Unlock()
			if h != nil && msg.Params != nil {
				// invoke handler asynchronously to avoid blocking receive loop
				go func(handler func(params json.RawMessage), p json.RawMessage) {
					defer func() {
						// recover to avoid crashing receive loop
						_ = recover()
					}()
					handler(p)
				}(h, *msg.Params)
			}
		}
	}
}
