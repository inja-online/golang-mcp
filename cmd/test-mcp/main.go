package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// JSONRPCRequest represents a JSON-RPC request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InitializeParams represents the parameters for MCP initialization.
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]string      `json:"clientInfo"`
}

// InitializeResult represents the result of MCP initialization.
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]string      `json:"serverInfo"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <path-to-mcp-go-binary>\n", os.Args[0])
		os.Exit(1)
	}

	binaryPath := os.Args[1]
	fmt.Printf("Testing MCP server at: %s\n\n", binaryPath)

	cmd := exec.Command(binaryPath)
	cmd.Env = append(os.Environ(), "DEBUG_MCP=true")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdin pipe: %v\n", err)
		os.Exit(1)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdout pipe: %v\n", err)
		os.Exit(1)
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stderr pipe: %v\n", err)
		os.Exit(1)
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[STDERR] %s\n", scanner.Text())
		}
	}()

	scanner := bufio.NewScanner(stdout)
	encoder := json.NewEncoder(stdin)

	fmt.Println("=== Test 1: Initialize ===")
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    make(map[string]interface{}),
			ClientInfo: map[string]string{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	if err := encoder.Encode(initReq); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding initialize request: %v\n", err)
		os.Exit(1)
	}
	if _, err := stdin.Write([]byte("\n")); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdin: %v\n", err)
		os.Exit(1)
	}

	var initResp JSONRPCResponse
	if !scanner.Scan() {
		fmt.Fprintf(os.Stderr, "No response to initialize\n")
		os.Exit(1)
	}

	if err := json.Unmarshal(scanner.Bytes(), &initResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing initialize response: %v\n", err)
		os.Exit(1)
	}

	if initResp.Error != nil {
		fmt.Printf("❌ Initialize failed: %s\n", initResp.Error.Message)
		os.Exit(1)
	}

	var initResult InitializeResult
	if err := json.Unmarshal(initResp.Result, &initResult); err != nil {
		fmt.Printf("❌ Error parsing initialize result: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Initialize successful\n")
	fmt.Printf("   Protocol Version: %s\n", initResult.ProtocolVersion)
	if serverInfo, ok := initResult.ServerInfo["name"]; ok {
		fmt.Printf("   Server Name: %s\n", serverInfo)
	}

	fmt.Println("\n=== Test 2: List Tools ===")
	toolsReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	if err := encoder.Encode(toolsReq); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding tools/list request: %v\n", err)
		os.Exit(1)
	}
	if _, err := stdin.Write([]byte("\n")); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdin: %v\n", err)
		os.Exit(1)
	}

	if !scanner.Scan() {
		fmt.Fprintf(os.Stderr, "No response to tools/list\n")
		os.Exit(1)
	}

	var toolsResp JSONRPCResponse
	if err := json.Unmarshal(scanner.Bytes(), &toolsResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing tools/list response: %v\n", err)
		os.Exit(1)
	}

	if toolsResp.Error != nil {
		fmt.Printf("❌ tools/list failed: %s\n", toolsResp.Error.Message)
		os.Exit(1)
	}

	var toolsResult map[string]interface{}
	if err := json.Unmarshal(toolsResp.Result, &toolsResult); err != nil {
		fmt.Printf("❌ Error parsing tools/list result: %v\n", err)
		os.Exit(1)
	}

	if tools, ok := toolsResult["tools"].([]interface{}); ok {
		fmt.Printf("✅ Found %d tools\n", len(tools))
		if len(tools) > 0 {
			fmt.Printf("   First tool: %v\n", tools[0])
		}
	} else {
		fmt.Printf("⚠️  Unexpected tools result format\n")
	}

	fmt.Println("\n=== Test 3: List Resources ===")
	resourcesReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "resources/list",
	}

	if err := encoder.Encode(resourcesReq); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding resources/list request: %v\n", err)
		os.Exit(1)
	}
	if _, err := stdin.Write([]byte("\n")); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdin: %v\n", err)
		os.Exit(1)
	}

	if !scanner.Scan() {
		fmt.Fprintf(os.Stderr, "No response to resources/list\n")
		os.Exit(1)
	}

	var resourcesResp JSONRPCResponse
	if err := json.Unmarshal(scanner.Bytes(), &resourcesResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing resources/list response: %v\n", err)
		os.Exit(1)
	}

	if resourcesResp.Error != nil {
		fmt.Printf("❌ resources/list failed: %s\n", resourcesResp.Error.Message)
		os.Exit(1)
	}

	var resourcesResult map[string]interface{}
	if err := json.Unmarshal(resourcesResp.Result, &resourcesResult); err != nil {
		fmt.Printf("❌ Error parsing resources/list result: %v\n", err)
		os.Exit(1)
	}

	if resources, ok := resourcesResult["resources"].([]interface{}); ok {
		fmt.Printf("✅ Found %d resources\n", len(resources))
		if len(resources) > 0 {
			fmt.Printf("   First resource: %v\n", resources[0])
		}
	} else {
		fmt.Printf("⚠️  Unexpected resources result format\n")
	}

	fmt.Println("\n=== Test 4: List Prompts ===")
	promptsReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "prompts/list",
	}

	if err := encoder.Encode(promptsReq); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding prompts/list request: %v\n", err)
		os.Exit(1)
	}
	if _, err := stdin.Write([]byte("\n")); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdin: %v\n", err)
		os.Exit(1)
	}

	if !scanner.Scan() {
		fmt.Fprintf(os.Stderr, "No response to prompts/list\n")
		os.Exit(1)
	}

	var promptsResp JSONRPCResponse
	if err := json.Unmarshal(scanner.Bytes(), &promptsResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing prompts/list response: %v\n", err)
		os.Exit(1)
	}

	if promptsResp.Error != nil {
		fmt.Printf("❌ prompts/list failed: %s\n", promptsResp.Error.Message)
		os.Exit(1)
	}

	var promptsResult map[string]interface{}
	if err := json.Unmarshal(promptsResp.Result, &promptsResult); err != nil {
		fmt.Printf("❌ Error parsing prompts/list result: %v\n", err)
		os.Exit(1)
	}

	if prompts, ok := promptsResult["prompts"].([]interface{}); ok {
		fmt.Printf("✅ Found %d prompts\n", len(prompts))
		if len(prompts) > 0 {
			fmt.Printf("   First prompt: %v\n", prompts[0])
		}
	} else {
		fmt.Printf("⚠️  Unexpected prompts result format\n")
	}

	fmt.Println("\n=== All Tests Passed! ===")
}
