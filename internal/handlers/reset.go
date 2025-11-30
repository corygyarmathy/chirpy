package handlers

import "net/http"

func (api *API) Reset(w http.ResponseWriter, r *http.Request) {
	api.FileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
