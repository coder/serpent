package serpent_test

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	serpent "github.com/coder/serpent"
)

func TestOptionSet_ParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("SimpleString", func(t *testing.T) {
		t.Parallel()

		var workspaceName serpent.String

		os := serpent.OptionSet{
			serpent.Option{
				Name:          "Workspace Name",
				Value:         &workspaceName,
				Flag:          "workspace-name",
				FlagShorthand: "n",
			},
		}

		var err error
		err = os.FlagSet().Parse([]string{"--workspace-name", "foo"})
		require.NoError(t, err)
		require.EqualValues(t, "foo", workspaceName)

		err = os.FlagSet().Parse([]string{"-n", "f"})
		require.NoError(t, err)
		require.EqualValues(t, "f", workspaceName)
	})

	t.Run("StringArray", func(t *testing.T) {
		t.Parallel()

		var names serpent.StringArray

		os := serpent.OptionSet{
			serpent.Option{
				Name:          "name",
				Value:         &names,
				Flag:          "name",
				FlagShorthand: "n",
			},
		}

		err := os.SetDefaults()
		require.NoError(t, err)

		err = os.FlagSet().Parse([]string{"--name", "foo", "--name", "bar"})
		require.NoError(t, err)
		require.EqualValues(t, []string{"foo", "bar"}, names)
	})

	t.Run("ExtraFlags", func(t *testing.T) {
		t.Parallel()

		var workspaceName serpent.String

		os := serpent.OptionSet{
			serpent.Option{
				Name:  "Workspace Name",
				Value: &workspaceName,
			},
		}

		err := os.FlagSet().Parse([]string{"--some-unknown", "foo"})
		require.Error(t, err)
	})

	t.Run("RegexValid", func(t *testing.T) {
		t.Parallel()

		var regexpString serpent.Regexp

		os := serpent.OptionSet{
			serpent.Option{
				Name:  "RegexpString",
				Value: &regexpString,
				Flag:  "regexp-string",
			},
		}

		err := os.FlagSet().Parse([]string{"--regexp-string", "$test^"})
		require.NoError(t, err)
	})

	t.Run("RegexInvalid", func(t *testing.T) {
		t.Parallel()

		var regexpString serpent.Regexp

		os := serpent.OptionSet{
			serpent.Option{
				Name:  "RegexpString",
				Value: &regexpString,
				Flag:  "regexp-string",
			},
		}

		err := os.FlagSet().Parse([]string{"--regexp-string", "(("})
		require.Error(t, err)
	})
}

func TestOptionSet_ParseEnv(t *testing.T) {
	t.Parallel()

	t.Run("SimpleString", func(t *testing.T) {
		t.Parallel()

		var workspaceName serpent.String

		os := serpent.OptionSet{
			serpent.Option{
				Name:  "Workspace Name",
				Value: &workspaceName,
				Env:   "WORKSPACE_NAME",
			},
		}

		err := os.ParseEnv([]serpent.EnvVar{
			{Name: "WORKSPACE_NAME", Value: "foo"},
		})
		require.NoError(t, err)
		require.EqualValues(t, "foo", workspaceName)
	})

	t.Run("EmptyValue", func(t *testing.T) {
		t.Parallel()

		var workspaceName serpent.String

		os := serpent.OptionSet{
			serpent.Option{
				Name:    "Workspace Name",
				Value:   &workspaceName,
				Default: "defname",
				Env:     "WORKSPACE_NAME",
			},
		}

		err := os.SetDefaults()
		require.NoError(t, err)

		err = os.ParseEnv(serpent.ParseEnviron([]string{"CODER_WORKSPACE_NAME="}, "CODER_"))
		require.NoError(t, err)
		// An explicitly empty environment variable should override the
		// default value, allowing users to clear a default.
		require.EqualValues(t, "", workspaceName)
	})

	t.Run("StringSlice", func(t *testing.T) {
		t.Parallel()

		var actual serpent.StringArray
		expected := []string{"foo", "bar", "baz"}

		os := serpent.OptionSet{
			serpent.Option{
				Name:  "name",
				Value: &actual,
				Env:   "NAMES",
			},
		}

		err := os.SetDefaults()
		require.NoError(t, err)

		err = os.ParseEnv([]serpent.EnvVar{
			{Name: "NAMES", Value: "foo,bar,baz"},
		})
		require.NoError(t, err)
		require.EqualValues(t, expected, actual)
	})

	t.Run("StructMapStringString", func(t *testing.T) {
		t.Parallel()

		var actual serpent.Struct[map[string]string]
		expected := map[string]string{"foo": "bar", "baz": "zap"}

		os := serpent.OptionSet{
			serpent.Option{
				Name:  "labels",
				Value: &actual,
				Env:   "LABELS",
			},
		}

		err := os.SetDefaults()
		require.NoError(t, err)

		err = os.ParseEnv([]serpent.EnvVar{
			{Name: "LABELS", Value: `{"foo":"bar","baz":"zap"}`},
		})
		require.NoError(t, err)
		require.EqualValues(t, expected, actual.Value)
	})

	t.Run("Homebrew", func(t *testing.T) {
		t.Parallel()

		var agentToken serpent.String

		os := serpent.OptionSet{
			serpent.Option{
				Name:  "Agent Token",
				Value: &agentToken,
				Env:   "AGENT_TOKEN",
			},
		}

		err := os.ParseEnv([]serpent.EnvVar{
			{Name: "HOMEBREW_AGENT_TOKEN", Value: "foo"},
		})
		require.NoError(t, err)
		require.EqualValues(t, "foo", agentToken)
	})
}

func TestOptionSet_JsonMarshal(t *testing.T) {
	t.Parallel()

	// This unit test ensures if the source optionset is missing the option
	// and cannot determine the type, it will not panic. The unmarshal will
	// succeed with a best effort.
	t.Run("MissingSrcOption", func(t *testing.T) {
		t.Parallel()

		var str serpent.String = "something"
		var arr serpent.StringArray = []string{"foo", "bar"}
		opts := serpent.OptionSet{
			serpent.Option{
				Name:  "StringOpt",
				Value: &str,
			},
			serpent.Option{
				Name:  "ArrayOpt",
				Value: &arr,
			},
		}
		data, err := json.Marshal(opts)
		require.NoError(t, err, "marshal option set")

		tgt := serpent.OptionSet{}
		err = json.Unmarshal(data, &tgt)
		require.NoError(t, err, "unmarshal option set")
		for i := range opts {
			compareOptionsExceptValues(t, opts[i], tgt[i])
			require.Empty(t, tgt[i].Value.String(), "unknown value types are empty")
		}
	})

	t.Run("RegexCase", func(t *testing.T) {
		t.Parallel()

		val := serpent.Regexp(*regexp.MustCompile(".*"))
		opts := serpent.OptionSet{
			serpent.Option{
				Name:    "Regex",
				Value:   &val,
				Default: ".*",
			},
		}
		data, err := json.Marshal(opts)
		require.NoError(t, err, "marshal option set")

		var foundVal serpent.Regexp
		newOpts := serpent.OptionSet{
			serpent.Option{
				Name:  "Regex",
				Value: &foundVal,
			},
		}
		err = json.Unmarshal(data, &newOpts)
		require.NoError(t, err, "unmarshal option set")

		require.EqualValues(t, opts[0].Value.String(), newOpts[0].Value.String())
	})
}

func TestOptionSet_DefaultFn(t *testing.T) {
	t.Parallel()
	var verbose serpent.Bool
	var logLevel serpent.String
	var setByEnv serpent.String
	os := serpent.OptionSet{
		{
			Name:    "verbose",
			Env:     "VERBOSE",
			Value:   &verbose,
			Default: "false",
		},
		{
			Name:  "log-level",
			Value: &logLevel,
			DefaultFn: func() string {
				if verbose.Value() {
					return "debug"
				}
				return "info"
			},
		},
		{
			Name:  "set-overridden",
			Value: &setByEnv,
			Env:   "SET_OVERRIDDEN",
			DefaultFn: func() string {
				return "set-by-default-fn"
			},
		},
	}
	// Simulate VERBOSE=true from env
	err := os.ParseEnv([]serpent.EnvVar{{Name: "VERBOSE", Value: "true"}, {Name: "SET_OVERRIDDEN", Value: "set-by-env"}})
	require.NoError(t, err)
	err = os.SetDefaults()
	require.NoError(t, err)
	require.Equal(t, "debug", logLevel.String())
	require.Equal(t, "set-by-env", setByEnv.String())
	require.Equal(t, os.ByName("log-level").Default, "debug")
	require.Equal(t, os.ByName("set-overridden").Value.String(), "set-by-env")
	require.Equal(t, os.ByName("set-overridden").Default, "set-by-default-fn")
}

// TestOptionSet_DefaultFnRace tests the racing behavior of DefaultFns when
// they reference each other
//
// In this test if the DefaultFns are not properly isolated, then defaults for
// the earlier values affect the later ones.
// The DefaultFn does not support referencing other option defaults.
func TestOptionSet_DefaultFnRace(t *testing.T) {
	t.Parallel()
	var (
		def serpent.String
		a   serpent.String
		b   serpent.String
		c   serpent.String
	)
	os := serpent.OptionSet{
		{
			Name:    "default",
			Default: "default", // Even this you cannot use
			Value:   &def,
		},
		{
			Name:      "a",
			Value:     &a,
			DefaultFn: func() string { return def.String() + "a" },
		},
		{
			Name:      "b",
			Value:     &b,
			DefaultFn: func() string { return a.String() + "b" },
		},
		{
			Name:      "c",
			Value:     &c,
			DefaultFn: func() string { return b.String() + "c" },
		},
	}
	err := os.SetDefaults()
	require.NoError(t, err)

	require.Equal(t, a.String(), "a")
	require.Equal(t, os.ByName("a").Default, "a")

	require.Equal(t, b.String(), "b")
	require.Equal(t, os.ByName("b").Default, "b")

	require.Equal(t, c.String(), "c")
	require.Equal(t, os.ByName("c").Default, "c")
}

func compareOptionsExceptValues(t *testing.T, exp, found serpent.Option) {
	t.Helper()

	require.Equalf(t, exp.Name, found.Name, "option name %q", exp.Name)
	require.Equalf(t, exp.Description, found.Description, "option description %q", exp.Name)
	require.Equalf(t, exp.Required, found.Required, "option required %q", exp.Name)
	require.Equalf(t, exp.Flag, found.Flag, "option flag %q", exp.Name)
	require.Equalf(t, exp.FlagShorthand, found.FlagShorthand, "option flag shorthand %q", exp.Name)
	require.Equalf(t, exp.Env, found.Env, "option env %q", exp.Name)
	require.Equalf(t, exp.YAML, found.YAML, "option yaml %q", exp.Name)
	require.Equalf(t, exp.Default, found.Default, "option default %q", exp.Name)
	require.Equalf(t, exp.ValueSource, found.ValueSource, "option value source %q", exp.Name)
	require.Equalf(t, exp.Hidden, found.Hidden, "option hidden %q", exp.Name)
	require.Equalf(t, exp.Annotations, found.Annotations, "option annotations %q", exp.Name)
	require.Equalf(t, exp.Group, found.Group, "option group %q", exp.Name)
	// UseInstead is the same comparison problem, just check the length
	require.Equalf(t, len(exp.UseInstead), len(found.UseInstead), "option use instead %q", exp.Name)
}
