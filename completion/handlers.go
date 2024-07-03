package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coder/serpent"
)

func EnumHandler(choices ...string) serpent.CompletionHandlerFunc {
	return func(inv *serpent.Invocation) []string {
		return choices
	}
}

// FileHandler returns a handler that completes files, using the
// given filter func, which may be nil.
func FileHandler(filter func(info os.FileInfo) bool) serpent.CompletionHandlerFunc {
	return func(inv *serpent.Invocation) []string {
		out := make([]string, 0, 32)
		curWord := inv.CurWord
		dir, _ := filepath.Split(curWord)
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
				cur = fmt.Sprintf("%s%s/", dir, info.Name())
			} else {
				cur = fmt.Sprintf("%s%s", dir, info.Name())
			}

			if strings.HasPrefix(cur, curWord) {
				out = append(out, cur)
			}
		}
		return out
	}
}
