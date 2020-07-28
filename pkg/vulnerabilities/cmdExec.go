package vulnerabilities

import (
	"fmt"
	"html"
	"regexp"
)

// CmdExec is the function that searches PHP code for common RCE vulnerabilities patterns
func CmdExec(content []byte) (VulnResults, error) {
	var vulnResults VulnResults

	signatures := []string{
		`system\(.*\$_.*\)`,
		`shell_exec\(.*\$_.*\)`,
		`pass_thru\(.*\$_.*\)`,
		`proc_open\(.*\$_.*\)`,
		`popen\(.*\$_.*\)`,
		`eval\(.*\$_.*\)`,
		`assert\(.*\$_.*\)`,
	}

	filter := "stripslashes|escape|prepare|esc_|sanitize|isset|int|htmlentities|htmlspecial|intval|wp_strip"
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
