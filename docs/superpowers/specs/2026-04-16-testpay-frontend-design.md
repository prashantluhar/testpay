# TestPay — Frontend Design Spec
**Date:** 2026-04-16
**Status:** Approved
**Companion doc:** [2026-04-16-testpay-design.md](./2026-04-16-testpay-design.md) (product + backend spec)

---

## 1. Scope

This spec covers the **Next.js dashboard** and the **backend auth additions** required to support it. It is the companion to the backend implementation already landed on `feat/backend-implementation` (24 tasks, Phases 1–11). The dashboard consumes the Control API defined in the product spec §6 and adds a minimal in-house auth layer.

**In scope**
- 7 dashboard screens (Overview, Scenarios list, Scenario Editor, Logs, Log Detail drawer, Settings, Login/Signup)
- Minimal in-house auth: email + password, JWT in httpOnly cookie, no email verification, no password reset
- Backend additions: new migration, `User` store methods, auth handlers, JWT middleware
- Embedded local mode (static export bundled in the Go binary) and hosted mode (Vercel)

**Out of scope (defer to v2)**
- Email verification, password reset, magic link
- MFA / TOTP
- Team members, invites, RBAC (workspace/user schema is ready — UI only)
- Billing screens / Stripe integration
- OAuth providers (GitHub, Google)

---

## 2. Tech Stack

| Layer | Choice | Why |
|---|---|---|
| Framework | Next.js 14 (App Router) | Already decided in product spec §8 |
| Language | TypeScript | |
| Rendering | Static export (`output: 'export'`) | Same build for local-embedded and hosted-Vercel |
| UI primitives | shadcn/ui | Already decided |
| Styling | Tailwind CSS | Already decided |
| Data fetching | SWR | Built-in polling, caching, revalidation |
| Forms | react-hook-form + zod | Standard combo; zod schemas double as type sources |
| Toasts | sonner (shadcn default) | |
| Icons | lucide-react (shadcn default) | |
| Tests | Vitest + @testing-library/react | |
| E2E (optional, v2) | Playwright | Not required for MVP |

No Next.js API routes, no SSR, no edge functions. Everything is client-side against the Go Control API.

---

## 3. Rendering Model

Next.js is configured with `output: 'export'` so the build produces a pure static bundle in `web/out/`. That bundle is the deliverable for both modes:

| Mode | How the bundle is served |
|---|---|
| **Local** | Go binary `go:embed`s `web/out/` and serves it at `:7701` via `http.FileServer`. The Go binary is self-contained — no Node runtime needed at runtime. |
| **Hosted** | Same `web/out/` deployed to Vercel as static assets. Vercel handles CDN + HTTPS. |

All API calls go to the same-origin Go backend. The base URL is resolved at runtime:

```ts
// web/lib/api.ts
const API_BASE =
  typeof window === 'undefined'
    ? '' // build-time, no calls made
    : (window.location.origin === 'http://localhost:7701'
        ? 'http://localhost:7700' // local dev — dashboard and API on different ports
        : ''); // hosted — same origin, relative paths
```

CORS config in hosted mode allows only the dashboard's own origin (already modelled in `testpay.prod.yaml`).

---

## 4. Auth — Shape

**Local mode**: no auth. Dashboard assumes the single local workspace (`LocalWorkspaceID`). The auth routes (`/login`, `/signup`) are not reachable — middleware short-circuits to the dashboard.

**Hosted mode**: email + password.
- Signup: creates a new `User` row AND a `Workspace` row (one user = one workspace in MVP; team v2 preserves the schema).
- Login: password check → JWT issued → set as httpOnly, Secure, SameSite=Lax cookie `testpay_session`, 30-day expiry.
- Logout: clears the cookie (server sends `Set-Cookie` with past Expires).
- Session refresh: JWT carries `exp`; frontend silently re-auths on 401 by calling `/api/auth/me`, on failure redirects to `/login`.

The JWT payload is minimal:
```json
{
  "sub": "<user_id>",
  "workspace_id": "<workspace_id>",
  "exp": <unix>,
  "iat": <unix>
}
```

Signed with HS256 using `cfg.Auth.JWTSecret` (env var `JWT_SECRET`, already in YAML config).

---

## 5. Backend Additions (Required)

The current backend has only API-key auth. The dashboard requires a real user/session model. The following backend changes must land **before** the dashboard implementation plan starts.

### 5.1 Schema migration (`000002_auth.up.sql`)

```sql
-- Add password + future session tracking
ALTER TABLE users
    ADD COLUMN password_hash TEXT NOT NULL DEFAULT '',
    ADD COLUMN last_login_at TIMESTAMPTZ;

-- Enforce that password is set for hosted-mode users
-- (local seed user has empty hash; it's never used for login)
CREATE INDEX idx_users_email ON users(email);
```

Corresponding `000002_auth.down.sql`:
```sql
DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS password_hash;
```

### 5.2 Store interface extensions (`internal/store/store.go`)

```go
// Users
CreateUser(ctx context.Context, u *User, passwordHash string) error
GetUserByEmail(ctx context.Context, email string) (*User, string, error) // returns user + hash
GetUserByID(ctx context.Context, id string) (*User, error)
UpdateUserLastLogin(ctx context.Context, id string, at time.Time) error
```

The `User` model already exists in `internal/store/models.go` — only the store methods are new. Add `PasswordHash string` is NOT added to the `User` struct (we keep the hash out of the struct so it's never accidentally marshalled into JSON). The hash is returned as a separate return value from `GetUserByEmail`.

### 5.3 Auth handlers (`internal/api/handlers/auth.go`)

```
POST /api/auth/signup   { email, password }  → 201 { user, workspace } + Set-Cookie
POST /api/auth/login    { email, password }  → 200 { user, workspace } + Set-Cookie
POST /api/auth/logout                        → 204 + Set-Cookie (expired)
GET  /api/auth/me                            → 200 { user, workspace }  (validates cookie)
```

Password rules (minimal):
- Signup: 8+ chars, no other validation
- Hash: bcrypt cost 12
- On signup: create Workspace with slug = email's local-part + random suffix (`alice-f3a`), random 32-byte hex as `api_key`
- Failed login: 401 with a generic message (no "user not found" leak)
- Rate limit: reuse the existing `rate_limit.requests_per_minute` config (already in YAML); no extra limit per email for MVP

### 5.4 Session middleware (`internal/api/middleware/session.go`)

New middleware that validates the `testpay_session` cookie, parses the JWT, and puts `workspace_id` and `user_id` into context. Complements (does not replace) the existing `Auth` middleware:

```go
// Chain order in server.go becomes:
r.Use(middleware.RequestID)
r.Use(middleware.Logger(cfg.Environment, "testpay"))
r.Use(middleware.GatewayResolver)
r.Use(middleware.Session(cfg.Auth.JWTSecret))   // new — best-effort parse
r.Use(middleware.Auth(cfg.Server.Mode, cfg.Auth.APIKey)) // keeps API-key auth for mock endpoints
```

`middleware.Session` is best-effort — it reads and validates the cookie if present but does not reject. Individual `/api/*` routes that require auth call a helper `middleware.RequireSession(w, r)` at the top. Routes under `/:gateway/v1/*` ignore the session (they use API-key auth).

In **local mode**, `Session` middleware short-circuits: it injects the `LocalWorkspaceID` into context without looking at cookies.

### 5.5 Frontend-visible helper endpoint

`GET /api/auth/me` returns the current user + workspace for session bootstrap on every page load. Returns 401 if no valid session.

### 5.6 Config additions

Already present in Task 19's config. No changes required — `auth.jwt_secret_env` already references `JWT_SECRET`.

---

## 6. Screens

### 6.1 Routes and access rules

| Route | Access | Notes |
|---|---|---|
| `/login` | Hosted only, unauthenticated | Redirect to `/` if already logged in |
| `/signup` | Hosted only, unauthenticated | Same |
| `/` | Authenticated (hosted) / always (local) | Overview |
| `/scenarios` | Authenticated | |
| `/scenarios/new` | Authenticated | Blank editor |
| `/scenarios/[id]` | Authenticated | Edit existing |
| `/logs` | Authenticated | Log Detail opens as drawer; no separate route |
| `/settings` | Authenticated | |

A root layout component checks `/api/auth/me` on mount. If 401 and mode is hosted: redirect to `/login`. If mode is local: skip the check and assume authenticated.

### 6.2 Shell — sidebar + topbar

**Sidebar (left, 240px):**
- Logo + workspace slug
- Nav items: Overview · Scenarios · Logs · Settings
- Footer: user email + sign out (hosted) / "local mode" badge (local)

**Topbar (thin, right of sidebar):**
- Breadcrumb (e.g. `Scenarios / Retry storm`)
- Environment badge (`local` / `hosted`)
- Live endpoint URL with copy button (`http://localhost:7700/stripe` or `https://api.testpay.dev/ws_acme`)

### 6.3 Overview (`/`)

Polished stat cards + dense live feed.

- **Stat cards (3):** Requests today · Active scenarios · Success rate (last hour)
- **Live request feed** (polling `/api/logs?limit=50&since=<last_id>` every 2s via SWR `refreshInterval: 2000`):
  - Monospace table rows: status chip · method · path · gateway · duration · time-ago
  - New rows slide in at top; oldest falls off at 50
  - Click row → opens Log Detail drawer
- **Workspace endpoint URL card** with copy button

### 6.4 Scenarios (`/scenarios`)

- Table: name · gateway · steps · default? · last run · actions (run / edit / delete)
- `+ New scenario` button → `/scenarios/new`
- `Run` action opens a confirmation modal, then calls `POST /api/scenarios/:id/run` and shows a toast
- `Set as default` inline action calls `PUT /api/scenarios/:id` with `is_default: true`

### 6.5 Scenario Editor (`/scenarios/new`, `/scenarios/[id]`)

Two-pane layout:

**Left pane — form:**
- Name, description, gateway (radio: stripe / razorpay / agnostic)
- Webhook delay (ms)
- Default checkbox
- **Steps list** (the centerpiece):
  - Each step row: event (dropdown: charge / refund / capture) + outcome (searchable select with all ~30 failure modes from `engine.FailureMode`) + optional error code
  - Add step · drag to reorder · remove
  - Step previews show the resulting HTTP status via a small helper (hardcoded map matching `modeToResult` in engine.go)
- Save / Cancel buttons

**Right pane — JSON preview:**
- Read-only syntax-highlighted JSON of the current scenario
- "Copy" button — for pasting into CLI scripts
- "Copy curl" button — generates `curl -X POST http://localhost:7700/api/scenarios/:id/run`

Form state via react-hook-form; validation via zod schema that mirrors `store.Scenario`. Outcome list is generated at build time from a TS const (`web/lib/failure-modes.ts`) kept in sync with `internal/engine/modes.go` — document this in a README note.

### 6.6 Logs (`/logs`)

Dense filterable table.

- **Filters (sticky top):**
  - Gateway: all / stripe / razorpay / agnostic
  - Status: all / 2xx / 4xx / 5xx
  - Time: last 15m / 1h / 24h / custom
  - Search (path contains)
- **Table columns (monospace):** time · status · method · path · gateway · duration · request ID (truncated)
- Row click → Log Detail drawer (slide-over from right, 640px)
- Pagination: offset-based, `Load more` button (simple — no virtualized list for MVP)

### 6.7 Log Detail drawer

Opened from a row on `/logs` or `/` live feed. No dedicated URL.

- **Tabs:** Request · Response · Webhook · Timeline
  - Request: headers (collapsible), body (pretty-printed JSON)
  - Response: headers, body, status, duration
  - Webhook: target URL, delivery status, per-attempt log (matches `AttemptLog` model)
  - Timeline: horizontal timeline — received → executed → response → webhook attempts
- **Actions:**
  - Replay button → `POST /api/logs/:id/replay`, shows toast on result
  - Copy trace ID
  - Copy as curl (reconstructs the original request)

### 6.8 Settings (`/settings`)

- **Workspace card:** slug, API key (masked, reveal + copy)
- **Endpoint card:** base URL by gateway (`http://localhost:7700/stripe`, etc.)
- **Appearance card:** dark / light / system toggle
- **Account card** (hosted only): email, sign out
- No dangerous actions for MVP (no delete account, no rotate API key — defer to v2)

### 6.9 Login (`/login`) and Signup (`/signup`)

Hosted mode only. Both are single-column centered cards:

- Email + password fields (zod validation)
- `Sign in` / `Create account` button
- Cross-link to the other route
- Generic error messages (no enumeration)
- On success: cookie is set server-side; frontend calls `/api/auth/me` then redirects to `/`

---

## 7. Component Architecture

**Shell (`app/(dashboard)/layout.tsx`):**
- Fetches `GET /api/auth/me` via SWR on mount
- Renders `<Sidebar />` + `<Topbar />` + `{children}`
- Provides context: `{ user, workspace, mode }`

**Reusable components (`web/components/`):**

| Component | Purpose |
|---|---|
| `StatCard` | Overview metrics |
| `LogRow` | One dense monospace row for live feed / logs table |
| `LogDetailDrawer` | Slide-over for full request detail |
| `StatusChip` | Color-coded HTTP status (green 2xx, amber 4xx, red 5xx) |
| `GatewayBadge` | `stripe` / `razorpay` / `agnostic` pill |
| `ScenarioStepEditor` | Add/reorder/remove steps; inline outcome picker |
| `OutcomePicker` | Searchable combobox of all failure modes |
| `JsonViewer` | Read-only pretty-printed JSON with copy |
| `CopyButton` | Used in many places |
| `ApiKeyReveal` | Masked + reveal toggle |
| `ConfirmModal` | Used for destructive actions (delete scenario, run scenario) |

**shadcn primitives used:** Button, Input, Label, Select, Checkbox, Dialog, Drawer (Sheet), Table, Badge, Tabs, DropdownMenu, Tooltip, Toast (via sonner), Form.

---

## 8. Data Fetching Patterns

All API access through `web/lib/api.ts` which exposes:

```ts
// Typed fetch wrapper — throws on non-2xx, returns parsed JSON
export async function api<T>(path: string, init?: RequestInit): Promise<T>;

// SWR-ready fetcher
export const swrFetcher = <T>(path: string) => api<T>(path);
```

Hooks in `web/lib/hooks.ts`:

```ts
useMe()            // SWR /api/auth/me
useScenarios()     // SWR /api/scenarios
useScenario(id)    // SWR /api/scenarios/:id
useLogs(filters)   // SWR /api/logs?...  — refreshInterval=2000 on live view, none on /logs
useLog(id)         // SWR /api/logs/:id
useWorkspace()     // SWR /api/workspace
```

Mutations are plain `api()` calls inside event handlers, followed by `mutate(key)` to revalidate SWR.

On any 401 response from the API layer:
- Clear SWR cache for `/api/auth/me`
- Redirect to `/login` (hosted mode)
- In local mode a 401 should be impossible; log to console and toast

---

## 9. Routing / Auth Guard

Single guard in `app/(dashboard)/layout.tsx`:

```tsx
const { data: me, error, isLoading } = useMe();

useEffect(() => {
  if (error?.status === 401 && mode === 'hosted') {
    router.push('/login');
  }
}, [error, mode]);

if (isLoading) return <LoaderShell />;
if (error && mode === 'hosted') return null; // redirecting
return <Shell user={me.user} workspace={me.workspace}>{children}</Shell>;
```

`mode` comes from a build-time env var `NEXT_PUBLIC_TESTPAY_MODE`. Local-mode builds set this to `local`; Vercel deploys set it to `hosted`.

---

## 10. Visual Design Principles

**Polished shell:**
- Rounded corners (6–8px), 1px borders with low-contrast color
- Whitespace-generous padding on cards (16–20px)
- Sans-serif: Inter (via `next/font/google`)
- Dark mode default; palette anchored on shadcn's `slate` scale with accent `emerald`

**Dense data areas:**
- Monospace: JetBrains Mono (via `next/font/google`)
- Table rows ≤ 32px tall
- Color-coded statuses: green (2xx), amber (4xx), red (5xx), gray (unknown)
- No zebra striping — alternate by subtle border
- JSON viewers indent 2-space, never wrap lines by default (horizontal scroll)

Theme tokens defined in `tailwind.config.ts` so both modes share the same semantic names (`--bg-surface`, `--text-muted`, etc.).

---

## 11. Folder Structure

```
web/
├── app/
│   ├── (auth)/
│   │   ├── layout.tsx                 # minimal centered shell, no sidebar
│   │   ├── login/page.tsx
│   │   └── signup/page.tsx
│   ├── (dashboard)/
│   │   ├── layout.tsx                 # sidebar + topbar + auth guard
│   │   ├── page.tsx                   # Overview
│   │   ├── scenarios/
│   │   │   ├── page.tsx               # List
│   │   │   ├── new/page.tsx
│   │   │   └── [id]/page.tsx
│   │   ├── logs/page.tsx
│   │   └── settings/page.tsx
│   ├── layout.tsx                     # root html/body, theme provider, toast root
│   └── globals.css
├── components/
│   ├── ui/                            # shadcn generated (do not edit directly)
│   ├── shell/
│   │   ├── sidebar.tsx
│   │   └── topbar.tsx
│   ├── overview/
│   │   ├── stat-card.tsx
│   │   └── live-feed.tsx
│   ├── scenarios/
│   │   ├── scenario-form.tsx
│   │   ├── scenario-step-editor.tsx
│   │   ├── outcome-picker.tsx
│   │   └── json-preview.tsx
│   ├── logs/
│   │   ├── logs-table.tsx
│   │   ├── log-filters.tsx
│   │   └── log-detail-drawer.tsx
│   └── common/
│       ├── status-chip.tsx
│       ├── gateway-badge.tsx
│       ├── copy-button.tsx
│       ├── json-viewer.tsx
│       └── confirm-modal.tsx
├── lib/
│   ├── api.ts                         # fetch wrapper
│   ├── hooks.ts                       # SWR hooks
│   ├── auth.ts                        # session helpers
│   ├── failure-modes.ts               # mirror of engine/modes.go
│   ├── types.ts                       # mirrors store/models.go
│   ├── schemas.ts                     # zod schemas
│   └── utils.ts                       # cn(), formatters
├── tests/
│   └── components/                    # Vitest + RTL unit tests
├── public/
│   └── favicon.ico
├── next.config.js                     # output: 'export'
├── tailwind.config.ts
├── tsconfig.json
├── package.json
└── pnpm-lock.yaml
```

---

## 12. Error Handling

- **Network errors:** `api()` wrapper throws `ApiError { status, message, body }`. SWR puts it in `error`. Pages render an inline `<ErrorState />` component with retry.
- **Form validation errors:** react-hook-form surfaces field-level errors under inputs. No global banner.
- **Mutation errors:** sonner toast with error message.
- **Auth errors (401):** redirect to `/login` (hosted) or toast (local).
- **Unexpected runtime errors:** Next.js `app/error.tsx` with a reload button.

No Sentry/error tracking for MVP (can wire later via the `integrations.sentry` config block).

---

## 13. Testing Approach

- **Unit tests (Vitest + RTL):** Each component in `components/common/` and `components/scenarios/` has tests for render + interaction. Target 70% frontend coverage (the backend 90% gate only applies to Go — separate frontend gate configured later if needed).
- **API client tests:** Mock `fetch` and verify `api()` wrapper handles 2xx, 4xx, 5xx, network errors, cookie pass-through.
- **Form validation tests:** zod schemas tested directly.
- **No E2E for MVP.** Playwright can be added in v2 when auth flows stabilise.

---

## 14. Local/Hosted Differences

| Concern | Local | Hosted |
|---|---|---|
| Auth screens | Hidden | Required |
| Sidebar footer | "local mode" badge | user email + sign out |
| API base URL | `http://localhost:7700` (separate port) | same origin |
| Endpoint URL shown | `http://localhost:7700/:gateway` | `https://api.testpay.dev/ws_<slug>/:gateway` |
| Auth guard | Skipped | Active |
| Signup flow | N/A | Creates user + workspace |
| `NEXT_PUBLIC_TESTPAY_MODE` | `local` | `hosted` |

Both modes ship from the same build — only the env var differs.

---

## 15. Open Decisions (to revisit during implementation)

- **CSRF:** JWT in httpOnly SameSite=Lax cookie provides baseline protection. Adding a double-submit CSRF token is easy if threat model changes — not in MVP.
- **CSP headers:** set via Next.js `headers()` config only when hosted (Vercel) — local mode is trusted.
- **Bundle size budget:** target <250KB gzipped JS on initial load. shadcn + SWR + Next.js alone is ~180KB; leaves room.
- **Accessibility:** shadcn primitives are accessible by default. Explicit aria-labels on icon-only buttons. No screen-reader audit in MVP.

---

## 16. Backend Work Summary (for planning)

Before the dashboard plan starts, the backend needs these items as their own task batch:

1. Migration `000002_auth.up/down.sql` — add `password_hash`, `last_login_at`, `idx_users_email`
2. Store methods: `CreateUser`, `GetUserByEmail`, `GetUserByID`, `UpdateUserLastLogin`
3. Handlers: `/api/auth/signup`, `/api/auth/login`, `/api/auth/logout`, `/api/auth/me`
4. Middleware: `Session(jwtSecret)` with local-mode short-circuit + `RequireSession(w, r)` helper
5. Wire into `server.go` middleware chain
6. Mount auth routes under `/api/auth/*`
7. Tests: store unit tests, handler tests, JWT round-trip test
8. Update CLI `start` command to check `JWT_SECRET` env var is set when `mode=hosted`

Password hashing: `golang.org/x/crypto/bcrypt` (new dep). JWT: `github.com/golang-jwt/jwt/v5` (new dep).

These 8 items should be the first phase of the implementation plan, before any frontend task.

---

## 17. Frontend Work Summary (for planning)

The dashboard implementation will roughly map to:

1. Scaffold `web/` — Next.js + TS + Tailwind + shadcn init + eslint/prettier
2. API client + SWR hooks + types + zod schemas
3. Shell (sidebar + topbar + auth guard + theme)
4. Auth screens (login, signup)
5. Settings screen
6. Overview screen (stat cards + live feed)
7. Scenarios list
8. Scenario editor (the big one — step builder, json preview, outcome picker)
9. Logs screen
10. Log Detail drawer
11. Static export build + embedding into Go binary
12. CI: frontend lint + test + build
13. Vercel deploy config

Exact task decomposition belongs in the plan doc.
