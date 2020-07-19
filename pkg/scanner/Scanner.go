package scanner

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"strings"

	vulnerabilities "github.com/Paraflare/Echidna/pkg/scanner/vulnerabilities"
)

// Results is a struct for storing the results of every vulnerable file that was scanned within a plugins archive
type Results struct {
	Plugin  string
	Results map[string][]vulnerabilities.VulnResults
}

// ZipScan opens zip files, finds PHP files and hands them over to vulnerability
// modules for bug hunting.
func ZipScan(zipPath string, scanResults *Results) error {

	files, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("Could not open zip file %s in scan() function with error\n%s", zipPath, err)
	}
	defer files.Close()

	for _, file := range files.File {

		if strings.HasSuffix(file.Name, ".php") {
			r, err := file.Open()
			if err != nil {
				return fmt.Errorf("Could not read file %s in scan() with error\n%s", file.Name, err)
			}
			defer r.Close()

			var content []byte
			content, err = ioutil.ReadAll(r)
			if err != nil {
				return fmt.Errorf("Could not read %s contents in scan with error\n%s", file.Name, err)
			}

			vulns, err := vulnerabilities.XSSscan(content)
			if err != nil {
				return fmt.Errorf("error Scanning file %s in ZipScan() with error\n%s", file.Name, err)
			}

			if vulns.Matches != nil {
				vulns.File = file.Name
				scanResults.Results["XSS"] = append(scanResults.Results["XSS"], vulns)
			}
		}
	}

	return nil
}
