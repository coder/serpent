package completion

import (
	"io"
	"path/filepath"

	home "github.com/mitchellh/go-homedir"
)

type fish struct {
	goos        string
	programName string
}

var _ Shell = &fish{}

func Fish(goos string, programName string) Shell {
	return &fish{goos: goos, programName: programName}
}

func (f *fish) Name() string {
	return "fish"
}

func (f *fish) InstallPath() (string, error) {
	homeDir, err := home.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config/fish/completions/", f.programName+".fish"), nil
}

func (f *fish) WriteCompletion(w io.Writer) error {
	return writeConfig(w, fishCompletionTemplate, f.programName)
}

func (f *fish) ProgramName() string {
	return f.programName
}

const fishCompletionTemplate = `
function _{{.Name}}_completions
	# Capture the full command line as an array
	set -l args (commandline -opc)
	set -l current (commandline -ct)
    COMPLETION_MODE=1 $args $current
end

# Setup Fish to use the function for completions for '{{.Name}}'
complete -c {{.Name}} -f -a '(_{{.Name}}_completions)'
`
