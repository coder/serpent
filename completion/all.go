package completion

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/coder/serpent"
)

const (
	BashShell string = "bash"
	FishShell string = "fish"
	ZShell    string = "zsh"
)

var shellCompletionByName = map[string]func(io.Writer, string) error{
	BashShell: GenerateBashCompletion,
	FishShell: GenerateFishCompletion,
	ZShell:    GenerateZshCompletion,
}

func ShellOptions(choice *string) *serpent.Enum {
	return serpent.EnumOf(choice, BashShell, FishShell, ZShell)
}

func ShellHandler() serpent.CompletionHandlerFunc {
	return EnumHandler(BashShell, FishShell, ZShell)
}

func GetCompletion(writer io.Writer, shell string, cmdName string) error {
	fn, ok := shellCompletionByName[shell]
	if !ok {
		return fmt.Errorf("unknown shell %q", shell)
	}
	fn(writer, cmdName)
	return nil
}

func GetUserShell() (string, error) {
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
