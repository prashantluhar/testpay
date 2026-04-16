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
	"github.com/rs/zerolog"
)

func ListScenarios(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "ListScenarios").Msg("handler entry")
		list, err := s.ListScenarios(ctx, WorkspaceIDFromRequest(r, s))
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "ListScenarios").Msg("store error")
			http.Error(w, `{"error":"failed to list scenarios"}`, 500)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "ListScenarios").Int("count", len(list)).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

func CreateScenario(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		zerolog.Ctx(ctx).Info().Str("handler", "CreateScenario").Msg("handler entry")
		var sc store.Scenario
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "CreateScenario").Msg("invalid body")
			http.Error(w, `{"error":"invalid body"}`, 400)
			return
		}
		sc.ID = uuid.NewString()
		sc.WorkspaceID = WorkspaceIDFromRequest(r, s)
		if err := s.CreateScenario(ctx, &sc); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "CreateScenario").Str("scenario_id", sc.ID).Msg("store error")
			http.Error(w, `{"error":"failed to create scenario"}`, 500)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "CreateScenario").Str("scenario_id", sc.ID).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(sc)
	}
}

func GetScenario(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "GetScenario").Str("scenario_id", id).Msg("handler entry")
		sc, err := s.GetScenario(ctx, id)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "GetScenario").Str("scenario_id", id).Msg("store error")
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "GetScenario").Str("scenario_id", id).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sc)
	}
}

func UpdateScenario(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "UpdateScenario").Str("scenario_id", id).Msg("handler entry")
		var sc store.Scenario
		if err := json.NewDecoder(r.Body).Decode(&sc); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "UpdateScenario").Str("scenario_id", id).Msg("invalid body")
			http.Error(w, `{"error":"invalid body"}`, 400)
			return
		}
		sc.ID = id
		if err := s.UpdateScenario(ctx, &sc); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "UpdateScenario").Str("scenario_id", id).Msg("store error")
			http.Error(w, `{"error":"update failed"}`, 500)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "UpdateScenario").Str("scenario_id", id).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sc)
	}
}

func DeleteScenario(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "DeleteScenario").Str("scenario_id", id).Msg("handler entry")
		if err := s.DeleteScenario(ctx, id); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "DeleteScenario").Str("scenario_id", id).Msg("store error")
			http.Error(w, `{"error":"delete failed"}`, 500)
			return
		}
		zerolog.Ctx(ctx).Info().Str("handler", "DeleteScenario").Str("scenario_id", id).Msg("handler exit")
		w.WriteHeader(204)
	}
}

func RunScenario(s store.Store, eng *engine.Engine, reg *adapters.Registry, d *webhook.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")
		zerolog.Ctx(ctx).Info().Str("handler", "RunScenario").Str("scenario_id", id).Msg("handler entry")
		sc, err := s.GetScenario(ctx, id)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "RunScenario").Str("scenario_id", id).Msg("scenario lookup failed")
			http.Error(w, `{"error":"scenario not found"}`, 404)
			return
		}
		run := &store.ScenarioRun{
			ID:         uuid.NewString(),
			ScenarioID: sc.ID,
			Status:     "running",
		}
		if err := s.CreateScenarioRun(ctx, run); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "RunScenario").Str("run_id", run.ID).Msg("failed to create scenario run")
		}
		// Execute each step
		for i := range sc.Steps {
			eng.Execute(sc, i)
		}
		now := time.Now()
		run.Status = "completed"
		run.CompletedAt = &now
		if err := s.UpdateScenarioRun(ctx, run); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("handler", "RunScenario").Str("run_id", run.ID).Msg("failed to update scenario run")
		}
		zerolog.Ctx(ctx).Info().Str("handler", "RunScenario").Str("scenario_id", sc.ID).Str("run_id", run.ID).Int("steps", len(sc.Steps)).Msg("handler exit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(run)
	}
}
