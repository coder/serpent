package completion

import (
	"io"
	"path/filepath"

	home "github.com/mitchellh/go-homedir"
)

type zsh struct {
	goos        string
	programName string
}

var _ Shell = &zsh{}

func Zsh(goos string, programName string) Shell {
	return &zsh{goos: goos, programName: programName}
}

// Name implements Shell.
func (z *zsh) Name() string {
	return "zsh"
}

// InstallPath implements Shell.
func (z *zsh) InstallPath() (string, error) {
	homeDir, err := home.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".zshrc"), nil
}

// WriteCompletion implements Shell.
func (z *zsh) WriteCompletion(w io.Writer) error {
	return configTemplateWriter(w, zshCompletionTemplate, z.programName)
}

// ProgramName implements Shell.
func (z *zsh) ProgramName() string {
	return z.programName
}

const zshCompletionTemplate = `
_{{.Name}}_completions() {
	local -a args completions
	args=("${words[@]:1:$#words}")
	completions=($(COMPLETION_MODE=1 "{{.Name}}" "${args[@]}"))
	compadd -a completions
}
compdef _{{.Name}}_completions {{.Name}}
`
