package vulnerabilities

import (
	"archive/zip"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
)

var modules = map[string]func([]byte) (VulnResults, error){
	"XSS":  XSS,
	"SQLI": SQLI,
	"CMDEXEC": CMDEXEC,
}

// Results is a struct for storing the results of every vulnerable file that was scanned within a plugins archive
type Results struct {
	Plugin  string
	Modules map[string][]VulnResults
}

// ZipScan opens zip files, finds PHP files and hands them over to vulnerability
// modules for bug hunting.
func ZipScan(ctx context.Context, zipPath string, fileResults *Results) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		files, err := zip.OpenReader(zipPath)
		if err != nil {
			// log.WithFields(log.Fields{
			// 	"file":  zipPath,
			// 	"error": err,
			// }).Error("Could not open Zip file. Skipping..")
			return fmt.Errorf("scanner.go:ZipScan() - failed to open zip file with zip.OpenReader(%v)", err)
		}
		defer files.Close()

		for _, file := range files.File {
			// Before we check each file, check if our context has been cancelled
			// So we can close and free up the zip file for deletion by the cleanup function
			select {
			case <-ctx.Done():
				return nil
			default:
				if strings.HasSuffix(file.Name, ".php") {
					r, err := file.Open()
					if err != nil {
						// log.WithFields(log.Fields{
						// 	"file":  file.Name,
						// 	"error": err,
						// }).Warn("Could not open php file. Skipping..")
						continue
					}
					defer r.Close()

					var content []byte
					content, err = ioutil.ReadAll(r)
					if err != nil {
						// log.WithFields(log.Fields{
						// 	"file":  file.Name,
						// 	"error": err,
						// }).Warn("Could not read php file. Skipping..")
						continue
					}

					for module, moduleFunc := range modules {
						vulns, err := moduleFunc(content)
						if err != nil {
							// log.WithFields(log.Fields{
							// 	"file":  file.Name,
							// 	"error": err,
							// }).Warn("Error returned while scanning file for XSS. Skipping..")
							continue
						}

						if vulns.Matches != nil {
							vulns.File = file.Name
							fileResults.Modules[module] = append(fileResults.Modules[module], vulns)
						}
					}
				}
			}

		}
		return nil
	}
}
