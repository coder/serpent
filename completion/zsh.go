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

func (z *zsh) Name() string {
	return "zsh"
}

func (z *zsh) InstallPath() (string, error) {
	homeDir, err := home.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".zshrc"), nil
}

func (z *zsh) WriteCompletion(w io.Writer) error {
	return writeConfig(w, zshCompletionTemplate, z.programName)
}

func (z *zsh) ProgramName() string {
	return z.programName
}

const zshCompletionTemplate = `
_{{.Name}}_completions() {
	local -a args completions
	args=("${words[@]:1:$#words}")
	completions=(${(f)"$(COMPLETION_MODE=1 "{{.Name}}" "${args[@]}")"})
	compadd -a completions
}
compdef _{{.Name}}_completions {{.Name}}
`
