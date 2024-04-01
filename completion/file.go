package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coder/serpent"
)

// FileHandler returns a handler that completes files, using the
// given filter func, which may be nil.
func FileHandler(next serpent.HandlerFunc, filter func(info *os.FileInfo) bool) serpent.HandlerFunc {
	return func(inv *serpent.Invocation) error {
		words := inv.Args

		var curWord string
		if len(words) > 0 {
			curWord = words[len(words)-1]
		}

		dir := filepath.Dir(curWord)
		if dir == "" {
			dir = "."
		}

		f, err := os.Open(dir)
		if err != nil {
			return err
		}
		defer f.Close()

		infos, err := f.Readdir(0)
		if err != nil {
			return err
		}

		for _, info := range infos {
			if filter != nil && !filter(&info) {
				continue
			}

			if !strings.HasPrefix(info.Name(), curWord) {
				continue
			}

			fmt.Fprintf(inv.Stdout, "%s\n", info.Name())
		}

		return next(inv)
	}
}
