package lsp

import (
	"context"
	"sync"
)

// Manager manages multiple LSP sessions keyed by workspace/root URI.
type Manager struct {
	mu       sync.Mutex
	sessions map[string]SessionHandle
	// TODO: add lifecycle management fields, restart counters, etc.
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]SessionHandle),
	}
}

type SessionOptions struct {
	GoplsPath   string
	Args        []string
	Env         map[string]string
	MaxRestarts int
	MaxMemoryMB int
}

type SessionHandle interface {
	RootURI() string
	Request(ctx context.Context, method string, params interface{}, result interface{}) error
	Notify(ctx context.Context, method string, params interface{}) error
	SubscribeDiagnostics(ctx context.Context, ch chan<- PublishDiagnosticsParams) (unsubscribe func(), err error)
	Shutdown(ctx context.Context) error
}

func (m *Manager) StartSession(ctx context.Context, rootURI string, opts SessionOptions) (SessionHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: start gopls process, wire transport and client, create session object
	if h, ok := m.sessions[rootURI]; ok {
		return h, nil
	}
	// placeholder: no-op session for now
	var sess SessionHandle = nil
	m.sessions[rootURI] = sess
	return sess, nil
}

func (m *Manager) GetSession(rootURI string) (SessionHandle, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.sessions[rootURI]
	return h, ok
}

func (m *Manager) ShutdownSession(ctx context.Context, rootURI string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: gracefully shutdown session and cleanup resources
	if h, ok := m.sessions[rootURI]; ok && h != nil {
		_ = h.Shutdown(ctx) // ignore error for stub
	}
	delete(m.sessions, rootURI)
	return nil
}
