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
