package echidna

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Paraflare/Echidna/pkg/wp"
)

var ctx context.Context
var pluginList *wp.Plugins

// Execute is the entry point for echidna
func Execute() {
	// Create directories if they dont exist
	err := createEchidnaDirs()
	if err != nil {
		log.Fatal(err)
	}
	// set up context for cancelling goroutines
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	// set up goroutine to catch CTRL+C and execute cleanup
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		fmt.Println("Ctrl+C detected. Cancelling scanning goroutines and current web requests.")
		// Give the goroutines time to return and free up access to the zip files
		// so when we delete them we have access
		time.Sleep(2 * time.Second)
		fmt.Println("Attempting to remove current/ directory")
		deleteCurrentDir()
		os.Exit(0)
	}()

	greeting()

	// PluginList exported allows the web server to retrieve its properties for display
	pluginList, err = wp.NewPlugins(ctx)
	if err != nil {
		log.Fatal(err)
	}

	webStart()
}
