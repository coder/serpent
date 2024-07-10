package completion

import (
	"fmt"
	"io"
	"text/template"
)

const zshCompletionTemplate = `
_{{.Name}}_completions() {
	local -a args completions
	args=("${words[@]:1:$#words}")
	completions=($(COMPLETION_MODE=1 "{{.Name}}" "${args[@]}"))
	compadd -a completions
}

compdef _{{.Name}}_completions {{.Name}}
`

func GenerateZshCompletion(
	w io.Writer,
	rootCmdName string,
) error {
	tmpl, err := template.New("zsh").Parse(zshCompletionTemplate)
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
