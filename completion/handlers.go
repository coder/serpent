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
		return listFiles(inv.CurWord, filter)
	}
}

// FileListHandler returns a handler that completes a list of comma-separated,
// file names, using the given filter func, which may be nil.
func FileListHandler(filter func(info os.FileInfo) bool) serpent.CompletionHandlerFunc {
	return func(inv *serpent.Invocation) []string {
		curWord := strings.TrimLeft(inv.CurWord, `"`)
		if curWord == "" {
			return listFiles("", filter)
		}
		parts := strings.Split(curWord, ",")
		out := listFiles(parts[len(parts)-1], filter)
		for i, s := range out {
			parts[len(parts)-1] = s
			out[i] = strings.Join(parts, ",")
		}
		return out
	}
}

func listFiles(word string, filter func(info os.FileInfo) bool) []string {
	// Avoid reallocating for each of the first few files we see.
	out := make([]string, 0, 16)

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
