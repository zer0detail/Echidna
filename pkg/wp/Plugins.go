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
	"golang.org/x/sync/semaphore"
)

const pluginAPI string = "https://api.wordpress.org/plugins/info/1.2/?action=query_plugins&request[per_page]=400&request[page]="

var (
	sem          = semaphore.NewWeighted(int64(100))
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
	Plugins []Plugin
	URI     string
	client  requests.HTTPClient

	resMu        sync.Mutex
	FilesScanned int
	VulnsFound   int
	LatestVuln   vulnerabilities.Results
	Vulns        []vulnerabilities.Results

	skipMu  sync.Mutex
	Skipped int

	Queue   chan *Plugin
	Results chan vulnerabilities.Results
}

// NewPlugins is the constructor for creating a new *Plugins object with initial data
func NewPlugins(ctx context.Context) (*Plugins, error) {
	var plugins Plugins

	plugins.URI = pluginAPI
	plugins.Info.Page = 1
	plugins.client = requests.NewHTTPClient()

	return &plugins, nil
}

// Scan is the main word press plugin scanner function that controls
// the execution flow of a full scan. Satisfies the scanner interface.
func (w *Plugins) Scan(ctx context.Context, errCh chan error) {

	var wg sync.WaitGroup
	// First retrieve all of the plugins from the WordPress store by
	// iterating over every page available
	fmt.Printf("Requesting plugin information from %d pages\n", w.Info.Pages)
	for w.Info.Page <= w.Info.Pages {
		select {
		case <-ctx.Done():
			return
		default:

			// TryAcquire rather than acquire so we dont block and continue to loop and
			// check for context closure
			openSlot := sem.TryAcquire(int64(1))
			if openSlot {
				wg.Add(1)
				go func(ctx context.Context, errCh chan error) {
					defer wg.Done()
					w.Info.Page++
					w.addPlugins(ctx, errCh)
					// If we cancel context Scan() will return and sem will be destroyed
					// but this go func will still try to Release and cause a panic.
					// So check if we are cancelled first.
					select {
					case <-ctx.Done():
						return
					default:
						sem.Release(1)
					}
				}(ctx, errCh)
			}
		}
	}
	fmt.Println("All requests sent. Waiting for all replies to return.")
	// Wait for all page information to return
	wg.Wait()
	// Channel for vulnerability scanning workers to receive a new plugin to scan on
	w.Queue = make(chan *Plugin, 10)
	w.Results = make(chan vulnerabilities.Results, 10)
	defer close(w.Queue)
	defer close(w.Results)
	// Spawn 100 zipScanner goroutines that will listen on the Queue and scan plugins
	// that are passed down the channel by w.Download()
	for workers := 1; workers <= 50; workers++ {
		go scanWorker(ctx, errCh, w.Queue, w.Results)
		go resultsWorker(ctx, errCh, w, w.Results)
	}
	// Loop until we have scanned ALL plugins
	for len(w.Plugins) > 0 {

		select {
		// every  time we get back to the top of the loop do a non-blocking check of
		// the errors channel. This is so we can constantly check for failed goroutines.
		// without hanging on a blocking channel read.
		case <-ctx.Done():
			return
		default:
			// TryAcquire rather than acquire so we dont block and continue to loop and
			// check for context closure
			openSlot := sem.TryAcquire(int64(1))
			if openSlot {
				w.printStatus()
				// Choose a random plugin
				randPluginIndex := randomPicker.Intn(len(w.Plugins))
				plugin := w.Plugins[randPluginIndex]

				w.RemovePlugin(randPluginIndex)

				// Refresh the http client every 1000 requests.
				// It stops alot of errors
				if w.FilesScanned%1000 == 0 && w.FilesScanned != 0 {
					w.client = requests.NewHTTPClient()
				}

				go func(ctx context.Context, errCh chan error, queue chan *Plugin) {
					plugin.Download(ctx, w, errCh, queue)
					// If we cancel context Scan() will return and sem will be destroyed
					// but this go func will still try to Release and cause a panic.
					// So check if we are cancelled first.
					select {
					case <-ctx.Done():
						return
					default:
						sem.Release(1)
					}
				}(ctx, errCh, w.Queue)
			}
		}
	}

	fmt.Println("Finished scanning all plugins. Happy Hunting!")
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

	// tm.Flush()

	fmt.Printf("Plugin count %5d\t", len(w.Plugins))
	fmt.Printf("Plugins Scanned: %5d\t", w.FilesScanned)
	fmt.Printf("Vulns found: %5d\t", len(w.Vulns))
	fmt.Printf("Plugins Skipped (due to errors): %5d\t", w.Skipped)
	fmt.Printf("Pages: %d/%d\t", w.Info.Page, w.Info.Pages)
	fmt.Printf("files in queue: %d\n", len(w.Queue))
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
func (w *Plugins) addPlugins(ctx context.Context, errChan chan error) {

	// create a temporary Plugins object to get the next list of 250 plugins
	// well then pull out the NextPluginList.Plugins slice and store it back into
	// the main PluginList for the scanner
	var nextPluginList Plugins

	uri := w.URI + strconv.Itoa(w.Info.Page)

	rawPluginList, err := requests.SendRequest(ctx, w.client, uri)
	if err != nil {
		errChan <- fmt.Errorf("Plugins.go:addPlugins() error while performing SendRequest(%s) with error: %s", uri, err)
	}
	err = json.Unmarshal(rawPluginList, &nextPluginList)
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

	body, err := requests.SendRequest(ctx, w.client, uri)
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
