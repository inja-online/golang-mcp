package lsp

import (
	"fmt"
	"strings"
)

// Position represents a position in a text document (zero-based).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a text range in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// TextDocumentIdentifier identifies a text document by URI.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// Diagnostic represents an LSP diagnostic message.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity,omitempty"`
	Code     string `json:"code,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

// PublishDiagnosticsParams is sent by the server to the client to publish diagnostics.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// FilePathToURI converts a local file path to a file:// URI.
// NOTE: This is a minimal implementation for stubbing purposes.
func FilePathToURI(path string) string {
	// TODO: handle platform-specific encoding, absolute paths, and escaping
	if strings.HasPrefix(path, "file://") {
		return path
	}
	return "file://" + path
}

// URIToFilePath converts a file:// URI back to a local file path.
func URIToFilePath(uri string) (string, error) {
	// TODO: robustly parse different URI forms and handle escaping
	if strings.HasPrefix(uri, "file://") {
		return strings.TrimPrefix(uri, "file://"), nil
	}
	return "", fmt.Errorf("unsupported URI scheme or invalid URI: %s", uri)
}
