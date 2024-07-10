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
		return ListFiles(inv.CurWord, filter)
	}
}

func FileListHandler(filter func(info os.FileInfo) bool) serpent.CompletionHandlerFunc {
	return func(inv *serpent.Invocation) []string {
		curWord := strings.TrimLeft(inv.CurWord, `"`)
		if curWord == "" {
			return ListFiles("", filter)
		}
		parts := strings.Split(curWord, ",")
		out := ListFiles(parts[len(parts)-1], filter)
		// prepend := strings.Join(parts[:len(parts)-1], ",")
		for i, s := range out {
			parts[len(parts)-1] = s
			out[i] = strings.Join(parts, ",")
		}
		return out
	}
}

func ListFiles(word string, filter func(info os.FileInfo) bool) []string {
	out := make([]string, 0, 32)

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
			cur = fmt.Sprintf("%s%s/", dir, info.Name())
		} else {
			cur = fmt.Sprintf("%s%s", dir, info.Name())
		}

		if strings.HasPrefix(cur, word) {
			out = append(out, cur)
		}
	}
	return out
}
