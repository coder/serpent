package main

import (
	"os"
	"strings"

	"github.com/coder/serpent"
	"github.com/coder/serpent/completion"
)

// InstallCommand returns a serpent command that helps
// a user configure their shell to use serpent's completion.
func InstallCommand() *serpent.Command {
	defaultShell, err := completion.GetUserShell()
	if err != nil {
		defaultShell = "bash"
	}

	var shell string
	return &serpent.Command{
		Use:   "completion",
		Short: "Generate completion scripts for the given shell.",
		Handler: func(inv *serpent.Invocation) error {
			completion.GetCompletion(inv.Stdout, shell, inv.Command.Parent.Name())
			return nil
		},
		Options: serpent.OptionSet{
			{
				Flag:              "shell",
				FlagShorthand:     "s",
				Default:           defaultShell,
				Description:       "The shell to generate a completion script for.",
				Value:             completion.ShellOptions(&shell),
				CompletionHandler: completion.ShellHandler(),
			},
		},
	}
}

func main() {
	var (
		print    bool
		upper    bool
		fileType string
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
						Name:              "type",
						Value:             serpent.EnumOf(&fileType, "binary", "text"),
						Flag:              "type",
						Description:       "The type of file.",
						CompletionHandler: completion.EnumHandler("binary", "text"),
					},
				},
				CompletionHandler: completion.FileHandler(func(info os.FileInfo) bool {
					return true
				}),
				Middleware: serpent.RequireNArgs(1),
			},
			InstallCommand(),
		},
	}

	inv := cmd.Invoke().WithOS()

	err := inv.Run()
	if err != nil {
		panic(err)
	}
}
