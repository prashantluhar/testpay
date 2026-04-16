package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/prashantluhar/testpay/internal/store"
)

func GetWorkspace(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := s.GetWorkspaceBySlug(r.Context(), "local")
		if err != nil {
			http.Error(w, `{"error":"workspace not found"}`, 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ws)
	}
}
