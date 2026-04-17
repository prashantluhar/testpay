# TestPay

[![Build](https://img.shields.io/github/actions/workflow/status/prashantluhar/testpay/ci.yml?branch=main&label=build&logo=github)](https://github.com/prashantluhar/testpay/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/prashantluhar/testpay?logo=codecov)](https://codecov.io/gh/prashantluhar/testpay)
[![Go Report Card](https://goreportcard.com/badge/github.com/prashantluhar/testpay)](https://goreportcard.com/report/github.com/prashantluhar/testpay)
[![Go Reference](https://pkg.go.dev/badge/github.com/prashantluhar/testpay.svg)](https://pkg.go.dev/github.com/prashantluhar/testpay)
[![Go Version](https://img.shields.io/github/go-mod/go-version/prashantluhar/testpay?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/prashantluhar/testpay/pkgs/container/testpay)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**Postman for Payments.** A mock payment gateway and failure-simulation tool that lets developers test every real-world payment edge case — locally and in CI — without touching production systems.

**Live demo** — [dashboard](https://testpay-web.onrender.com) · [API](https://testpay-przk.onrender.com) (Render free tier; first request after idle is slow, see [Free-tier quirks](#free-tier-quirks-to-expect)).

---

## What's in this repo

- **Go backend** — mock gateway + simulation engine + Control API + webhook dispatcher. Single binary. See [internal/README.md](internal/README.md).
- **Next.js dashboard** — 7-screen UI to create scenarios, inspect logs, replay requests. Embedded in the Go binary for local mode; deployable to Vercel for hosted mode. See [web/README.md](web/README.md).
- **Deploy manifests** — Docker, docker-compose, Kubernetes, Vercel config — under `deploy/`.
- **CI pipelines** — GitHub Actions for test, coverage gate, Docker image, release. Under `.github/workflows/`.

---

## Why TestPay

- Sandbox environments never replicate real failure modes
- Edge cases like duplicate webhooks, bank timeouts, and async state transitions are impossible to trigger on demand
- Payment bugs only surface in production — after customers are affected

TestPay gives you a mock gateway that behaves exactly like Stripe, Razorpay, or any other payment processor — including every way they can fail.

---

## Features

- **28 failure modes** across bank, PG, webhook, redirect/3DS, charge anomalies, and async state transitions
- **Named scenarios** — save sequences of failure modes as replayable test fixtures
- **Full request logging** — every request, header, response, and webhook delivery logged to Postgres
- **Webhook debugger** — inspect delivery attempts, retry history, and payloads
- **Zero code change** — point your Stripe SDK at `localhost:7700/stripe` and it just works
- **Gateway-agnostic engine** — Stripe, Razorpay, and a generic "agnostic" adapter today; more coming
- **Embedded dashboard** — `./testpay start` serves both the API (`:7700`) and the dashboard (`:7701`) from one binary

---

## Quick Start — Local

**Prerequisites:** Go 1.24+, Node 20+ (for building the dashboard), Postgres 16+ running somewhere

### Option A — you already have Postgres running

```bash
git clone https://github.com/prashantluhar/testpay.git
cd testpay

# 1. Build the dashboard static bundle (one-time)
cd web
pnpm install
pnpm build                   # emits web/out/
cd ..

# 2. Create the database (skip if it already exists)
createdb -h localhost -U postgres testpay

# 3. Start the server — migrations run automatically on boot
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/testpay?sslmode=disable"
go run ./cmd/testpay start --config deploy/config/testpay.local.yaml
```

### Option B — fresh Postgres via docker-compose

```bash
git clone https://github.com/prashantluhar/testpay.git
cd testpay

cd web && pnpm install && pnpm build && cd ..
docker compose -f deploy/docker/docker-compose.yml up -d postgres

export DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay?sslmode=disable"
go run ./cmd/testpay start --config deploy/config/testpay.local.yaml
```

Open:
- **Dashboard:** http://localhost:7701
- **Mock gateway API:** http://localhost:7700 (point your app here)
- **Control API:** http://localhost:7700/api/* (scenarios, logs, auth)

---

## Point Your App at the Mock

```bash
# Stripe — one env var, no code changes
STRIPE_BASE_URL=http://localhost:7700/stripe

# Razorpay
RAZORPAY_BASE_URL=http://localhost:7700/razorpay

# Any other gateway
YOUR_GATEWAY_BASE_URL=http://localhost:7700/v1
```

All requests are logged to Postgres and surfaced in the dashboard's **Logs** page.

---

## Build a Single Binary (production)

```bash
cd web && pnpm install && pnpm build && cd ..
go build -o bin/testpay ./cmd/testpay

# Run
./bin/testpay start --config deploy/config/testpay.local.yaml
```

The Go binary embeds `web/out/` via `go:embed`, so the single file contains the full dashboard.

---

## Testing

```bash
# All Go unit tests (no database required)
go test ./...

# Integration tests (requires Postgres)
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/testpay?sslmode=disable"
go test ./internal/store/postgres/... -v

# Coverage (40% floor enforced in CI, scoped to tested packages only)
make coverage-check

# Frontend tests
cd web && pnpm test
```

---

## CLI

`testpay` is a single Cobra binary. All commands either boot the server (`start`) or talk to a running local server over HTTP at `localhost:7700`.

### Available commands

| Command | What it does |
|---|---|
| `testpay start` | Boot the server + embedded dashboard. Defaults to API `:7700` and dashboard `:7701`. |
| `testpay start --no-dashboard` | API only — useful when you're running `pnpm dev` for the dashboard separately. |
| `testpay start --config <path>` | Load a specific YAML config (e.g. `deploy/config/testpay.local.yaml`). |
| `testpay scenario list` | `GET /api/scenarios` — dump the workspace's scenarios as JSON. |
| `testpay scenario run <id>` | `POST /api/scenarios/{id}/run` — execute a saved scenario and print the result. |
| `testpay logs` | `GET /api/logs` — recent request logs, JSON. |
| `testpay logs --follow` / `-f` | Tails new log rows by polling every second. Ctrl-C to stop. |

### Try it out

```bash
# 1. Start Postgres
docker compose -f deploy/docker/docker-compose.yml up -d postgres

# 2. Build the dashboard bundle (first time — needed for go:embed)
cd web && pnpm install && pnpm build && cd ..

# 3. Build the binary
go build -o bin/testpay ./cmd/testpay

# 4. Start the server in one terminal
export DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay?sslmode=disable"
./bin/testpay start --config deploy/config/testpay.local.yaml
```

In a second terminal:

```bash
# List scenarios (empty at first)
./bin/testpay scenario list

# Send a mock request so there's something to see in logs
curl -X POST http://localhost:7700/stripe/v1/charges \
  -H "Content-Type: application/json" \
  -d '{"amount":5000,"currency":"usd"}'

# Dump recent logs
./bin/testpay logs

# Tail mode — hit the mock a few more times in another terminal
./bin/testpay logs --follow
```

To test `scenario run`, create one first through the dashboard at http://localhost:7701/scenarios, note the ID, then `./bin/testpay scenario run <id>`.

### Limitations

- **Local-only.** The CLI hardcodes `http://localhost:7700` in `cli/scenario.go` and `cli/logs.go`. It does not talk to hosted deployments (Render, etc.). Add a `--base-url` flag + `TESTPAY_API_KEY` env if you need that.
- **Raw JSON output.** Responses are not formatted; pipe through `jq` for readability.

---

## What to Test Manually

With the server running (`http://localhost:7701`):

1. **Auth flow (hosted mode only)** — `/signup` → creates workspace + user, logs you in. `/login` → validates credentials.
2. **Overview page (`/`)** — stat cards populate after you hit the mock; live feed polls every 2s.
3. **Scenarios (`/scenarios`)** — create a scenario with multiple failure-mode steps, save, run, delete.
4. **Scenario Editor** — visual step builder with 28 outcomes grouped by category; JSON preview updates live.
5. **Logs (`/logs`)** — send some requests to `/stripe/v1/charges`, see them appear; click a row for full request/response/webhook detail.
6. **Log Detail drawer** — inspect headers, body, webhook payload across tabs; use the Replay button.
7. **Settings (`/settings`)** — mask/reveal/copy API key, copy endpoint URLs, toggle theme.

**Smoke test with curl:**
```bash
# Hit the mock
curl -X POST http://localhost:7700/stripe/v1/charges \
  -H "Content-Type: application/json" \
  -d '{"amount":5000,"currency":"usd"}'

# Watch the log flow into the dashboard at http://localhost:7701/logs
```

---

## Architecture

```
testpay/
├── cmd/testpay/          # CLI entrypoint
├── cli/                  # Cobra commands (start, scenario, logs)
├── internal/             # Go backend — see internal/README.md
│   ├── engine/           # PG-agnostic simulation engine
│   ├── adapters/         # Stripe, Razorpay, Agnostic adapters
│   ├── store/            # Postgres data layer
│   ├── webhook/          # Webhook dispatcher + retry
│   ├── api/              # HTTP server, middleware, handlers
│   └── observability/    # zerolog setup
├── web/                  # Next.js dashboard — see web/README.md
├── deploy/
│   ├── config/           # Per-env YAML config files
│   ├── docker/           # Dockerfile + compose
│   └── k8s/              # Kubernetes manifests
├── docs/superpowers/     # Design specs + implementation plans
└── .github/workflows/    # CI + release
```

The simulation engine is gateway-agnostic. Gateway-specific adapters translate requests/responses to the right wire format. Adding a new gateway is ~200 lines.

All HTTP requests pass through a middleware chain that logs full headers, bodies, and response times to Postgres — giving you complete observability into every simulated transaction. See [internal/README.md](internal/README.md) for deep detail.

---

## Project Status

| Component | Status |
|---|---|
| Backend core (engine, adapters, middleware, webhook, store) | ✅ Complete |
| Control API (scenarios, sessions, logs, webhooks) | ✅ Complete |
| Mock gateway endpoints (Stripe, Razorpay, Agnostic) | ✅ Complete |
| Auth (signup/login/logout/me, JWT cookie) | ✅ Complete |
| Observability (trace IDs, per-function logs, slow-query logging) | ✅ Complete |
| YAML config + per-env files | ✅ Complete |
| Dashboard (all 7 screens) | ✅ Complete |
| Docker + Kubernetes + CI + release automation | ✅ Complete |
| Embedded dashboard in Go binary | ✅ Complete |
| CLI (start, scenario list/run, logs --follow) | ✅ Complete |

---

## Sub-project READMEs

- [`internal/README.md`](internal/README.md) — backend architecture, package tour, auth, config, testing
- [`web/README.md`](web/README.md) — frontend stack, screens, run modes, deployment

---

## Deployment

**Dashboard:**
- **Local:** embedded in the Go binary — run `./testpay start` and open http://localhost:7701
- **Hosted:** deploy `web/` to Vercel or Render Static Site. Set `NEXT_PUBLIC_TESTPAY_MODE=hosted` and `NEXT_PUBLIC_API_BASE=https://your-api-host`.

**Backend:**
- **Docker:** `docker compose -f deploy/docker/docker-compose.yml up -d`
- **Kubernetes:** `kubectl apply -f deploy/k8s/`
- **Render + Neon (free tier):** see [Hosted Deploy](#hosted-deploy-render--neon-free-tier) below
- **Fly.io / any PaaS:** `go build` the binary + Postgres

---

## Hosted Deploy (Render + Neon, Free Tier)

End-to-end recipe for a public, $0 demo instance. Architecture:

```
testpay-web (Render static site) ──► testpay-api (Render web service, Docker)
                                              │
                                              ▼
                                      Neon Postgres 16
```

Full design + trade-offs: `docs/superpowers/specs/2026-04-17-render-neon-deploy-design.md`.

### Prerequisites

- GitHub account with this repo pushed
- Free account on [Neon](https://neon.tech)
- Free account on [Render](https://render.com)

### 1. Neon — create the database

1. Sign up at neon.tech, create a new project — pick Postgres 16, region closest to you.
2. Create a database named `neondb` (the default role `neondb_owner` is created automatically).
3. On the project dashboard → **Connection Details** → toggle **Pooled connection** ON, copy the connection string. It looks like:
   ```
   postgresql://neondb_owner:***@ep-xxxxx-pooler.<region>.aws.neon.tech/neondb?sslmode=require
   ```
   Keep this URL secret — never commit it.

### 2. Render — API web service

1. Render dashboard → **New → Web Service** → connect GitHub and pick the `testpay` repo.
2. Fill the form:

   | Field | Value |
   |---|---|
   | Name | `testpay-api` |
   | Language | **Docker** |
   | Branch | `main` |
   | Region | same region as Neon (e.g. **Singapore**) |
   | Root Directory | *(blank)* |
   | Dockerfile Path | `./deploy/docker/Dockerfile` |
   | Docker Build Context Directory | `.` |
   | Instance Type | **Free** |
   | Health Check Path | *(blank)* |
   | Auto-Deploy | On Commit |

3. Scroll to **Advanced → Environment Variables** and add:

   | Key | Value |
   |---|---|
   | `DATABASE_URL` | Neon pooled URL from step 1 |
   | `JWT_SECRET` | `openssl rand -hex 32` output |
   | `API_KEY` | `openssl rand -hex 32` output |
   | `CORS_ALLOWED_ORIGINS` | `https://testpay-web.onrender.com` |

   Do **not** set `PORT` — Render injects it automatically.

4. Click **Deploy Web Service**. The Docker build takes 3-5 minutes on cold cache.

### 3. Render — dashboard static site

1. Render dashboard → **New → Static Site** → same repo.
2. Fill the form:

   | Field | Value |
   |---|---|
   | Name | `testpay-web` |
   | Branch | `main` |
   | Root Directory | `web` |
   | Build Command | `pnpm install && pnpm build` |
   | Publish Directory | `out` |

3. **Environment Variables** (baked in at build time):

   | Key | Value |
   |---|---|
   | `NEXT_PUBLIC_API_BASE` | `https://testpay-api.onrender.com` |
   | `NEXT_PUBLIC_TESTPAY_MODE` | `hosted` |

4. Click **Create Static Site**.

### 4. Smoke test

```bash
API=https://testpay-api.onrender.com
WEB=https://testpay-web.onrender.com

# Visit the dashboard, sign up, log in
open "$WEB/signup"

# After signup, grab your API key from the Settings page, then:
curl -X POST "$API/stripe/v1/charges" \
  -H "Authorization: Bearer $YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"amount":5000,"currency":"usd"}'
```

The mock charge should appear on the dashboard's `/logs` page.

### Free-tier quirks to expect

- **Cold start — first request after 15 min idle is slow.** `testpay-api` sleeps on Render free tier; waking from sleep takes ~30-60 seconds (Docker image unpack, Go binary boot, Neon DB wake all stack up). If you haven't hit the service in a while, expect the first `curl` / page load to hang for up to a minute. Every subsequent request until the next idle window is fast. You can keep it warm with an external uptime pinger (e.g. Uptime Kuma, cron-job.org) at your own quota expense.
- **Neon auto-suspends after ~5 min idle.** Adds ~500 ms to the first DB query after a suspend. Negligible after that.
- **Neon storage cap: 0.5 GB.** Plenty for demo traffic. The `request_logs` table is the heaviest; if you stress-test, manually `TRUNCATE request_logs, webhook_logs;` via the Neon SQL editor to reclaim space.
- **One Render free web service.** If the binary crashes you'll get 502s until the next deploy. Push a fix and it auto-redeploys.

### Abuse protection (enforced server-side)

Because the demo runs on shared free-tier resources, the binary rate-limits itself so a noisy caller can't burn the monthly quota for everyone:

| Scope | Limit | What happens when exceeded |
|---|---|---|
| **Mock endpoints, per-IP** | 30 requests / minute (burst 10) | `HTTP 429`, `Retry-After: 60` header |
| **Mock endpoints, global** | 150 requests / minute combined | `HTTP 429` |
| **Signup / login, per-IP** | 10 requests / minute (burst 3) | `HTTP 429` |
| **Workspace API key required** | mock routes reject any call without a valid workspace `Authorization: Bearer <api_key>` | `HTTP 401` |
| **Session auth on `/api/*`** | dashboard routes require a valid JWT cookie | `HTTP 401 login required` |

All numbers live in `deploy/config/testpay.render.yaml` (`rate_limit` block) and are enforced by `internal/api/middleware/ratelimit.go` using an in-memory token-bucket keyed by client IP (with `X-Forwarded-For` resolution since Render terminates TLS at its proxy). Self-hosted? Bump the numbers in your own config — the middleware is limit-agnostic.

---

## Repository Setup (one-time)

1. **Codecov** — sign in at https://codecov.io with GitHub, add the repo, copy the upload token, add as GitHub secret `CODECOV_TOKEN`
2. **Go Report Card** — first visit https://goreportcard.com/report/github.com/prashantluhar/testpay to generate the report
3. **pkg.go.dev** — push a tagged release (`git tag v0.1.0 && git push --tags`), the badge becomes live within minutes
4. **GitHub Actions** — already wired via `.github/workflows/ci.yml`; no setup needed

---

## Branching & Releases

We follow [GitHub Flow](https://docs.github.com/en/get-started/quickstart/github-flow) with a protected `main` branch.

- All work happens on `feature/`, `fix/`, `chore/`, or `hotfix/` branches
- Open a PR — CI must pass (90% coverage gate, lint, build, all tests)
- Squash-merge to `main`
- Release: maintainer tags `vX.Y.Z` → release workflow builds and publishes the Docker image to GHCR

See [CONTRIBUTING.md](CONTRIBUTING.md) for full details.

---

## License

MIT License — see [LICENSE](LICENSE).

The hosted version of TestPay (testpay.dev) is a commercial product. The source code is MIT-licensed and free to use, modify, and self-host.

---

## Contributing

Issues and PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
