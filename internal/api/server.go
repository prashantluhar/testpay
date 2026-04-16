package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prashantluhar/testpay/internal/adapters"
	"github.com/prashantluhar/testpay/internal/api/handlers"
	"github.com/prashantluhar/testpay/internal/api/middleware"
	"github.com/prashantluhar/testpay/internal/config"
	"github.com/prashantluhar/testpay/internal/engine"
	"github.com/prashantluhar/testpay/internal/store"
	"github.com/prashantluhar/testpay/internal/webhook"
)

func NewServer(cfg *config.Config, s store.Store) *http.Server {
	r := chi.NewRouter()

	eng := engine.New()
	reg := adapters.NewRegistry()
	dispatcher := webhook.NewDispatcher(
		cfg.Webhook.MaxAttempts,
		time.Duration(cfg.Webhook.BaseDelayMs)*time.Millisecond,
	)

	// Global middleware
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(cfg.Environment, "testpay"))
	r.Use(middleware.GatewayResolver)
	r.Use(middleware.Auth(cfg.Server.Mode, cfg.Auth.APIKey))

	// Mock gateway endpoints (Stripe, Razorpay, Agnostic)
	mockHandler := handlers.NewMock(eng, reg, s, dispatcher)
	r.Mount("/stripe", http.StripPrefix("/stripe", mockHandler))
	r.Mount("/razorpay", http.StripPrefix("/razorpay", mockHandler))
	r.Mount("/v1", http.StripPrefix("/v1", mockHandler))

	// Control API
	r.Route("/api", func(r chi.Router) {
		r.Get("/workspace", handlers.GetWorkspace(s))
		r.Route("/scenarios", func(r chi.Router) {
			r.Get("/", handlers.ListScenarios(s))
			r.Post("/", handlers.CreateScenario(s))
			r.Get("/{id}", handlers.GetScenario(s))
			r.Put("/{id}", handlers.UpdateScenario(s))
			r.Delete("/{id}", handlers.DeleteScenario(s))
			r.Post("/{id}/run", handlers.RunScenario(s, eng, reg, dispatcher))
		})
		r.Route("/sessions", func(r chi.Router) {
			r.Post("/", handlers.CreateSession(s))
			r.Delete("/{id}", handlers.DeleteSession(s))
		})
		r.Route("/logs", func(r chi.Router) {
			r.Get("/", handlers.ListLogs(s))
			r.Get("/{id}", handlers.GetLog(s))
			r.Post("/{id}/replay", handlers.ReplayLog(s, eng, reg, dispatcher))
		})
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/test", handlers.TestWebhook(dispatcher))
			r.Get("/{id}/status", handlers.GetWebhookStatus(s))
		})
	})

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
	}
}
