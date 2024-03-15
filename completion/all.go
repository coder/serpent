package completion

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/coder/serpent"
)

func getUserShell() (string, error) {
	// Attempt to get the SHELL environment variable first
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell), nil
	}

	// Fallback: Look up the current user and parse /etc/passwd
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	// Open and parse /etc/passwd
	passwdFile, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(passwdFile), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, currentUser.Username+":") {
			parts := strings.Split(line, ":")
			if len(parts) > 6 {
				return filepath.Base(parts[6]), nil // The shell is typically the 7th field
			}
		}
	}

	return "", fmt.Errorf("default shell not found")
}

func IsCompletionMode(inv *serpent.Invocation) bool {
	_, ok := inv.Environ.Lookup("COMPLETION_MODE")
	return ok
}

// Request is a completion request from the shell.
type Request struct {
	Words []string
}

// Middleware returns a serpent middleware function that
// hijacks completion requests and generates completion.
//
// Commands may use the "extra" function to provide additional
// completion words based on the current request.
func Middleware(
	extra func(*Request, *serpent.Invocation) []string,
) serpent.MiddlewareFunc {
	return func(next serpent.HandlerFunc) serpent.HandlerFunc {
		return func(inv *serpent.Invocation) error {
			if !IsCompletionMode(inv) {
				return next(inv)
			}

			r := &Request{}

			r.Words = inv.Args

			var curWord string
			if len(r.Words) > 0 {
				curWord = r.Words[len(r.Words)-1]
			}

			var allResps []string
			for _, cmd := range inv.Command.Children {
				allResps = append(allResps, cmd.Name())
			}

			for _, opt := range inv.Command.Options {
				allResps = append(allResps, "--"+opt.Flag)
			}

			if extra != nil {
				allResps = append(allResps, extra(r, inv)...)
			}

			// fmt.Fprintf(
			// 	inv.Stderr, "%v: osArgs: %v, words: %v, curWord: %s\n",
			// 	inv.Command.Name(),
			// 	os.Args, r.Words, curWord,
			// )

			for _, resp := range allResps {
				if !strings.HasPrefix(resp, curWord) {
					continue
				}

				fmt.Fprintf(inv.Stdout, "%s\n", resp)
			}
			return nil
		}
	}
}

func rootCommand(cmd *serpent.Command) *serpent.Command {
	for cmd.Parent != nil {
		cmd = cmd.Parent
	}
	return cmd
}

// InstallCommand returns a serpent command that helps
// a user configure their shell to use serpent's completion.
func InstallCommand() *serpent.Command {
	defaultShell, err := getUserShell()
	if err != nil {
		defaultShell = "bash"
	}

	var shell string
	return &serpent.Command{
		Use:   "completion",
		Short: "Generate completion scripts for the given shell.",
		Handler: func(inv *serpent.Invocation) error {
			switch shell {
			case "bash":
				return GenerateBashCompletion(inv.Stdout, rootCommand(inv.Command))
			case "fish":
				return GenerateFishCompletion(inv.Stdout, rootCommand(inv.Command))
			default:
				return fmt.Errorf("unsupported shell: %s", shell)
			}
		},
		Options: serpent.OptionSet{
			{
				Flag:          "shell",
				FlagShorthand: "s",
				Default:       defaultShell,
				Description:   "The shell to generate a completion script for.",
				Value: serpent.EnumOf(
					&shell,
					"bash",
					"fish",
				),
			},
		},
	}
}
