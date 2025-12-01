// Package server handles server operations & routing
package server

import (
	"net/http"

	"github.com/corygyarmathy/chirpy/internal/handlers"
)

func NewMux(api *handlers.API) *http.ServeMux {
	const filepathRoot = "."

	mux := http.NewServeMux()

	mux.Handle("/app/",
		api.MetricsMiddleware(
			http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))),
		),
	)

	mux.HandleFunc("GET /api/healthz", handlers.Readiness)
	mux.HandleFunc("POST /api/chirps", api.CreateChirp)
	mux.HandleFunc("POST /api/users", api.CreateUser)
	mux.HandleFunc("GET /admin/metrics", api.Metrics)
	mux.HandleFunc("POST /admin/reset", api.Reset)

	return mux
}
