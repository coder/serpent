package serpent

import (
	"strings"

	"golang.org/x/xerrors"
)

type useArg struct {
	// name is the name of the argument.
	// or "bob" in arg "[bob]"
	name string
	// array is true if the argument is an array.
	// or "true" in arg "[bob...]"
	array bool
	// required is true if the argument is required.
	// or "true" in arg "<bob>"
	required bool
}

func parseUse(use string) []useArg {
	words := strings.Fields(use)
	args := make([]useArg, 0, len(words))
	for _, word := range words {
		if len(word) < 2 {
			continue
		}

		isOptional := word[0] == '[' && word[len(word)-1] == ']'
		isRequired := word[0] == '<' && word[len(word)-1] == '>'
		if !isOptional && !isRequired {
			continue
		}

		name := word[1 : len(word)-1]

		const ellipse = "..."
		isArray := strings.HasSuffix(name, ellipse)

		if isArray {
			name = name[:len(name)-len(ellipse)]
		}

		arg := useArg{
			name:     name,
			array:    isArray,
			required: isRequired,
		}

		args = append(args, arg)
	}

	return args
}

// enforceUse returns a middleware that enforces that a command
// was invoked according to its use. It does not validate flags.
func enforceUse() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(inv *Invocation) error {
			if inv.Command.Use == "" {
				return xerrors.Errorf("command has no use")
			}

			return next(inv)
		}
	}
}
