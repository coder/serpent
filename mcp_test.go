package serpent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestToolAndResourceFields(t *testing.T) {
	// Test that a command with neither Tool nor Resource is not MCP-enabled
	cmd := &Command{
		Use:   "regular",
		Short: "Regular command with no MCP fields",
	}
	if cmd.IsMCPEnabled() {
		t.Error("Command without Tool or Resource should not be MCP-enabled")
	}

	// Test that a command with Tool is MCP-enabled
	toolCmd := &Command{
		Use:   "tool-cmd",
		Short: "Command with Tool field",
		Tool:  "example-tool",
	}
	if !toolCmd.IsMCPEnabled() {
		t.Error("Command with Tool should be MCP-enabled")
	}

	// Test that a command with Resource is MCP-enabled
	resourceCmd := &Command{
		Use:      "resource-cmd",
		Short:    "Command with Resource field",
		Resource: "example-resource",
	}
	if !resourceCmd.IsMCPEnabled() {
		t.Error("Command with Resource should be MCP-enabled")
	}

	// Test that a command cannot have both Tool and Resource
	invalidCmd := &Command{
		Use:      "invalid-cmd",
		Short:    "Command with both Tool and Resource",
		Tool:     "example-tool",
		Resource: "example-resource",
	}

	if err := invalidCmd.init(); err == nil {
		t.Error("Command with both Tool and Resource should fail initialization")
	}
}

func TestMCPServerSetup(t *testing.T) {
	// Create a root command with subcommands having Tool and Resource
	root := &Command{
		Use:   "root",
		Short: "Root command",
	}

	toolCmd := &Command{
		Use:   "tool-cmd",
		Short: "Tool command",
		Tool:  "test-tool",
		Handler: func(inv *Invocation) error {
			fmt.Fprintln(inv.Stdout, "Tool executed!")
			return nil
		},
	}

	resourceCmd := &Command{
		Use:      "resource-cmd",
		Short:    "Resource command",
		Resource: "test-resource",
		Handler: func(inv *Invocation) error {
			fmt.Fprintln(inv.Stdout, `{"result": "Resource data"`)
			return nil
		},
	}

	templatedResourceCmd := &Command{
		Use:      "templated-cmd",
		Short:    "Templated resource command",
		Resource: "test/{param}",
		Handler: func(inv *Invocation) error {
			fmt.Fprintln(inv.Stdout, `{"template": "Resource template"}`)
			return nil
		},
	}

	root.AddSubcommands(toolCmd, resourceCmd, templatedResourceCmd)

	// Create a server with the root command
	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	server := NewMCPServer(root, stdin, stdout, stderr)

	// Check if tools were indexed correctly
	if len(server.toolCmds) != 1 {
		t.Errorf("Expected 1 tool command, got %d", len(server.toolCmds))
	}

	if len(server.resourceCmds) != 1 {
		t.Errorf("Expected 1 resource command, got %d", len(server.resourceCmds))
	}

	if len(server.resourceTemplates) != 1 {
		t.Errorf("Expected 1 resource template, got %d", len(server.resourceTemplates))
	}

	if cmd, ok := server.toolCmds["test-tool"]; !ok || cmd != toolCmd {
		t.Error("Tool command not properly indexed")
	}

	if cmd, ok := server.resourceCmds["test-resource"]; !ok || cmd != resourceCmd {
		t.Error("Resource command not properly indexed")
	}

	if cmd, ok := server.resourceTemplates["test/{param}"]; !ok || cmd != templatedResourceCmd {
		t.Error("Resource template command not properly indexed")
	}
}

func TestJSONSchemaGeneration(t *testing.T) {
	// Create a command with various option types
	cmd := &Command{
		Use:   "test-schema",
		Short: "Command for testing schema generation",
		Options: OptionSet{
			{
				Flag:        "string-flag",
				Description: "A string flag",
				Value:       StringOf(new(string)),
			},
			{
				Flag:        "bool-flag",
				Description: "A boolean flag",
				Value:       BoolOf(new(bool)),
				Required:    true,
			},
			{
				Flag:        "file-path",
				Description: "A file path",
				Value:       StringOf(new(string)),
			},
			{
				Flag:        "enum-choice",
				Description: "An enum choice",
				Value:       EnumOf(new(string), "option1", "option2", "option3"),
			},
		},
	}

	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	server := NewMCPServer(cmd, stdin, stdout, stderr)
	schema, err := server.generateJSONSchema(cmd)

	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	var schemaObj map[string]interface{}
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		t.Fatalf("Generated schema is not valid JSON: %v", err)
	}

	// Validate schema structure
	if schemaObj["type"] != "object" {
		t.Error("Schema should have type 'object'")
	}

	properties, ok := schemaObj["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties map")
	}

	// Check if required fields are properly identified
	required, ok := schemaObj["required"].([]interface{})
	if !ok {
		t.Fatal("Schema should have required array")
	}

	foundRequired := false
	for _, r := range required {
		if r == "bool-flag" {
			foundRequired = true
			break
		}
	}
	if !foundRequired {
		t.Error("Required flag not found in required list")
	}

	// Check if properties have correct types
	filePathProp, ok := properties["file-path"].(map[string]interface{})
	if !ok {
		t.Fatal("file-path property not found or not an object")
	}
	if filePathProp["format"] != "path" {
		t.Errorf("file-path should have format 'path', got %v", filePathProp["format"])
	}

	enumProp, ok := properties["enum-choice"].(map[string]interface{})
	if !ok {
		t.Fatal("enum-choice property not found or not an object")
	}

	enumValues, ok := enumProp["enum"].([]interface{})
	if !ok || len(enumValues) != 3 {
		t.Errorf("enum-choice should have enum array with 3 values, got %v", enumProp["enum"])
	}
}

func TestMCPServerRun(t *testing.T) {
	// Create a simple command for testing
	cmd := &Command{
		Use:   "test",
		Short: "Test command",
	}

	toolCmd := &Command{
		Use:   "tool-cmd",
		Short: "Tool command",
		Tool:  "test-tool",
		Handler: func(inv *Invocation) error {
			fmt.Fprintln(inv.Stdout, "Tool executed!")
			return nil
		},
	}

	cmd.AddSubcommands(toolCmd)

	// Setup the server with buffers
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","clientInfo":{"name":"test-client","version":"1.0.0"},"capabilities":{}}}
{"jsonrpc":"2.0","id":2,"method":"notifications/initialized","params":{}}
{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}
`
	stdin := strings.NewReader(input)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	server := NewMCPServer(cmd, stdin, stdout, stderr)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Run the server (it will stop when stdin is drained or context is cancelled)
	err := server.Run(ctx)
	if err != nil {
		t.Fatalf("Server run failed: %v", err)
	}

	// Check output
	output := stdout.String()

	// Verify we got the expected responses
	if !strings.Contains(output, `"protocolVersion"`) {
		t.Error("Missing protocol version in initialize response")
	}

	if !strings.Contains(output, `"tools":[`) {
		t.Error("Missing tools list in response")
	}

	if !strings.Contains(output, `"test-tool"`) {
		t.Error("Tool name not found in response")
	}
}

func TestLenientParameterHandling(t *testing.T) {
	// Create a simple command for testing
	cmd := &Command{
		Use:   "test",
		Short: "Test command",
	}

	toolCmd := &Command{
		Use:   "tool-cmd",
		Short: "Tool command",
		Tool:  "test-tool",
		Handler: func(inv *Invocation) error {
			fmt.Fprintln(inv.Stdout, "Tool executed!")
			return nil
		},
	}

	cmd.AddSubcommands(toolCmd)

	// Test the unmarshalParamsLenient function directly
	testCases := []struct {
		name      string
		params    json.RawMessage
		expectErr bool
	}{
		{
			name:      "Missing params (nil)",
			params:    nil,
			expectErr: false,
		},
		{
			name:      "Empty params object",
			params:    json.RawMessage(`{}`),
			expectErr: false,
		},
		{
			name:      "Null params",
			params:    json.RawMessage(`null`),
			expectErr: false,
		},
		{
			name:      "Empty params array",
			params:    json.RawMessage(`[]`),
			expectErr: false,
		},
		{
			name:      "Partial params",
			params:    json.RawMessage(`{"protocolVersion":"2025-03-26"}`),
			expectErr: false,
		},
		{
			name:      "Invalid params format",
			params:    json.RawMessage(`"invalid"`),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := UnmarshalParamsLenient[InitializeParams](tc.params); err != nil {
				if !tc.expectErr {
					t.Errorf("Expected no error but got: %v", err)
				}
				return
			}

			if tc.expectErr {
				t.Error("Expected error but got none")
			}
		})
	}
}
