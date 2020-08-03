package echidna

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Paraflare/Echidna/pkg/wp"
	flags "github.com/jessevdk/go-flags"
)

// targetScanner is the struct that holds the target interface which we inject with
// whatever target we want to scan. My attempt at DI.
type targetScanner struct {
	Target  target
	Started bool
}

type target interface {
	Scan(context.Context, chan error)
}

func newScanner(t target) *targetScanner {
	return &targetScanner{
		Target: t,
	}
}

var (
	scanner     *targetScanner
	errChan     = make(chan error)
	ctx, cancel = context.WithCancel(context.Background())
)

// Execute is the entry point for echidna
func Execute() {
	// Create directories if they dont exist
	err := createEchidnaDirs()
	if err != nil {
		log.Fatal(err)
	}

	greeting()
	go setupCloseHandler(ctx, cancel)
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
		scanner = newScanner(plugins)
	case opts.Themes:
		log.Fatal("This functionality isn't built yet. We only have WordPress Plugins for now.")
	default:
		fmt.Println("No target selected. Creating a WordPress Plugin Scanner as a default.")
		plugins, err := wp.NewPlugins(ctx)
		if err != nil {
			log.Fatal(err)
		}
		scanner = newScanner(plugins)
	}

	// if the user selected web (-w or --web) from the commandline then start
	// the webserver, otherwise kick off the cli version.
	if opts.Web {
		webStart()
	} else {
		scanner.Started = true
		scanner.Target.Scan(ctx, errChan)
	}

	select {
	case <-ctx.Done():
		for {
			// context was canceled infinite loop while we wait
			// for the closehandler to do its things
		}
	default:
		os.Exit(0)
	}

}
