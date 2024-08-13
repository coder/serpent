package completion

import (
	"io"
	"path/filepath"

	home "github.com/mitchellh/go-homedir"
)

type bash struct {
	goos        string
	programName string
}

var _ Shell = &bash{}

func Bash(goos string, programName string) Shell {
	return &bash{goos: goos, programName: programName}
}

// Name implements Shell.
func (b *bash) Name() string {
	return "bash"
}

// InstallPath implements Shell.
func (b *bash) InstallPath() (string, error) {
	homeDir, err := home.Dir()
	if err != nil {
		return "", err
	}
	if b.goos == "darwin" {
		return filepath.Join(homeDir, ".bash_profile"), nil
	}
	return filepath.Join(homeDir, ".bashrc"), nil
}

// WriteCompletion implements Shell.
func (b *bash) WriteCompletion(w io.Writer) error {
	return configTemplateWriter(w, bashCompletionTemplate, b.programName)
}

// ProgramName implements Shell.
func (b *bash) ProgramName() string {
	return b.programName
}

const bashCompletionTemplate = `
_generate_{{.Name}}_completions() {
    # Capture the line excluding the command, and everything after the current word
    local args=("${COMP_WORDS[@]:1:COMP_CWORD}")

    # Set COMPLETION_MODE and call the command with the arguments, capturing the output
    local completions=$(COMPLETION_MODE=1 "{{.Name}}" "${args[@]}")

    # Use the command's output to generate completions for the current word
    COMPREPLY=($(compgen -W "$completions" -- "${COMP_WORDS[COMP_CWORD]}"))

    # Ensure no files are shown, even if there are no matches
    if [ ${#COMPREPLY[@]} -eq 0 ]; then
        COMPREPLY=()
    fi
}
# Setup Bash to use the function for completions for '{{.Name}}'
complete -F _generate_{{.Name}}_completions {{.Name}}
`
