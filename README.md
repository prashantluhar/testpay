# TestPay

[![Build](https://img.shields.io/github/actions/workflow/status/prashantluhar/testpay/ci.yml?branch=main&label=build&logo=github)](https://github.com/prashantluhar/testpay/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/prashantluhar/testpay?logo=codecov)](https://codecov.io/gh/prashantluhar/testpay)
[![Go Report Card](https://goreportcard.com/badge/github.com/prashantluhar/testpay)](https://goreportcard.com/report/github.com/prashantluhar/testpay)
[![Go Reference](https://pkg.go.dev/badge/github.com/prashantluhar/testpay.svg)](https://pkg.go.dev/github.com/prashantluhar/testpay)
[![Go Version](https://img.shields.io/github/go-mod/go-version/prashantluhar/testpay?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/prashantluhar/testpay?logo=github)](https://github.com/prashantluhar/testpay/releases)
[![Docker](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/prashantluhar/testpay/pkgs/container/testpay)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**Postman for Payments.** A mock payment gateway and failure simulation tool that lets developers test every real-world payment edge case — locally and in CI — without touching production systems.

---

## Why TestPay

- Sandbox environments never replicate real failure modes
- Edge cases like duplicate webhooks, bank timeouts, and async state transitions are impossible to trigger on demand
- Payment bugs only surface in production — after customers are affected

TestPay gives you a mock gateway that behaves exactly like Stripe, Razorpay, or any other payment processor — including every way they can fail.

---

## Features

- **50+ failure modes** — bank declines, PG server errors, webhook anomalies, redirect/3DS flows, async state transitions
- **Named scenarios** — save sequences of failure modes as replayable test fixtures
- **CI-ready** — trigger scenarios from GitHub Actions, pytest, Jest, or any test suite
- **Full request logging** — every request, header, response, and webhook delivery logged to Postgres
- **Webhook debugger** — inspect delivery attempts, retry history, and payloads
- **Zero code change** — point your Stripe SDK at `localhost:7700/stripe` and it just works
- **Gateway-agnostic engine** — Stripe and Razorpay adapters today; more coming

---

## Quick Start (Local)

**Prerequisites:** Docker, Go 1.22+

```bash
# 1. Clone
git clone https://github.com/prashantluhar/testpay.git
cd testpay

# 2. Start Postgres
docker compose up -d

# 3. Run migrations + start the server
export DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay?sslmode=disable"
go run ./cmd/testpay start

# Mock server:  http://localhost:7700
# Dashboard:    http://localhost:7701
```

**Point your app at the mock:**
```bash
# Stripe users — one env var change, zero code changes
STRIPE_BASE_URL=http://localhost:7700/stripe

# Razorpay users
RAZORPAY_BASE_URL=http://localhost:7700/razorpay
```

---

## Running Scenarios

```bash
# List available scenarios
testpay scenario list

# Run a built-in scenario
testpay scenario run scn_retry_storm

# Run in CI with assertion
testpay scenario run scn_pending_then_failed --assert status=failed

# Tail live logs
testpay logs --follow
```

---

## CI Integration (GitHub Actions)

```yaml
jobs:
  payment-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: testpay
          POSTGRES_USER: testpay
          POSTGRES_PASSWORD: testpay
        ports: ["5432:5432"]

    steps:
      - uses: actions/checkout@v4

      - name: Start TestPay
        run: |
          export DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay?sslmode=disable"
          go run ./cmd/testpay start &
          sleep 2

      - name: Run payment scenarios
        run: |
          testpay scenario run scn_retry_storm --assert webhook_received
          testpay scenario run scn_webhook_missing --assert no_webhook
          testpay scenario run scn_double_charge --assert idempotency_key_used
```

---

## Testing

```bash
# All unit tests (no database required)
go test ./...

# Integration tests (requires Postgres)
export TEST_DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay_test?sslmode=disable"
go test ./internal/store/postgres/... -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Architecture

```
testpay/
├── cmd/testpay/         # CLI entrypoint
├── internal/
│   ├── engine/          # PG-agnostic simulation engine
│   ├── adapters/        # Stripe, Razorpay, Agnostic adapters
│   ├── store/           # Postgres data layer
│   ├── webhook/         # Webhook dispatcher + retry
│   └── api/             # HTTP server, middleware, handlers
├── web/                 # Next.js dashboard (embedded in binary)
└── docker-compose.yml
```

The simulation engine is gateway-agnostic. Gateway-specific adapters translate requests/responses to the right wire format. Adding a new gateway is ~200 lines.

All requests pass through a middleware chain that logs full headers, bodies, and response times to Postgres — giving you complete observability into every simulated transaction.

---

## Failure Modes

| Category | Examples |
|---|---|
| Bank failures | `bank_decline_hard`, `bank_decline_soft`, `bank_server_down`, `bank_timeout` |
| PG failures | `pg_server_error`, `pg_timeout`, `pg_rate_limited`, `pg_maintenance` |
| Webhook anomalies | `webhook_missing`, `webhook_delayed`, `webhook_duplicate`, `webhook_malformed` |
| Redirect/3DS | `redirect_success`, `redirect_abandoned`, `redirect_timeout`, `redirect_failed` |
| Charge anomalies | `double_charge`, `amount_mismatch`, `partial_success` |
| Async transitions | `pending_then_failed`, `pending_then_success`, `failed_then_success` |

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
