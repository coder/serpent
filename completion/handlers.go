package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coder/serpent"
)

// FileHandler returns a handler that completes file names, using the
// given filter func, which may be nil.
func FileHandler(filter func(info os.FileInfo) bool) serpent.CompletionHandlerFunc {
	return func(inv *serpent.Invocation) []string {
		var out []string
		_, word := inv.CurWords()

		dir, _ := filepath.Split(word)
		if dir == "" {
			dir = "."
		}
		f, err := os.Open(dir)
		if err != nil {
			return out
		}
		defer f.Close()
		if dir == "." {
			dir = ""
		}

		infos, err := f.Readdir(0)
		if err != nil {
			return out
		}

		for _, info := range infos {
			if filter != nil && !filter(info) {
				continue
			}

			var cur string
			if info.IsDir() {
				cur = fmt.Sprintf("%s%s%c", dir, info.Name(), os.PathSeparator)
			} else {
				cur = fmt.Sprintf("%s%s", dir, info.Name())
			}

			if strings.HasPrefix(cur, word) {
				out = append(out, cur)
			}
		}
		return out
	}
}
