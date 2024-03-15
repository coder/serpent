# completion

The `completion` package extends `serpent` to allow applications to generate rich auto-completions.


## Protocol

The completion scripts call out to the serpent command to generate
completions. The convention is to pass the exact args and flags (or
cmdline) of the in-progress command with a `COMPLETION_MODE=1` environment variable. That environment variable lets the command know to generate completions instead of running the command.



Because of this, the middleware must be installed on every command.
For example:

```go
	inv := cmd.Invoke().WithOS()
	if completion.IsCompletionMode(inv) {
		cmd.Walk(
			func(cmd *serpent.Command) {
				// Do not want to waste compute or error on flags.
				cmd.RawArgs = true
				cmd.Handler = completion.Middleware(nil)(cmd.Handler)
			},
		)
	}
	err := inv.Run()
```