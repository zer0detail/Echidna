package echidna

import (
	"log"

	"github.com/Paraflare/Echidna/pkg/wp"
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
	errChan := make(chan error)
	go wp.AllPluginScan(errChan)
	err := <-errChan
	log.Fatal(err)
}
