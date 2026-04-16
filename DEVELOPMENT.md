# Development Workflow

Two processes run separately so you can iterate on the frontend with hot reload while keeping the backend stable (or vice versa). The dashboard on `:7701` talks to the API on `:7700` cross-origin — CORS is already wired.

---

## Prerequisites

- **Go 1.24+** on PATH
- **Node 20+** and **pnpm**. If pnpm is missing: `npm i -g pnpm`
- **Postgres** running on `localhost:5432` with a database named `testpay` and a user who can connect

PowerShell env var (per shell session):
```powershell
$env:DATABASE_URL = "postgres://postgres:root@localhost:5432/testpay?sslmode=disable"
```

---

## One-time setup

```powershell
# From the repo root
cd web
pnpm install
cd ..
```

Optional: create the DB if you haven't.
```powershell
$env:PGPASSWORD = "root"
createdb -h localhost -U postgres testpay
```

---

## Run them separately

### Terminal 1 — Backend (Go API on :7700)

```powershell
cd C:\Users\ADMIN\Desktop\work\testpay
$env:DATABASE_URL = "postgres://postgres:root@localhost:5432/testpay?sslmode=disable"
go run ./cmd/testpay start --config deploy\config\testpay.local.yaml
```

Migrations run automatically on boot. The server stays up until you `Ctrl+C`.

Edit any Go file → `Ctrl+C` → re-run the command. No frontend rebuild needed.

> **Tip:** `go run` recompiles every time, which is slow on cold start. For faster dev loops, use a file watcher like [`air`](https://github.com/air-verse/air):
> ```powershell
> go install github.com/air-verse/air@latest
> air -c .air.toml   # create .air.toml as needed; default covers most cases
> ```

### Terminal 2 — Frontend dev server (Next.js on :7701)

```powershell
cd C:\Users\ADMIN\Desktop\work\testpay\web
pnpm dev
```

Open `http://localhost:7701`. Edit any `.tsx`/`.ts`/`.css` → the browser hot-reloads instantly. No backend restart.

**Note:** `pnpm dev` serves the app in dev mode with HMR. The `:7701` port is set in `package.json`.

### Terminal 3 — Optional: load-test script

```powershell
cd C:\Users\ADMIN\Desktop\work\testpay
.\scripts\load-test.ps1 -Users 3 -RequestsPerUser 2
```

---

## When to rebuild

| You edited | What to do |
|---|---|
| Go code under `internal/`, `cli/`, `cmd/` | `Ctrl+C` + re-run backend (Terminal 1) |
| Frontend under `web/app/`, `web/components/`, `web/lib/` | Nothing — `pnpm dev` hot-reloads automatically |
| `web/next.config.js`, `web/tailwind.config.ts`, `web/tsconfig.json` | `Ctrl+C` + re-run `pnpm dev` |
| New shadcn primitive (`pnpm dlx shadcn add ...`) | Nothing — dev server picks it up |
| Go DB migration (`internal/store/postgres/migrations/*.sql`) | `Ctrl+C` + re-run backend (applies on boot) |

---

## Single-binary mode (for shipping)

If you want to package everything into one executable (what `go build` produces):

```powershell
cd web
pnpm build            # emits web/out/
cd ..
go build -o bin\testpay.exe ./cmd/testpay
.\bin\testpay.exe start --config deploy\config\testpay.local.yaml
```

This serves the dashboard at `:7701` embedded via `//go:embed` from `web/embed.go`. Use this for local smoke tests of the production layout, not for day-to-day UI iteration.

---

## Logging

The backend emits structured logs in local mode (colored console format). Every request carries a `trace_id` field so you can follow one request end-to-end. Typical flow:

```
INF handler entry         handler=Signup     step=entry       trace_id=abc
INF request body parsed   step=body_parsed   body_bytes=47    trace_id=abc
INF workspace resolved    step=workspace_resolved             trace_id=abc workspace_id=...
INF engine step executed  step=engine_executed outcome=success trace_id=abc
INF response written      step=response_sent response_status=200 duration_ms=12 trace_id=abc
INF webhook scheduled     step=webhook_scheduled target_url=... trace_id=abc
INF mock request completed step=mock_exit    duration_ms=13   trace_id=abc
```

Error paths include `err=...` and the same `trace_id`. Grep by `trace_id` to follow a single request.

To filter: `... | grep trace_id=abc` or use `jq` in JSON-mode (set `logging.format: json` in the YAML config).

---

## Common tasks

```powershell
# Reset the DB
$env:PGPASSWORD = "root"
dropdb -h localhost -U postgres testpay; createdb -h localhost -U postgres testpay

# Connect a psql shell
psql -h localhost -U postgres -d testpay

# Tail just simulation logs
# (add a line grep filter in the server terminal)

# Run all Go tests
go test ./...

# Run all frontend tests
cd web; pnpm test; cd ..

# Kill a stuck backend process
taskkill /F /IM testpay.exe
```

---

## Troubleshooting

**"Black screen" on the dashboard**
- Check the F12 Console for errors
- Most common cause: backend not running or on the wrong port

**"unknown gateway" from curl**
- Use `:7700` for the API (not `:7701`)
- URL format: `/stripe/v1/...`, `/razorpay/v1/...`, or `/v1/...`

**Webhooks not firing**
- Set a webhook URL in **Settings → Webhook destinations** (one per gateway), OR include an `X-Webhook-URL` header in the request
- Check `request_logs` and `webhook_logs` in Postgres to confirm persistence

**"JWT_SECRET required when mode=hosted"**
- Either run in local mode (default), or `$env:JWT_SECRET = "some-random-string"` before starting
