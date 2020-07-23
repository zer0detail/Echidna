package echidna

import (
	"log"

	"github.com/Paraflare/Echidna/pkg/wp"
)

var pluginList *wp.Plugins

// Execute is the entry point for echidna
func Execute() {
	// Create directories if they dont exist
	err := createEchidnaDirs()
	if err != nil {
		log.Fatal(err)
	}

	greeting()

	// PluginList exported allows the web server to retrieve its properties for display
	pluginList, err = wp.NewPlugins(ctx)
	if err != nil {
		log.Fatal(err)
	}

	webStart()
}
