package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/coder/serpent"
	"github.com/coder/serpent/completion"
)

// installCommand returns a serpent command that helps
// a user configure their shell to use serpent's completion.
func installCommand() *serpent.Command {
	var shell string
	return &serpent.Command{
		Use:   "completion [--shell <shell>]",
		Short: "Generate completion scripts for the given shell.",
		Handler: func(inv *serpent.Invocation) error {
			defaultShell, err := completion.DetectUserShell(inv.Command.Parent.Name())
			if err != nil {
				return fmt.Errorf("Could not detect user shell, please specify a shell using `--shell`")
			}
			return defaultShell.WriteCompletion(inv.Stdout)
		},
		Options: serpent.OptionSet{
			{
				Flag:          "shell",
				FlagShorthand: "s",
				Description:   "The shell to generate a completion script for.",
				Value:         completion.ShellOptions(&shell),
			},
		},
	}
}

func main() {
	var (
		print    bool
		upper    bool
		fileType string
		fileArr  []string
		types    []string
	)
	cmd := serpent.Command{
		Use:   "completetest <text>",
		Short: "Prints the given text to the console.",
		Options: serpent.OptionSet{
			{
				Name:        "different",
				Value:       serpent.BoolOf(&upper),
				Flag:        "different",
				Description: "Do the command differently.",
			},
		},
		Handler: func(inv *serpent.Invocation) error {
			if len(inv.Args) == 0 {
				inv.Stderr.Write([]byte("error: missing text\n"))
				os.Exit(1)
			}

			text := inv.Args[0]
			if upper {
				text = strings.ToUpper(text)
			}

			inv.Stdout.Write([]byte(text))
			return nil
		},
		Children: []*serpent.Command{
			{
				Use:   "sub",
				Short: "A subcommand",
				Handler: func(inv *serpent.Invocation) error {
					inv.Stdout.Write([]byte("subcommand"))
					return nil
				},
				Options: serpent.OptionSet{
					{
						Name:        "upper",
						Value:       serpent.BoolOf(&upper),
						Flag:        "upper",
						Description: "Prints the text in upper case.",
					},
				},
			},
			{
				Use: "file <file>",
				Handler: func(inv *serpent.Invocation) error {
					return nil
				},
				Options: serpent.OptionSet{
					{
						Name:        "print",
						Value:       serpent.BoolOf(&print),
						Flag:        "print",
						Description: "Print the file.",
					},
					{
						Name:        "type",
						Value:       serpent.EnumOf(&fileType, "binary", "text"),
						Flag:        "type",
						Description: "The type of file.",
					},
					{
						Name:        "extra",
						Flag:        "extra",
						Description: "Extra files.",
						Value:       serpent.StringArrayOf(&fileArr),
					},
					{
						Name:  "types",
						Flag:  "types",
						Value: serpent.EnumArrayOf(&types, "binary", "text"),
					},
				},
				CompletionHandler: completion.FileHandler(nil),
				Middleware:        serpent.RequireNArgs(1),
			},
			installCommand(),
		},
	}

	inv := cmd.Invoke().WithOS()

	err := inv.Run()
	if err != nil {
		panic(err)
	}
}
