package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/corygyarmathy/chirpy/internal/database"
	"github.com/corygyarmathy/chirpy/internal/handlers"
	"github.com/corygyarmathy/chirpy/internal/server"
	_ "github.com/lib/pq"
)

func main() {
	const port = "8080"
	const filepathRoot = "."
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("DB SQL open error: %v\n", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil && err == nil {
			log.Fatalf("DB error: %v", cerr)
		}
	}()

	dbQueries := database.New(db)
	api := handlers.New(dbQueries)

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.metricsMiddleware(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
