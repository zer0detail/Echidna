package wp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
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
	numOfWorkers = 50
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
	for workers := 1; workers <= numOfWorkers; workers++ {
		go requests.DownloadWorker(ctx, workers, DownloadQueue, ScanQueue, errCh, requests.NewHTTPClient())
		go scanWorker(ctx, errCh, &(w.FilesScanned), &(w.Skipped), &(w.ScannedPlugins), ScanQueue, Results, done)
		go resultsWorker(ctx, errCh, w, Results, done)

	}

	w.ScannedPlugins = make([]Plugin, len(w.Plugins))
	// Loop until we have scanned ALL plugins
	for len(w.Plugins) > 0 {
		// Choose a random plugin
		randPluginIndex := randomPicker.Intn(len(w.Plugins))
		w.ScannedPlugins = append(w.ScannedPlugins, w.Plugins[randPluginIndex])
		w.RemovePlugin(randPluginIndex)
		index := len(w.ScannedPlugins) - 1
		go w.ScannedPlugins[index].scan(errCh, index, DownloadQueue)
	}
	w.Timer = time.Now()
	for (w.FilesScanned + w.Skipped) != len(w.ScannedPlugins) {
		select {
		// every  time we get back to the top of the loop do a non-blocking check of
		// the background context to see if we should cancel or not. We cancel if someone
		// pressed ctrl+c.
		case <-ctx.Done():
			return
		default:
			<-done
			w.printStatus()
		}
	}

	fmt.Println("Finished scanning all plugins. Happy Hunting!")
}

func (w *Plugins) queryAllStorePages(ctx context.Context, errCh chan error) {
	color.Blue.Println("-----------------------------------------------------")
	color.Blue.Printf("|")
	color.Green.Printf(" [+] ")
	fmt.Printf("Number of Worker routines: %d\n", numOfWorkers)
	color.Blue.Printf("|")
	color.Green.Printf(" [+] ")
	fmt.Printf("Analysis Modules:\t\t")
	for module := range vulnerabilities.Modules {
		fmt.Printf(" %s ", module)
	}
	fmt.Printf("\n")
	color.Blue.Printf("|")
	color.Green.Printf(" [+] ")
	fmt.Printf("Plugin store page:\t %d\n", w.Info.Pages)
	color.Blue.Printf("|")
	color.Green.Printf(" [+] ")
	fmt.Printf("Total Plugins to scan:\t %d", len(w.Plugins))

	// Create buffered channels for worker requests and results
	reqCh := make(chan string, w.Info.Pages)
	resultCh := make(chan []byte, w.Info.Pages)
	defer close(reqCh)
	defer close(resultCh)

	// Spin up request workers to receive uri's to request
	for i := 0; i <= numOfWorkers; i++ {
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
			color.Blue.Printf("\r|")
			color.Green.Printf(" [+] ")
			fmt.Printf("Total Plugins to scan:\t %d", len(w.Plugins))
			pluginBody := <-resultCh
			wg.Add(1)
			go w.addPlugins(ctx, pluginBody, errCh, &wg)
		}
	}
	color.Blue.Printf("\r|")
	color.Green.Printf(" [+] ")
	fmt.Printf("Total Plugins to scan:\t %d\n", len(w.Plugins))
	color.Blue.Printf("-----------------------------------------------------\n")
	wg.Wait()
}

func (w *Plugins) printStatus() {

	elapsed := time.Since(w.Timer).Seconds()

	fmt.Printf("\rPlugins Scanned: ")
	color.LightGreen.Printf("%5d    ", w.FilesScanned)
	fmt.Printf("Vulns found:")
	color.Green.Printf("%5d    ", w.VulnsFound)
	fmt.Printf("Plugins Skipped (due to errors): ")
	color.Red.Printf("%5d     ", w.Skipped)
	fmt.Printf("Plugins Scanned Per Second: ")
	color.Gray.Printf("%0.1f       ", (float64((w.FilesScanned + w.Skipped)) / elapsed))
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
