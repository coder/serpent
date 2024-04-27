package serpent

import (
	"fmt"
)

// CompletionModeEnv is a special environment variable that is
// set when the command is being run in completion mode.
const CompletionModeEnv = "COMPLETION_MODE"

// IsCompletionMode returns true if the command is being run in completion mode.
func (inv *Invocation) IsCompletionMode() bool {
	_, ok := inv.Environ.Lookup(CompletionModeEnv)
	return ok
}

// DefaultCompletionHandler returns a handler that prints all
// known flags and subcommands.
func DefaultCompletionHandler(next HandlerFunc) HandlerFunc {
	return func(inv *Invocation) error {
		// words := inv.Args

		// var curWord string
		// if len(words) > 0 {
		// 	curWord = words[len(words)-1]
		// }

		var allResps []string
		for _, cmd := range inv.Command.Children {
			allResps = append(allResps, cmd.Name())
		}

		for _, opt := range inv.Command.Options {
			allResps = append(allResps, "--"+opt.Flag)
		}

		for _, resp := range allResps {
			// if !strings.HasPrefix(resp, curWord) {
			// 	continue
			// }

			fmt.Fprintf(inv.Stdout, "%s\n", resp)
		}
		return next(inv)
	}
}
