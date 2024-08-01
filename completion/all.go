package completion

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/coder/serpent"
)

const (
	BashShell  string = "bash"
	FishShell  string = "fish"
	ZShell     string = "zsh"
	Powershell string = "powershell"
)

var shellCompletionByName = map[string]func(io.Writer, string) error{
	BashShell:  generateCompletion(bashCompletionTemplate),
	FishShell:  generateCompletion(fishCompletionTemplate),
	ZShell:     generateCompletion(zshCompletionTemplate),
	Powershell: generateCompletion(pshCompletionTemplate),
}

func ShellOptions(choice *string) *serpent.Enum {
	return serpent.EnumOf(choice, BashShell, FishShell, ZShell, Powershell)
}

func WriteCompletion(writer io.Writer, shell string, cmdName string) error {
	fn, ok := shellCompletionByName[shell]
	if !ok {
		return fmt.Errorf("unknown shell %q", shell)
	}
	fn(writer, cmdName)
	return nil
}

func DetectUserShell() (string, error) {
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

func generateCompletion(
	scriptTemplate string,
) func(io.Writer, string) error {
	return func(w io.Writer, rootCmdName string) error {
		tmpl, err := template.New("script").Parse(scriptTemplate)
		if err != nil {
			return fmt.Errorf("parse template: %w", err)
		}

		err = tmpl.Execute(
			w,
			map[string]string{
				"Name": rootCmdName,
			},
		)
		if err != nil {
			return fmt.Errorf("execute template: %w", err)
		}

		return nil
	}
}
