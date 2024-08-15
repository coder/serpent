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

func (b *bash) Name() string {
	return "bash"
}

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

func (b *bash) WriteCompletion(w io.Writer) error {
	return writeConfig(w, bashCompletionTemplate, b.programName)
}

func (b *bash) ProgramName() string {
	return b.programName
}

const bashCompletionTemplate = `
_generate_{{.Name}}_completions() {
    local args=("${COMP_WORDS[@]:1:COMP_CWORD}")

    declare -a output
    mapfile -t output < <(COMPLETION_MODE=1 "{{.Name}}" "${args[@]}")

    declare -a completions
    mapfile -t completions < <( compgen -W "$(printf '%q ' "${output[@]}")" -- "$2" )

    local comp
    COMPREPLY=()
    for comp in "${completions[@]}"; do
        COMPREPLY+=("$(printf "%q" "$comp")")
    done
}
# Setup Bash to use the function for completions for '{{.Name}}'
complete -F _generate_{{.Name}}_completions {{.Name}}
`
