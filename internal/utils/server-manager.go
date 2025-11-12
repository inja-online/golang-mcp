package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/inja-online/golang-mcp/internal/config"
)

// ServerInfo holds information about a running server
type ServerInfo struct {
	ID         string
	Name       string
	PID        int
	Command    string
	Args       []string
	WorkingDir string
	StartTime  time.Time
	Process    *exec.Cmd
	Logs       *RingBuffer
	StdoutLogs *RingBuffer
	StderrLogs *RingBuffer
	ExitCode   *int
	Status     string // "running", "stopped", "error"
	Metadata   map[string]interface{}
	ctx        context.Context
	cancel     context.CancelFunc
	logMutex   sync.RWMutex
}

type safeBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func newSafeBuffer() *safeBuffer {
	return &safeBuffer{}
}

func (sb *safeBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *safeBuffer) ReadAndReset() []byte {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	data := sb.buf.Bytes()
	result := make([]byte, len(data))
	copy(result, data)
	sb.buf.Reset()
	return result
}

func (sb *safeBuffer) Len() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Len()
}

type safeWriter struct {
	buf *safeBuffer
}

func (sw *safeWriter) Write(p []byte) (n int, err error) {
	sw.buf.mu.Lock()
	n, err = sw.buf.buf.Write(p)
	sw.buf.mu.Unlock()
	return n, err
}

// RingBuffer is a circular buffer for storing strings.
type RingBuffer struct {
	buffer []string
	size   int
	head   int
	count  int
	mu     sync.RWMutex
}

// NewRingBuffer creates a new ring buffer with the specified size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buffer: make([]string, size),
		size:   size,
	}
}

// Add adds a line to the ring buffer
func (rb *RingBuffer) Add(line string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.head] = line
	rb.head = (rb.head + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	}
}

// GetAll returns all log lines in order
func (rb *RingBuffer) GetAll() []string {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make([]string, rb.count)
	for i := 0; i < rb.count; i++ {
		idx := (rb.head - rb.count + i + rb.size) % rb.size
		result[i] = rb.buffer[idx]
	}
	return result
}

// GetRecent returns the most recent n lines
func (rb *RingBuffer) GetRecent(n int) []string {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if n > rb.count {
		n = rb.count
	}

	result := make([]string, n)
	for i := 0; i < n; i++ {
		idx := (rb.head - n + i + rb.size) % rb.size
		result[i] = rb.buffer[idx]
	}
	return result
}

// ServerManager manages multiple MCP servers.
type ServerManager struct {
	servers sync.Map
}

// NewServerManager creates a new server manager
func NewServerManager() *ServerManager {
	return &ServerManager{}
}

// StartServer starts a Go server in the background
func (sm *ServerManager) StartServer(ctx context.Context, cfg *config.Config, id, name, command string, args []string, workingDir string, envVars map[string]string, logSize int) (*ServerInfo, error) {
	if logSize <= 0 {
		logSize = 1000 // Default log size
	}

	// Create context for this server
	serverCtx, cancel := context.WithCancel(ctx)

	// Create command
	cmd := exec.CommandContext(serverCtx, command, args...)

	// Set working directory
	if workingDir != "" {
		cmd.Dir = workingDir
	} else {
		cmd.Dir = cfg.WorkingDirectory
	}

	// Set up environment
	env := os.Environ()
	goEnv := cfg.GetGoEnv()
	for k, v := range goEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	stdoutBuf := newSafeBuffer()
	stderrBuf := newSafeBuffer()
	cmd.Stdout = &safeWriter{buf: stdoutBuf}
	cmd.Stderr = &safeWriter{buf: stderrBuf}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	// Create server info
	serverInfo := &ServerInfo{
		ID:         id,
		Name:       name,
		PID:        cmd.Process.Pid,
		Command:    command,
		Args:       args,
		WorkingDir: workingDir,
		StartTime:  time.Now(),
		Process:    cmd,
		Logs:       NewRingBuffer(logSize),
		StdoutLogs: NewRingBuffer(logSize),
		StderrLogs: NewRingBuffer(logSize),
		Status:     "running",
		Metadata:   make(map[string]interface{}),
		ctx:        serverCtx,
		cancel:     cancel,
	}

	// Store server
	sm.servers.Store(id, serverInfo)

	// Start log collection goroutine
	go sm.collectLogs(serverInfo, stdoutBuf, stderrBuf)

	// Start monitoring goroutine
	go sm.monitorServer(serverInfo)

	return serverInfo, nil
}

func (sm *ServerManager) collectLogs(serverInfo *ServerInfo, stdoutBuf, stderrBuf *safeBuffer) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-serverInfo.ctx.Done():
			return
		case <-ticker.C:
			if stdoutBuf.Len() > 0 {
				data := stdoutBuf.ReadAndReset()
				lines := bytes.Split(data, []byte("\n"))
				for _, line := range lines {
					if len(line) > 0 {
						lineStr := string(line)
						serverInfo.Logs.Add(lineStr)
						serverInfo.StdoutLogs.Add(lineStr)
					}
				}
			}

			if stderrBuf.Len() > 0 {
				data := stderrBuf.ReadAndReset()
				lines := bytes.Split(data, []byte("\n"))
				for _, line := range lines {
					if len(line) > 0 {
						lineStr := string(line)
						serverInfo.Logs.Add(lineStr)
						serverInfo.StderrLogs.Add(lineStr)
					}
				}
			}
		}
	}
}

// monitorServer monitors the server process
func (sm *ServerManager) monitorServer(serverInfo *ServerInfo) {
	err := serverInfo.Process.Wait()

	serverInfo.logMutex.Lock()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			serverInfo.ExitCode = &exitCode
			serverInfo.Status = "stopped"
		} else {
			serverInfo.Status = "error"
		}
	} else {
		exitCode := 0
		serverInfo.ExitCode = &exitCode
		serverInfo.Status = "stopped"
	}
	serverInfo.logMutex.Unlock()
}

// StopServer stops a running server by ID.
func (sm *ServerManager) StopServer(id string, force bool) error {
	value, ok := sm.servers.Load(id)
	if !ok {
		return fmt.Errorf("server not found: %s", id)
	}

	serverInfo := value.(*ServerInfo)

	serverInfo.logMutex.RLock()
	status := serverInfo.Status
	serverInfo.logMutex.RUnlock()

	if status != "running" {
		return fmt.Errorf("server is not running: %s", id)
	}

	if force {
		if err := serverInfo.Process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill server: %w", err)
		}
	} else {
		serverInfo.cancel()
		if err := serverInfo.Process.Process.Signal(syscall.SIGTERM); err != nil {
			return serverInfo.Process.Process.Kill()
		}
	}

	return nil
}

// ListServers returns a list of all servers
func (sm *ServerManager) ListServers() []*ServerInfo {
	var servers []*ServerInfo
	sm.servers.Range(func(key, value interface{}) bool {
		serverInfo := value.(*ServerInfo)
		servers = append(servers, serverInfo)
		return true
	})
	return servers
}

// GetServer returns a server by ID
func (sm *ServerManager) GetServer(id string) (*ServerInfo, error) {
	value, ok := sm.servers.Load(id)
	if !ok {
		return nil, fmt.Errorf("server not found: %s", id)
	}
	return value.(*ServerInfo), nil
}

// GetServerLogs returns logs from a server
func (sm *ServerManager) GetServerLogs(id string, recent int) ([]string, error) {
	serverInfo, err := sm.GetServer(id)
	if err != nil {
		return nil, err
	}

	serverInfo.logMutex.RLock()
	defer serverInfo.logMutex.RUnlock()

	if recent > 0 {
		return serverInfo.Logs.GetRecent(recent), nil
	}
	return serverInfo.Logs.GetAll(), nil
}
