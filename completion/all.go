package completion

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/coder/serpent"

	"github.com/natefinch/atomic"
)

const (
	completionStartTemplate = `# ============ BEGIN {{.Name}} COMPLETION ============`
	completionEndTemplate   = `# ============ END {{.Name}} COMPLETION ==============`
)

type Shell interface {
	Name() string
	InstallPath() (string, error)
	WriteCompletion(io.Writer) error
	ProgramName() string
}

const (
	ShellBash       string = "bash"
	ShellFish       string = "fish"
	ShellZsh        string = "zsh"
	ShellPowershell string = "powershell"
)

func ShellByName(shell, programName string) (Shell, error) {
	switch shell {
	case ShellBash:
		return Bash(runtime.GOOS, programName), nil
	case ShellFish:
		return Fish(runtime.GOOS, programName), nil
	case ShellZsh:
		return Zsh(runtime.GOOS, programName), nil
	case ShellPowershell:
		return Powershell(runtime.GOOS, programName), nil
	default:
		return nil, fmt.Errorf("unsupported shell %q", shell)
	}
}

func ShellOptions(choice *string) *serpent.Enum {
	return serpent.EnumOf(choice, ShellBash, ShellFish, ShellZsh, ShellPowershell)
}

func DetectUserShell(programName string) (Shell, error) {
	// Attempt to get the SHELL environment variable first
	if shell := os.Getenv("SHELL"); shell != "" {
		return ShellByName(filepath.Base(shell), "")
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
				return ShellByName(filepath.Base(parts[6]), programName) // The shell is typically the 7th field
			}
		}
	}

	return nil, fmt.Errorf("default shell not found")
}

func writeConfig(
	w io.Writer,
	cfgTemplate string,
	programName string,
) error {
	tmpl, err := template.New("script").Parse(cfgTemplate)
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

func InstallShellCompletion(shell Shell) error {
	path, err := shell.InstallPath()
	if err != nil {
		return fmt.Errorf("get install path: %w", err)
	}
	var headerBuf bytes.Buffer
	err = writeConfig(&headerBuf, completionStartTemplate, shell.ProgramName())
	if err != nil {
		return fmt.Errorf("generate header: %w", err)
	}

	var footerBytes bytes.Buffer
	err = writeConfig(&footerBytes, completionEndTemplate, shell.ProgramName())
	if err != nil {
		return fmt.Errorf("generate footer: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	f, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read ssh config failed: %w", err)
	}

	before, after, err := templateConfigSplit(headerBuf.Bytes(), footerBytes.Bytes(), f)
	if err != nil {
		return err
	}

	outBuf := bytes.Buffer{}
	_, _ = outBuf.Write(before)
	if len(before) > 0 {
		_, _ = outBuf.Write([]byte("\n"))
	}
	_, _ = outBuf.Write(headerBuf.Bytes())
	err = shell.WriteCompletion(&outBuf)
	if err != nil {
		return fmt.Errorf("generate completion: %w", err)
	}
	_, _ = outBuf.Write(footerBytes.Bytes())
	_, _ = outBuf.Write([]byte("\n"))
	_, _ = outBuf.Write(after)

	err = atomic.WriteFile(path, &outBuf)
	if err != nil {
		return fmt.Errorf("write completion: %w", err)
	}

	return nil
}

func templateConfigSplit(header, footer, data []byte) (before, after []byte, err error) {
	startCount := bytes.Count(data, header)
	endCount := bytes.Count(data, footer)
	if startCount > 1 || endCount > 1 {
		return nil, nil, fmt.Errorf("Malformed config file: multiple config sections")
	}

	startIndex := bytes.Index(data, header)
	endIndex := bytes.Index(data, footer)
	if startIndex == -1 && endIndex != -1 {
		return data, nil, fmt.Errorf("Malformed config file: missing completion header")
	}
	if startIndex != -1 && endIndex == -1 {
		return data, nil, fmt.Errorf("Malformed config file: missing completion footer")
	}
	if startIndex != -1 && endIndex != -1 {
		if startIndex > endIndex {
			return data, nil, fmt.Errorf("Malformed config file: completion header after footer")
		}
		// Include leading and trailing newline, if present
		start := startIndex
		if start > 0 {
			start--
		}
		end := endIndex + len(footer)
		if end < len(data) {
			end++
		}
		return data[:start], data[end:], nil
	}
	return data, nil, nil
}
