package serpent_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	serpent "github.com/coder/serpent"
	"github.com/coder/serpent/completion"
	"github.com/stretchr/testify/require"
)

func TestCompletion(t *testing.T) {
	t.Parallel()

	cmd := func() *serpent.Command { return sampleCommand(t) }

	t.Run("SubcommandList", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "altfile\nfile\nrequired-flag\ntoupper\n", io.Stdout.String())
	})

	t.Run("SubcommandNoPartial", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("f")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "altfile\nfile\nrequired-flag\ntoupper\n", io.Stdout.String())
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
		i := cmd().Invoke("required-flag", "-")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-array\n--req-bool\n--req-enum\n--req-enum-array\n--req-string\n", io.Stdout.String())
	})

	t.Run("ListFlagsAfterArg", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("altfile", "-")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "doesntexist.go\n--extra\n", io.Stdout.String())
	})

	t.Run("FlagExhaustive", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "--req-bool", "--req-string", "foo bar", "--req-array", "asdf", "--req-array", "qwerty", "-")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-array\n--req-enum\n--req-enum-array\n", io.Stdout.String())
	})

	t.Run("FlagShorthand", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "-b", "-s", "foo bar", "-a", "asdf", "-")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-array\n--req-enum\n--req-enum-array\n", io.Stdout.String())
	})

	t.Run("NoOptDefValueFlag", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("--verbose", "-")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--prefix\n", io.Stdout.String())
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

	t.Run("EnumEqualsOK", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "--req-enum", "--req-enum=")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-enum=foo\n--req-enum=bar\n--req-enum=qux\n", io.Stdout.String())
	})

	t.Run("EnumEqualsBeginQuotesOK", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "--req-enum", "--req-enum=\"")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-enum=foo\n--req-enum=bar\n--req-enum=qux\n", io.Stdout.String())
	})

	t.Run("EnumArrayOK", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "--req-enum-array", "")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "foo\nbar\nqux\n", io.Stdout.String())
	})

	t.Run("EnumArrayEqualsOK", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "--req-enum-array=")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-enum-array=foo\n--req-enum-array=bar\n--req-enum-array=qux\n", io.Stdout.String())
	})

	t.Run("EnumArrayEqualsBeginQuotesOK", func(t *testing.T) {
		t.Parallel()
		i := cmd().Invoke("required-flag", "--req-enum-array=\"")
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, "--req-enum-array=foo\n--req-enum-array=bar\n--req-enum-array=qux\n", io.Stdout.String())
	})

}

func TestFileCompletion(t *testing.T) {
	t.Parallel()

	cmd := func() *serpent.Command { return sampleCommand(t) }

	t.Run("DirOK", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		i := cmd().Invoke("file", tempDir)
		i.Environ.Set(serpent.CompletionModeEnv, "1")
		io := fakeIO(i)
		err := i.Run()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s%c\n", tempDir, os.PathSeparator), io.Stdout.String())
	})

	t.Run("EmptyDirOK", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir() + string(os.PathSeparator)
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
					if strings.HasSuffix(str, string(os.PathSeparator)) {
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

func TestCompletionInstall(t *testing.T) {
	t.Parallel()

	t.Run("InstallingNew", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "fake.sh")
		shell := &fakeShell{baseInstallDir: dir, programName: "fake"}

		err := completion.InstallShellCompletion(shell)
		require.NoError(t, err)
		contents, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, "# ============ BEGIN fake COMPLETION ============\nFAKE_COMPLETION\n# ============ END fake COMPLETION ==============\n", string(contents))
	})

	cases := []struct {
		name     string
		input    []byte
		expected []byte
		errMsg   string
	}{
		{
			name:     "InstallingAppend",
			input:    []byte("FAKE_SCRIPT"),
			expected: []byte("FAKE_SCRIPT\n# ============ BEGIN fake COMPLETION ============\nFAKE_COMPLETION\n# ============ END fake COMPLETION ==============\n"),
		},
		{
			name:     "InstallReplaceBeginning",
			input:    []byte("# ============ BEGIN fake COMPLETION ============\nOLD_COMPLETION\n# ============ END fake COMPLETION ==============\nFAKE_SCRIPT\n"),
			expected: []byte("# ============ BEGIN fake COMPLETION ============\nFAKE_COMPLETION\n# ============ END fake COMPLETION ==============\nFAKE_SCRIPT\n"),
		},
		{
			name:     "InstallReplaceMiddle",
			input:    []byte("FAKE_SCRIPT\n# ============ BEGIN fake COMPLETION ============\nOLD_COMPLETION\n# ============ END fake COMPLETION ==============\nFAKE_SCRIPT\n"),
			expected: []byte("FAKE_SCRIPT\n# ============ BEGIN fake COMPLETION ============\nFAKE_COMPLETION\n# ============ END fake COMPLETION ==============\nFAKE_SCRIPT\n"),
		},
		{
			name:     "InstallReplaceEnd",
			input:    []byte("FAKE_SCRIPT\n# ============ BEGIN fake COMPLETION ============\nOLD_COMPLETION\n# ============ END fake COMPLETION ==============\n"),
			expected: []byte("FAKE_SCRIPT\n# ============ BEGIN fake COMPLETION ============\nFAKE_COMPLETION\n# ============ END fake COMPLETION ==============\n"),
		},
		{
			name:   "InstallNoFooter",
			input:  []byte("FAKE_SCRIPT\n# ============ BEGIN fake COMPLETION ============\nOLD_COMPLETION\n"),
			errMsg: "missing completion footer",
		},
		{
			name:   "InstallNoHeader",
			input:  []byte("OLD_COMPLETION\n# ============ END fake COMPLETION ==============\n"),
			errMsg: "missing completion header",
		},
		{
			name:   "InstallBadOrder",
			input:  []byte("# ============ END fake COMPLETION ==============\nFAKE_COMPLETION\n# ============ BEGIN fake COMPLETION =============="),
			errMsg: "header after footer",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "fake.sh")
			err := os.WriteFile(path, tc.input, 0o644)
			require.NoError(t, err)

			shell := &fakeShell{baseInstallDir: dir, programName: "fake"}
			err = completion.InstallShellCompletion(shell)
			if tc.errMsg != "" {
				require.ErrorContains(t, err, tc.errMsg)
				return
			} else {
				require.NoError(t, err)
				contents, err := os.ReadFile(path)
				require.NoError(t, err)
				require.Equal(t, tc.expected, contents)
			}
		})
	}
}

type fakeShell struct {
	baseInstallDir string
	programName    string
}

func (f *fakeShell) ProgramName() string {
	return f.programName
}

var _ completion.Shell = &fakeShell{}

func (f *fakeShell) InstallPath() (string, error) {
	return filepath.Join(f.baseInstallDir, "fake.sh"), nil
}

func (f *fakeShell) Name() string {
	return "Fake"
}

func (f *fakeShell) WriteCompletion(w io.Writer) error {
	_, err := w.Write([]byte("\nFAKE_COMPLETION\n"))
	return err
}
