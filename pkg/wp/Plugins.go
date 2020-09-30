package wp

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/Paraflare/Echidna/pkg/requests"
	"github.com/Paraflare/Echidna/pkg/vulnerabilities"
)

const pluginAPI string = "https://api.wordpress.org/plugins/info/1.2/?action=query_plugins&request[per_page]=400&request[page]="

var (
	seed         = rand.NewSource(time.Now().Unix())
	randomPicker = rand.New(seed)
)

// Plugins struct holds the WordPress Plugins information and satisfies
// the scanner interface.
type Plugins struct {
	mu   sync.Mutex
	Info struct {
		Page    int `json:"page"`
		Pages   int `json:"pages"`
		Results int `json:"results"`
	} `json:"info"`
	Plugins        []Plugin
	ScannedPlugins []Plugin
	URI            string

	resMu      sync.Mutex
	VulnsFound int
	LatestVuln vulnerabilities.Results
	Vulns      []vulnerabilities.Results

	Skipped      int
	FilesScanned int
	Timer        time.Time
}

// NewPlugins is the constructor for creating a new *Plugins object with initial data
func NewPlugins(ctx context.Context) (*Plugins, error) {
	var plugins Plugins

	plugins.URI = pluginAPI
	plugins.Info.Page = 1

	return &plugins, nil
}

// Scan is the main word press plugin scanner function that controls
// the execution flow of a full scan. Satisfies the scanner interface.
func (w *Plugins) Scan(ctx context.Context, errCh chan error) {

	w.queryAllStorePages(ctx, errCh)
	// Channel for vulnerability scanning workers to receive a new plugin to scan on
	DownloadQueue := make(chan requests.PluginReq, len(w.Plugins))
	ScanQueue := make(chan int, len(w.Plugins))
	Results := make(chan vulnerabilities.Results, len(w.Plugins))
	done := make(chan int, len(w.Plugins))
	defer close(DownloadQueue)
	defer close(ScanQueue)
	defer close(Results)
	defer close(done)
	// Spawn worker goroutines that will listen on the Queue and scan plugins
	// that are passed down the channel by w.Download()
	for workers := 1; workers <= 20; workers++ {
		go requests.DownloadWorker(ctx, workers, DownloadQueue, ScanQueue, errCh, requests.NewHTTPClient())
		go scanWorker(ctx, errCh, &(w.FilesScanned), &(w.Skipped), &(w.ScannedPlugins), ScanQueue, Results, done)
		go resultsWorker(ctx, errCh, w, Results, done)

	}

	w.ScannedPlugins = make([]Plugin, len(w.Plugins))
	// Loop until we have scanned ALL plugins
	for len(w.Plugins) > 0 {

		select {
		// every  time we get back to the top of the loop do a non-blocking check of
		// the errors channel. This is so we can constantly check for failed goroutines.
		// without hanging on a blocking channel read.
		case <-ctx.Done():
			return
		default:
			fmt.Printf("\rPlugins left to scan: %6d", len(w.Plugins))
			// Choose a random plugin
			randPluginIndex := randomPicker.Intn(len(w.Plugins))
			w.ScannedPlugins = append(w.ScannedPlugins, w.Plugins[randPluginIndex])
			w.RemovePlugin(randPluginIndex)
			index := len(w.ScannedPlugins) - 1
			go w.ScannedPlugins[index].scan(errCh, index, DownloadQueue)

		}
	}
	fmt.Printf("\rPlugins left add to worker queue: %6d\n", len(w.Plugins))
	w.Timer = time.Now()
	for (w.FilesScanned + w.Skipped) != len(w.ScannedPlugins) {
		<-done
		w.printStatus()
	}

	fmt.Println("Finished scanning all plugins. Happy Hunting!")
}

func (w *Plugins) queryAllStorePages(ctx context.Context, errCh chan error) {
	fmt.Printf("Requesting plugin information from %d pages\n", w.Info.Pages)

	// Create buffered channels for worker requests and results
	reqCh := make(chan string, w.Info.Pages)
	resultCh := make(chan []byte, w.Info.Pages)
	defer close(reqCh)
	defer close(resultCh)

	// Spin up request workers to receive uri's to request
	for i := 0; i <= 20; i++ {
		go requests.ReqWorker(ctx, i, reqCh, resultCh, errCh, requests.NewHTTPClient())
	}
	// Send all of the requests to the workers
	for w.Info.Page <= w.Info.Pages {
		select {
		case <-ctx.Done():
			return
		default:
			w.Info.Page++
			uri := w.URI + strconv.Itoa(w.Info.Page)
			reqCh <- uri
		}
	}
	var wg sync.WaitGroup
	for i := 1; i <= w.Info.Pages; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			pluginBody := <-resultCh
			wg.Add(1)
			go w.addPlugins(ctx, pluginBody, errCh, &wg)
		}
	}
	fmt.Println("All requests sent. Waiting for all results to return before proceeding.")
	wg.Wait()
	fmt.Println("All requests have been returned. Beginning scan")
}

func (w *Plugins) printStatus() {
	// tm.Clear()

	// tm.MoveCursor(1, 1)
	// tm.Printf("Plugin count: %d\t", len(w.Plugins))
	// tm.Printf("Files Scanned: %d\t", w.FilesScanned)
	// tm.Printf("Vulnerable Plugins so far: %d\n", len(w.Vulns))
	// tm.Printf("\n\t\t\tLatest Vulnerable plugin - %s\n", w.LatestVuln.Plugin)
	// for k := range w.LatestVuln.Modules {
	// 	tm.Printf("\n\t\t%s\n\t\t\t%s", k, w.LatestVuln.Modules[k])
	// }
	elapsed := time.Since(w.Timer).Seconds()
	// tm.Flush()
	fmt.Printf("\rPlugins Scanned: %5d    ", w.FilesScanned)
	fmt.Printf("Vulns found: %5d    ", w.VulnsFound)
	fmt.Printf("Plugins Skipped (due to errors): %5d     ", w.Skipped)
	fmt.Printf("Plugins Scanned Per Second: %0.1f", (float64((w.FilesScanned + w.Skipped)) / elapsed))
}

// Page returns the page property and satisfies the scanner interface.
func (w *Plugins) Page() int { return w.Info.Page }

// Pages returns the pages property and satisfies the scanner interface.
func (w *Plugins) Pages() int { return w.Info.Pages }

// TotalFiles returns the Results property and satisfies the scanner interface
func (w *Plugins) TotalFiles() int { return w.Info.Results }

// ScannedCount returns the FileScanned property and satisfiies the scanner interface
func (w *Plugins) ScannedCount() int { return w.FilesScanned }

// addPlugins retrieves the next page of plugins from the WP store and adds them to the plugins object for scanning
func (w *Plugins) addPlugins(ctx context.Context, pluginBody []byte, errChan chan error, wg *sync.WaitGroup) {

	defer wg.Done()
	// create a temporary Plugins object to get the next list of 250 plugins
	// well then pull out the NextPluginList.Plugins slice and store it back into
	// the main PluginList for the scanner
	var nextPluginList Plugins

	err := json.Unmarshal(pluginBody, &nextPluginList)
	if err != nil {
		errChan <- fmt.Errorf("Plugins.go:addPlugins() error while performing json.Unmarshal with error: %s", err)
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, plugin := range nextPluginList.Plugins {
		w.Plugins = append((*w).Plugins, plugin)
	}
}

// AddInfo retrieve the number of pages in the WP plugin store and the number of plugins to scan
func (w *Plugins) AddInfo(ctx context.Context) error {

	// Create a temporary Plugins value to store the results of the web request in.
	// we'll then pull out the .Info properties to append to our actual
	// main plugin list that the scanner is using.
	var info Plugins

	uri := w.URI + strconv.Itoa(w.Info.Page)

	body, err := requests.SendRequest(ctx, requests.NewHTTPClient(), uri)
	if err != nil {
		// Die if we can't get WordPress Plugin info. There's no way to continue without it
		return fmt.Errorf("Plugins.go:AddInfo() - call to requests.SendRequest(ctx, %s) error: %s", uri, err.Error())
	}
	err = json.Unmarshal(body, &info)
	if err != nil {
		// Die if we can't get WordPress Plugin info. There's no way to continue without it
		return fmt.Errorf("Plugins.go:AddInfo() - while attempting to unmarshal json from the body with error: %s", err.Error())
	}
	w.Info = info.Info

	return nil
}

// RemovePlugin takes the random plugin we've just selected and removes it from the plugins object so we dont scan it again later.
func (w *Plugins) RemovePlugin(i int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Plugins[len(w.Plugins)-1], w.Plugins[i] = w.Plugins[i], w.Plugins[len(w.Plugins)-1]
	w.Plugins = w.Plugins[:len(w.Plugins)-1]

}
