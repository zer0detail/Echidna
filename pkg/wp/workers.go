package wp

import (
	"context"
	"os"

	"github.com/Paraflare/Echidna/pkg/vulnerabilities"
)

// Scan will call the vulnerability packages scanning function to check each file for vulns
// if it finds vulns the plugin will be moved to the inspect/ folder with the results stored
// with it as a .txt file with the same name
func scanWorker(ctx context.Context, errCh chan error, plugins *Plugins, workQueue chan *Plugin, resultsQueue chan *vulnerabilities.Results) {

	for p := range workQueue {
		scanResults := vulnerabilities.Results{
			Plugin:  p.Name,
			Modules: make(map[string][]vulnerabilities.VulnResults),
		}

		err := vulnerabilities.ZipScan(ctx, p.OutPath, &scanResults)
		if err != nil {
			errCh <- err
			removeZip(p.OutPath, errCh)
			incSkipped(plugins)
			continue
		}
		if len(scanResults.Modules) > 0 {
			err := p.moveToInspect(&scanResults)
			if err != nil {
				errCh <- err
				removeZip(p.OutPath, errCh)
				incSkipped(plugins)
				continue
			}
			err = p.saveResults(&scanResults)
			if err != nil {
				errCh <- err
				incSkipped(plugins)
				continue
			}

			resultsQueue <- &scanResults
		}
		plugins.scanMu.Lock()
		plugins.FilesScanned++
		plugins.scanMu.Unlock()
		removeZip(p.OutPath, errCh)
	}

}

func resultsWorker(ctx context.Context, errCh chan error, plugins *Plugins, resultsQueue chan *vulnerabilities.Results) {

	for result := range resultsQueue {

		plugins.resMu.Lock()

		plugins.LatestVuln = *result
		plugins.VulnsFound++
		plugins.Vulns = append(plugins.Vulns, *result)

		plugins.resMu.Unlock()
	}

}

func removeZip(path string, errCh chan error) {
	err := os.Remove(path)
	if err != nil {
		errCh <- err
	}
}

func incSkipped(plugins *Plugins) {
	plugins.scanMu.Lock()
	plugins.Skipped++
	plugins.scanMu.Unlock()
}
