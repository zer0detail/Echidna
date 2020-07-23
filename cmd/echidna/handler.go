package echidna

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// Start is exported to allow /cmd/echidna/main.go execute() to start the web app side
func webStart() {
	http.HandleFunc("/", echidnaStatus)
	http.HandleFunc("/begin", beginScanning)
	http.ListenAndServe("localhost:8080", nil)
}

func echidnaStatus(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("web\\main.html")
	if err != nil {
		log.Fatal(err)
	}
	err = html.Execute(w, Scanner)
}

func beginScanning(w http.ResponseWriter, r *http.Request) {
	go Scanner.Target.Scan(ctx)
	fmt.Println("Scanner started..")
	Scanner.Started = true
	http.Redirect(w, r, "/", http.StatusFound)
}
