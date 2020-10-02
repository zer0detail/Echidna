package vulnerabilities

import (
	"fmt"
	"html"
	"regexp"
)

// LFI is the function that searches PHP code for common Local File Inclusion vulnerability patterns
func LFI(content []byte) (VulnResults, error) {
	var vulnResults VulnResults

	signatures := []string{
		`[ \n\t\r]require\(.*\$_.*\)`,
		`[ \n\t\r]require_once\(.*\$_.*\)`,
		`[ \n\t\r]include_once\(.*\$_.*\)`,
		`[ \n\t\r]include\(.*\$_.*\)`,
		`[ \n\t\r]fopen\(.*\$_.*\)`,
		`[ \n\t\r]file_get_contents\(.*\$_.*\)`,
	}

	filter := "stripslashes|escape|prepare|esc_|sanitize|isset|int|htmlentities|htmlspecial|intval|wp_strip|init_crypt"
	reFilter, err := regexp.Compile(filter)
	if err != nil {
		return vulnResults, fmt.Errorf("cmdExec.go:CmdExec() - error compiling Cmd Execution filter in cmdExec() with error\n%s", err)
	}

	for _, signature := range signatures {
		re, err := regexp.Compile(signature)
		if err != nil {
			return vulnResults, fmt.Errorf("cmdExec.go:CmdExec() - error compiling signature %s with error\n%s", signature, err)
		}
		matches := re.FindAllString(string(content), -1)
		for _, match := range matches {
			filteredMatches := reFilter.FindAllString(match, 1)
			if len(filteredMatches) != 0 {
				continue
			} else {
				match := html.UnescapeString(match)
				vulnResults.Matches = append(vulnResults.Matches, match)
			}
		}
	}

	return vulnResults, nil

}
