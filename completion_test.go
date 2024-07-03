package serpent_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	serpent "github.com/coder/serpent"
	"github.com/stretchr/testify/require"
)

func TestCompletion(t *testing.T) {
	t.Parallel()

	cmd := func() *serpent.Command { return SampleCommand(t) }

	t.Run("SubcommandList", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "file\nrequired-flag\ntoupper\n--prefix\n--verbose\n", io.Stdout.String())
	})

	t.Run("SubcommandComplete", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "required-flag\n", io.Stdout.String())
	})

	t.Run("ListFlags", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-bool\n--req-enum\n--req-string\n", io.Stdout.String())
	})

	t.Run("FlagExhaustive", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "--req-bool", "--req-string", "foo bar")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-enum\n", io.Stdout.String())
	})

	t.Run("EnumOK", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "--req-enum", "")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "foo\nbar\nqux\n", io.Stdout.String())
	})
}

func TestFileCompletion(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	cmd := func() *serpent.Command { return SampleCommand(t) }

	t.Run("DirOK", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		i := cmd().Invoke("file", tempDir)
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, tempDir+"/\n", io.Stdout.String())
	})

	t.Run("EmptyDirOK", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir() + "/"
		i := cmd().Invoke("file", tempDir)
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "", io.Stdout.String())
	})

	cases := []struct {
		name     string
		realPath string
		paths    []string
	}{
		{
			name:     "CurDirOK",
			realPath: ".",
			paths:    []string{"", "./", "././"},
		},
		{
			name:     "PrevDirOK",
			realPath: "..",
			paths:    []string{"../", ".././"},
		},
		{
			name:     "RootOK",
			realPath: "/",
			paths:    []string{"/", "/././"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, path := range tc.paths {
				i := cmd().Invoke("file", path)
				i.Environ.Set(serpent.CompletionModeEnv, "1")
				io := fakeIO(i)
				err := i.Run()
				require.NoError(t, err)
				output := strings.Split(io.Stdout.String(), "\n")
				output = output[:len(output)-1]
				for _, str := range output {
					if strings.HasSuffix(str, "/") {
						require.DirExists(t, str)
					} else {
						require.FileExists(t, str)
					}
				}
				files, err := os.ReadDir(tc.realPath)
				require.NoError(t, err)
				require.Equal(t, len(files), len(output))
			}
		})
	}
}
