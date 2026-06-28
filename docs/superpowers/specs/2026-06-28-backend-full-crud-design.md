# Backend Full CRUD Design

**Date:** 2026-06-28  
**Branch:** feature/eventlineup-mvp  
**Status:** Approved

## Problem

The backend has List/Create/Delete for most resources but is missing Get-by-ID and Update (PATCH) endpoints. The slot update endpoint uses PUT rather than PATCH. There is no Swagger documentation.

## Approach

Extend in place — add missing operations at each layer following the established 4-layer pattern (handler → usecase → repository interface → DB). No structural changes.

---

## CRUD Gap (before this work)

| Resource | List | Get | Create | Update | Delete |
|----------|:----:|:---:|:------:|:------:|:------:|
| DJs      | ✅   | ❌  | ✅     | ❌     | ✅     |
| Events   | ✅   | ✅  | ✅     | ❌     | ✅     |
| Stages   | ✅   | ❌  | ✅     | ❌     | ✅     |
| Slots    | ✅   | ❌  | ✅     | ✅ PUT | ✅     |

---

## New Routes

```
GET    /api/djs/:id
PATCH  /api/djs/:id                        — update name, genre_tags
PATCH  /api/events/:id                     — update name, venue_name, start_date, end_date, genres
GET    /api/events/:id/stages/:stage_id
PATCH  /api/events/:id/stages/:stage_id    — update name, color
GET    /api/events/:id/slots/:slot_id
PATCH  /api/events/:id/slots/:slot_id      — same fields as before, replaces PUT
GET    /swagger/*any                       — Swagger UI
```

PATCH semantics: full replacement of all mutable fields. No partial/pointer logic.

---

## Architecture

### Domain — `internal/domain/repository/interfaces.go`

```
DJRepository:    + Get(ctx, id) (DJ, error)
                 + Update(ctx, dj DJ) (DJ, error)
EventRepository: + Update(ctx, e Event) (Event, error)
StageRepository: + Get(ctx, id, eventID string) (Stage, error)
                 + Update(ctx, s Stage) (Stage, error)
SlotRepository:  + Get(ctx, id, eventID string) (Slot, error)
```

### Infrastructure — `internal/infrastructure/database/`

New SQL queries in `queries.go`:
- `queryDJGet` — SELECT by id
- `queryDJUpdate` — UPDATE name, genre_tags RETURNING *
- `queryEventUpdate` — UPDATE name, venue_name, start_date, end_date, genres RETURNING *
- `queryStageGet` — SELECT by id AND event_id
- `queryStageUpdate` — UPDATE name, color RETURNING *
- `querySlotGet` — SELECT with JOIN (same shape as querySlotList) WHERE id AND event_id

Repository implementations: `Get`/`Update` added to each of the four repo files. Not-found detection: `pgx.ErrNoRows` → `apperrors.ErrNotFound`.

### Use cases — `internal/usecase/`

Thin pass-throughs for each new repo method across `dj`, `event`, `stage`, `slot` packages.

### Handlers — `internal/interfaces/http/`

- `dj_handler.go`: +`GET /djs/:id`, +`PATCH /djs/:id`
- `event_handler.go`: +`PATCH /events/:id`
- `stage_handler.go`: +`GET /events/:id/stages/:stage_id`, +`PATCH /events/:id/stages/:stage_id`
- `slot_handler.go`: PUT→PATCH, +`GET /events/:id/slots/:slot_id`
- All handler functions get swaggo annotations (`@Summary`, `@Tags`, `@Param`, `@Success`, `@Failure`, `@Router`)

### Router — `internal/interfaces/http/router.go`

- Add `"PATCH"` to CORS `AllowMethods`
- Mount Swagger UI at `GET /swagger/*any`

### Main — `cmd/app/main.go`

- Add swaggo top-level annotations (`@title`, `@version`, `@description`, `@host`, `@BasePath`)
- Import `_ "eventlineup/docs"`

### Generated — `backend/docs/`

Run `swag init -g cmd/app/main.go` from `backend/` to produce `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`. Commit generated files.

---

## Swagger

Libraries: `github.com/swaggo/swag`, `github.com/swaggo/gin-swagger`, `github.com/swaggo/files`

Error responses documented: 400, 404, 500.

---

## Tests

Integration tests, real DB, skipped when `DATABASE_URL` not set.

New:
- `TestGetDJ` — 200 valid ID, 404 unknown UUID
- `TestPatchDJ` — 200 updates name+tags, 404 unknown UUID
- `TestPatchEvent` — 200 updates all mutable fields
- `TestGetStage` — 200 valid ID, 404 unknown UUID
- `TestPatchStage` — 200 updates name+color, 404 unknown UUID
- `TestGetSlot` — 200 valid ID, 404 unknown UUID

Updated:
- `TestUpdateSlot` — `http.MethodPut` → `http.MethodPatch`

---

## Files Changed

| File | Change |
|------|--------|
| `go.mod` / `go.sum` | add swaggo deps |
| `internal/domain/repository/interfaces.go` | +6 method signatures |
| `internal/infrastructure/database/queries.go` | +6 SQL queries |
| `internal/infrastructure/database/dj_repository.go` | +Get, +Update |
| `internal/infrastructure/database/event_repository.go` | +Update |
| `internal/infrastructure/database/stage_repository.go` | +Get, +Update |
| `internal/infrastructure/database/slot_repository.go` | +Get |
| `internal/usecase/dj/usecase.go` | +Get, +Update |
| `internal/usecase/event/usecase.go` | +Update |
| `internal/usecase/stage/usecase.go` | +Get, +Update |
| `internal/usecase/slot/usecase.go` | +Get |
| `internal/interfaces/http/dj_handler.go` | +get, +patch, swaggo on all |
| `internal/interfaces/http/event_handler.go` | +patch, swaggo on all |
| `internal/interfaces/http/stage_handler.go` | +get, +patch, swaggo on all |
| `internal/interfaces/http/slot_handler.go` | PUT→PATCH, +get, swaggo on all |
| `internal/interfaces/http/router.go` | +PATCH to CORS, +Swagger route |
| `cmd/app/main.go` | swaggo annotations, import docs |
| `internal/interfaces/http/dj_handler_test.go` | +TestGetDJ, +TestPatchDJ |
| `internal/interfaces/http/event_handler_test.go` | +TestPatchEvent |
| `internal/interfaces/http/stage_handler_test.go` | +TestGetStage, +TestPatchStage |
| `internal/interfaces/http/slot_handler_test.go` | PUT→PATCH, +TestGetSlot |
| `docs/` | generated by swag init |
