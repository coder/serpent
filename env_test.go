package serpent_test

import (
	"reflect"
	"testing"

	serpent "github.com/coder/serpent"
)

func TestFilterNamePrefix(t *testing.T) {
	t.Parallel()
	type args struct {
		environ []string
		prefix  string
	}
	tests := []struct {
		name string
		args args
		want serpent.Environ
	}{
		{"empty", args{[]string{}, "SHIRE"}, nil},
		{
			"ONE",
			args{
				[]string{
					"SHIRE_BRANDYBUCK=hmm",
				},
				"SHIRE_",
			},
			[]serpent.EnvVar{
				{Name: "BRANDYBUCK", Value: "hmm"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := serpent.ParseEnviron(tt.args.environ, tt.args.prefix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterNamePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
