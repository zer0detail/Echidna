package echidna

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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
	err := createDirs()
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

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Ctrl+C detected. Cleaning up current scan directory.")
		time.Sleep(3 * time.Second)
		deleteCurrentDir()
		os.Exit(0)
	}()
}

func createDirs() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get current working directory in CreateDirs() with error\n%s", err)
	}

	current := filepath.Join(dir, "current")
	err = os.MkdirAll(current, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not create directory %s in createdirs() with error\n%s", current, err)
	}
	inspect := filepath.Join(dir, "inspect")
	err = os.MkdirAll(inspect, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not create directory %s in createdirs() with error\n%s", inspect, err)
	}

	return nil

}

func deleteCurrentDir() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get current working directory in deleteCurrentDir(). Please manually remove the folder")
	}

	current := filepath.Join(dir, "current")
	err = os.RemoveAll(current)
	if err != nil {
		log.Fatal("Failed to remove Current/ directory and some subfiles. Please remove them manually")
	}
}
