# completion

The `completion` package extends `serpent` to allow applications to generate rich auto-completions.


## Protocol

The completion scripts call out to the serpent command to generate
completions. The convention is to pass the exact args and flags (or
cmdline) of the in-progress command with a `COMPLETION_MODE=1` environment variable. That environment variable lets the command know to generate completions instead of running the command.
By default, completions will be generated based on available flags and subcommands. Additional completions can be added by supplying a `CompletionHandlerFunc` on an Option or Command.