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
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func echidnaStatus(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("web\\main.html")
	if err != nil {
		log.Fatal(err)
	}
	err = html.Execute(w, scanner)
	if err != nil {
		log.Fatal(err)
	}
}

func beginScanning(w http.ResponseWriter, r *http.Request) {
	go scanner.Target.Scan(ctx, errChan)
	fmt.Println("Scanner started..")
	scanner.Started = true
	http.Redirect(w, r, "/", http.StatusFound)
}
