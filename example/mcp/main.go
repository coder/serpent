package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/coder/serpent"
)

func main() {
	// Create a root command
	root := &serpent.Command{
		Use:   "mcp-example",
		Short: "Example MCP server",
		Long:  "An example of how to use the MCP functionality in serpent.",
	}

	var repeats int64 = 2

	// Add a command that will be exposed as an MCP tool
	echoCmd := &serpent.Command{
		Use:   "echo [message]",
		Short: "Echo a message",
		Tool:  "echo", // This makes the command available as an MCP tool
		Options: []serpent.Option{
			{
				Name:        "repeat",
				Flag:        "repeat", // Add the Flag field so it's exposed in JSON Schema
				Description: "Number of times to repeat the message.",
				Default:     "2",
				Value:       serpent.Int64Of(&repeats),
			},
		},
		Handler: func(inv *serpent.Invocation) error {
			message := "Hello, World!"
			if len(inv.Args) > 0 {
				message = strings.Join(inv.Args, " ")
			}
			for i := int64(0); i < repeats; i++ {
				if _, err := fmt.Fprintln(inv.Stdout, message); err != nil {
					return err
				}
			}
			return nil
		},
	}
	root.AddSubcommands(echoCmd)

	// Add a command that will be exposed as an MCP resource
	versionCmd := &serpent.Command{
		Use:      "version",
		Short:    "Get version information",
		Resource: "version", // This makes the command available as an MCP resource
		Handler: func(inv *serpent.Invocation) error {
			version := map[string]string{
				"version": "1.0.0",
				"name":    "serpent-mcp-example",
				"author":  "Coder",
			}
			encoder := json.NewEncoder(inv.Stdout)
			return encoder.Encode(version)
		},
	}
	root.AddSubcommands(versionCmd)

	// Add a command that will not be exposed via MCP
	hiddenCmd := &serpent.Command{
		Use:   "hidden",
		Short: "This command is not exposed via MCP",
		Handler: func(inv *serpent.Invocation) error {
			_, err := fmt.Fprintln(inv.Stdout, "This command is not exposed via MCP")
			return err
		},
	}
	root.AddSubcommands(hiddenCmd)

	// Add the MCP command to the root command
	root.AddSubcommands(serpent.MCPCommand())

	// Run the command
	if err := root.Invoke(os.Args[1:]...).WithOS().Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
