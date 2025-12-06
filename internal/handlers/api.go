// Package handlers contains all command handlers, and API struct
package handlers

import (
	"sync/atomic"

	"github.com/corygyarmathy/chirpy/internal/database"
)

type API struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}

func New(db *database.Queries, platform string, jwtSecret string, polkaKey string) *API {
	return &API{
		FileserverHits: atomic.Int32{},
		DB:             db,
		platform:       platform,
		jwtSecret:      jwtSecret,
		polkaKey:       polkaKey,
	}
}
