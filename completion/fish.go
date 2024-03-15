package completion

import (
	"fmt"
	"io"
	"text/template"

	"github.com/coder/serpent"
)

const fishCompletionTemplate = `
function _{{.Name}}_completions
	# Capture the full command line as an array
	set -l args (commandline -o)

    COMPLETION_MODE=1 $args
end

# Setup Fish to use the function for completions for '{{.Name}}'
complete -c {{.Name}} -f -a '(_{{.Name}}_completions)'

`

func GenerateFishCompletion(
	w io.Writer,
	rootCmd *serpent.Command,
) error {
	tmpl, err := template.New("fish").Parse(fishCompletionTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	err = tmpl.Execute(
		w,
		map[string]string{
			"Name": rootCmd.Name(),
		},
	)
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
