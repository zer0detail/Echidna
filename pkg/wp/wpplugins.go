package wp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Paraflare/Echidna/pkg/requests"
	"github.com/Paraflare/Echidna/pkg/scanner"
	"github.com/Paraflare/Echidna/pkg/scanner/vulnerabilities"
	"github.com/gookit/color"
)

const pluginAPI string = "https://api.wordpress.org/plugins/info/1.2/?action=query_plugins&request[per_page]=400&request[page]="

// Plugins unmarshals the Word Press plugins API results
type Plugins struct {
	Info struct {
		Page    int `json:"page"`
		Pages   int `json:"pages"`
		Results int `json:"results"`
	} `json:"info"`
	Plugins []Plugin
	URI     string
}

// NewPlugins is the constructor for creating a new *Plugins object with initial data
func NewPlugins() *Plugins {
	var plugins Plugins

	plugins.URI = pluginAPI
	plugins.Info.Page = 1

	plugins.AddInfo()

	return &plugins
}

// AddPlugins retrieves the next page of plugins from the WP store and adds them to the plugins object for scanning
func (w *Plugins) AddPlugins() error {

	var nextPluginList Plugins

	uri := w.URI + strconv.Itoa(w.Info.Page)

	rawPluginList, err := requests.SendRequest(uri)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rawPluginList, &nextPluginList)
	if err != nil {
		return err
	}
	for _, plugin := range nextPluginList.Plugins {
		w.Plugins = append((*w).Plugins, plugin)
	}

	return nil
}

// AddInfo retrieve the number of pages in the WP plugin store and the number of plugins to scan
func (w *Plugins) AddInfo() error {

	var info Plugins

	uri := w.URI + strconv.Itoa(w.Info.Page)

	body, err := requests.SendRequest(uri)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &info)
	if err != nil {
		return err
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
		return fmt.Errorf("Could not get current working directory in setOutPath() with error\n%s", err)
	}

	p.OutPath = dir + string(os.PathSeparator) + "current" + string(os.PathSeparator) + fileName

	return nil
}

func (p *Plugin) setDaysSinceUpdate() error {

	timeLayout := "2006-01-02 3:04pm MST"

	lastUpdateTime, err := time.Parse(timeLayout, p.LastUpdated)
	if err != nil {
		return fmt.Errorf("Failed to parse time in setDaysSinceUpdate() with error\n%s", err)
	}

	p.DaysSinceLastUpdate = strconv.Itoa(int(time.Now().Sub(lastUpdateTime).Hours() / 24))

	return nil
}

// VulnScan downloads the plugin, scans each php file for vulnerabilitys and sets it aside
// for later inspection if somethings found. Otherwise the plugin is then deleted.
func (p *Plugin) VulnScan(filesScanned *int) (bool, error) {
	*filesScanned++
	err := p.setDaysSinceUpdate()
	if err != nil {
		return false, err
	}
	err = p.setOutPath()
	if err != nil {
		return false, err
	}

	err = requests.Download(p.OutPath, p.DownloadLink)
	if err != nil {
		return false, err
	}

	scanResults := scanner.Results{
		Plugin:  p.Name,
		Results: make(map[string][]vulnerabilities.VulnResults),
	}

	err = scanner.ZipScan(p.OutPath, &scanResults)
	if err != nil {
		return false, err
	}

	if len(scanResults.Results) > 0 {
		err := p.moveToInspect()
		if err != nil {
			return false, err
		}
		err = p.saveResults(&scanResults)
		if err != nil {
			return false, err
		}
		color.Green.Printf("Potential Vulnerabilities found in plugin: %s\n", p.Name)
		color.Green.Println("Moving plugin to inspect/ folder.")
		return true, nil

	}
	err = os.Remove(p.OutPath)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (p *Plugin) moveToInspect() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get current working directory in setOutPath() with error\n%s", err)
	}

	p.inspectPath = dir + string(os.PathSeparator) + "inspect" + string(os.PathSeparator) + p.FileName
	err = os.Rename(p.OutPath, p.inspectPath)
	if err != nil {
		return fmt.Errorf("Could not move %s to inspect folder with error\n%s", p.Name, err)
	}

	return nil
}

func (p *Plugin) saveResults(results *scanner.Results) error {
	file, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		return fmt.Errorf("Could not MarshalIndent() results for file %s with error\n%s", p.Name, err)
	}

	err = ioutil.WriteFile(p.inspectPath+".txt", file, 0644)
	if err != nil {
		return fmt.Errorf("Could not save results file for %s with error \n%s", p.Name, err)
	}

	return nil
}
