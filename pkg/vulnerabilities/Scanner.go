package vulnerabilities

import (
	"archive/zip"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

var modules = map[string]func([]byte) (VulnResults, error){
	"XSS":  XSS,
	"SQLI": SQLI,
}

// Results is a struct for storing the results of every vulnerable file that was scanned within a plugins archive
type Results struct {
	Plugin  string
	Modules map[string][]VulnResults
}

// ZipScan opens zip files, finds PHP files and hands them over to vulnerability
// modules for bug hunting.
func ZipScan(zipPath string, fileResults *Results) error {

	files, err := zip.OpenReader(zipPath)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  zipPath,
			"error": err,
		}).Error("Could not open Zip file. Skipping..")
		return err
	}
	defer files.Close()

	for _, file := range files.File {

		if strings.HasSuffix(file.Name, ".php") {
			r, err := file.Open()
			if err != nil {
				log.WithFields(log.Fields{
					"file":  file.Name,
					"error": err,
				}).Warn("Could not open php file. Skipping..")
				continue
			}
			defer r.Close()

			var content []byte
			content, err = ioutil.ReadAll(r)
			if err != nil {
				log.WithFields(log.Fields{
					"file":  file.Name,
					"error": err,
				}).Warn("Could not read php file. Skipping..")
				continue
			}

			for module, moduleFunc := range modules {
				vulns, err := moduleFunc(content)
				if err != nil {
					log.WithFields(log.Fields{
						"file":  file.Name,
						"error": err,
					}).Warn("Error returned while scanning file for XSS. Skipping..")
					continue
				}

				if vulns.Matches != nil {
					vulns.File = file.Name
					fileResults.Modules[module] = append(fileResults.Modules[module], vulns)
				}
			}

		}
	}

	return nil
}
