package handlers

import (
	"fmt"
	"log"
	"net/http"
)

func (api *API) Metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := api.FileserverHits.Load()
	body := fmt.Sprintf(`
<html>
    <body>
	      <h1>Welcome, Chirpy Admin</h1>
	      <p>Chirpy has been visited %v times!</p>
    </body>
<html>
		`, hits)
	_, err := w.Write([]byte(body))
	if err != nil {
		log.Fatal("Failed to write http body")
	}
}

func (api *API) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.FileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}
