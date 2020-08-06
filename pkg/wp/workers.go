package wp

import (
	"context"
	"os"

	"github.com/Paraflare/Echidna/pkg/vulnerabilities"
)

// Scan will call the vulnerability packages scanning function to check each file for vulns
// if it finds vulns the plugin will be moved to the inspect/ folder with the results stored
// with it as a .txt file with the same name
func scanWorker(ctx context.Context, errCh chan error, filesScanned *int, skipped *int, plugins *[]Plugin, scanQueue chan int, resultsQueue chan vulnerabilities.Results, done chan int) {

	for i := range scanQueue {
		select {
		case <-ctx.Done():
			return
		default:
			*filesScanned++
			p := (*plugins)[i]
			scanResults := vulnerabilities.Results{
				Plugin:  p.Name,
				Modules: make(map[string][]vulnerabilities.VulnResults),
			}

			err := vulnerabilities.ZipScan(ctx, p.OutPath, &scanResults)
			if err != nil {
				errCh <- err
				*skipped++
				removeZip(p.OutPath, errCh)
				continue
			}
			if len(scanResults.Modules) > 0 {
				err := p.moveToInspect(&scanResults)
				if err != nil {
					errCh <- err
					*skipped++
					removeZip(p.OutPath, errCh)
					continue
				}
				err = p.saveResults(&scanResults)
				if err != nil {
					*skipped++
					errCh <- err
					continue
				}

				resultsQueue <- scanResults
			}

			removeZip(p.OutPath, errCh)
			done <- 1
		}

	}

}

func resultsWorker(ctx context.Context, errCh chan error, plugins *Plugins, resultsQueue chan vulnerabilities.Results, done chan int) {

	for result := range resultsQueue {
		select {
		case <-ctx.Done():
			return
		default:
			plugins.resMu.Lock()

			plugins.LatestVuln = result
			plugins.VulnsFound++
			plugins.Vulns = append(plugins.Vulns, result)

			plugins.resMu.Unlock()

			done <- 1
		}
	}

}

func removeZip(path string, errCh chan error) {
	err := os.Remove(path)
	if err != nil {
		errCh <- err
	}
}