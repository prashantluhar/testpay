# TestPay Backend

Go implementation of the TestPay simulation engine, adapters, HTTP API, middleware stack, webhook dispatcher, and Postgres data layer. All packages live under `internal/` — consumed only by `cmd/testpay` and `cli/`.

See the top-level [README.md](../README.md) for what TestPay is and how to run it end-to-end.

---

## Package tour

| Package | Responsibility |
|---|---|
| `internal/config` | YAML + env loader. `Config` struct covers server, database, logging, webhook, auth, CORS, cloud, integrations, rate-limit. Secrets referenced via `*_env` keys — never inlined in YAML. |
| `internal/engine` | The failure-mode simulator. `FailureMode` constants + `Engine.Execute(scenario, stepIndex)` returning `*Result` (HTTP status, error code, webhook flags, pending state). Gateway-agnostic — the same engine drives Stripe, Razorpay, and Agnostic adapters. |
| `internal/adapters` | Gateway-specific wire-format translators. `Adapter` interface + `Registry` that routes by URL prefix. Sub-packages `stripe/`, `razorpay/`, `agnostic/` each implement `BuildResponse` and `BuildWebhookPayload`. |
| `internal/store` | `Store` interface + domain models (Workspace, User, Scenario, Session, RequestLog, WebhookLog). |
| `internal/store/postgres` | pgx/v5 implementation with `embed.FS` migrations (auto-applied at boot). Slow-query logging for any call > 100ms. |
| `internal/webhook` | Webhook dispatcher with exponential backoff retry. `DispatchAsync` fires in a goroutine and persists per-attempt logs. |
| `internal/api` | Chi router wiring, middleware chain, all HTTP handlers. |
| `internal/api/middleware` | RequestID, Logger (trace-id context), GatewayResolver, Session (JWT cookie), Auth (API key), Capture (response snapshot). |
| `internal/api/handlers` | `mock.go` (gateway simulation), `auth.go` (signup/login/logout/me), `scenarios.go`, `sessions.go`, `logs.go`, `webhooks.go`, `workspace.go`. |
| `internal/observability` | `zerolog` global logger setup based on YAML config. |

---

## HTTP surface

### Mock gateway endpoints (developer's app calls these)

```
POST   /stripe/v1/*         # Stripe-shaped responses
POST   /razorpay/v1/*       # Razorpay-shaped responses
POST   /v1/*                # Agnostic JSON shape
```

Every request runs the configured default scenario (or an active session scenario) through the engine and returns the shaped response.

### Control API (dashboard / CLI / CI)

```
# Auth (hosted mode)
POST   /api/auth/signup     # {email, password} → creates user + workspace, sets cookie
POST   /api/auth/login      # validates credentials, sets cookie
POST   /api/auth/logout     # clears cookie
GET    /api/auth/me         # current user + workspace

# Scenarios
GET    /api/scenarios
POST   /api/scenarios
GET    /api/scenarios/:id
PUT    /api/scenarios/:id
DELETE /api/scenarios/:id
POST   /api/scenarios/:id/run

# Sessions (pin a scenario to the mock endpoint for a window)
POST   /api/sessions
DELETE /api/sessions/:id

# Logs
GET    /api/logs
GET    /api/logs/:id
POST   /api/logs/:id/replay

# Webhooks
POST   /api/webhooks/test
GET    /api/webhooks/:id/status

# Workspace
GET    /api/workspace
```

---

## Middleware chain (order matters)

```
chi.Recoverer
  └─ RequestID               # UUID trace id in context + X-Request-ID header
     └─ Logger               # zerolog in context with trace_id, env, service fields
        └─ GatewayResolver   # infers gateway from URL path → context
           └─ Session        # JWT cookie → user_id + workspace_id in context (local mode injects LocalWorkspaceID)
              └─ Auth        # API key check (hosted mode, for mock endpoints)
                 └─ handler
```

Every handler, store call, webhook attempt, and error path uses `zerolog.Ctx(ctx)` — so every log line carries `trace_id`, `env`, `service`, and whatever fields the call site added.

---

## Auth model

- **Local mode** (`cfg.Server.Mode == "local"`) — no auth. `Session` middleware injects `LocalWorkspaceID` and lets everything through. Auth screens are hidden in the dashboard.
- **Hosted mode** — email + password. Signup creates one user + one workspace. JWT (HS256) signed with `JWT_SECRET`, stored in an httpOnly `testpay_session` cookie, 30-day expiry. Bcrypt cost 12. No email verification, no reset — defer to v2.

The CLI refuses to start in hosted mode if `JWT_SECRET` is unset.

---

## Configuration

Loaded from three sources, highest precedence first:

1. Environment variables (e.g. `PORT`, `DATABASE_URL`, `API_KEY`)
2. YAML config file (`--config deploy/config/testpay.<env>.yaml`)
3. Built-in defaults (in `config.defaults()`)

The YAML file references env var **names** via `*_env` keys (e.g. `database.url_env: DATABASE_URL`) — secrets are never stored in YAML.

Per-environment files live in `deploy/config/`:
- `testpay.local.yaml` — dev laptop
- `testpay.dev.yaml` — shared dev
- `testpay.test.yaml` — CI
- `testpay.stage.yaml` — pre-prod
- `testpay.prod.yaml` — production

---

## Running locally

```bash
# From the repo root
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/testpay?sslmode=disable"
go run ./cmd/testpay start --config deploy/config/testpay.local.yaml
```

Migrations run automatically on startup. A default workspace (slug `local`, id `00000000-0000-0000-0000-000000000001`) is seeded if it doesn't exist.

---

## Testing

```bash
# Unit tests (no DB)
go test ./...

# Integration tests (requires Postgres)
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/testpay?sslmode=disable"
go test ./internal/store/postgres/... -v

# Coverage — CI gate is 90%
make coverage-check

# Lint
go vet ./...
```

### What's covered by tests

- **Engine** — `success`, `bank_decline_hard`, `webhook_missing`, multi-step sequencing, out-of-bounds index wrap
- **Adapters** — Stripe/Razorpay/Agnostic response shapes + webhook payload shapes, registry resolution
- **Middleware** — RequestID propagation, Logger edge+handler logs, Auth local/hosted, capture, Session cookie validation, RequireSession 401
- **Handlers** — mock handler integration, scenario list, auth signup/login happy path + wrong password
- **Store** — all Postgres CRUD methods (integration-gated on `TEST_DATABASE_URL`)
- **Webhook** — dispatcher success + retry on failure
- **Config** — YAML load + env overrides + missing-secret error

---

## Adding a new gateway

1. Create `internal/adapters/<name>/adapter.go` implementing the `Adapter` interface (`Name`, `BuildResponse`, `BuildWebhookPayload`)
2. Register it in `internal/adapters/registry.go` — one line
3. Add the URL prefix mount in `internal/api/server.go` — one line
4. Write `adapter_test.go` covering at least success + one decline

The engine never changes.

---

## Observability

All logs go to stdout as JSON (or console format in local mode per YAML). Every line carries:

- `trace_id` — correlates all logs for one request
- `env` — `local` / `dev` / `stage` / `prod`
- `service` — `testpay`
- `time` / `level` / `msg` — standard zerolog fields

Fluent bit, Datadog agent, Splunk OTel collector, etc. pick up stdout — no app-side shipping code.

Slow Postgres queries (>100ms) and errors log at `warn`/`error` with the query name + duration. Webhook dispatcher logs every attempt with target, status, and duration.

---

## Branch & release

See [`../CONTRIBUTING.md`](../CONTRIBUTING.md).
