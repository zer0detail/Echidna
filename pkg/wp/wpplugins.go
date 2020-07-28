package wp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
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
	Info struct {
		Page    int `json:"page"`
		Pages   int `json:"pages"`
		Results int `json:"results"`
	} `json:"info"`
	Plugins      []Plugin
	URI          string
	FilesScanned int
	VulnsFound   int
	Vulns        []struct{}
}

// NewPlugins is the constructor for creating a new *Plugins object with initial data
func NewPlugins(ctx context.Context) (*Plugins, error) {
	var plugins Plugins

	plugins.URI = pluginAPI
	plugins.Info.Page = 1

	err := plugins.AddInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &plugins, nil
}

// Scan is the main word press plugin scanner function that controls
// the execution flow of a full scan. Satisfies the scanner interface.
func (w *Plugins) Scan(ctx context.Context, errChan chan error) {

	// Loop until we have scanned ALL plugins
	for w.FilesScanned != w.Info.Results {

		select {
		// every  time we get back to the top of the loop do a non-blocking check of
		// the errors channel. This is so we can constantly check for failed goroutines.
		// without hanging on a blocking channel read.
		case <-ctx.Done():
			return
		default:
			// if we have plugins, scan them
			if len(w.Plugins) > 0 {
				err := sem.Acquire(ctx, 1)
				if err != nil {
					errChan <- fmt.Errorf("wpplugins.go:Scan() - sem.Acquire(ctx, 1) failed with error\n%s", err)
				}
				// Choose a random plugin
				randPluginIndex := randomPicker.Intn(len(w.Plugins))
				plugin := w.Plugins[randPluginIndex]
				// Remove it from the list
				w.RemovePlugin(randPluginIndex)

				go func(ctx context.Context, errChan chan error) {
					plugin.VulnScan(ctx, &w.FilesScanned, errChan)
					sem.Release(1)
				}(ctx, errChan)

				// color.Yellow.Print("Plugin count: ")
				// color.Gray.Printf("%d\t", len(w.Plugins))
				// color.Yellow.Print("Files Scanned: ")
				// color.Gray.Printf("%d\n", w.FilesScanned)
			}

			// If we haven't finished pulling the list of plugins from the store, grab another page and
			// add it to PluginList.Plugins
			if w.Info.Page <= w.Info.Pages {
				err := sem.Acquire(ctx, 1)
				if err != nil {
					errChan <- fmt.Errorf("wpplugins.go:Scan() - sem.Acquire(ctx, 1) failed with error\n%s", err)
				}
				go func(ctx context.Context, errChan chan error) {
					w.Info.Page++
					w.addPlugins(ctx, errChan)
					sem.Release(1)
				}(ctx, errChan)
			}
		}
	}
	fmt.Println("Finished scanning all plugins. Happy Hunting!")
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

	rawPluginList, err := requests.SendRequest(ctx, uri)
	if err != nil {
		errChan <- err
		return
	}
	err = json.Unmarshal(rawPluginList, &nextPluginList)
	if err != nil {

		errChan <- err
		return
	}
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

	body, err := requests.SendRequest(ctx, uri)
	if err != nil {
		// Die if we can't get WordPress Plugin info. There's no way to continue without it
		return fmt.Errorf("wpplugins.go:AddInfo() - call to requests.SendRequest(ctx, %s)\n%s", uri, err.Error())
	}
	err = json.Unmarshal(body, &info)
	if err != nil {
		// Die if we can't get WordPress Plugin info. There's no way to continue without it
		return fmt.Errorf("wpplugins.go:AddInfo() - while attempting to unmarshal json from the body with error:\n%s", err.Error())
	}
	w.Info = info.Info

	return nil
}

// RemovePlugin takes the random plugin we've just selected and removes it from the plugins object so we dont scan it again later.
func (w *Plugins) RemovePlugin(i int) {
	w.Plugins[len(w.Plugins)-1], w.Plugins[i] = w.Plugins[i], w.Plugins[len(w.Plugins)-1]
	w.Plugins = w.Plugins[:len(w.Plugins)-1]

}

// Plugin struct holds the information for each plugin we iterate through
type Plugin struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Version string `json:"version"`
	Author  string `json:"author"`
	Ratings struct {
		Num1 int `json:"1"`
		Num2 int `json:"2"`
		Num3 int `json:"3"`
		Num4 int `json:"4"`
		Num5 int `json:"5"`
	} `json:"ratings,omitempty"`
	NumRatings             int    `json:"num_ratings"`
	SupportThreads         int    `json:"support_threads"`
	SupportThreadsResolved int    `json:"support_threads_resolved"`
	ActiveInstalls         int    `json:"active_installs"`
	Downloaded             int    `json:"downloaded"`
	LastUpdated            string `json:"last_updated"`
	Added                  string `json:"added"`
	Homepage               string `json:"homepage"`
	Description            string `json:"description"`
	ShortDescription       string `json:"short_description"`
	DownloadLink           string `json:"download_link"`
	DonateLink             string `json:"donate_link"`
	Icons                  struct {
		OneX string `json:"1x"`
		TwoX string `json:"2x"`
	} `json:"icons,omitempty"`
	DaysSinceLastUpdate string
	OutPath             string
	FileName            string
	inspectPath         string
}

func (p *Plugin) setOutPath() error {
	fileName := strings.Split(p.DownloadLink, "/")[4]
	fileName = strconv.Itoa(p.ActiveInstalls) + "_" + p.DaysSinceLastUpdate + "_" + fileName

	p.FileName = fileName

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("wpplugins.go:setOutPath() - os.Getwd() with error\n%s", err)
	}

	p.OutPath = dir + string(os.PathSeparator) + "current" + string(os.PathSeparator) + fileName

	return nil
}

func (p *Plugin) setDaysSinceUpdate() error {

	timeLayout := "2006-01-02 3:04pm MST"

	lastUpdateTime, err := time.Parse(timeLayout, p.LastUpdated)
	if err != nil {
		return fmt.Errorf("wpplugins.go:setDaysSinceUpdate() time.Parse(%s, %s) failed with error\n%s", timeLayout, p.LastUpdated, err)
	}

	p.DaysSinceLastUpdate = strconv.Itoa(int(time.Since(lastUpdateTime).Hours() / 24))

	return nil
}

// VulnScan downloads the plugin, scans each php file for vulnerabilitys and sets it aside
// for later inspection if somethings found. Otherwise the plugin is then deleted.
func (p *Plugin) VulnScan(ctx context.Context, filesScanned *int, errChan chan error) {
	*filesScanned++
	err := p.setDaysSinceUpdate()
	if err != nil {
		errChan <- err
	}
	err = p.setOutPath()
	if err != nil {
		errChan <- err
	}

	err = requests.Download(ctx, p.OutPath, p.DownloadLink)
	if err != nil {
		errChan <- err
	}

	scanResults := vulnerabilities.Results{
		Plugin:  p.Name,
		Modules: make(map[string][]vulnerabilities.VulnResults),
	}

	err = vulnerabilities.ZipScan(ctx, p.OutPath, &scanResults)
	if err != nil {
		return
	}

	if len(scanResults.Modules) > 0 {
		err := p.moveToInspect(&scanResults)
		if err != nil {
			errChan <- err
		}
		err = p.saveResults(&scanResults)
		if err != nil {
			errChan <- err
		}
		// color.Green.Printf("Potential Vulnerabilities found in plugin: %s\n", p.Name)
		// color.Green.Println("Moving plugin to inspect/ folder.")
		return

	}

	err = os.Remove(p.OutPath)
	if err != nil {
		errChan <- err
	}
}

func (p *Plugin) moveToInspect(results *vulnerabilities.Results) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("wpplugins.go:moveToInspect() - Could not get current working directory with os.Getwd() with error\n%s", err)
	}
	// Split results out into the results for each vulnerability module (e.g xss, sqli, etc)
	for k := range results.Modules {
		p.inspectPath = dir + string(os.PathSeparator) + "inspect" + string(os.PathSeparator) + k
		// if a folder for that vuln module doesnt already exist, create it.
		if _, err := os.Stat(p.inspectPath); os.IsNotExist(err) {
			err = os.MkdirAll(p.inspectPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("wpplugins.go:moveToInspect() - Failed to create directory %s with error\n%s", p.inspectPath, err)
			}
		}

		outfile := p.inspectPath + string(os.PathSeparator) + p.FileName

		src, err := os.Open(p.OutPath)
		if err != nil {
			return fmt.Errorf("wpplugins.go:moveToInspect() - failed to os.Open(%s) with error\n%s", p.OutPath, err)
		}
		defer src.Close()
		dst, err := os.Create(outfile)
		if err != nil {
			return fmt.Errorf("wpplugins.go:moveToInspect() - failed to os.Create(%s) with error \n%s", p.inspectPath, err)
		}
		defer dst.Close()

		_, err = io.Copy(src, dst)
		if err != nil {
			return fmt.Errorf("wpplugins.go:moveToInspect() - Could not move %s to inspect folder with error\n%s", p.Name, err)
		}
	}
	return nil
}

func (p *Plugin) saveResults(results *vulnerabilities.Results) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("wpplugins.go:saveResults() - Could not get current working directory with os.Getwd() with error\n%s", err)
	}
	for k := range results.Modules {
		file, err := json.MarshalIndent(results.Modules[k], "", " ")
		if err != nil {
			return fmt.Errorf("wpplugins.go:saveResults() - Could not MarshalIndent() results for file %s with error\n%s", p.Name, err)
		}
		p.inspectPath = dir + string(os.PathSeparator) + "inspect" + string(os.PathSeparator) + k + string(os.PathSeparator) + p.FileName
		err = ioutil.WriteFile(p.inspectPath+".txt", file, 0644)
		if err != nil {
			return fmt.Errorf("wpplugins.go:saveResults() - Could not save results file for %s with error \n%s", p.Name, err)
		}
	}

	return nil
}
