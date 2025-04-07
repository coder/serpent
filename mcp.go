package serpent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"golang.org/x/xerrors"
)

// JSONRPC2 message types
type JSONRPC2Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPC2Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPC2Error  `json:"error,omitempty"`
}

type JSONRPC2Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPC2Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// MCP protocol message types
type InitializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
	Capabilities map[string]any `json:"capabilities"`
}

type InitializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
	Capabilities map[string]any `json:"capabilities"`
}

type ListToolsParams struct {
	Cursor string `json:"cursor,omitempty"`
}

type ListToolsResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema,omitempty"`
}

type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}

type ListResourcesParams struct {
	Cursor string `json:"cursor,omitempty"`
}

type ListResourcesResult struct {
	Resources  []Resource `json:"resources"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Size        int    `json:"size,omitempty"`
}

type ListResourceTemplatesParams struct {
	Cursor string `json:"cursor,omitempty"`
}

type ListResourceTemplatesResult struct {
	ResourceTemplates []ResourceTemplate `json:"resourceTemplates"`
	NextCursor        string             `json:"nextCursor,omitempty"`
}

type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ReadResourceParams struct {
	URI string `json:"uri"`
}

type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// JSON-RPC 2.0 error codes
const (
	// Standard JSON-RPC error codes
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603

	// MCP specific error codes
	ErrorCodeResourceNotFound    = -32002
	ErrorCodeResourceUnavailable = -32001
	ErrorCodeToolNotFound        = -32100
	ErrorCodeToolUnavailable     = -32101
)

// MCPServer represents an MCP server that can handle tool invocations and resource access
type MCPServer struct {
	rootCmd           *Command
	stdin             io.Reader
	stdout            io.Writer
	stderr            io.Writer
	cmdFinder         CommandFinder
	toolCmds          map[string]*Command
	resourceCmds      map[string]*Command
	resourceTemplates map[string]*Command // Maps URI templates to commands
	initialized       bool                // Track if the server has been initialized
	protocolVersion   string              // Protocol version negotiated during initialization
}

// CommandFinder is a function that finds a command by name
type CommandFinder func(rootCmd *Command, name string) *Command

// DefaultCommandFinder is the default implementation of CommandFinder
func DefaultCommandFinder(rootCmd *Command, name string) *Command {
	parts := strings.Split(name, " ")
	cmd := rootCmd

	for _, part := range parts {
		found := false
		for _, child := range cmd.Children {
			if child.Name() == part {
				cmd = child
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	return cmd
}

// NewMCPServer creates a new MCP server
func NewMCPServer(rootCmd *Command, stdin io.Reader, stdout, stderr io.Writer) *MCPServer {
	server := &MCPServer{
		rootCmd:           rootCmd,
		stdin:             stdin,
		stdout:            stdout,
		stderr:            stderr,
		cmdFinder:         DefaultCommandFinder,
		toolCmds:          make(map[string]*Command),
		resourceCmds:      make(map[string]*Command),
		resourceTemplates: make(map[string]*Command),
		protocolVersion:   "2025-03-26", // Default to latest version
	}

	// Index all commands with Tool or Resource fields
	rootCmd.Walk(func(cmd *Command) {
		if cmd.Tool != "" {
			server.toolCmds[cmd.Tool] = cmd
		}
		if cmd.Resource != "" {
			if strings.Contains(cmd.Resource, "{") && strings.Contains(cmd.Resource, "}") {
				// This is a URI template
				server.resourceTemplates[cmd.Resource] = cmd
			} else {
				// This is a static resource URI
				server.resourceCmds[cmd.Resource] = cmd
			}
		}
	})

	return server
}

// Run starts the MCP server
func (s *MCPServer) Run(ctx context.Context) error {
	// Check if context is already done
	select {
	case <-ctx.Done():
		return nil
	default:
		// Continue with normal operation
	}

	// Create a buffered reader for stdin
	reader := bufio.NewReader(s.stdin)

	// Process requests until context is done or EOF
	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			return nil
		default:
			// Continue processing
		}

		// Try to read a line with a non-blocking approach
		var line string

		// Use a channel to communicate when a line is read
		lineCh := make(chan string, 1)
		errCh := make(chan error, 1)

		// Start a goroutine to read a line
		go func() {
			text, err := reader.ReadString('\n')
			if err != nil {
				errCh <- err
				return
			}
			lineCh <- strings.TrimSpace(text)
		}()

		// Wait for either a line to be read, an error, or context cancellation
		select {
		case <-ctx.Done():
			// Context was canceled, exit gracefully
			return nil
		case err := <-errCh:
			if err == io.EOF {
				// End of input, exit normally
				return nil
			}
			// Other error
			return xerrors.Errorf("reading stdin: %w", err)
		case line = <-lineCh:
			// Line was read successfully, process it
		}

		// Parse the JSON-RPC request
		var req JSONRPC2Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendErrorResponse(nil, ErrorCodeParseError, "Failed to parse JSON-RPC request", nil)
			continue
		}

		// Ensure this is a JSON-RPC 2.0 request
		if req.JSONRPC != "2.0" {
			s.sendErrorResponse(req.ID, ErrorCodeInvalidRequest, "Invalid JSON-RPC version, expected 2.0", nil)
			continue
		}

		// Handle the request based on the method
		switch req.Method {
		case "initialize":
			s.handleInitialize(req)
		case "notifications/initialized":
			s.handleInitialized(req)
		case "ping":
			s.handlePing(req)
		case "tools/list":
			if !s.initialized {
				s.sendErrorResponse(req.ID, ErrorCodeInvalidRequest, "Server not initialized", nil)
				continue
			}
			s.handleListTools(req)
		case "tools/call":
			if !s.initialized {
				s.sendErrorResponse(req.ID, ErrorCodeInvalidRequest, "Server not initialized", nil)
				continue
			}
			s.handleCallTool(req)
		case "resources/list":
			if !s.initialized {
				s.sendErrorResponse(req.ID, ErrorCodeInvalidRequest, "Server not initialized", nil)
				continue
			}
			s.handleListResources(req)
		case "resources/templates/list":
			if !s.initialized {
				s.sendErrorResponse(req.ID, ErrorCodeInvalidRequest, "Server not initialized", nil)
				continue
			}
			s.handleListResourceTemplates(req)
		case "resources/read":
			if !s.initialized {
				s.sendErrorResponse(req.ID, ErrorCodeInvalidRequest, "Server not initialized", nil)
				continue
			}
			s.handleReadResource(req)
		default:
			s.sendErrorResponse(req.ID, ErrorCodeMethodNotFound, fmt.Sprintf("Method not found: %s", req.Method), nil)
		}
	}
}

// handlePing handles the ping method, responding with an empty result
func (s *MCPServer) handlePing(req JSONRPC2Request) {
	s.sendSuccessResponse(req.ID, struct{}{})
}

// handleListTools handles the tools/list method
func (s *MCPServer) handleListTools(req JSONRPC2Request) {
	if _, err := UnmarshalParamsLenient[ListToolsParams](req.Params); err != nil {
		s.sendErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid parameters", nil)
		return
	}

	tools := make([]Tool, 0, len(s.toolCmds))
	for name, cmd := range s.toolCmds {
		// Generate a proper JSON Schema from the command's options
		schema, err := s.generateJSONSchema(cmd)
		if err != nil {
			fmt.Fprintf(s.stderr, "Failed to generate schema for tool %s: %v\n", name, err)
			schema = json.RawMessage(`{}`)
		}

		tools = append(tools, Tool{
			Name:        name,
			Description: cmd.Short,
			InputSchema: schema,
		})
	}

	response := ListToolsResult{
		Tools: tools,
		// We're not implementing pagination for now
		NextCursor: "",
	}
	s.sendSuccessResponse(req.ID, response)
}

// handleListResources handles the resources/list method
func (s *MCPServer) handleListResources(req JSONRPC2Request) {
	if _, err := UnmarshalParamsLenient[ListResourcesParams](req.Params); err != nil {
		s.sendErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid parameters", nil)
		return
	}

	resources := make([]Resource, 0, len(s.resourceCmds))
	for uri, cmd := range s.resourceCmds {
		resources = append(resources, Resource{
			URI:         uri,
			Name:        cmd.Name(),
			Description: cmd.Short,
			MimeType:    "application/json", // Default MIME type
		})
	}

	response := ListResourcesResult{
		Resources: resources,
		// We're not implementing pagination for now
		NextCursor: "",
	}
	s.sendSuccessResponse(req.ID, response)
}

// handleListResourceTemplates handles the resources/templates/list method
func (s *MCPServer) handleListResourceTemplates(req JSONRPC2Request) {
	_, err := UnmarshalParamsLenient[ListResourceTemplatesParams](req.Params)
	if err != nil {
		s.sendErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid parameters", nil)
		return
	}

	templates := make([]ResourceTemplate, 0, len(s.resourceTemplates))
	for uriTemplate, cmd := range s.resourceTemplates {
		templates = append(templates, ResourceTemplate{
			URITemplate: uriTemplate,
			Name:        cmd.Name(),
			Description: cmd.Short,
			MimeType:    "application/json", // Default MIME type
		})
	}

	response := ListResourceTemplatesResult{
		ResourceTemplates: templates,
		// We're not implementing pagination for now
		NextCursor: "",
	}
	s.sendSuccessResponse(req.ID, response)
}

// handleCallTool handles the tools/call method
func (s *MCPServer) handleCallTool(req JSONRPC2Request) {
	params, err := UnmarshalParamsLenient[CallToolParams](req.Params)
	if err != nil {
		s.sendErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid parameters", nil)
		return
	}

	cmd, ok := s.toolCmds[params.Name]
	if !ok {
		s.sendErrorResponse(req.ID, ErrorCodeToolNotFound, fmt.Sprintf("Tool not found: %s", params.Name), nil)
		return
	}

	// Create a new invocation with captured stdout/stderr
	var stdout, stderr strings.Builder
	inv := cmd.Invoke()
	inv.Stdout = &stdout
	inv.Stderr = &stderr

	// Parse the arguments as a map and convert to command-line args
	var args map[string]any
	if err := json.Unmarshal(params.Arguments, &args); err != nil {
		s.sendErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid arguments format", nil)
		return
	}
	// Convert the arguments map to command-line args
	var cmdArgs []string

	// Check for positional arguments (using "_" as the key)
	if posArgs, ok := args["_"]; ok {
		switch val := posArgs.(type) {
		case string:
			cmdArgs = append(cmdArgs, val)
		case []any:
			for _, item := range val {
				cmdArgs = append(cmdArgs, fmt.Sprintf("%v", item))
			}
		default:
			cmdArgs = append(cmdArgs, fmt.Sprintf("%v", val))
		}
		// Remove the "_" key so it's not processed as a flag
		delete(args, "_")
	}

	// Process remaining arguments as flags
	for k, v := range args {
		switch val := v.(type) {
		case bool:
			if val {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--%s", k))
			} else {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--%s=false", k))
			}
		case []any:
			for _, item := range val {
				cmdArgs = append(cmdArgs, fmt.Sprintf("--%s=%v", k, item))
			}
		default:
			cmdArgs = append(cmdArgs, fmt.Sprintf("--%s=%v", k, v))
		}
	}

	inv.Args = cmdArgs

	// Run the command
	err = inv.Run()

	// Prepare the response following MCP specification
	var content []ToolContent
	if stdout.Len() > 0 {
		content = append(content, ToolContent{
			Type: "text",
			Text: stdout.String(),
		})
	}

	if stderr.Len() > 0 {
		content = append(content, ToolContent{
			Type: "text",
			Text: stderr.String(),
		})
	}

	// If no content but error, add error message
	if len(content) == 0 && err != nil {
		content = append(content, ToolContent{
			Type: "text",
			Text: err.Error(),
		})
	}

	// If still no content, add empty result
	if len(content) == 0 {
		content = append(content, ToolContent{
			Type: "text",
			Text: "",
		})
	}

	response := CallToolResult{
		Content: content,
		IsError: err != nil,
	}

	s.sendSuccessResponse(req.ID, response)
}

// handleReadResource handles the resources/read method
func (s *MCPServer) handleReadResource(req JSONRPC2Request) {
	params, err := UnmarshalParamsLenient[ReadResourceParams](req.Params)
	if err != nil {
		s.sendErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid parameters", nil)
		return
	}

	// First check if this is a direct resource URI match
	cmd, ok := s.resourceCmds[params.URI]
	if !ok {
		// If not a direct match, check if it matches any URI template
		for template, templateCmd := range s.resourceTemplates {
			// Very basic template matching - would need more complex handling for real URI templates
			if matched, _ := path.Match(template, params.URI); matched {
				cmd = templateCmd
				ok = true
				break
			}
		}

		if !ok {
			s.sendErrorResponse(req.ID, ErrorCodeResourceNotFound, fmt.Sprintf("Resource not found: %s", params.URI), nil)
			return
		}
	}

	// Create a new invocation with captured stdout
	var stdout strings.Builder
	inv := cmd.Invoke()
	inv.Stdout = &stdout

	// Run the command
	if err := inv.Run(); err != nil {
		s.sendErrorResponse(req.ID, ErrorCodeResourceUnavailable, fmt.Sprintf("Resource unavailable: %s", err.Error()), nil)
		return
	}

	// Create the response with the text content
	response := ReadResourceResult{
		Contents: []ResourceContent{
			{
				URI:      params.URI,
				MimeType: "application/json", // Assuming JSON by default
				Text:     stdout.String(),
			},
		},
	}

	s.sendSuccessResponse(req.ID, response)
}

// sendSuccessResponse sends a successful JSON-RPC response
func (s *MCPServer) sendSuccessResponse(id any, result any) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		s.sendErrorResponse(id, ErrorCodeInternalError, "Failed to marshal result", nil)
		return
	}

	response := JSONRPC2Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultBytes,
	}

	s.sendResponse(response)
}

// generateJSONSchema generates a JSON Schema for a command's options
func (s *MCPServer) generateJSONSchema(cmd *Command) (json.RawMessage, error) {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
		"required":   []string{},
	}

	properties := schema["properties"].(map[string]any)
	requiredList := schema["required"].([]string)

	// Process each option in the command
	for _, opt := range cmd.Options {
		// Skip options that aren't exposed as flags
		if opt.Flag == "" {
			continue
		}
		// Skip hidden options
		if opt.Hidden {
			continue
		}

		property := map[string]any{
			"description": opt.Description,
		}

		// Determine JSON Schema type using pflag.Value.Type()
		valueType := opt.Value.Type()

		switch valueType {
		case "string":
			property["type"] = "string"
			// Special handling for file paths
			if opt.Flag == "file-path" {
				property["format"] = "path"
			}
		case "bool":
			property["type"] = "boolean"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			property["type"] = "integer" // Use integer for whole numbers
		case "float32", "float64":
			property["type"] = "number"
		case "ip", "ipMask", "ipNet", "count": // Specific pflag types
			// Count is integer, others are strings
			if valueType == "count" {
				property["type"] = "integer"
			} else {
				property["type"] = "string"
			}
		case "duration":
			property["type"] = "string" // Represent duration as string (e.g., "1h", "30m")
			property["format"] = "duration"
		// Handle slice types
		case "stringSlice":
			property["type"] = "array"
			property["items"] = map[string]any{"type": "string"}
		case "boolSlice":
			property["type"] = "array"
			property["items"] = map[string]any{"type": "boolean"}
		case "intSlice", "int32Slice", "int64Slice", "uintSlice":
			property["type"] = "array"
			property["items"] = map[string]any{"type": "integer"}
		case "float32Slice", "float64Slice":
			property["type"] = "array"
			property["items"] = map[string]any{"type": "number"}
		case "ipSlice":
			property["type"] = "array"
			property["items"] = map[string]any{"type": "string"}
		case "durationSlice":
			property["type"] = "array"
			property["items"] = map[string]any{"type": "string", "format": "duration"}
		case "stringArray", "stringToString", "stringToInt", "stringToInt64": // More pflag types
			// stringArray is like stringSlice
			// Map types are complex, represent as object for now
			if valueType == "stringArray" {
				property["type"] = "array"
				property["items"] = map[string]any{"type": "string"}
			} else {
				property["type"] = "object"
				property["additionalProperties"] = map[string]any{
					"type": "string", // Default to string value type for maps
				}
				if valueType == "stringToInt" || valueType == "stringToInt64" {
					property["additionalProperties"] = map[string]any{
						"type": "integer",
					}
				}
			}
		// Handle custom serpent types
		default:
			// Check for known serpent custom types (Enum, EnumArray)
			if enum, ok := opt.Value.(*Enum); ok {
				property["type"] = "string"
				property["enum"] = enum.Choices
			} else if enumArray, ok := opt.Value.(*EnumArray); ok {
				property["type"] = "array"
				property["items"] = map[string]any{
					"type": "string",
					"enum": enumArray.Choices,
				}
			} else {
				// Fallback for unknown types
				property["type"] = "string"
				fmt.Fprintf(s.stderr, "Warning: Unknown pflag type '%s' for option '%s', defaulting to string\n", valueType, opt.Flag)
			}
		}

		// Add the property definition
		properties[opt.Flag] = property

		// Add to required list if Required is true AND Default is not set
		// (as per comment in option.go)
		if opt.Required && opt.Default == "" {
			requiredList = append(requiredList, opt.Flag)
		}
	}

	// Update required field only if it's not empty
	if len(requiredList) > 0 {
		schema["required"] = requiredList
	} else {
		// Remove the empty required array if no options are required
		delete(schema, "required")
	}

	return json.MarshalIndent(schema, "", "  ") // Use MarshalIndent for readability
}

// sendErrorResponse sends an error JSON-RPC response
func (s *MCPServer) sendErrorResponse(id any, code int, message string, data any) {
	var dataBytes json.RawMessage
	if data != nil {
		var err error
		dataBytes, err = json.Marshal(data)
		if err != nil {
			// If we can't marshal the data, just ignore it
			dataBytes = nil
		}
	}

	response := JSONRPC2Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPC2Error{
			Code:    code,
			Message: message,
			Data:    dataBytes,
		},
	}

	s.sendResponse(response)
}

// sendResponse sends a JSON-RPC response to stdout
func (s *MCPServer) sendResponse(response JSONRPC2Response) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(s.stderr, "Failed to marshal response: %v\n", err)
		return
	}

	fmt.Fprintln(s.stdout, string(responseBytes))
}

// unmarshalParamsLenient unmarshals JSON parameters in a lenient way, handling omitted optional parameters.
// If the raw JSON is empty, null, or "[]", it will use the default zero value for the target struct.
// If the JSON contains invalid data (not empty), the original error is returned.
// UnmarshalParamsLenient unmarshals JSON parameters in a lenient way, handling omitted optional parameters.
// If the raw JSON is empty, null, or "[]", it will return the zero value for type T.
// If the JSON contains invalid data (not empty), the original error is returned.
func UnmarshalParamsLenient[T any](data json.RawMessage) (T, error) {
	var result T

	// If params is nil, empty, or just whitespace/empty array, use zero value
	if len(data) == 0 || string(data) == "null" || string(data) == "{}" || string(data) == "[]" {
		// Return the zero value of T
		return result, nil
	}

	// Otherwise, try to unmarshal normally
	err := json.Unmarshal(data, &result)
	if err != nil {
		// Only return error if the JSON is not empty and contains invalid data
		return result, err
	}

	return result, nil
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(req JSONRPC2Request) {
	params, err := UnmarshalParamsLenient[InitializeParams](req.Params)
	if err != nil {
		s.sendErrorResponse(req.ID, ErrorCodeInvalidParams, "Invalid parameters", nil)
		return
	}

	// Negotiate protocol version
	if params.ProtocolVersion != "" {
		// For now, we just accept the client's protocol version
		// In a real implementation, you would compare with supported versions
		s.protocolVersion = params.ProtocolVersion
	}

	// Determine if we have tools and resources
	hasTools := len(s.toolCmds) > 0
	hasResources := len(s.resourceCmds) > 0 || len(s.resourceTemplates) > 0

	// Create capabilities object following MCP 2025-03-26 spec
	capabilities := map[string]any{}

	if hasTools {
		capabilities["tools"] = map[string]any{
			"listChanged": false, // We don't support list change notifications yet
		}
	}

	if hasResources {
		capabilities["resources"] = map[string]any{
			"listChanged": false, // We don't support list change notifications yet
			"subscribe":   false, // We don't support subscriptions yet
		}
	}

	// Create the response
	result := InitializeResult{
		ProtocolVersion: s.protocolVersion,
		ServerInfo: struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "serpent-mcp",
			Version: "1.0.0",
		},
		Capabilities: capabilities,
	}

	// Send the response
	s.sendSuccessResponse(req.ID, result)
}

// handleInitialized handles the initialized notification
func (s *MCPServer) handleInitialized(req JSONRPC2Request) {
	// Mark the server as initialized
	s.initialized = true

	// No response needed for a notification
	// But we'll send a success response anyway since our request has an ID
	if req.ID != nil {
		s.sendSuccessResponse(req.ID, struct{}{})
	}
}

// MCPCommand creates a generic command that can run any serpent command as an MCP server
func MCPCommand() *Command {
	return &Command{
		Use:   "mcp [command]",
		Short: "Run a command as an MCP server",
		Long: `Run a command as a Model Context Protocol (MCP) server over stdio.

This command allows any serpent command to be exposed as an MCP server, which can
provide tools and resources to MCP clients. The server communicates using JSON-RPC 2.0
over stdin/stdout.

If a command name is provided, that specific command will be run as an MCP server.
Otherwise, the root command will be used.

Commands with a Tool field set can be invoked as MCP tools.
Commands with a Resource field set can be accessed as MCP resources.
Commands with neither Tool nor Resource set will not be accessible via MCP.`,
		Handler: func(inv *Invocation) error {
			rootCmd := inv.Command
			if rootCmd.Parent != nil {
				// Find the root command
				for rootCmd.Parent != nil {
					rootCmd = rootCmd.Parent
				}
			}

			// If a command name is provided, use that as the root
			if len(inv.Args) > 0 {
				cmdName := strings.Join(inv.Args, " ")
				cmd := DefaultCommandFinder(rootCmd, cmdName)
				if cmd == nil {
					return xerrors.Errorf("command not found: %s", cmdName)
				}
				rootCmd = cmd
			}

			// Create and run the MCP server
			server := NewMCPServer(rootCmd, inv.Stdin, inv.Stdout, inv.Stderr)
			return server.Run(inv.Context())
		},
	}
}
