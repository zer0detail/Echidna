package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/Paraflare/Echidna/pkg/wp"
)

// Start is exported to allow /cmd/echidna/main.go execute() to start the web app side
func Start() {
	http.HandleFunc("/", echidnaStatus)
	http.HandleFunc("/begin", beginScanning)
	http.ListenAndServe("localhost:8080", nil)
}

func echidnaStatus(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("web\\main.html")
	if err != nil {
		log.Fatal(err)
	}
	err = html.Execute(w, wp.PluginList)
}

func beginScanning(w http.ResponseWriter, r *http.Request) {
	go wp.AllPluginScan()
	fmt.Println("Scanner started..")
	wp.PluginList.Started = true
	http.Redirect(w, r, "/", http.StatusFound)
}
