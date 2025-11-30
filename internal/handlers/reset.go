package handlers

import (
	"log"
	"net/http"
)

func (api *API) Reset(w http.ResponseWriter, r *http.Request) {
	api.FileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hits reset to 0")); err != nil {
		log.Printf("Error writing data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
