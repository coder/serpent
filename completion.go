package serpent

import (
	"github.com/spf13/pflag"
)

// CompletionModeEnv is a special environment variable that is
// set when the command is being run in completion mode.
const CompletionModeEnv = "COMPLETION_MODE"

// IsCompletionMode returns true if the command is being run in completion mode.
func (inv *Invocation) IsCompletionMode() bool {
	_, ok := inv.Environ.Lookup(CompletionModeEnv)
	return ok
}

// DefaultCompletionHandler is a handler that prints all  known flags and
// subcommands that haven't been exhaustively set.
func DefaultCompletionHandler(inv *Invocation) []string {
	var allResps []string
	for _, cmd := range inv.Command.Children {
		allResps = append(allResps, cmd.Name())
	}
	for _, opt := range inv.Command.Options {
		_, isSlice := opt.Value.(pflag.SliceValue)
		if opt.ValueSource == ValueSourceNone ||
			opt.ValueSource == ValueSourceDefault ||
			isSlice {
			allResps = append(allResps, "--"+opt.Flag)
		}
	}
	return allResps
}
