package echidna

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Paraflare/Echidna/pkg/wp"
	"github.com/gookit/color"
	"golang.org/x/sync/semaphore"
)

var (
	sem          = semaphore.NewWeighted(int64(70))
	seed         = rand.NewSource(time.Now().Unix())
	randomPicker = rand.New(seed)
)

func init() {
	// Create directories if they dont exist
	err := createEchidnaDirs()
	if err != nil {
		log.Fatal(err)
	}
	setupCloseHandler()
}

// Execute is the entry point for echidna
func Execute() {

	pluginList := wp.NewPlugins()
	// Initial Plugin Object setup. Add the reusable http client and the wp plugins URI properties
	color.Yellow.Printf("Found %d pages of plugins to scan..\n", pluginList.Info.Pages)

	var filesScanned int
	ctx := context.Background()

	for filesScanned != pluginList.Info.Results {

		// if we have plugins, scan them
		if len(pluginList.Plugins) > 0 {
			sem.Acquire(ctx, 1)
			// Choose a random plugin
			randPluginIndex := randomPicker.Intn(len(pluginList.Plugins))
			plugin := pluginList.Plugins[randPluginIndex]
			// Remove it from the list
			pluginList.RemovePlugin(randPluginIndex)

			go func() {
				plugin.VulnScan(&filesScanned)
				sem.Release(1)
			}()

			color.Yellow.Print("Plugin count: ")
			color.Gray.Printf("%d\t", len(pluginList.Plugins))
			color.Yellow.Print("Files Scanned: ")
			color.Gray.Printf("%d\n", filesScanned)
		}

		// If we haven't finished pulling the list of plugins from the store, grab another page and
		// add it to pluginList.Plugins
		if pluginList.Info.Page <= pluginList.Info.Pages {
			sem.Acquire(ctx, 1)
			go func() {
				pluginList.AddPlugins()
				sem.Release(1)
			}()
			// if err != nil {
			// 	log.Fatal(err)
			// }
		}

	}
	fmt.Println("Finished scanning all plugins. Happy Hunting!")
}
