package vulnerabilities

// VulnResults type is for each vuln module to output its results to.
// Each one is appended to the files "Results" object
type VulnResults struct {
	File    string
	Matches []string
}
