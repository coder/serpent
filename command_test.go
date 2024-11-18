package serpent_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	serpent "github.com/coder/serpent"
	"github.com/coder/serpent/completion"
)

// ioBufs is the standard input, output, and error for a command.
type ioBufs struct {
	Stdin  bytes.Buffer
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

// fakeIO sets Stdin, Stdout, and Stderr to buffers.
func fakeIO(i *serpent.Invocation) *ioBufs {
	var b ioBufs
	i.Stdout = &b.Stdout
	i.Stderr = &b.Stderr
	i.Stdin = &b.Stdin
	return &b
}

func sampleCommand(t *testing.T) *serpent.Command {
	t.Helper()
	var (
		verbose    bool
		lower      bool
		prefix     string
		reqBool    bool
		reqStr     string
		reqArr     []string
		reqEnumArr []string
		fileArr    []string
		enumStr    string
	)
	enumChoices := []string{"foo", "bar", "qux"}
	return &serpent.Command{
		Use: "root [subcommand]",
		Options: serpent.OptionSet{
			serpent.Option{
				Name:    "verbose",
				Flag:    "verbose",
				Default: "false",
				Value:   serpent.BoolOf(&verbose),
			},
			serpent.Option{
				Name:  "verbose-old",
				Flag:  "verbode-old",
				Value: serpent.BoolOf(&verbose),
			},
			serpent.Option{
				Name:  "prefix",
				Flag:  "prefix",
				Value: serpent.StringOf(&prefix),
			},
		},
		Children: []*serpent.Command{
			{
				Use:   "required-flag --req-bool=true --req-string=foo",
				Short: "Example with required flags",
				Options: serpent.OptionSet{
					serpent.Option{
						Name:          "req-bool",
						Flag:          "req-bool",
						FlagShorthand: "b",
						Value:         serpent.BoolOf(&reqBool),
						Required:      true,
					},
					serpent.Option{
						Name:          "req-string",
						Flag:          "req-string",
						FlagShorthand: "s",
						Value: serpent.Validate(serpent.StringOf(&reqStr), func(value *serpent.String) error {
							ok := strings.Contains(value.String(), " ")
							if !ok {
								return xerrors.Errorf("string must contain a space")
							}
							return nil
						}),
						Required: true,
					},
					serpent.Option{
						Name:  "req-enum",
						Flag:  "req-enum",
						Value: serpent.EnumOf(&enumStr, enumChoices...),
					},
					serpent.Option{
						Name:          "req-array",
						Flag:          "req-array",
						FlagShorthand: "a",
						Value:         serpent.StringArrayOf(&reqArr),
					},
					serpent.Option{
						Name:  "req-enum-array",
						Flag:  "req-enum-array",
						Value: serpent.EnumArrayOf(&reqEnumArr, enumChoices...),
					},
				},
				HelpHandler: func(i *serpent.Invocation) error {
					_, _ = i.Stdout.Write([]byte("help text.png"))
					return nil
				},
				Handler: func(i *serpent.Invocation) error {
					_, _ = i.Stdout.Write([]byte(fmt.Sprintf("%s-%t", reqStr, reqBool)))
					return nil
				},
			},
			{
				Use:   "toupper [word]",
				Short: "Converts a word to upper case",
				Middleware: serpent.Chain(
					serpent.RequireNArgs(1),
				),
				Aliases: []string{"up"},
				Options: serpent.OptionSet{
					serpent.Option{
						Name:  "lower",
						Flag:  "lower",
						Value: serpent.BoolOf(&lower),
					},
				},
				Handler: func(i *serpent.Invocation) error {
					_, _ = i.Stdout.Write([]byte(prefix))
					w := i.Args[0]
					if lower {
						w = strings.ToLower(w)
					} else {
						w = strings.ToUpper(w)
					}
					_, _ = i.Stdout.Write(
						[]byte(
							w,
						),
					)
					if verbose {
						_, _ = i.Stdout.Write([]byte("!!!"))
					}
					return nil
				},
			},
			{
				Use: "file <file>",
				Handler: func(inv *serpent.Invocation) error {
					return nil
				},
				CompletionHandler: completion.FileHandler(func(info os.FileInfo) bool {
					return true
				}),
				Middleware: serpent.RequireNArgs(1),
			},
			{
				Use: "altfile",
				Handler: func(inv *serpent.Invocation) error {
					return nil
				},
				Options: serpent.OptionSet{
					{
						Name:        "extra",
						Flag:        "extra",
						Description: "Extra files.",
						Value:       serpent.StringArrayOf(&fileArr),
					},
				},
				CompletionHandler: func(i *serpent.Invocation) []string {
					return []string{"doesntexist.go"}
				},
			},
		},
	}
}

func TestCommand(t *testing.T) {
	t.Parallel()

	cmd := func() *serpent.Command { return sampleCommand(t) }

	t.Run("SimpleOK", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("toupper", "hello")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "HELLO", io.Stdout.String())
	})

	t.Run("Alias", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"up", "hello",
		)
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)

		require.Equal(t, "HELLO", io.Stdout.String())
	})

	t.Run("BadArgs", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"toupper",
		)
		io := fakeIO(i)
		err := i.Run()
		require.Empty(t, io.Stdout.String())
		require.Error(t, err)
	})

	t.Run("NoSubcommand", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"na",
		)
		io := fakeIO(i)
		err := i.Run()
		require.Error(t, err)
		require.Contains(t, io.Stderr.String(), "unknown subcommand")
	})

	t.Run("UnknownFlags", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"toupper", "--unknown",
		)
		io := fakeIO(i)
		err := i.Run()
		require.Empty(t, io.Stdout.String())
		require.Error(t, err)
	})

	t.Run("Verbose", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"--verbose", "toupper", "hello",
		)
		io := fakeIO(i)
		require.NoError(t, i.Run())
		require.Equal(t, "HELLO!!!", io.Stdout.String())
	})

	t.Run("Verbose=", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"--verbose=true", "toupper", "hello",
		)
		io := fakeIO(i)
		require.NoError(t, i.Run())
		require.Equal(t, "HELLO!!!", io.Stdout.String())
	})

	t.Run("PrefixSpace", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"--prefix", "conv: ", "toupper", "hello",
		)
		io := fakeIO(i)
		require.NoError(t, i.Run())
		require.Equal(t, "conv: HELLO", io.Stdout.String())
	})

	t.Run("GlobalFlagsAnywhere", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"toupper", "--prefix", "conv: ", "hello", "--verbose",
		)
		io := fakeIO(i)
		require.NoError(t, i.Run())
		require.Equal(t, "conv: HELLO!!!", io.Stdout.String())
	})

	t.Run("LowerVerbose", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"toupper", "--verbose", "hello", "--lower",
		)
		io := fakeIO(i)
		require.NoError(t, i.Run())
		require.Equal(t, "hello!!!", io.Stdout.String())
	})

	t.Run("ParsedFlags", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"toupper", "--verbose", "hello", "--lower",
		)
		_ = fakeIO(i)
		require.NoError(t, i.Run())
		require.Equal(t,
			"true",
			i.ParsedFlags().Lookup("verbose").Value.String(),
		)
	})

	t.Run("NoDeepChild", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"root", "level", "level", "toupper", "--verbose", "hello", "--lower",
		)
		fio := fakeIO(i)
		require.Error(t, i.Run(), fio.Stdout.String())
	})

	t.Run("RequiredFlagsMissing", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"required-flag",
		)
		fio := fakeIO(i)
		err := i.Run()
		require.Error(t, err, fio.Stdout.String())
		require.ErrorContains(t, err, "Missing values")
	})

	t.Run("RequiredFlagsMissingWithHelp", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"required-flag",
			"--help",
		)
		fio := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Contains(t, fio.Stdout.String(), "help text.png")
	})

	t.Run("RequiredFlagsMissingBool", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"required-flag", "--req-string", "foo bar",
		)
		fio := fakeIO(i)
		err := i.Run()
		require.Error(t, err, fio.Stdout.String())
		require.ErrorContains(t, err, "Missing values for the required flags: req-bool")
	})

	t.Run("RequiredFlagsMissingString", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"required-flag", "--req-bool", "true",
		)
		fio := fakeIO(i)
		err := i.Run()
		require.Error(t, err, fio.Stdout.String())
		require.ErrorContains(t, err, "Missing values for the required flags: req-string")
	})

	t.Run("RequiredFlagsInvalid", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"required-flag", "--req-string", "nospace",
		)
		fio := fakeIO(i)
		err := i.Run()
		require.Error(t, err, fio.Stdout.String())
		require.ErrorContains(t, err, "string must contain a space")
	})

	t.Run("RequiredFlagsOK", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke(
			"required-flag", "--req-bool", "true", "--req-string", "foo bar",
		)
		fio := fakeIO(i)
		err := i.Run()
		require.NoError(t, err, fio.Stdout.String())
	})

	t.Run("DeprecatedCommand", func(t *testing.T) {
		t.Parallel()

		deprecatedCmd := &serpent.Command{
			Use:        "deprecated-cmd",
			Deprecated: "This command is deprecated and will be removed in the future.",
			Handler: func(i *serpent.Invocation) error {
				_, _ = i.Stdout.Write([]byte("Running deprecated command"))
				return nil
			},
		}

		i := deprecatedCmd.Invoke()
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		expectedWarning := fmt.Sprintf("WARNING: %q is deprecated!. %s\n", deprecatedCmd.Use, deprecatedCmd.Deprecated)
		require.Equal(t, io.Stderr.String(), expectedWarning)
		require.Contains(t, io.Stdout.String(), "Running deprecated command")
	})
}

func TestCommand_DeepNest(t *testing.T) {
	t.Parallel()
	cmd := &serpent.Command{
		Use: "1",
		Children: []*serpent.Command{
			{
				Use: "2",
				Children: []*serpent.Command{
					{
						Use: "3",
						Handler: func(i *serpent.Invocation) error {
							_, _ = i.Stdout.Write([]byte("3"))
							return nil
						},
					},
				},
			},
		},
	}
	inv := cmd.Invoke("2", "3")
	stdio := fakeIO(inv)
	err := inv.Run()
	require.NoError(t, err)
	require.Equal(t, "3", stdio.Stdout.String())
}

func TestCommand_FlagOverride(t *testing.T) {
	t.Parallel()
	var flag string

	cmd := &serpent.Command{
		Use: "1",
		Options: serpent.OptionSet{
			{
				Name:  "flag",
				Flag:  "f",
				Value: serpent.DiscardValue,
			},
		},
		Children: []*serpent.Command{
			{
				Use: "2",
				Options: serpent.OptionSet{
					{
						Name:  "flag",
						Flag:  "f",
						Value: serpent.StringOf(&flag),
					},
				},
				Handler: func(i *serpent.Invocation) error {
					return nil
				},
			},
		},
	}

	err := cmd.Invoke("2", "--f", "mhmm").Run()
	require.NoError(t, err)

	require.Equal(t, "mhmm", flag)
}

func TestCommand_MiddlewareOrder(t *testing.T) {
	t.Parallel()

	mw := func(letter string) serpent.MiddlewareFunc {
		return func(next serpent.HandlerFunc) serpent.HandlerFunc {
			return (func(i *serpent.Invocation) error {
				_, _ = i.Stdout.Write([]byte(letter))
				return next(i)
			})
		}
	}

	cmd := &serpent.Command{
		Use:   "toupper [word]",
		Short: "Converts a word to upper case",
		Middleware: serpent.Chain(
			mw("A"),
			mw("B"),
			mw("C"),
		),
		Handler: (func(i *serpent.Invocation) error {
			return nil
		}),
	}

	i := cmd.Invoke(
		"hello", "world",
	)
	io := fakeIO(i)
	require.NoError(t, i.Run())
	require.Equal(t, "ABC", io.Stdout.String())
}

func TestCommand_RawArgs(t *testing.T) {
	t.Parallel()

	cmd := func() *serpent.Command {
		return &serpent.Command{
			Use: "root",
			Options: serpent.OptionSet{
				{
					Name:  "password",
					Flag:  "password",
					Value: serpent.StringOf(new(string)),
				},
			},
			Children: []*serpent.Command{
				{
					Use:     "sushi <args...>",
					Short:   "Throws back raw output",
					RawArgs: true,
					Handler: (func(i *serpent.Invocation) error {
						if v := i.ParsedFlags().Lookup("password").Value.String(); v != "codershack" {
							return xerrors.Errorf("password %q is wrong!", v)
						}
						_, _ = i.Stdout.Write([]byte(strings.Join(i.Args, " ")))
						return nil
					}),
				},
			},
		}
	}

	t.Run("OK", func(t *testing.T) {
		// Flag parsed before the raw arg command should still work.
		t.Parallel()

		i := cmd().Invoke(
			"--password", "codershack", "sushi", "hello", "--verbose", "world",
		)
		io := fakeIO(i)
		require.NoError(t, i.Run())
		require.Equal(t, "hello --verbose world", io.Stdout.String())
	})

	t.Run("BadFlag", func(t *testing.T) {
		// Verbose before the raw arg command should fail.
		t.Parallel()

		i := cmd().Invoke(
			"--password", "codershack", "--verbose", "sushi", "hello", "world",
		)
		io := fakeIO(i)
		require.Error(t, i.Run())
		require.Empty(t, io.Stdout.String())
	})

	t.Run("NoPassword", func(t *testing.T) {
		// Flag parsed before the raw arg command should still work.
		t.Parallel()
		i := cmd().Invoke(
			"sushi", "hello", "--verbose", "world",
		)
		_ = fakeIO(i)
		require.Error(t, i.Run())
	})
}

func TestCommand_RootRaw(t *testing.T) {
	t.Parallel()
	cmd := &serpent.Command{
		RawArgs: true,
		Handler: func(i *serpent.Invocation) error {
			_, _ = i.Stdout.Write([]byte(strings.Join(i.Args, " ")))
			return nil
		},
	}

	inv := cmd.Invoke("hello", "--verbose", "--friendly")
	stdio := fakeIO(inv)
	err := inv.Run()
	require.NoError(t, err)

	require.Equal(t, "hello --verbose --friendly", stdio.Stdout.String())
}

func TestCommand_HyphenHyphen(t *testing.T) {
	t.Parallel()
	var verbose bool
	cmd := &serpent.Command{
		Handler: (func(i *serpent.Invocation) error {
			_, _ = i.Stdout.Write([]byte(strings.Join(i.Args, " ")))
			if verbose {
				return xerrors.New("verbose should not be true because flag after --")
			}
			return nil
		}),
		Options: serpent.OptionSet{
			{
				Name:  "verbose",
				Flag:  "verbose",
				Value: serpent.BoolOf(&verbose),
			},
		},
	}

	inv := cmd.Invoke("--", "--verbose", "--friendly")
	stdio := fakeIO(inv)
	err := inv.Run()
	require.NoError(t, err)

	require.Equal(t, "--verbose --friendly", stdio.Stdout.String())
}

func TestCommand_ContextCancels(t *testing.T) {
	t.Parallel()

	var gotCtx context.Context

	cmd := &serpent.Command{
		Handler: (func(i *serpent.Invocation) error {
			gotCtx = i.Context()
			if err := gotCtx.Err(); err != nil {
				return xerrors.Errorf("unexpected context error: %w", i.Context().Err())
			}
			return nil
		}),
	}

	err := cmd.Invoke().Run()
	require.NoError(t, err)

	require.Error(t, gotCtx.Err())
}

func TestCommand_Help(t *testing.T) {
	t.Parallel()

	cmd := func() *serpent.Command {
		return &serpent.Command{
			Use: "root",
			HelpHandler: (func(i *serpent.Invocation) error {
				_, _ = i.Stdout.Write([]byte("abdracadabra"))
				return nil
			}),
			Handler: (func(i *serpent.Invocation) error {
				return xerrors.New("should not be called")
			}),
		}
	}

	t.Run("DefaultHandler", func(t *testing.T) {
		t.Parallel()

		c := cmd()
		c.HelpHandler = nil
		err := c.Invoke("--help").Run()
		require.NoError(t, err)
	})

	t.Run("Long", func(t *testing.T) {
		t.Parallel()

		inv := cmd().Invoke("--help")
		stdio := fakeIO(inv)
		err := inv.Run()
		require.NoError(t, err)

		require.Contains(t, stdio.Stdout.String(), "abdracadabra")
	})

	t.Run("Short", func(t *testing.T) {
		t.Parallel()

		inv := cmd().Invoke("-h")
		stdio := fakeIO(inv)
		err := inv.Run()
		require.NoError(t, err)

		require.Contains(t, stdio.Stdout.String(), "abdracadabra")
	})
}

func TestCommand_SliceFlags(t *testing.T) {
	t.Parallel()

	cmd := func(want ...string) *serpent.Command {
		var got []string
		return &serpent.Command{
			Use: "root",
			Options: serpent.OptionSet{
				{
					Name:    "arr",
					Flag:    "arr",
					Default: "bad,bad,bad",
					Value:   serpent.StringArrayOf(&got),
				},
			},
			Handler: (func(i *serpent.Invocation) error {
				require.Equal(t, want, got)
				return nil
			}),
		}
	}

	err := cmd("good", "good", "good").Invoke("--arr", "good", "--arr", "good", "--arr", "good").Run()
	require.NoError(t, err)

	err = cmd("bad", "bad", "bad").Invoke().Run()
	require.NoError(t, err)
}

func TestCommand_EmptySlice(t *testing.T) {
	t.Parallel()

	cmd := func(want ...string) *serpent.Command {
		var got []string
		return &serpent.Command{
			Use: "root",
			Options: serpent.OptionSet{
				{
					Name:    "arr",
					Flag:    "arr",
					Default: "def,def,def",
					Env:     "ARR",
					Value:   serpent.StringArrayOf(&got),
				},
			},
			Handler: (func(i *serpent.Invocation) error {
				require.Equal(t, want, got)
				return nil
			}),
		}
	}

	// Base-case, uses default.
	err := cmd("def", "def", "def").Invoke().Run()
	require.NoError(t, err)

	// Empty-env uses default, too.
	inv := cmd("def", "def", "def").Invoke()
	inv.Environ.Set("ARR", "")
	require.NoError(t, err)

	// Reset to nothing at all via flag.
	inv = cmd().Invoke("--arr", "")
	inv.Environ.Set("ARR", "cant see")
	err = inv.Run()
	require.NoError(t, err)

	// Reset to a specific value with flag.
	inv = cmd("great").Invoke("--arr", "great")
	inv.Environ.Set("ARR", "")
	err = inv.Run()
	require.NoError(t, err)
}

func TestCommand_DefaultsOverride(t *testing.T) {
	t.Parallel()

	test := func(name string, want string, fn func(t *testing.T, inv *serpent.Invocation)) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var (
				got    string
				config serpent.YAMLConfigPath
			)
			cmd := &serpent.Command{
				Options: serpent.OptionSet{
					{
						Name:    "url",
						Flag:    "url",
						Default: "def.com",
						Env:     "URL",
						Value:   serpent.StringOf(&got),
						YAML:    "url",
					},
					{
						Name:  "url-deprecated",
						Flag:  "url-deprecated",
						Env:   "URL_DEPRECATED",
						Value: serpent.StringOf(&got),
					},
					{
						Name:    "config",
						Flag:    "config",
						Default: "",
						Value:   &config,
					},
				},
				Handler: (func(i *serpent.Invocation) error {
					_, _ = fmt.Fprintf(i.Stdout, "%s", got)
					return nil
				}),
			}

			inv := cmd.Invoke()
			stdio := fakeIO(inv)
			fn(t, inv)
			err := inv.Run()
			require.NoError(t, err)
			require.Equal(t, want, stdio.Stdout.String())
		})
	}

	test("DefaultOverNothing", "def.com", func(t *testing.T, inv *serpent.Invocation) {})

	test("FlagOverDefault", "good.com", func(t *testing.T, inv *serpent.Invocation) {
		inv.Args = []string{"--url", "good.com"}
	})

	test("EnvOverDefault", "good.com", func(t *testing.T, inv *serpent.Invocation) {
		inv.Environ.Set("URL", "good.com")
	})

	test("FlagOverEnv", "good.com", func(t *testing.T, inv *serpent.Invocation) {
		inv.Environ.Set("URL", "bad.com")
		inv.Args = []string{"--url", "good.com"}
	})

	test("FlagOverYAML", "good.com", func(t *testing.T, inv *serpent.Invocation) {
		fi, err := os.CreateTemp(t.TempDir(), "config.yaml")
		require.NoError(t, err)
		defer fi.Close()

		_, err = fi.WriteString("url: bad.com")
		require.NoError(t, err)

		inv.Args = []string{"--config", fi.Name(), "--url", "good.com"}
	})

	test("EnvOverYAML", "good.com", func(t *testing.T, inv *serpent.Invocation) {
		fi, err := os.CreateTemp(t.TempDir(), "config.yaml")
		require.NoError(t, err)
		defer fi.Close()

		_, err = fi.WriteString("url: bad.com")
		require.NoError(t, err)

		inv.Environ.Set("URL", "good.com")
	})

	test("YAMLOverDefault", "good.com", func(t *testing.T, inv *serpent.Invocation) {
		fi, err := os.CreateTemp(t.TempDir(), "config.yaml")
		require.NoError(t, err)
		defer fi.Close()

		_, err = fi.WriteString("url: good.com")
		require.NoError(t, err)

		inv.Args = []string{"--config", fi.Name()}
	})

	test("AltFlagOverDefault", "good.com", func(t *testing.T, inv *serpent.Invocation) {
		inv.Args = []string{"--url-deprecated", "good.com"}
	})
}

func TestCommand_OptionsWithSharedValue(t *testing.T) {
	t.Parallel()

	var got string
	makeCmd := func(def, altDef string) *serpent.Command {
		got = ""
		return &serpent.Command{
			Options: serpent.OptionSet{
				{
					Name:    "url",
					Flag:    "url",
					Env:     "URL",
					Default: def,
					Value:   serpent.StringOf(&got),
				},
				{
					Name:    "alt-url",
					Flag:    "alt-url",
					Env:     "ALT_URL",
					Default: altDef,
					Value:   serpent.StringOf(&got),
				},
			},
			Handler: (func(i *serpent.Invocation) error {
				return nil
			}),
		}
	}

	// Check proper value propagation.
	err := makeCmd("def.com", "def.com").Invoke().Run()
	require.NoError(t, err, "default values are same")
	require.Equal(t, "def.com", got)

	err = makeCmd("def.com", "").Invoke().Run()
	require.NoError(t, err, "other default value is empty")
	require.Equal(t, "def.com", got)

	err = makeCmd("def.com", "").Invoke("--url", "sup").Run()
	require.NoError(t, err)
	require.Equal(t, "sup", got)

	err = makeCmd("def.com", "").Invoke("--alt-url", "hup").Run()
	require.NoError(t, err)
	require.Equal(t, "hup", got)

	// Both flags are given, last wins.
	err = makeCmd("def.com", "").Invoke("--url", "sup", "--alt-url", "hup").Run()
	require.NoError(t, err)
	require.Equal(t, "hup", got)

	// Both flags are given, last wins #2.
	err = makeCmd("", "def.com").Invoke("--alt-url", "hup", "--url", "sup").Run()
	require.NoError(t, err)
	require.Equal(t, "sup", got)

	// Both flags are given, option type priority wins.
	inv := makeCmd("def.com", "").Invoke("--alt-url", "hup")
	inv.Environ.Set("URL", "sup")
	err = inv.Run()
	require.NoError(t, err)
	require.Equal(t, "hup", got)

	// Both flags are given, option type priority wins #2.
	inv = makeCmd("", "def.com").Invoke("--url", "sup")
	inv.Environ.Set("ALT_URL", "hup")
	err = inv.Run()
	require.NoError(t, err)
	require.Equal(t, "sup", got)

	// Catch invalid configuration.
	err = makeCmd("def.com", "alt-def.com").Invoke().Run()
	require.Error(t, err, "default values are different")
}
