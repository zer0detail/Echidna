package web

import (
	"fmt"
	"net/http"

	"github.com/Paraflare/Echidna/pkg/wp"
)

// Start is exported to allow /cmd/echidna/main.go execute() to start the web app side
func Start() {
	http.HandleFunc("/", echidnaStatus)
	http.ListenAndServe("localhost:8080", nil)
}

func echidnaStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Current pages retrieved: %d\n", wp.PluginList.Info.Page)
	fmt.Fprintf(w, "Current plugins scanned: %d\n", wp.PluginList.FilesScanned)
}
