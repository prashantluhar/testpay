# TestPay Dashboard

Next.js 14 App Router dashboard for TestPay. 7 screens: Overview, Scenarios, Scenario Editor, Logs, Log Detail drawer, Settings, Login/Signup.

Runs in two modes from the **same static bundle**:

- **Local** — embedded in the Go binary via `go:embed web/out/`, served on `:7701`. No auth. Connects to the Go API on `:7700`.
- **Hosted** — deployed to Vercel as static assets. Auth required. Same-origin API on `api.testpay.dev`.

See the top-level [README.md](../README.md) for what TestPay is.

---

## Tech stack

| Layer | Choice |
|---|---|
| Framework | Next.js 14 App Router (static export, `output: 'export'`) |
| Language | TypeScript |
| Styling | Tailwind CSS |
| UI primitives | shadcn/ui + Radix |
| Data fetching | SWR (polling + caching) |
| Forms | react-hook-form + zod |
| Toasts | sonner |
| Icons | lucide-react |
| Tests | Vitest + @testing-library/react |
| Package manager | pnpm |

No Next.js API routes, no SSR, no edge functions. All data access goes through the Go Control API.

---

## Prerequisites

- **Node 20+** and **pnpm** (install: `npm i -g pnpm`)
- The Go backend running (or at least reachable) on `http://localhost:7700` for the dev server to talk to

---

## Run modes

### 1. Develop the UI (hot reload)

The Go backend must be running. In a separate terminal:
```bash
cd web
pnpm install           # one-time
pnpm dev               # dev server on http://localhost:7701
```

The dev server hot-reloads on file changes. API calls go to `http://localhost:7700` (hardcoded when origin is `localhost:7701`).

### 2. Production build (static export)

```bash
cd web
pnpm install
pnpm build             # emits web/out/
```

`web/out/` is what gets embedded in the Go binary and what ships to Vercel.

### 3. Preview the production build locally

```bash
cd web
pnpm start             # serves web/out/ on http://localhost:7701
```

### 4. Via the Go binary (single-process local mode)

```bash
# From the repo root, after pnpm build
go run ./cmd/testpay start --config deploy/config/testpay.local.yaml
```

The Go binary starts the mock API on `:7700` and serves the embedded dashboard on `:7701`.

---

## Screens

| Route | Purpose |
|---|---|
| `/login` | Hosted mode only — sign in with email + password |
| `/signup` | Hosted mode only — create workspace + user |
| `/` | Overview — stat cards + live request feed (polls every 2s) |
| `/scenarios` | List, create, run, delete scenarios |
| `/scenarios/new` | Visual step builder with 28 failure-mode outcomes + JSON preview |
| `/scenarios/[id]` | Edit an existing scenario |
| `/logs` | Filterable request log table; click a row → slide-over drawer |
| `/settings` | Workspace slug, API key (masked/reveal/copy), endpoint URLs, theme toggle |

---

## What to test manually

1. **Sign up / log in (hosted mode)** — create a new account; confirm cookie is set and redirect to `/` works.
2. **Make a mock request** to `http://localhost:7700/stripe/v1/charges` (curl or your app). Watch the live feed on `/` update within 2s.
3. **Create a scenario** on `/scenarios/new` — add 3 steps with different failure modes, save, see it in the list.
4. **Run a scenario** — click ▶ on a scenario row; verify the toast + a run row appears.
5. **Inspect a log** — open `/logs`, click a row; check Request/Response/Webhook tabs in the drawer.
6. **Replay a log** — click Replay in the drawer; verify the toast.
7. **Copy endpoint URLs + API key** on `/settings` — verify clipboard.
8. **Theme toggle** — switch between light/dark/system on `/settings`; verify it persists in localStorage.

---

## Folder structure

```
web/
├── app/
│   ├── (auth)/                  # login + signup (no sidebar shell)
│   │   ├── layout.tsx
│   │   ├── login/page.tsx
│   │   └── signup/page.tsx
│   ├── (dashboard)/             # auth-guarded, sidebar shell
│   │   ├── layout.tsx           # calls /api/auth/me; redirects to /login on 401
│   │   ├── page.tsx             # Overview
│   │   ├── scenarios/
│   │   │   ├── page.tsx
│   │   │   ├── new/page.tsx
│   │   │   └── [id]/
│   │   │       ├── page.tsx     # static-export wrapper
│   │   │       └── client.tsx
│   │   ├── logs/page.tsx
│   │   └── settings/page.tsx
│   ├── layout.tsx               # root: fonts + theme + Toaster
│   └── globals.css              # Tailwind + shadcn CSS vars
├── components/
│   ├── ui/                      # shadcn primitives (generated)
│   ├── shell/                   # Sidebar, Topbar
│   ├── common/                  # StatusChip, GatewayBadge, CopyButton, JsonViewer,
│   │                            #  ApiKeyReveal, ConfirmModal, ErrorState, ThemeProvider
│   ├── overview/                # StatCard, LiveFeed
│   ├── scenarios/               # OutcomePicker, ScenarioStepEditor, ScenarioForm
│   └── logs/                    # LogFilters, LogsTable, LogDetailDrawer
├── lib/
│   ├── api.ts                   # typed fetch wrapper + ApiError + swrFetcher
│   ├── hooks.ts                 # SWR hooks (useMe, useScenarios, useLogs, …)
│   ├── types.ts                 # TypeScript mirrors of Go store models
│   ├── failure-modes.ts         # 28-outcome list mirroring internal/engine/modes.go
│   ├── schemas.ts               # zod schemas for forms (login, signup, scenario)
│   └── utils.ts                 # cn() helper for className merging
├── tests/
│   ├── setup.ts
│   └── api.test.ts
├── embed.go                     # //go:embed all:out  — consumed by cli/start.go
├── next.config.js               # output: 'export'
├── tailwind.config.ts
├── tsconfig.json
├── vitest.config.ts
├── components.json              # shadcn config
├── package.json
└── pnpm-lock.yaml
```

---

## Testing

```bash
pnpm test              # one-shot run
pnpm test:watch        # watch mode
```

Current coverage:
- `lib/api.ts` — fetch wrapper happy-path, non-2xx → ApiError, credentials pass-through

The UI relies on TypeScript + shadcn's accessibility defaults for most quality guarantees; E2E tests (Playwright) are deferred to v2.

---

## Data flow

```
Dashboard (React + SWR)
    │
    ▼   fetch(/api/*, { credentials: 'include' })
    │
Go Control API (:7700)
    │
    ▼   pgxpool
    │
Postgres
```

All API calls are same-origin in hosted mode (Vercel → `api.testpay.dev` via proxy). In local dev the dashboard is on `:7701` and the API is on `:7700` — `lib/api.ts` detects this and redirects requests accordingly.

---

## Auth flow (hosted mode)

1. User visits any dashboard route
2. `app/(dashboard)/layout.tsx` mounts, calls `GET /api/auth/me` via SWR
3. 401 → `router.push('/login')`
4. Login form posts to `/api/auth/login` with `credentials: 'include'`
5. Backend sets `testpay_session` httpOnly cookie + returns `{user, workspace}`
6. Dashboard re-mounts, `useMe()` succeeds, user sees the shell

In local mode: `NEXT_PUBLIC_TESTPAY_MODE=local` → the guard is skipped entirely.

---

## Embedding in the Go binary

`web/embed.go`:
```go
package web
import "embed"
//go:embed all:out
var Assets embed.FS
```

`cli/start.go` serves `Assets` on `:7701` using `http.FileServer(http.FS(fs.Sub(Assets, "out")))`. **`web/out/` must exist at Go build time** — CI runs `pnpm build` before `go build`.

---

## Deployment (hosted mode)

Deploy `web/` to Vercel:

- Root directory: `web`
- Build command: `pnpm build`
- Output directory: `out`
- Install command: `pnpm install --frozen-lockfile`
- Env var: `NEXT_PUBLIC_TESTPAY_MODE=hosted`

CORS on the Go backend must allow the Vercel origin — configured via `cors.allowed_origins` in `testpay.prod.yaml`.

---

## Keeping in sync with the backend

Two files intentionally mirror backend contracts and must be updated when the backend changes:

- **`lib/types.ts`** — mirrors `internal/store/models.go` (struct field names preserve PascalCase from Go JSON defaults)
- **`lib/failure-modes.ts`** — mirrors `internal/engine/modes.go` (the 28 `FailureMode` constants)

A future CI step could diff these automatically; for now treat them as manually-sync'd sources.
