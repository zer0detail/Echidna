package echidna

import (
	"log"

	"github.com/Paraflare/Echidna/web"
	"github.com/gookit/color"
)

func init() {
	// Create directories if they dont exist
	err := createEchidnaDirs()
	if err != nil {
		log.Fatal(err)
	}
	setupCloseHandler()
	greeting()
}

// Execute is the entry point for echidna
func Execute() {
	go web.Start()
	// go wp.AllPluginScan()
	for {

	}

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
