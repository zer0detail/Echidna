package echidna

import (
	"context"
	"fmt"
	"log"

	"github.com/Paraflare/Echidna/pkg/wp"
	flags "github.com/jessevdk/go-flags"
)

type scanner struct {
	Target  target
	Started bool
}

type target interface {
	Scan(context.Context, chan error)
}

func newScanner(t target) *scanner {
	return &scanner{
		Target: t,
	}
}

// Scanner is the struct that holds the target interface which we inject with
// whatever target we want to scan. My attempt at DI.
var (
	Scanner *scanner
	errChan = make(chan error)
)

// Execute is the entry point for echidna
func Execute() {
	// Create directories if they dont exist
	err := createEchidnaDirs()
	if err != nil {
		log.Fatal(err)
	}

	greeting()
	setupCloseHandler()
	go errorHandler(ctx, errChan)

	_, err = flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	// Inject our target into the scanner based on the users choice (-p/--plugin or -t/--theme)
	// Select WordPress Plugins as a target if nothing is selected
	switch {
	case opts.Plugins:
		fmt.Println("Preparing WordPress Plugin Scanner.")
		plugins, err := wp.NewPlugins(ctx)
		if err != nil {
			log.Fatal(err)
		}
		Scanner = newScanner(plugins)
	case opts.Themes:
		log.Fatal("This functionality isn't built yet. We only have WordPress Plugins for now.")
	default:
		fmt.Println("No target selected. Creating a WordPress Plugin Scanner as a default.")
		plugins, err := wp.NewPlugins(ctx)
		if err != nil {
			log.Fatal(err)
		}
		Scanner = newScanner(plugins)
	}

	// if the user selected web (-w or --web) from the commandline then start
	// the webserver, otherwise kick off the cli version.
	if opts.Web {
		webStart()
	} else {
		Scanner.Started = true
		Scanner.Target.Scan(ctx, errChan)
	}

}
