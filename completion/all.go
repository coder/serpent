package completion

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/coder/serpent"
)

type Shell interface {
	Name() string
	InstallPath() (string, error)
	WriteCompletion(io.Writer) error
}

const (
	ShellBash       string = "bash"
	ShellFish       string = "fish"
	ShellZsh        string = "zsh"
	ShellPowershell string = "powershell"
)

func ShellByName(shell, programName string) Shell {
	switch shell {
	case ShellBash:
		return Bash(runtime.GOOS, programName)
	case ShellFish:
		return Fish(runtime.GOOS, programName)
	case ShellZsh:
		return Zsh(runtime.GOOS, programName)
	case ShellPowershell:
		return Powershell(runtime.GOOS, programName)
	default:
		return nil
	}
}

func ShellOptions(choice *string) *serpent.Enum {
	return serpent.EnumOf(choice, ShellBash, ShellFish, ShellZsh, ShellPowershell)
}

func DetectUserShell(programName string) (Shell, error) {
	// Attempt to get the SHELL environment variable first
	if shell := os.Getenv("SHELL"); shell != "" {
		return ShellByName(filepath.Base(shell), ""), nil
	}

	// Fallback: Look up the current user and parse /etc/passwd
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	// Open and parse /etc/passwd
	passwdFile, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(passwdFile), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, currentUser.Username+":") {
			parts := strings.Split(line, ":")
			if len(parts) > 6 {
				return ShellByName(filepath.Base(parts[6]), programName), nil // The shell is typically the 7th field
			}
		}
	}

	return nil, fmt.Errorf("default shell not found")
}

func generateCompletion(
	scriptTemplate string,
) func(io.Writer, string) error {
	return func(w io.Writer, programName string) error {
		tmpl, err := template.New("script").Parse(scriptTemplate)
		if err != nil {
			return fmt.Errorf("parse template: %w", err)
		}

		err = tmpl.Execute(
			w,
			map[string]string{
				"Name": programName,
			},
		)
		if err != nil {
			return fmt.Errorf("execute template: %w", err)
		}

		return nil
	}
}

func InstallShellCompletion(shell Shell) error {
	path, err := shell.InstallPath()
	if err != nil {
		return fmt.Errorf("get install path: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create and append to file: %w", err)
	}
	defer f.Close()

	return shell.WriteCompletion(f)
}
