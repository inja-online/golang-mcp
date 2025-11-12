package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// rwPipe composes a ReadCloser and WriteCloser into a single
// io.ReadWriteCloser. Tests create two pipes and cross-wire them
// so that writes on one endpoint are readable by the other.
type rwPipe struct {
	r io.ReadCloser
	w io.WriteCloser
}

func (p *rwPipe) Read(b []byte) (int, error)  { return p.r.Read(b) }
func (p *rwPipe) Write(b []byte) (int, error) { return p.w.Write(b) }
func (p *rwPipe) Close() error {
	if err := p.r.Close(); err != nil {
		_ = p.w.Close()
		return err
	}
	return p.w.Close()
}

// makeDuplex returns two connected io.ReadWriteCloser endpoints (a, b).
// Writing to a writes into b's read side and vice-versa.
func makeDuplex() (a, b io.ReadWriteCloser) {
	pr1, pw1 := io.Pipe() // connects pw1 -> pr1
	pr2, pw2 := io.Pipe() // connects pw2 -> pr2

	// Endpoint a: reads what b wrote (pr2), writes to b (pw1)
	a = &rwPipe{r: pr2, w: pw1}
	// Endpoint b: reads what a wrote (pr1), writes to a (pw2)
	b = &rwPipe{r: pr1, w: pw2}
	return a, b
}

// helper to create a simple params raw JSON
func rawParams(s string) *json.RawMessage {
	r := json.RawMessage([]byte(s))
	return &r
}

// TestTransportSendAndRead_RoundTrip: connect two transports using a duplex pipe,
// send a simple rpcMessage from sender.Send and verify receiver.Read returns the
// same fields (Method/JSONRPC/Params).
func TestTransportSendAndRead_RoundTrip(t *testing.T) {
	a, b := makeDuplex()
	defer a.Close()
	defer b.Close()

	sender := newTransport(a)
	receiver := newTransport(b)

	msg := &rpcMessage{
		JSONRPC: "2.0",
		Method:  "example/echo",
		Params:  rawParams(`{"hello":"world","n":1}`),
	}

	// send in a goroutine to mimic asynchronous writer
	sendErrCh := make(chan error, 1)
	go func() {
		sendErrCh <- sender.Send(context.Background(), msg)
	}()

	recv, err := receiver.Read()
	if err != nil {
		t.Fatalf("receiver.Read error: %v", err)
	}

	if err := <-sendErrCh; err != nil {
		t.Fatalf("sender.Send error: %v", err)
	}

	if recv.Method != msg.Method {
		t.Fatalf("method mismatch: got %q want %q", recv.Method, msg.Method)
	}
	if recv.JSONRPC != msg.JSONRPC {
		t.Fatalf("jsonrpc mismatch: got %q want %q", recv.JSONRPC, msg.JSONRPC)
	}
	if recv.Params == nil {
		t.Fatalf("params missing")
	}
	if strings.TrimSpace(string(*recv.Params)) != strings.TrimSpace(string(*msg.Params)) {
		t.Fatalf("params mismatch: got %s want %s", string(*recv.Params), string(*msg.Params))
	}
}

// TestTransportSendConcurrent: ensure concurrent Send calls from multiple
// goroutines serialize correctly. Launch N goroutines each calling Send with
// distinct messages, then sequentially Read N messages on the receiver side and
// validate all messages arrived intact and without interleaving.
func TestTransportSendConcurrent(t *testing.T) {
	const N = 50

	a, b := makeDuplex()
	defer a.Close()
	defer b.Close()

	sender := newTransport(a)
	receiver := newTransport(b)

	var wg sync.WaitGroup
	wg.Add(N)

	// Prepare messages
	for i := 0; i < N; i++ {
		i := i
		go func() {
			defer wg.Done()
			msg := &rpcMessage{
				JSONRPC: "2.0",
				Method:  "concurrent/send",
				Params:  rawParams(fmt.Sprintf(`{"i":%d}`, i)),
			}
			if err := sender.Send(context.Background(), msg); err != nil {
				// We don't fail the test from goroutine; capture via t.Errorf
				t.Errorf("Send(%d) error: %v", i, err)
			}
		}()
	}

	// Do not wait for all senders to finish before reading.
	// If we wait here, a single blocked writer can deadlock the whole test
	// because the pipe's buffer is limited. Start reading concurrently and
	// then wait for goroutines to finish after reading.

	// Read N messages sequentially and collect i values we see.
	seen := make(map[int]bool)
	for j := 0; j < N; j++ {
		msg, err := receiver.Read()
		if err != nil {
			t.Fatalf("Read #%d error: %v", j, err)
		}
		if msg.Params == nil {
			t.Fatalf("message #%d missing params", j)
		}
		// parse params as map to get i
		var pm map[string]int
		if err := json.Unmarshal(*msg.Params, &pm); err != nil {
			t.Fatalf("unmarshal params #%d: %v", j, err)
		}
		i, ok := pm["i"]
		if !ok {
			// maybe json numbers decode as float if structure differs; attempt decode via interface
			var im map[string]interface{}
			if err := json.Unmarshal(*msg.Params, &im); err == nil {
				if v, ok := im["i"].(float64); ok {
					i = int(v)
				} else {
					t.Fatalf("params #%d missing 'i' or invalid type", j)
				}
			} else {
				t.Fatalf("params #%d missing 'i'", j)
			}
		}
		if seen[i] {
			t.Fatalf("duplicate message for i=%d", i)
		}
		seen[i] = true
	}

	// Wait for all senders to complete to ensure goroutines didn't encounter errors.
	wg.Wait()

	// ensure all messages 0..N-1 were observed
	for i := 0; i < N; i++ {
		if !seen[i] {
			t.Fatalf("missing message for i=%d", i)
		}
	}
}

// TestTransportReadMalformedContentLength: feed malformed header(s) into a
// transport reader and assert Read returns an error describing invalid
// Content-Length.
func TestTransportReadMalformedContentLength(t *testing.T) {
	a, b := makeDuplex()
	defer a.Close()
	defer b.Close()

	receiver := newTransport(a)

	// Write malformed header directly to the peer writer (b) from a goroutine
	// so the reader can start consuming without deadlocking the test.
	writeErrCh := make(chan error, 1)
	go func() {
		_, err := b.Write([]byte("Content-Length: abc\r\n\r\n{\"jsonrpc\":\"2.0\",\"method\":\"x\"}"))
		if cerr := b.Close(); err == nil {
			err = cerr
		}
		writeErrCh <- err
		close(writeErrCh)
	}()

	_, err := receiver.Read()
	if err == nil {
		t.Fatalf("expected error for malformed Content-Length, got nil")
	}
	if !strings.Contains(err.Error(), "invalid Content-Length") {
		t.Fatalf("unexpected error; want invalid Content-Length got: %v", err)
	}
}

// TestTransportReadTruncatedBody: write a correct Content-Length header but only
// write a partial body (less bytes than header indicates) and ensure Read
// returns an io.ErrUnexpectedEOF or an error wrapping read failure.
func TestTransportReadTruncatedBody(t *testing.T) {
	a, b := makeDuplex()
	defer a.Close()
	// will close writer below

	receiver := newTransport(a)

	// Prepare a body longer than what we will send
	body := `{"jsonrpc":"2.0","method":"truncated","params":{"x":1,"y":2}}`
	// Intentionally set Content-Length larger than actual bytes written.
	totalLen := len(body) + 10
	header := "Content-Length: " + strconv.Itoa(totalLen) + "\r\n\r\n"
	partial := body[:len(body)/2] // write only half

	// Write header+partial body then close writer to simulate truncation.
	// Do the write from a goroutine so receiver.Read can run concurrently and
	// avoid deadlocking when pipe buffers are full.
	writeErrCh := make(chan error, 1)
	go func() {
		_, err := b.Write([]byte(header + partial))
		if cerr := b.Close(); err == nil {
			err = cerr
		}
		writeErrCh <- err
		close(writeErrCh)
	}()

	_, err := receiver.Read()
	if err == nil {
		t.Fatalf("expected error for truncated body, got nil")
	}
	// Expect the underlying error to be io.ErrUnexpectedEOF (from io.ReadFull) or EOF
	if !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected error type for truncated body: %v", err)
	}
}
