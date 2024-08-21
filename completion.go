package serpent

import (
	"strings"

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

// DefaultCompletionHandler is a handler that prints all the subcommands, or
// all the options that haven't been exhaustively set, if the current word
// starts with a dash.
func DefaultCompletionHandler(inv *Invocation) []string {
	_, cur := inv.CurWords()
	var allResps []string
	if strings.HasPrefix(cur, "-") {
		for _, opt := range inv.Command.Options {
			_, isSlice := opt.Value.(pflag.SliceValue)
			if opt.ValueSource == ValueSourceNone ||
				opt.ValueSource == ValueSourceDefault ||
				isSlice {
				allResps = append(allResps, "--"+opt.Flag)
			}
		}
	} else {
		for _, cmd := range inv.Command.Children {
			allResps = append(allResps, cmd.Name())
		}
	}
	return allResps
}
