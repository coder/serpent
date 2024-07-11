package completion

import (
	"fmt"
	"io"
	"text/template"
)

const pshCompletionTemplate = `

# Escaping output sourced from:
# https://github.com/spf13/cobra/blob/e94f6d0dd9a5e5738dca6bce03c4b1207ffbc0ec/powershell_completions.go#L47
filter _{{.Name}}_escapeStringWithSpecialChars {
` + "    $_ -replace '\\s|#|@|\\$|;|,|''|\\{|\\}|\\(|\\)|\"|`|\\||<|>|&','`$&'" + `
}

$_{{.Name}}_completions = {
    param(
        $wordToComplete,
        $commandAst,
        $cursorPosition
    )
    # Legacy space handling sourced from:
	# https://github.com/spf13/cobra/blob/e94f6d0dd9a5e5738dca6bce03c4b1207ffbc0ec/powershell_completions.go#L107
	if ($PSVersionTable.PsVersion -lt [version]'7.2.0' -or
        ($PSVersionTable.PsVersion -lt [version]'7.3.0' -and -not [ExperimentalFeature]::IsEnabled("PSNativeCommandArgumentPassing")) -or
        (($PSVersionTable.PsVersion -ge [version]'7.3.0' -or [ExperimentalFeature]::IsEnabled("PSNativeCommandArgumentPassing")) -and
         $PSNativeCommandArgumentPassing -eq 'Legacy')) {
        $Space =` + "' `\"`\"'" + `
    } else {
        $Space = ' ""'
    }
	$Command = $commandAst.ToString().Substring(0, $cursorPosition - 1)
	if ($wordToComplete -ne "" ) {
        $wordToComplete = $Command.Split(" ")[-1]
    } else {
        $Command = $Command + $Space
    }
    # Get completions by calling the command with the COMPLETION_MODE environment variable set to 1
    "$Command" | Out-File -Append -FilePath "out.log"
    $env:COMPLETION_MODE = 1
    Invoke-Expression $Command | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        "$_" | _{{.Name}}_escapeStringWithSpecialChars
    }
    rm env:COMPLETION_MODE
}

Register-ArgumentCompleter -CommandName {{.Name}} -ScriptBlock $_{{.Name}}_completions
`

func GeneratePowershellCompletion(
	w io.Writer,
	rootCmdName string,
) error {
	tmpl, err := template.New("powershell").Parse(pshCompletionTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	err = tmpl.Execute(
		w,
		map[string]string{
			"Name": rootCmdName,
		},
	)
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
