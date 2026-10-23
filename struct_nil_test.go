package serpent_test

import (
	"testing"

	"github.com/coder/serpent"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestStructAnyYAML(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, input string
		want        any
	}{
		{"mapping", "name: coder", map[string]interface{}{"name": "coder"}},
		{"sequence", "[one, two]", []interface{}{"one", "two"}},
		{"scalar", "hello", "hello"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var value serpent.Struct[any]
			require.NotPanics(t, func() { require.NoError(t, yaml.Unmarshal([]byte(tt.input), &value)) })
			require.Equal(t, tt.want, value.Value)
		})
	}
}

func TestStructSliceEmptyYAML(t *testing.T) {
	t.Parallel()
	value := serpent.Struct[[]string]{Value: []string{"previous"}}
	require.NoError(t, yaml.Unmarshal([]byte("[]"), &value))
	require.Nil(t, value.Value)
}
