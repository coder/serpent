package completion

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
