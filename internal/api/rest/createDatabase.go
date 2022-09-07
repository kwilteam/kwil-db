package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kwilteam/kwil-db/internal/api/types"
	"github.com/rs/zerolog/log"
)

// CreateDatabase Handler for CreateDatabase
func (h *Handler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var db types.CreateDatabaseMsg

	// Validate JSON
	err := json.NewDecoder(r.Body).Decode(&db)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create database
	err = h.Service.CreateDatabase(context.Background(), &db)
	if err != nil {
		log.Error().Err(err).Msg("failed to create database")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
