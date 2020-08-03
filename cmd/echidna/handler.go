package echidna

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type handleWrapper struct {
	ctx   context.Context
	errCh chan error
	scanner *targetScanner
}

// Start is exported to allow /cmd/echidna/main.go execute() to start the web app side
func webStart(ctx context.Context, errCh chan error, scanner *targetScanner) {
	handler := handleWrapper {
		ctx: ctx,
		errCh: errCh,
		scanner: scanner,
	}
	http.HandleFunc("/", handler.echidnaStatus)
	http.HandleFunc("/begin", handler.beginScanning)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *handleWrapper) echidnaStatus(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("web\\main.html")
	if err != nil {
		log.Fatal(err)
	}
	err = html.Execute(w, h.scanner)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *handleWrapper) beginScanning(w http.ResponseWriter, r *http.Request) {
	go h.scanner.Target.Scan(h.ctx, h.errCh)
	fmt.Println("Scanner started..")
	h.scanner.Started = true
	http.Redirect(w, r, "/", http.StatusFound)
}
