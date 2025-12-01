package handlers

import (
	"log"
	"net/http"
)

func Readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Printf("readiness: error writing response: %v", err)
		return
	}
}
