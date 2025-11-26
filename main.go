package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	const port = "8080"
	const filepathRoot = "."

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
		if err != nil {
			log.Fatal("Failed to write http body")
		}

	})

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(s.ListenAndServe())
}
