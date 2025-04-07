# Serpent MCP Server Example

This example demonstrates how to use the Model Context Protocol (MCP) functionality in Serpent to create a command-line tool that can also be used as an MCP server.

## What is MCP?

The Model Context Protocol (MCP) is a protocol for communication between AI models and external tools or resources. It allows AI models to invoke tools and access resources provided by MCP servers.

## How to Use

### Running as a CLI Tool

You can run the example as a normal CLI tool:

```bash
# Echo a message
go run main.go echo "Hello, World!"

# Get version information
go run main.go version

# Show help
go run main.go --help
```

### Running as an MCP Server

You can run the example as an MCP server using the `mcp` subcommand:

```bash
go run main.go mcp
```

This will start an MCP server that listens on stdin/stdout for JSON-RPC 2.0 requests.

## MCP Protocol

### Lifecycle

The MCP server follows the standard MCP lifecycle:

1. The client sends an `initialize` request to the server
2. The server responds with its capabilities
3. The client sends an `initialized` notification
4. After this, normal message exchange can begin

All MCP methods will return an error if called before the initialization process is complete.

### Methods

The MCP server implements the following JSON-RPC 2.0 methods:

- `initialize`: Initializes the MCP server and returns its capabilities
- `notifications/initialized`: Notifies the server that initialization is complete
- `ping`: Simple ping method to check server availability
- `tools/list`: Lists all available tools
- `tools/call`: Invokes a tool with the given arguments
- `resources/list`: Lists all available resources
- `resources/templates/list`: Lists all available resource templates
- `resources/read`: Accesses a resource with the given URI

### Example Requests

Here are some example JSON-RPC 2.0 requests you can send to the MCP server:

#### Initialize

```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","clientInfo":{"name":"manual-test-client","version":"1.0.0"},"capabilities":{}}}
```

Response:
```json
{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"tools":true,"resources":true}}}
```

#### Initialized

```json
{"jsonrpc":"2.0","id":2,"method":"notifications/initialized"}
```

#### List Tools

```json
{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}
```

#### List Resources

```json
{"jsonrpc":"2.0","id":4,"method":"resources/list","params":{}}
```

#### Invoke Tool

```json
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"echo","arguments":{"_":"Hello from MCP!"}}}
```

#### Access Resource

```json
{"jsonrpc":"2.0","id":6,"method":"resources/read","params":{"uri":"version"}}
```

### Complete Initialization Example

Here's a complete example of the initialization process:

```json
// Client sends initialize request
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","clientInfo":{"name":"manual-test-client","version":"1.0.0"},"capabilities":{}}}

// Server responds with capabilities
{"jsonrpc":"2.0","id":1,"result":{"capabilities":{"tools":true,"resources":true}}}

// Client sends initialized notification
{"jsonrpc":"2.0","id":2,"method":"notifications/initialized"}

// Server acknowledges (optional, since initialized is technically a notification)
{"jsonrpc":"2.0","id":2,"result":{}}

// Now client can use MCP methods
{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}
```

## How to Implement MCP in Your Own Commands

To implement MCP in your own Serpent commands:

1. Add the `Tool` field to commands that should be invokable as MCP tools
2. Add the `Resource` field to commands that should be accessible as MCP resources
3. Add the MCP command to your root command using `root.AddMCPCommand()`

Example:

```go
// Create a command that will be exposed as an MCP tool
echoCmd := &serpent.Command{
    Use:   "echo [message]",
    Short: "Echo a message",
    Tool:  "echo", // This makes the command available as an MCP tool
    Handler: func(inv *serpent.Invocation) error {
        // Command implementation
    },
}

// Create a command that will be exposed as an MCP resource
versionCmd := &serpent.Command{
    Use:      "version",
    Short:    "Get version information",
    Resource: "version", // This makes the command available as an MCP resource
    Handler: func(inv *serpent.Invocation) error {
        // Command implementation
    },
}

// Add the MCP command to the root command
root.AddSubcommands(serpent.MCPCommand())
```

## Notes

- A command can have either a `Tool` field or a `Resource` field, but not both
- Commands with neither `Tool` nor `Resource` set will not be accessible via MCP
- The MCP server communicates using JSON-RPC 2.0 over stdin/stdout
