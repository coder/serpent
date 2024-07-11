package completion

const zshCompletionTemplate = `
_{{.Name}}_completions() {
	local -a args completions
	args=("${words[@]:1:$#words}")
	completions=($(COMPLETION_MODE=1 "{{.Name}}" "${args[@]}"))
	compadd -a completions
}

compdef _{{.Name}}_completions {{.Name}}
`
