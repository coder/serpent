package completion

import (
	"fmt"
	"io"
	"text/template"
)

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

func GenerateFishCompletion(
	w io.Writer,
	rootCmdName string,
) error {
	tmpl, err := template.New("fish").Parse(fishCompletionTemplate)
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
