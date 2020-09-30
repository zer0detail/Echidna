package echidna

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Paraflare/Echidna/pkg/wp"
)

// targetScanner is the struct that holds the target interface which we inject with
// whatever target we want to scan. My attempt at DI.
type targetScanner struct {
	Target  target
	Started bool
}

type target interface {
	Scan(context.Context, chan error)
	AddInfo(context.Context) error
}

func newScanner(t target) *targetScanner {
	return &targetScanner{
		Target: t,
	}
}

// Execute is the entry point for echidna
func Execute() {

	exitCh := make(chan bool, 1)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	var scanner *targetScanner
	//var opts options

	//opts.Parse()

	// Create directories if they dont exist
	err := createEchidnaDirs()
	if err != nil {
		log.Fatal(err)
	}

	greeting()
	go closeHandler(ctx, cancel, exitCh)
	go errorHandler(ctx, errCh)

	// not used until i decide to add a web UI or themes scanner
	// Inject our target into the scanner based on the users choice (-p/--plugin or -t/--theme)
	// Select WordPress Plugins as a target if nothing is selected
	// if *opts.Plugins {
	fmt.Println("Preparing WordPress Plugin Scanner.")
	plugins, err := wp.NewPlugins(ctx)
	if err != nil {
		log.Fatal(err)
	}
	scanner = newScanner(plugins)
	// }
	// if *opts.Themes {
	// 	log.Fatal("This functionality isn't built yet. We only have WordPress Plugins for now.")
	// }

	// Add the initial scanner information such as pages, # of objects to scan, etc
	err = scanner.Target.AddInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// not used until I decide to add web UI or theme scanner
	// if the user selected web (-w or --web) from the commandline then start
	// // the webserver, otherwise kick off the cli version.
	// if *opts.Web {
	// 	go webStart(ctx, errCh, scanner)
	// } else {
	// 	scanner.Started = true
	// 	go scanner.Target.Scan(ctx, errCh)
	// }
	scanner.Started = true
	go scanner.Target.Scan(ctx, errCh)

	// block here until we are finished or have received a cancel()
	select {
	case <-ctx.Done():
		fmt.Printf("\nExecution canceled. Waiting for close handler to perform cleanup.")
		<-exitCh
		fmt.Println("Cleanup complete.")
		os.Exit(0)
	case <-exitCh:
		os.Exit(0)
	}

}
