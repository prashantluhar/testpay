package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
)

func ListScenarios(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.ListScenarios(r.Context(), store.LocalWorkspaceID)
		if err != nil {
			http.Error(w, `{"error":"failed to list scenarios"}`, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

func CreateScenario(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sc store.Scenario
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			http.Error(w, `{"error":"invalid body"}`, 400)
			return
		}
		sc.ID = uuid.NewString()
		sc.WorkspaceID = store.LocalWorkspaceID
		if err := s.CreateScenario(r.Context(), &sc); err != nil {
			http.Error(w, `{"error":"failed to create scenario"}`, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(sc)
	}
}

func GetScenario(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sc, err := s.GetScenario(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sc)
	}
}

func UpdateScenario(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sc store.Scenario
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			http.Error(w, `{"error":"invalid body"}`, 400)
			return
		}
		sc.ID = chi.URLParam(r, "id")
		if err := s.UpdateScenario(r.Context(), &sc); err != nil {
			http.Error(w, `{"error":"update failed"}`, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sc)
	}
}

func DeleteScenario(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.DeleteScenario(r.Context(), chi.URLParam(r, "id")); err != nil {
			http.Error(w, `{"error":"delete failed"}`, 500)
			return
		}
		w.WriteHeader(204)
	}
}

func RunScenario(s store.Store, eng *engine.Engine, reg *adapters.Registry, d *webhook.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sc, err := s.GetScenario(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, `{"error":"scenario not found"}`, 404)
			return
		}
		run := &store.ScenarioRun{
			ID:         uuid.NewString(),
			ScenarioID: sc.ID,
			Status:     "running",
		}
		s.CreateScenarioRun(r.Context(), run)
		// Execute each step
		for i := range sc.Steps {
			eng.Execute(sc, i)
		}
		now := time.Now()
		run.Status = "completed"
		run.CompletedAt = &now
		s.UpdateScenarioRun(r.Context(), run)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(run)
	}
}
