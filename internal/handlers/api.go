// Package handlers contains all command handlers, and API struct
package handlers

import (
	"os"
	"sync/atomic"

	"github.com/corygyarmathy/chirpy/internal/database"
)

type API struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	platform       string
}

func New(db *database.Queries) *API {
	return &API{
		FileserverHits: atomic.Int32{},
		DB:             db,
		platform:       os.Getenv("PLATFORM"),
	}
}
