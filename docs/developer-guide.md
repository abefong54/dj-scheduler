# EventLineup — Developer Onboarding Guide

This guide gets a new engineer productive on **EventLineup (dj-scheduler)**: a full-stack app for planning DJ lineups and running event nights. It covers the architecture, how to run everything locally, the API surface, the auth/multi-tenancy model, the database, and the build/test tooling.

> For a feature-level, non-technical tour of the product, see [users-guide.md](./users-guide.md).

---

## 1. The big picture

EventLineup is two deployables backed by one PostgreSQL database:

- **Backend** — a Go HTTP API (Gin + pgx). Deployed on Railway.
- **Frontend** — an Angular single-page app (signals-based, Tailwind CSS). Deployed on Netlify.

```
Browser (Angular SPA)  ──HTTP/JSON──▶  Go API (Gin)  ──pgx──▶  PostgreSQL
        │                                   │
        └── Google OAuth sign-in ───────────┘
```

There are three classes of caller, which shapes the whole API:

| Caller | Auth | What they touch |
|--------|------|-----------------|
| **Organizer** (logged-in teacher/admin) | Google OAuth → organizer JWT | Full CRUD on their own events, stages, slots, DJs |
| **DJ** | A per-DJ portal token (Bearer, never in URL) | View own slots; confirm/flag own slots |
| **Public** | None | Read-only event schedule + single-slot share card |

---

## 2. Repository layout

```
dj-scheduler/
├── backend/                     # Go API
│   ├── cmd/
│   │   ├── app/main.go           # entry point → app.Run()
│   │   ├── migrate/main.go       # standalone migration runner
│   │   └── mintdevtoken/main.go  # TEST-only: prints a signed organizer JWT
│   ├── internal/
│   │   ├── domain/               # models, repository interfaces, app errors
│   │   ├── usecase/              # application logic (auth, dj, event, stage, slot, linenotify)
│   │   ├── infrastructure/       # config, crypto, database (pgx), googleauth
│   │   └── interfaces/http/      # Gin handlers, router, middleware
│   ├── migrations/               # 001–009 *.sql, applied in lexical order
│   ├── seed/seed_test.sql        # fixed-UUID fixtures for E2E
│   ├── docs/                     # swagger (swaggo) generated docs
│   ├── Dockerfile
│   └── template.env              # copy to .env (gitignored)
├── frontend/                    # Angular SPA
│   └── src/app/{admin,schedule,dj-portal,card,auth,shared,services,guards}
├── docker-compose.yml            # local Postgres + API
├── docker-compose.test.yml       # throwaway E2E stack
└── .github/workflows/            # backend-ci, frontend-ci, e2e
```

---

## 3. Backend architecture (Clean Architecture)

Dependencies point strictly **inward**. `domain` depends on nothing; `infrastructure` and `interfaces` depend on `domain` (and `usecase`); nothing inner depends on outer.

| Layer | Path | Responsibility |
|-------|------|----------------|
| Entry point | `cmd/app/main.go` | Calls `app.Run()`; carries swagger `@title` / `@BasePath /` annotations |
| Bootstrap / DI | `internal/app/app.go` | Manual constructor wiring of the whole graph |
| Domain — models | `internal/domain/model/` | Pure entity structs, no dependencies |
| Domain — ports | `internal/domain/repository/interfaces.go` | Repository interfaces |
| Domain — errors | `internal/domain/apperrors/errors.go` | Sentinel errors + `ConflictError` |
| Use cases | `internal/usecase/{auth,dj,event,stage,slot,linenotify}` | Application logic; depend only on repo interfaces + models |
| Interfaces (HTTP) | `internal/interfaces/http/` | Gin handlers, router, middleware — parse requests, call use cases, serialize JSON |
| Infrastructure | `internal/infrastructure/{config,crypto,database,googleauth}` | pgx repo implementations (all SQL), AES, Google OAuth, env config |
| Token | `internal/token/jwt.go` | Signs organizer JWTs (shared claim shape with auth middleware) |

### Dependency injection

There is **no DI framework** — wiring is explicit in `internal/app/app.go`:

1. `config.Load()` → `Validate()`, `ValidateGoogle()`, `ValidateLineNotify()` (fail-fast at startup).
2. `database.InitDB(cfg.DatabaseURL)` → `*pgxpool.Pool` (deferred `Close()`).
3. Construct repositories (`NewDJRepository`, `NewEventRepository`, `NewStageRepository`, `NewSlotRepository`, `NewOrganizerRepository`) — each holds the pool.
4. Construct use cases (`djuc.New(djRepo)`, `slotuc.New(slotRepo)`, `authuc.New(googleAuth, organizerRepo, cfg.JWTSecret, 24h)`, …).
5. Construct handlers, build the router via `NewRouter(...)`, then mount auth routes separately with `authHandler.Register(r)` (auth routes are unauthenticated).
6. `r.Run(":" + cfg.Port)`.

Use cases are cheap, stateless wrappers — some handlers get their own fresh instances (e.g. `publicHandler`, `shareHandler`), so don't assume singletons.

---

## 4. Domain models

`internal/domain/model/models.go` and `organizer.go`:

- **Organizer** — `ID, Email, Name, GoogleID, CreatedAt`. Created on first Google sign-in.
- **DJ** — `ID, Name, GenreTags []string, Certifications []string, IsStudent bool, CreatedAt`. The certification gate applies to students and is bypassed for graduates.
- **Event** — `ID, Name, VenueName, StartDate, EndDate, Genres []string, LineNotifyEnabled bool`. The raw LINE token is never exposed — only the boolean.
- **Stage** — `ID, EventID, Name, Color, DisplayOrder`.
- **Slot** — `ID, EventID, StageID, StageName, DjID, DjName, Genre, SlotDate, StartTime, EndTime, Notes, DJConfirmation *string`. `DJConfirmation` is a pointer so `null` round-trips in JSON (`"confirmed"` / `"flagged"` / `nil`).
- **PortalSlot** — the DJ-portal projection of a booking spanning events; carries `EventName, StageName, DJConfirmation` directly.

**Relationships:** Organizer 1—* Event; Organizer 1—* DJ; Event 1—* Stage; Event 1—* Slot; Stage 1—* Slot; DJ 0/1—* Slot (`slots.dj_id` nullable, `ON DELETE SET NULL`).

---

## 5. HTTP API surface

Router: `internal/interfaces/http/router.go`. Built with `gin.New()` (not `Default()`) deliberately, to avoid logging that could leak secrets. Global middleware: `gin.Recovery()`, `middleware.RequestLogger()`, and CORS (`cors.New(...)` allowing only `FRONTEND_URL`, methods GET/POST/PUT/PATCH/DELETE/OPTIONS, headers Content-Type + Authorization).

Two `/api` groups exist: a **public** one and a **protected** one (`api.Use(middleware.Auth(jwtSecret))`).

### Public routes (no auth) — `/api`

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/events/:id/public` | Event + stages + slots in one payload (shareable schedule) |
| GET | `/api/slots/:id/public` | Single slot + parent event (backs the share card) |
| GET | `/api/dj/portal` | DJ + their slots; portal token via `Authorization: Bearer <token>` |
| PATCH | `/api/dj/portal/slots/:slot_id` | DJ confirms/flags own slot; token-gated |

### Protected routes (organizer JWT) — `/api`

- **DJs:** `GET /api/djs` (filters: `certified_for`, `ready=true`), `POST /api/djs`, `GET /api/djs/:id`, `PATCH /api/djs/:id`, `DELETE /api/djs/:id`
- **DJ portal tokens:** `POST /api/djs/:id/token` (mints a portal token; regenerating revokes the prior one)
- **Events:** `GET/POST /api/events`, `GET/PATCH/DELETE /api/events/:id`, `POST /api/events/:id/clone`
- **Stages:** `GET/POST /api/events/:id/stages`, `GET/PATCH/DELETE /api/events/:id/stages/:stage_id`
- **Slots:** `GET/POST /api/events/:id/slots`, `GET/PATCH/DELETE /api/events/:id/slots/:slot_id`
- **LINE:** `PUT /api/events/:id/line-token`

### Auth routes (unauthenticated) — root group

- `GET /auth/google` — redirect to Google consent
- `GET /auth/google/callback` — exchange code → JWT, redirect to frontend

### Other root routes

- `GET /s/dj/:slotId` — server-rendered OG/Twitter meta page that redirects humans to the SPA card (tokenless; uses `html/template` for XSS-safe escaping)
- `GET /healthz` — liveness probe (used by compose healthchecks / CI)
- `GET /swagger/*any` — Swagger UI

> **Stale README note:** the clone endpoint is `POST /api/events/:id/clone` (the old README says `/duplicate`). Trust the router.

**Middleware applicability:** logger + CORS + recovery apply everywhere. `middleware.Auth(jwtSecret)` applies only to the protected `/api` group. The DJ-portal public routes do their own Bearer-token check inside the handler (`portalToken()`), not via the JWT middleware. **There is no scoping middleware** — tenant isolation is enforced in SQL (next section).

---

## 6. Auth & multi-tenancy

### Google OAuth → organizer JWT

1. `auth_handler.go` `redirect`: generates a 256-bit CSRF `state`, sets a short-lived (`5 min`) `oauth_state` cookie (HttpOnly, SameSite=Lax, `Secure` driven by `COOKIE_SECURE`), and redirects to Google's consent URL.
2. `googleauth/google.go`: `oauth2.Config` with scopes `openid email profile`; `UserFromCode` exchanges the code and fetches `userinfo` → `GoogleUser{GoogleID, Email, Name}`.
3. `callback`: validates `state == cookieState`, clears the cookie (one-time use), calls `uc.HandleCallback`, redirects to `FRONTEND_URL + /auth/callback?token=<jwt>`.
4. `usecase/auth/usecase.go` `HandleCallback`: `UserFromCode` → `upsertOrganizer` (find by GoogleID, create on first sign-in) → `token.Sign(...)`.
5. `token/jwt.go` `Sign`: **HS256** JWT, claims `{organizer_id, email, iat, exp}`, 24h TTL.
6. `middleware/auth.go`: validates the `Bearer` token, **rejects non-HMAC signing methods** (guards against alg-confusion), distinguishes expired vs invalid, and stores `organizer_id` in context under `middleware.OrganizerIDKey`. Handlers read it via `c.MustGet(middleware.OrganizerIDKey).(string)`.

### Organizer scoping (multi-tenancy)

Added in migration `004_organizer_scoping.sql`: `organizer_id UUID NOT NULL REFERENCES organizers(id)` on `events` and `djs`. The migration is non-destructive — it backfills existing rows to a synthetic **legacy organizer** (`google_id='legacy-pre-auth'`) before enforcing `NOT NULL`.

Isolation is enforced **in SQL**, in every protected query:
- events/djs filter `WHERE organizer_id = $n`;
- stages/slots join through the parent event (`JOIN events e … WHERE e.organizer_id = $n`).

A resource owned by another organizer reads as **404 (`ErrNotFound`), not 403** — this deliberately avoids leaking the existence of other tenants' rows (event UUIDs are otherwise public via share links). The `eventOwnedBy` / `queryEventOwned` helpers in `db.go` distinguish "owned but empty" (→ 200 empty list) from "not yours" (→ 404).

Tests: `internal/interfaces/http/scoping_test.go`, `database/organizer_repository_test.go`.

---

## 7. Key use-case features

- **Slot conflict detection** — `usecase/slot/conflict.go`. `CheckConflicts` runs two passes: DJ double-booked on any stage (`"dj_double_booked"`) and same-stage overlap (`"stage_overlap"`). It builds an absolute timeline (day ordinal × minutes-per-day + time-of-day) so cross-midnight sets (end ≤ start ⇒ +1 day) and cross-date overlaps work. Ends are exclusive (adjacent slots don't conflict). Returns `*apperrors.ConflictError`; the handler maps it to **409** with `{error, type, conflicting_slot_id}`. The DB also enforces this race-proof (migration 007); `slot_repository.go` maps PG error `23P01` to the same `ConflictError`.
- **DJ portal tokens** — `usecase/dj/portal.go` + migration 005. 32 random bytes hex-encoded; only the **SHA-256 hash** is stored, 90-day TTL. The raw token is returned once and never persisted or logged. Regenerating overwrites the hash (revokes the old). Lookups check `expires_at > now()`. Token travels in the `Authorization` header, never the URL.
- **Slot confirmation** — migration 006 adds `slots.dj_confirmation TEXT CHECK IN ('confirmed','flagged')` (NULL = no response). `SetDJConfirmation` is scoped to `dj_id`; a non-matching row returns `ErrForbidden` → 403.
- **LINE Notify** — `usecase/linenotify/usecase.go` + migration 008 (`events.line_notify_token_enc`). `SetToken` encrypts the raw token via AES-256-GCM before storage (or clears it). A non-owned event returns `ErrNotFound` → 403. The raw token is never returned.
- **DJ certifications** — migration 009 adds `certifications TEXT[]` + `is_student BOOLEAN`. List filters: `certified_for` (case-insensitive `unnest` match) and `ready` (`cardinality(certifications) > 0`), surfaced via `repository.DJListFilter`.

---

## 8. Database

Migrations live in `backend/migrations/` and apply in lexical order:

| File | Adds |
|------|------|
| `001_init.sql` | `pgcrypto`; tables `djs, events, stages, slots` (UUID PKs); FKs with cascade (`stages/slots → events ON DELETE CASCADE`, `slots → djs ON DELETE SET NULL`); base indexes |
| `002_genres.sql` | `events.genres TEXT[]`, `slots.genre TEXT` |
| `003_organizers.sql` | `organizers` table (unique `email`, unique `google_id`) |
| `004_organizer_scoping.sql` | `organizer_id` on events/djs; legacy-org backfill; NOT NULL; indexes |
| `005_dj_portal_tokens.sql` | `djs.portal_token_hash`, `portal_token_expires_at`; partial index |
| `006_slot_confirmation.sql` | `slots.dj_confirmation` with CHECK constraint |
| `007_slot_overlap_constraints.sql` | `btree_gist`; GiST EXCLUDE constraints `slots_stage_no_overlap` + `slots_dj_no_overlap` (partial, `WHERE dj_id IS NOT NULL`) using `tsrange` with +1 day for cross-midnight |
| `008_line_notify.sql` | `events.line_notify_token_enc TEXT` |
| `009_dj_certifications.sql` | `djs.certifications TEXT[]`, `is_student BOOLEAN DEFAULT true` |

**Repository pattern** (`internal/infrastructure/database/`): one repo struct per entity, each holding `*pgxpool.Pool`. **All SQL is centralized as constants in `queries.go`** — handlers and use cases never see SQL. pgx maps: `pgx.ErrNoRows` → `apperrors.ErrNotFound`; `pgconn.PgError` code `23P01` → `ConflictError`. `db.go` has `InitDB` (pool + ping) and the ownership helpers. Public/unscoped query variants back the share and public endpoints.

---

## 9. Configuration & secrets

`internal/infrastructure/config/config.go` reads plain environment variables (no framework):

| Env var | Required | Default | Notes |
|---------|----------|---------|-------|
| `DATABASE_URL` | **Yes** | — | Postgres DSN; app fatals if empty |
| `PORT` | No | `8080` | |
| `FRONTEND_URL` | No | `http://localhost:4200` | CORS origin + redirect base |
| `JWT_SECRET` | **Yes** | — | HMAC secret, **min 32 chars** |
| `COOKIE_SECURE` | No | `false` | `=="true"` → Secure OAuth cookies (prod) |
| `GOOGLE_CLIENT_ID` | **Yes** | — | all three required by `ValidateGoogle` |
| `GOOGLE_CLIENT_SECRET` | **Yes** | — | |
| `GOOGLE_REDIRECT_URL` | **Yes** | — | e.g. `http://localhost:8080/auth/google/callback` |
| `LINE_NOTIFY_ENCRYPTION_KEY` | **Yes** | — | 64-char hex = 32-byte AES-256 key |

Copy `backend/template.env` to `backend/.env` (gitignored). Generate secrets:

```bash
openssl rand -base64 48   # JWT_SECRET
openssl rand -hex 32      # LINE_NOTIFY_ENCRYPTION_KEY
```

**Crypto** (`internal/infrastructure/crypto/aes.go`): AES-256-GCM. `Encrypt` prepends a fresh random nonce and returns base64 (`nonce||ciphertext`); `Decrypt` reverses and authenticates. Requires exactly a 32-byte key. The **only thing encrypted at rest** is the per-event LINE Notify token. DJ portal tokens are SHA-256 *hashed* (not encrypted); JWTs are HMAC-signed.

> **Security reminder:** never log, print, or commit secrets — organizer email/name/`google_id`, OAuth codes, JWTs, portal tokens, and LINE tokens are all sensitive. `gin.New()` is used precisely to avoid default logging that could leak them.

---

## 10. Running it locally

### Option A — Docker Compose (fastest)

```bash
cp backend/template.env backend/.env   # fill in Google creds; set JWT_SECRET, LINE key
docker compose up --build
```

`docker-compose.yml` starts `db` (postgres:16-alpine, auto-applies `backend/migrations/` on first run) and `api` (builds `./backend`, loads `backend/.env`, points `DATABASE_URL` at the in-network `db` host). API on `:8080`.

### Option B — run the Go API directly

```bash
# Start just Postgres
docker compose up -d db

cd backend
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/eventlineup?sslmode=disable"
export JWT_SECRET="<32+ chars>"
export GOOGLE_CLIENT_ID=... GOOGLE_CLIENT_SECRET=... GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
export LINE_NOTIFY_ENCRYPTION_KEY="$(openssl rand -hex 32)"

go run ./cmd/migrate   # apply migrations
go run ./cmd/app        # start the API on :8080
```

### Frontend

```bash
cd frontend
npm install
npm start          # ng serve on http://localhost:4200
```

### Minting a dev organizer token (test-only)

To call protected endpoints without going through Google OAuth:

```bash
JWT_SECRET=<your-secret> go run ./cmd/mintdevtoken
# flags: -organizer-id (default seeded E2E org), -email, -ttl (24h), -secret
```

Send the printed token as `Authorization: Bearer <token>`. **This tool must never ship to production.**

---

## 11. Build, test & CI

### Commands

```bash
# Backend
go run ./cmd/app                    # run server
go run ./cmd/migrate                # apply migrations
go test -race -shuffle=on ./...     # tests (needs a real Postgres + DATABASE_URL)
golangci-lint run                   # lint (v2)

# Frontend
npm test                            # unit tests
npm run build                       # production build
```

Backend tests are integration tests that require a live Postgres. Seed fixtures (`backend/seed/seed_test.sql`) use fixed UUIDs (organizer `…0001`; DJ `…00d1` has raw portal token `e2e-portal-token-dj1`).

### CI (`.github/workflows/`)

- **backend-ci.yml** — three jobs:
  - **test** — spins up postgres:16, checks `go mod tidy` is clean, `go build`, `go vet`, runs migrations, then `go test -race -shuffle=on -count=1` with a **coverage floor ≥ 73%**.
  - **lint** — golangci-lint v2 (pinned `v2.12.2`); config in `backend/.golangci.yml` (errcheck, govet, ineffassign, staticcheck, unused, bodyclose, misspell).
  - **vuln** — `govulncheck`.
- **e2e.yml** — brings up `docker-compose.test.yml`, seeds the test data, runs the frontend Playwright suite (Chromium), uploads the report.
- **frontend-ci.yml** — frontend build/test/lint.

### API docs

Swagger (swaggo) is generated into `backend/docs/` and served at `/swagger/*any`. Annotations live on `cmd/app/main.go` and the individual handlers.

> **Tooling caveats:** the README and some seed comments reference `make` targets (`make e2e-token`, `make e2e-seed`) — **there is no Makefile**; those are stale. Use the `go run ./cmd/...` commands above.

---

## 12. Frontend orientation

Angular SPA, **standalone components + signals** for reactive state, Tailwind CSS, `ngx-translate` (EN + Traditional Chinese). Routes (`src/app/app.routes.ts`):

| Route | Component area | Audience |
|-------|----------------|----------|
| `/login`, `/auth/callback` | `auth/` | Public |
| `/admin/events`, `/admin/events/new`, `/admin/events/:id` | `admin/events/` | Organizer (auth guard) |
| `/admin/djs` | `admin/djs/` | Organizer (auth guard) |
| `/events/:id` | `schedule/` | Public schedule view |
| `/dj/portal` | `dj-portal/` | DJ (token link) |
| `/card/:slotId` | `card/` | Public promo set-card |

Admin routes are protected by `guards/auth.guard.ts`. The shared shell (`shared/admin-shell.component`) provides the sidebar; `app.component` provides the top bar (brand, language toggle, sign-out). API calls go through `services/`. Translations: `src/assets/i18n/{en,zh-TW}.json` (keep both files in sync — every key must exist in both).

> Note: `src/app/landing/` exists but is **not wired into the routes** and has no i18n strings — it's not currently reachable.

---

## 13. Conventions for contributors

- **Tests with every change.** Add tests for new cases — Go (`*_test.go`, table-driven where it fits) on the backend, Vitest specs on the frontend. CI enforces a backend coverage floor.
- **All SQL lives in `queries.go`.** Don't inline SQL in handlers or use cases.
- **Every protected query is organizer-scoped.** When adding an endpoint that touches tenant data, filter by `organizer_id` (directly or via the parent event) and return 404 (not 403) for other tenants' rows.
- **Secrets never leak.** Don't log tokens, OAuth codes, or organizer PII. Encrypt secrets at rest (AES helper) or hash them (portal tokens).
- **Tailwind:** use utility classes in markup; avoid `@apply` in component CSS (it blows the per-component budget and fails `ng build`).
- **i18n:** any new user-visible string needs a key in both `en.json` and `zh-TW.json`.
