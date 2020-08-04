package wp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Paraflare/Echidna/pkg/requests"
	"github.com/Paraflare/Echidna/pkg/vulnerabilities"
)

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

// Download downloads the plugins zip file and places it into the current/ folder. 
// It then sends the plugin info down the workQueue ready for the scanWorkers to pick up and scan.
func (p *Plugin) Download(ctx context.Context, plugins *Plugins, errChan chan error, workQueue chan *Plugin) {

	err := p.setDaysSinceUpdate()
	if err != nil {
		errChan <- fmt.Errorf("Plugin.go:Download() failed to setDaysSinceUpdate() with error: %s", err)
	}
	err = p.setOutPath()
	if err != nil {
		// This error will stop us from downloading the plugin properly. so we skip it early.
		errChan <- fmt.Errorf("Plugin.go:Download() failed to setOutPath() with error: %s", err)
		return
	}

	err = requests.Download(ctx, plugins.client, p.OutPath, p.DownloadLink)
	if err != nil {
		// Plugin downloads fail for various reasons. If we fail to download one just skip it
		// there's ~50,000 plugins atm, dropping some is ok.
		errChan <- fmt.Errorf("Plugin.go:Download() requests.Download(%s) failed with error %s", p.DownloadLink, err)
		plugins.scanMu.Lock()
		plugins.Skipped++
		plugins.scanMu.Unlock()
		return
	}

	// Block until we are cancelled or the queue opens up
	select {
	case <-ctx.Done():
		return
	case workQueue <- p:
		return
	}

}

func (p *Plugin) moveToInspect(results *vulnerabilities.Results) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Plugin.go:moveToInspect() - Could not get current working directory with os.Getwd() with error: %s", err)
	}
	// Split results out into the results for each vulnerability module (e.g xss, sqli, etc)
	for k := range results.Modules {
		p.inspectPath = dir + string(os.PathSeparator) + "inspect" + string(os.PathSeparator) + k
		// if a folder for that vuln module doesnt already exist, create it.
		if _, err := os.Stat(p.inspectPath); os.IsNotExist(err) {
			err = os.MkdirAll(p.inspectPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("Plugin.go:moveToInspect() - Failed to create directory %s with error: %s", p.inspectPath, err)
			}
		}

		outfile := p.inspectPath + string(os.PathSeparator) + p.FileName

		src, err := os.Open(p.OutPath)
		if err != nil {
			return fmt.Errorf("Plugin.go:moveToInspect() - failed to os.Open(%s) with error: %s", p.OutPath, err)
		}
		defer src.Close()
		dst, err := os.Create(outfile)
		if err != nil {
			return fmt.Errorf("Plugin.go:moveToInspect() - failed to os.Create(%s) with error: %s", p.inspectPath, err)
		}
		defer dst.Close()

		_, err = io.Copy(src, dst)
		if err != nil {
			return fmt.Errorf("Plugin.go:moveToInspect() - Could not io.Copy() %s to inspect folder with error: %s", p.Name, err)
		}
	}
	return nil
}

func (p *Plugin) saveResults(results *vulnerabilities.Results) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Plugin.go:saveResults() - Could not get current working directory with os.Getwd() with error: %s", err)
	}
	for k := range results.Modules {
		file, err := json.MarshalIndent(results.Modules[k], "", " ")
		if err != nil {
			return fmt.Errorf("Plugin.go:saveResults() - Could not MarshalIndent() results for file %s with error: %s", p.Name, err)
		}
		p.inspectPath = dir + string(os.PathSeparator) + "inspect" + string(os.PathSeparator) + k + string(os.PathSeparator) + p.FileName
		err = ioutil.WriteFile(p.inspectPath+".txt", file, 0644)
		if err != nil {
			return fmt.Errorf("Plugin.go:saveResults() - Could not save results file for %s with error: %s", p.Name, err)
		}
	}

	return nil
}

func (p *Plugin) setOutPath() error {

	fileName := strings.Split(p.DownloadLink, "/")[4]
	fileName = strconv.Itoa(p.ActiveInstalls) + "_" + p.DaysSinceLastUpdate + "_" + fileName
	p.FileName = fileName

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Plugin.go:setOutPath() - os.Getwd() with error: %s", err)
	}

	p.OutPath = dir + string(os.PathSeparator) + "current" + string(os.PathSeparator) + fileName

	return nil
}

func (p *Plugin) setDaysSinceUpdate() error {

	timeLayout := "2006-01-02 3:04pm MST"

	lastUpdateTime, err := time.Parse(timeLayout, p.LastUpdated)
	if err != nil {
		return fmt.Errorf("Plugin.go:setDaysSinceUpdate() time.Parse(%s, %s) failed with error: %s", timeLayout, p.LastUpdated, err)
	}

	p.DaysSinceLastUpdate = strconv.Itoa(int(time.Since(lastUpdateTime).Hours() / 24))

	return nil
}
