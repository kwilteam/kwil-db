package rest

import (
	"context"
	"encoding/json"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog/log"
	"net/http"
)

// CreateDatabase Handler for CreateDatabase
func (h *Handler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var db types.CreateDatabase

	// Validate JSON
	err := json.NewDecoder(r.Body).Decode(&db)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create database
	err = h.Service.KDB.CreateDatabase(context.Background(), &db)
	if err != nil {
		log.Error().Err(err).Msg("failed to create database")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
