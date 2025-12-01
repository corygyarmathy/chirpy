package handlers

import (
	"log"
	"net/http"
)

func (api *API) Reset(w http.ResponseWriter, r *http.Request) {
	// Only allow in dev
	if api.platform != "dev" {
		http.Error(w, "Reset is only allowed in dev environment.", http.StatusForbidden)
		return
	}

	api.FileserverHits.Store(0)

	if err := api.DB.ResetUsers(r.Context()); err != nil {
		log.Printf("reset: error resetting users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hits reset to 0")); err != nil {
		log.Printf("reset: error writing response: %v", err)
	}
}
