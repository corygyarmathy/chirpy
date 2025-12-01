package handlers

import (
	"log"
	"net/http"
)

func (api *API) Reset(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	response := []byte("Hits reset to 0")

	if api.platform != "dev" {
		statusCode = http.StatusForbidden
		w.WriteHeader(statusCode)
		response = []byte("Reset is only allowed in dev environment.")
		if _, err := w.Write(response); err != nil {
			log.Printf("reset: error writing response: %v", err)
			return
		}
		return
	}

	api.FileserverHits.Store(0)
	err := api.DB.ResetUsers(r.Context())
	if err != nil {
		statusCode = http.StatusInternalServerError
		log.Printf("reset: error resetting users: %v", err)
	}

	w.WriteHeader(statusCode)
	if _, err := w.Write(response); err != nil {
		log.Printf("reset: error writing response: %v", err)
		return
	}
}
