package serpent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseUse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		use      string
		expected []useArg
	}{
		{
			use: "<required> [optional] <requiredArray...> [optionalArray...]",
			expected: []useArg{
				{name: "required", array: false, required: true},
				{name: "optional", array: false, required: false},
				{name: "requiredArray", array: true, required: true},
				{name: "optionalArray", array: true, required: false},
			},
		},
		{
			use:      "<single> [singleOptional]",
			expected: []useArg{{name: "single", array: false, required: true}, {name: "singleOptional", array: false, required: false}},
		},
		{
			use:      "noBrackets noBracketsEither",
			expected: []useArg{},
		},
		{
			use:      "<incomplete [mismatched>",
			expected: []useArg{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.use, func(t *testing.T) {
			result := parseUse(tc.use)
			require.Equal(t, tc.expected, result)
		})
	}
}
