package echidna

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gookit/color"
)

var ctx context.Context

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
		fmt.Println("Ctrl+C detected. Cancelling GoRoutines.")
		cancel()
		time.Sleep(3 * time.Second)
		fmt.Println("Attempting to remove current/ directory")
		deleteCurrentDir()
		os.Exit(0)
	}()

	greeting()
	webStart()
}

func greeting() {
	color.Yellow.Println(`
	   _     _     _             
          | |   (_)   | |            
  ___  ___| |__  _  __| |_ __   __ _ 
 / _ \/ __| '_ \| |/ _' | '_ \ / _' |
|  __/ (__| | | | | (_| | | | | (_| |
 \___|\___|_| |_|_|\__,_|_| |_|\__,_|`)

	color.LightBlue.Println("Echidna Scanner running. Browse to http://127.0.0.1:8080 to view status.")
}
