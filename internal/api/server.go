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

	// Global middleware — CORS first so preflight requests short-circuit before
	// Auth/Session middleware tries to process OPTIONS.
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(cfg.Environment, "testpay"))
	r.Use(middleware.GatewayResolver)
	r.Use(middleware.Session(cfg.Auth.JWTSecret, cfg.Server.Mode))

	// Public health endpoint — unauthenticated, side-effect-free, cheap.
	// Used by uptime-ping services to keep the free-tier Render instance
	// warm (first call after 15 min idle costs ~30-60 s; pinging /healthz
	// every 5 min avoids that).
	r.Get("/healthz", handlers.Healthz())

	// Feedback endpoint — public so visitors on /docs can submit too.
	// Session context (if present) enriches the row with user/workspace ids.
	r.Post("/api/feedback", handlers.Feedback(s))

	// Mock gateway endpoints — one /{gateway}/* route per registered adapter.
	// The handler uses the original URL path to resolve the gateway; don't
	// strip the prefix. The mock handler itself enforces workspace auth via
	// Bearer api_key in hosted mode, so no extra middleware is needed here.
	// Rate limiting applies to mock endpoints (main traffic source) so
	// abusive callers can't blow through the free-tier hosting quota.
	mockHandler := handlers.NewMockWithMode(eng, reg, s, dispatcher, cfg.Server.Mode)
	perMin := cfg.RateLimit.RequestsPerMinute
	burst := cfg.RateLimit.Burst
	globalCap := perMin * 5 // cap all clients combined at 5× single-client rate
	mockLimiter := middleware.NewRateLimiter(perMin, burst, globalCap).Middleware
	for _, g := range reg.KnownGateways() {
		if g == "agnostic" {
			continue // agnostic is reached via /v1/*
		}
		r.With(mockLimiter).Handle("/"+g+"/*", mockHandler)
	}
	r.With(mockLimiter).Handle("/v1/*", mockHandler)

	// Signup/login are also common abuse targets — narrower limiter, per-IP
	// only (no global bucket, since legitimate signup traffic is low).
	authLimiter := middleware.NewRateLimiter(10, 3, 0).Middleware

	// Control API — /api/auth/* stays open; everything else requires a session.
	r.Route("/api", func(r chi.Router) {
		// Public auth routes.
		r.With(authLimiter).Post("/auth/signup", handlers.Signup(s, cfg.Auth.JWTSecret, cfg.Server.Mode))
		r.With(authLimiter).Post("/auth/login", handlers.Login(s, cfg.Auth.JWTSecret, cfg.Server.Mode))
		r.Post("/auth/logout", handlers.Logout(cfg.Server.Mode))
		r.Get("/auth/me", handlers.Me(s))

		// Authenticated routes — require a valid session cookie. Bearer api_key
		// auth is deliberately NOT accepted here; it's only for mock endpoints.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Get("/gateways", handlers.ListGateways(reg))
			r.Get("/workspace", handlers.GetWorkspace(s))
			r.Put("/workspace", handlers.UpdateWorkspace(s))
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
				r.Get("/", handlers.ListWebhooks(s))
				r.Post("/test", handlers.TestWebhook(dispatcher))
				r.Get("/{id}", handlers.GetWebhook(s))
				r.Get("/{id}/status", handlers.GetWebhookStatus(s))
			})
		})
	})

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
	}
}
