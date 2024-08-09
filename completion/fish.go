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

// UsesOwnFile implements Shell.
func (f *fish) UsesOwnFile() bool {
	return true
}

// Name implements Shell.
func (f *fish) Name() string {
	return "fish"
}

// InstallPath implements Shell.
func (f *fish) InstallPath() (string, error) {
	homeDir, err := home.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config/fish/completions/", f.programName+".fish"), nil
}

// WriteCompletion implements Shell.
func (f *fish) WriteCompletion(w io.Writer) error {
	return generateCompletion(fishCompletionTemplate)(w, f.programName)
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
