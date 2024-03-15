package main

import (
	"os"
	"strings"

	"github.com/coder/serpent"
)

func main() {
	var upper bool
	cmd := serpent.Command{
		Use:   "echo <text>",
		Short: "Prints the given text to the console.",
		Options: serpent.OptionSet{
			{
				Name:        "upper",
				Value:       serpent.BoolOf(&upper),
				Flag:        "upper",
				Description: "Prints the text in upper case.",
			},
		},
		Handler: func(inv *serpent.Invocation) error {
			if len(inv.Args) == 0 {
				inv.Stderr.Write([]byte("error: missing text\n"))
				os.Exit(1)
			}

			text := inv.Args[0]
			if upper {
				text = strings.ToUpper(text)
			}

			inv.Stdout.Write([]byte(text))
			return nil
		},
	}

	err := cmd.Invoke().WithOS().Run()
	if err != nil {
		panic(err)
	}
}
