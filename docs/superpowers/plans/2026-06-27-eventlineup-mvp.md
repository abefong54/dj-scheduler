# EventLineup MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a DJ event scheduling web app — Angular 17+ frontend, Go + Gin backend, Postgres database — replacing Excel-based scheduling with a shareable visual schedule.

**Architecture:** Angular SPA calls a Go/Gin REST API. Go connects to Postgres via pgx/v5. No auth for MVP — admin routes are open. Public schedule at `/events/:id` is read-only and shareable via LINE URL scheme. i18n via ngx-translate (EN + zh-TW).

**Tech Stack:** Angular 17+ (standalone), Tailwind CSS, ngx-translate, Go 1.22+, Gin, pgx/v5, Postgres 16, Docker Compose (local dev), Netlify (frontend), Railway (backend).

## Global Constraints

- Go package: `main` (single binary in `backend/`)
- Go router: Gin (`github.com/gin-gonic/gin`) — no `net/http` mux
- CORS middleware: `github.com/gin-contrib/cors`
- All API routes prefixed `/api`
- Backend port: `8080` (env var `PORT`, default `8080`)
- DB: `DATABASE_URL` env var; local value: `postgresql://eventlineup:eventlineup@localhost:5432/eventlineup`
- Angular: standalone components only, no NgModules
- Node/npm/ng: all commands prefixed with `export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"`
- All user-facing strings via ngx-translate keys — no hardcoded English/Chinese in templates
- i18n files: `frontend/src/assets/i18n/en.json` and `frontend/src/assets/i18n/zh-TW.json`
- Stage colors assigned from palette (cycling): `#6366F1` `#10B981` `#F59E0B` `#EF4444` `#8B5CF6` `#EC4899`
- DB tests skip when `DATABASE_URL` unset (`t.Skip`)
- Docker Compose local: `docker compose up` starts db + api together

---

## File Map

**Backend (`backend/`)**

| File | Responsibility |
|---|---|
| `backend/go.mod` | Module + dependencies |
| `backend/main.go` | Server start, CORS, all route registration |
| `backend/db.go` | `InitDB` — pgxpool connection |
| `backend/models.go` | `DJ`, `Event`, `Stage`, `Slot` structs |
| `backend/handlers_djs.go` | `listDJs`, `createDJ`, `deleteDJ` |
| `backend/handlers_events.go` | `listEvents`, `getEvent`, `createEvent`, `deleteEvent`, `duplicateEvent` |
| `backend/handlers_stages.go` | `listStages`, `createStage`, `deleteStage` |
| `backend/handlers_slots.go` | `listSlots`, `createSlot`, `deleteSlot` |
| `backend/db_test.go` | `TestInitDB` + `setupTestDB` helper (shared by all tests) |
| `backend/handlers_djs_test.go` | DJ handler tests |
| `backend/handlers_events_test.go` | Event handler tests |
| `backend/handlers_stages_test.go` | Stage handler tests |
| `backend/handlers_slots_test.go` | Slot handler tests |
| `backend/migrations/001_init.sql` | Schema: djs, events, stages, slots |
| `backend/Dockerfile` | Multi-stage Go build |
| `backend/railway.toml` | Railway deploy config |

**Root**

| File | Responsibility |
|---|---|
| `docker-compose.yml` | Local dev: db + api services |

**Frontend (`frontend/`)**

| File | Responsibility |
|---|---|
| `frontend/src/main.ts` | Bootstrap Angular app |
| `frontend/src/app/app.config.ts` | App providers: router, HTTP, translate |
| `frontend/src/app/app.component.ts` | Root: top nav + `<router-outlet>` |
| `frontend/src/app/app.routes.ts` | All route definitions |
| `frontend/src/app/services/api.service.ts` | All HTTP calls to Go backend |
| `frontend/src/environments/environment.ts` | `apiUrl: 'http://localhost:8080'` |
| `frontend/src/environments/environment.prod.ts` | `apiUrl: ''` (updated at deploy) |
| `frontend/src/assets/i18n/en.json` | English strings |
| `frontend/src/assets/i18n/zh-TW.json` | Traditional Chinese strings |
| `frontend/src/app/admin/events/events-list.component.ts` | Event list — upcoming/past tabs |
| `frontend/src/app/admin/events/event-new.component.ts` | Create event form |
| `frontend/src/app/admin/events/event-detail.component.ts` | Stages panel + slots list |
| `frontend/src/app/admin/djs/djs.component.ts` | DJ roster |
| `frontend/src/app/schedule/schedule.component.ts` | Public schedule grid |
| `frontend/netlify.toml` | Netlify build + SPA redirect |

---

### Task 1: Go project setup, DB connection, models

**Files:**
- Create: `backend/go.mod`
- Create: `backend/db.go`
- Create: `backend/models.go`
- Create: `backend/main.go`
- Create: `backend/db_test.go`

**Interfaces:**
- Produces: `InitDB(connStr string) (*pgxpool.Pool, error)`
- Produces: `DJ{ID, Name, GenreTags, CreatedAt}`, `Event{ID, Name, VenueName, StartDate, EndDate}`, `Stage{ID, EventID, Name, Color, DisplayOrder}`, `Slot{ID, EventID, StageID, StageName, DjID, DjName, SlotDate, StartTime, EndTime, Notes}`
- Produces: `registerRoutes(r *gin.Engine, pool *pgxpool.Pool)` stub (filled in later tasks)

---

- [ ] **Step 1: Create backend directory and Go module**

```bash
mkdir -p backend && cd backend
go mod init eventlineup
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
```

- [ ] **Step 2: Write failing DB test**

```go
// backend/db_test.go
package main

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInitDB(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	pool, err := InitDB(url)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

// setupTestDB is shared by all handler tests in this package.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	pool, err := InitDB(url)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
```

- [ ] **Step 3: Run test — expect FAIL**

```bash
cd backend && go test ./... -run TestInitDB -v
```
Expected: `FAIL — InitDB undefined`

- [ ] **Step 4: Write db.go, models.go, main.go**

```go
// backend/db.go
package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}
```

```go
// backend/models.go
package main

type DJ struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	GenreTags []string `json:"genre_tags"`
}

type Event struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	VenueName string `json:"venue_name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type Stage struct {
	ID           string `json:"id"`
	EventID      string `json:"event_id"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

type Slot struct {
	ID        string `json:"id"`
	EventID   string `json:"event_id"`
	StageID   string `json:"stage_id"`
	StageName string `json:"stage_name"`
	DjID      string `json:"dj_id"`
	DjName    string `json:"dj_name"`
	SlotDate  string `json:"slot_date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Notes     string `json:"notes"`
}
```

```go
// backend/main.go
package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	pool, err := InitDB(dbURL)
	if err != nil {
		log.Fatalf("InitDB: %v", err)
	}
	defer pool.Close()

	r := gin.Default()

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{frontendURL},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))

	registerRoutes(r, pool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(r.Run(":" + port))
}

func registerRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	// filled in by handler tasks
}
```

- [ ] **Step 5: Run test — expect PASS (with docker compose db running)**

```bash
cd backend && DATABASE_URL=postgresql://eventlineup:eventlineup@localhost:5432/eventlineup go test ./... -run TestInitDB -v
```
Expected: `PASS`

- [ ] **Step 6: Commit**

```bash
git add backend/
git commit -m "feat(backend): Go project setup, DB connection, models"
```

---

### Task 2: Database schema migration + Docker Compose

**Files:**
- Create: `backend/migrations/001_init.sql`
- Create: `docker-compose.yml`

**Interfaces:**
- Produces: tables `djs`, `events`, `stages`, `slots` in Postgres
- Produces: `docker compose up` starts db (auto-migrates) + api

---

- [ ] **Step 1: Write migration SQL**

```sql
-- backend/migrations/001_init.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS djs (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  genre_tags TEXT[] DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS events (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  venue_name  TEXT NOT NULL,
  start_date  DATE NOT NULL,
  end_date    DATE NOT NULL,
  created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS stages (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id      UUID REFERENCES events(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  color         TEXT NOT NULL DEFAULT '#6366F1',
  display_order INTEGER DEFAULT 0,
  created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS slots (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id   UUID REFERENCES events(id) ON DELETE CASCADE,
  stage_id   UUID REFERENCES stages(id) ON DELETE CASCADE,
  dj_id      UUID REFERENCES djs(id) ON DELETE SET NULL,
  slot_date  DATE NOT NULL,
  start_time TIME NOT NULL,
  end_time   TIME NOT NULL,
  notes      TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_stages_event ON stages(event_id);
CREATE INDEX IF NOT EXISTS idx_slots_event  ON slots(event_id);
CREATE INDEX IF NOT EXISTS idx_slots_date   ON slots(slot_date);
```

- [ ] **Step 2: Write docker-compose.yml**

```yaml
# docker-compose.yml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: eventlineup
      POSTGRES_PASSWORD: eventlineup
      POSTGRES_DB: eventlineup
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U eventlineup"]
      interval: 5s
      timeout: 5s
      retries: 5

  api:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgresql://eventlineup:eventlineup@db:5432/eventlineup
      PORT: "8080"
      FRONTEND_URL: "http://localhost:4200"
    depends_on:
      db:
        condition: service_healthy

volumes:
  pgdata:
```

- [ ] **Step 3: Start db and verify migration ran**

```bash
docker compose up db -d
sleep 3
docker compose exec db psql -U eventlineup -d eventlineup -c "\dt"
```
Expected: rows for `djs`, `events`, `stages`, `slots`

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/001_init.sql docker-compose.yml
git commit -m "feat(db): schema migration + docker-compose local dev"
```

---

### Task 3: DJs API handlers

**Files:**
- Create: `backend/handlers_djs.go`
- Create: `backend/handlers_djs_test.go`
- Modify: `backend/main.go` — add `registerDJRoutes` call in `registerRoutes`

**Interfaces:**
- Consumes: `DJ` from `models.go`, `pool` from `db.go`
- Produces: `registerDJRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool)` — registers `GET/POST /api/djs`, `DELETE /api/djs/:id`

---

- [ ] **Step 1: Write failing tests**

```go
// backend/handlers_djs_test.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func djRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerDJRoutes(r.Group("/api"), pool)
	return r
}

func TestListDJs(t *testing.T) {
	pool := setupTestDB(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/djs", nil)
	djRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var djs []DJ
	if err := json.NewDecoder(w.Body).Decode(&djs); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestCreateDJ(t *testing.T) {
	pool := setupTestDB(t)
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM djs WHERE name = 'test-dj'")
	})

	body, _ := json.Marshal(map[string]interface{}{
		"name":       "test-dj",
		"genre_tags": []string{"hip hop", "r&b"},
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/djs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	djRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var d DJ
	json.NewDecoder(w.Body).Decode(&d)
	if d.ID == "" {
		t.Fatal("expected ID in response")
	}
	if len(d.GenreTags) != 2 {
		t.Fatalf("expected 2 genre tags, got %v", d.GenreTags)
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

```bash
cd backend && go test ./... -run "TestListDJs|TestCreateDJ" -v
```
Expected: `FAIL — registerDJRoutes undefined`

- [ ] **Step 3: Write handlers_djs.go**

```go
// backend/handlers_djs.go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerDJRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	rg.GET("/djs", listDJs(pool))
	rg.POST("/djs", createDJ(pool))
	rg.DELETE("/djs/:id", deleteDJ(pool))
}

func listDJs(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT id, name, COALESCE(genre_tags, '{}') FROM djs ORDER BY name`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		djs := []DJ{}
		for rows.Next() {
			var d DJ
			if err := rows.Scan(&d.ID, &d.Name, &d.GenreTags); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			djs = append(djs, d)
		}
		c.JSON(http.StatusOK, djs)
	}
}

func createDJ(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var d DJ
		if err := c.ShouldBindJSON(&d); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if d.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
			return
		}
		if d.GenreTags == nil {
			d.GenreTags = []string{}
		}
		err := pool.QueryRow(context.Background(),
			`INSERT INTO djs (name, genre_tags) VALUES ($1, $2)
			 RETURNING id, name, COALESCE(genre_tags, '{}')`,
			d.Name, d.GenreTags).Scan(&d.ID, &d.Name, &d.GenreTags)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, d)
	}
}

func deleteDJ(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM djs WHERE id = $1`, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
```

- [ ] **Step 4: Wire into main.go**

```go
// backend/main.go — replace registerRoutes stub
func registerRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	api := r.Group("/api")
	registerDJRoutes(api, pool)
}
```

- [ ] **Step 5: Run — expect PASS**

```bash
cd backend && DATABASE_URL=postgresql://eventlineup:eventlineup@localhost:5432/eventlineup go test ./... -run "TestListDJs|TestCreateDJ" -v
```
Expected: `PASS`

- [ ] **Step 6: Commit**

```bash
git add backend/handlers_djs.go backend/handlers_djs_test.go backend/main.go
git commit -m "feat(backend): DJs CRUD API"
```

---

### Task 4: Events API handlers

**Files:**
- Create: `backend/handlers_events.go`
- Create: `backend/handlers_events_test.go`
- Modify: `backend/main.go` — add `registerEventRoutes` in `registerRoutes`

**Interfaces:**
- Consumes: `Event` from `models.go`
- Produces: `registerEventRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool)` — registers:
  - `GET /api/events` → list all events
  - `POST /api/events` → create event
  - `GET /api/events/:id` → get single event
  - `DELETE /api/events/:id` → delete event
  - `POST /api/events/:id/duplicate` → copy event + stages (not slots)

---

- [ ] **Step 1: Write failing tests**

```go
// backend/handlers_events_test.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func eventRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerEventRoutes(r.Group("/api"), pool)
	return r
}

// createTestEvent inserts a test event and registers cleanup.
func createTestEvent(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO events (name, venue_name, start_date, end_date)
		 VALUES ('test-event', 'Test Venue', '2026-07-25', '2026-07-26')
		 RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatalf("createTestEvent: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE id = $1", id)
	})
	return id
}

func TestListEvents(t *testing.T) {
	pool := setupTestDB(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	eventRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var events []Event
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestCreateEvent(t *testing.T) {
	pool := setupTestDB(t)
	t.Cleanup(func() {
		pool.Exec(context.Background(), "DELETE FROM events WHERE name = 'test-create-event'")
	})

	body, _ := json.Marshal(map[string]string{
		"name":       "test-create-event",
		"venue_name": "Test Club",
		"start_date": "2026-07-25",
		"end_date":   "2026-07-26",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	eventRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var e Event
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID == "" {
		t.Fatal("expected ID in response")
	}
}

func TestGetEvent(t *testing.T) {
	pool := setupTestDB(t)
	id := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+id, nil)
	eventRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var e Event
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID != id {
		t.Fatalf("expected id %s, got %s", id, e.ID)
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

```bash
cd backend && go test ./... -run "TestListEvents|TestCreateEvent|TestGetEvent" -v
```
Expected: `FAIL — registerEventRoutes undefined`

- [ ] **Step 3: Write handlers_events.go**

```go
// backend/handlers_events.go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerEventRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	rg.GET("/events", listEvents(pool))
	rg.POST("/events", createEvent(pool))
	rg.GET("/events/:id", getEvent(pool))
	rg.DELETE("/events/:id", deleteEvent(pool))
	rg.POST("/events/:id/duplicate", duplicateEvent(pool))
}

func listEvents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT id, name, venue_name, start_date::text, end_date::text
			 FROM events ORDER BY start_date DESC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		events := []Event{}
		for rows.Next() {
			var e Event
			rows.Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate)
			events = append(events, e)
		}
		c.JSON(http.StatusOK, events)
	}
}

func getEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e Event
		err := pool.QueryRow(context.Background(),
			`SELECT id, name, venue_name, start_date::text, end_date::text
			 FROM events WHERE id = $1`, c.Param("id")).
			Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, e)
	}
}

func createEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e Event
		if err := c.ShouldBindJSON(&e); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if e.Name == "" || e.VenueName == "" || e.StartDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name, venue_name, start_date required"})
			return
		}
		if e.EndDate == "" {
			e.EndDate = e.StartDate
		}
		err := pool.QueryRow(context.Background(),
			`INSERT INTO events (name, venue_name, start_date, end_date)
			 VALUES ($1,$2,$3,$4)
			 RETURNING id, name, venue_name, start_date::text, end_date::text`,
			e.Name, e.VenueName, e.StartDate, e.EndDate).
			Scan(&e.ID, &e.Name, &e.VenueName, &e.StartDate, &e.EndDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, e)
	}
}

func deleteEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM events WHERE id = $1`, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func duplicateEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		origID := c.Param("id")

		var orig Event
		err := pool.QueryRow(ctx,
			`SELECT name, venue_name, start_date::text, end_date::text
			 FROM events WHERE id = $1`, origID).
			Scan(&orig.Name, &orig.VenueName, &orig.StartDate, &orig.EndDate)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		var newEvent Event
		err = pool.QueryRow(ctx,
			`INSERT INTO events (name, venue_name, start_date, end_date)
			 VALUES ($1||' (copy)', $2, $3, $4)
			 RETURNING id, name, venue_name, start_date::text, end_date::text`,
			orig.Name, orig.VenueName, orig.StartDate, orig.EndDate).
			Scan(&newEvent.ID, &newEvent.Name, &newEvent.VenueName, &newEvent.StartDate, &newEvent.EndDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Copy stages (not slots — dates would need adjustment)
		rows, _ := pool.Query(ctx,
			`SELECT name, color, display_order FROM stages WHERE event_id = $1 ORDER BY display_order`, origID)
		defer rows.Close()
		for rows.Next() {
			var s Stage
			rows.Scan(&s.Name, &s.Color, &s.DisplayOrder)
			pool.Exec(ctx,
				`INSERT INTO stages (event_id, name, color, display_order) VALUES ($1,$2,$3,$4)`,
				newEvent.ID, s.Name, s.Color, s.DisplayOrder)
		}

		c.JSON(http.StatusCreated, newEvent)
	}
}
```

- [ ] **Step 4: Wire into main.go**

```go
// backend/main.go — update registerRoutes
func registerRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	api := r.Group("/api")
	registerDJRoutes(api, pool)
	registerEventRoutes(api, pool)
}
```

- [ ] **Step 5: Run — expect PASS**

```bash
cd backend && DATABASE_URL=postgresql://eventlineup:eventlineup@localhost:5432/eventlineup go test ./... -run "TestListEvents|TestCreateEvent|TestGetEvent" -v
```
Expected: `PASS`

- [ ] **Step 6: Commit**

```bash
git add backend/handlers_events.go backend/handlers_events_test.go backend/main.go
git commit -m "feat(backend): events CRUD + duplicate API"
```

---

### Task 5: Stages + Slots API handlers

**Files:**
- Create: `backend/handlers_stages.go`
- Create: `backend/handlers_stages_test.go`
- Create: `backend/handlers_slots.go`
- Create: `backend/handlers_slots_test.go`
- Modify: `backend/main.go` — add both route registrations

**Interfaces:**
- Consumes: `Stage`, `Slot` from `models.go`; `createTestEvent` from `handlers_events_test.go`
- Produces: `registerStageRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool)` — `GET/POST /api/events/:id/stages`, `DELETE /api/events/:id/stages/:stage_id`
- Produces: `registerSlotRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool)` — `GET/POST /api/events/:id/slots`, `DELETE /api/events/:id/slots/:slot_id`

---

- [ ] **Step 1: Write failing stage tests**

```go
// backend/handlers_stages_test.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func stageRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerStageRoutes(r.Group("/api"), pool)
	return r
}

func TestListStages(t *testing.T) {
	pool := setupTestDB(t)
	eventID := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/stages", nil)
	stageRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var stages []Stage
	if err := json.NewDecoder(w.Body).Decode(&stages); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestCreateStage(t *testing.T) {
	pool := setupTestDB(t)
	eventID := createTestEvent(t, pool)

	body, _ := json.Marshal(map[string]string{"name": "A 主舞台", "color": "#6366F1"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/stages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	stageRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var s Stage
	json.NewDecoder(w.Body).Decode(&s)
	if s.ID == "" {
		t.Fatal("expected ID in response")
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

```bash
cd backend && go test ./... -run "TestListStages|TestCreateStage" -v
```
Expected: `FAIL — registerStageRoutes undefined`

- [ ] **Step 3: Write handlers_stages.go**

```go
// backend/handlers_stages.go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerStageRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	rg.GET("/events/:id/stages", listStages(pool))
	rg.POST("/events/:id/stages", createStage(pool))
	rg.DELETE("/events/:id/stages/:stage_id", deleteStage(pool))
}

func listStages(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT id, event_id, name, color, display_order
			 FROM stages WHERE event_id = $1 ORDER BY display_order, name`,
			c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		stages := []Stage{}
		for rows.Next() {
			var s Stage
			rows.Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder)
			stages = append(stages, s)
		}
		c.JSON(http.StatusOK, stages)
	}
}

func createStage(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s Stage
		if err := c.ShouldBindJSON(&s); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if s.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
			return
		}
		if s.Color == "" {
			s.Color = "#6366F1"
		}
		err := pool.QueryRow(context.Background(),
			`INSERT INTO stages (event_id, name, color)
			 VALUES ($1,$2,$3)
			 RETURNING id, event_id, name, color, display_order`,
			c.Param("id"), s.Name, s.Color).
			Scan(&s.ID, &s.EventID, &s.Name, &s.Color, &s.DisplayOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, s)
	}
}

func deleteStage(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM stages WHERE id = $1 AND event_id = $2`,
			c.Param("stage_id"), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
```

- [ ] **Step 4: Write failing slot tests**

```go
// backend/handlers_slots_test.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func slotRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerSlotRoutes(r.Group("/api"), pool)
	return r
}

func createTestStage(t *testing.T, pool *pgxpool.Pool, eventID string) string {
	t.Helper()
	var id string
	pool.QueryRow(context.Background(),
		`INSERT INTO stages (event_id, name, color) VALUES ($1, 'Test Stage', '#6366F1') RETURNING id`,
		eventID).Scan(&id)
	return id
}

func TestListSlots(t *testing.T) {
	pool := setupTestDB(t)
	eventID := createTestEvent(t, pool)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/slots", nil)
	slotRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var slots []Slot
	if err := json.NewDecoder(w.Body).Decode(&slots); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// new event has no slots
	if len(slots) != 0 {
		t.Fatalf("expected empty slot list, got %d", len(slots))
	}
}

func TestCreateSlot(t *testing.T) {
	pool := setupTestDB(t)
	eventID := createTestEvent(t, pool)
	stageID := createTestStage(t, pool, eventID)

	body, _ := json.Marshal(map[string]string{
		"stage_id":   stageID,
		"slot_date":  "2026-07-25",
		"start_time": "16:00",
		"end_time":   "17:30",
		"notes":      "PA",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/slots", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	slotRouter(pool).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var s Slot
	json.NewDecoder(w.Body).Decode(&s)
	if s.ID == "" {
		t.Fatal("expected ID in response")
	}
}
```

- [ ] **Step 5: Run — expect FAIL**

```bash
cd backend && go test ./... -run "TestListSlots|TestCreateSlot" -v
```
Expected: `FAIL — registerSlotRoutes undefined`

- [ ] **Step 6: Write handlers_slots.go**

```go
// backend/handlers_slots.go
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerSlotRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool) {
	rg.GET("/events/:id/slots", listSlots(pool))
	rg.POST("/events/:id/slots", createSlot(pool))
	rg.DELETE("/events/:id/slots/:slot_id", deleteSlot(pool))
}

func listSlots(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(context.Background(), `
			SELECT sl.id, sl.event_id, sl.stage_id, st.name,
			       COALESCE(sl.dj_id::text,''), COALESCE(d.name,''),
			       sl.slot_date::text, sl.start_time::text, sl.end_time::text, COALESCE(sl.notes,'')
			FROM slots sl
			JOIN stages st ON st.id = sl.stage_id
			LEFT JOIN djs d ON d.id = sl.dj_id
			WHERE sl.event_id = $1
			ORDER BY sl.slot_date, sl.start_time`, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		slots := []Slot{}
		for rows.Next() {
			var s Slot
			rows.Scan(&s.ID, &s.EventID, &s.StageID, &s.StageName,
				&s.DjID, &s.DjName, &s.SlotDate, &s.StartTime, &s.EndTime, &s.Notes)
			slots = append(slots, s)
		}
		c.JSON(http.StatusOK, slots)
	}
}

func createSlot(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s Slot
		if err := c.ShouldBindJSON(&s); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if s.StageID == "" || s.SlotDate == "" || s.StartTime == "" || s.EndTime == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stage_id, slot_date, start_time, end_time required"})
			return
		}
		eventID := c.Param("id")
		// dj_id is optional — store NULL when empty
		err := pool.QueryRow(context.Background(),
			`INSERT INTO slots (event_id, stage_id, dj_id, slot_date, start_time, end_time, notes)
			 VALUES ($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7)
			 RETURNING id`,
			eventID, s.StageID, s.DjID, s.SlotDate, s.StartTime, s.EndTime, s.Notes).
			Scan(&s.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		s.EventID = eventID
		c.JSON(http.StatusCreated, s)
	}
}

func deleteSlot(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM slots WHERE id = $1 AND event_id = $2`,
			c.Param("slot_id"), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
```

- [ ] **Step 7: Wire both into main.go**

```go
// backend/main.go — final registerRoutes
func registerRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	api := r.Group("/api")
	registerDJRoutes(api, pool)
	registerEventRoutes(api, pool)
	registerStageRoutes(api, pool)
	registerSlotRoutes(api, pool)
}
```

- [ ] **Step 8: Run all backend tests**

```bash
cd backend && DATABASE_URL=postgresql://eventlineup:eventlineup@localhost:5432/eventlineup go test ./... -v
```
Expected: all `PASS`

- [ ] **Step 9: Commit**

```bash
git add backend/handlers_stages.go backend/handlers_stages_test.go \
        backend/handlers_slots.go backend/handlers_slots_test.go \
        backend/main.go
git commit -m "feat(backend): stages + slots CRUD API"
```

---

### Task 6: Dockerfile + Railway config

**Files:**
- Create: `backend/Dockerfile`
- Create: `backend/railway.toml`

**Interfaces:**
- Produces: `docker build -t eventlineup-backend ./backend` succeeds
- Produces: Railway can deploy using `backend/Dockerfile`

---

- [ ] **Step 1: Write Dockerfile**

```dockerfile
# backend/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

- [ ] **Step 2: Write railway.toml**

```toml
# backend/railway.toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"

[deploy]
startCommand = "./server"
healthcheckPath = "/api/events"
healthcheckTimeout = 10
```

- [ ] **Step 3: Build image locally**

```bash
docker build -t eventlineup-backend ./backend
```
Expected: `Successfully built` (or `writing image sha256:...`)

- [ ] **Step 4: Verify full stack via docker compose**

```bash
docker compose up --build -d
sleep 5
curl -s http://localhost:8080/api/events | head -c 50
```
Expected: `[]` (empty JSON array — no events yet)

- [ ] **Step 5: Commit**

```bash
git add backend/Dockerfile backend/railway.toml
git commit -m "feat(infra): Dockerfile + Railway config"
```

---

### Task 7: Angular scaffold + config + API service

**Files:**
- Create: `frontend/` (ng new)
- Create: `frontend/tailwind.config.js`
- Create: `frontend/src/styles.css` (Tailwind directives)
- Create: `frontend/src/environments/environment.ts`
- Create: `frontend/src/environments/environment.prod.ts`
- Create: `frontend/src/app/app.config.ts`
- Create: `frontend/src/app/app.routes.ts`
- Create: `frontend/src/app/app.component.ts`
- Create: `frontend/src/assets/i18n/en.json`
- Create: `frontend/src/assets/i18n/zh-TW.json`
- Create: `frontend/src/app/services/api.service.ts`

**Interfaces:**
- Produces: `ApiService` with methods for all API calls (see below)
- Produces: `routes` array with all 5 routes
- Produces: `ng build` succeeds

---

- [ ] **Step 1: Scaffold Angular project**

```bash
export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"
ng new frontend --routing=false --style=css --standalone --skip-git --defaults
cd frontend
npm install @ngx-translate/core @ngx-translate/http-loader
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init
```

- [ ] **Step 2: Configure Tailwind**

```js
// frontend/tailwind.config.js
module.exports = {
  content: ["./src/**/*.{html,ts}"],
  theme: { extend: {} },
  plugins: [],
}
```

Replace contents of `frontend/src/styles.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

- [ ] **Step 3: Environment files**

```typescript
// frontend/src/environments/environment.ts
export const environment = { apiUrl: 'http://localhost:8080' };
```

```typescript
// frontend/src/environments/environment.prod.ts
export const environment = { apiUrl: '' };
```

- [ ] **Step 4: App config**

```typescript
// frontend/src/app/app.config.ts
import { ApplicationConfig } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { provideTranslateService, TranslateLoader } from '@ngx-translate/core';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';
import { HttpClient } from '@angular/common/http';
import { routes } from './app.routes';

export function HttpLoaderFactory(http: HttpClient) {
  return new TranslateHttpLoader(http, '/assets/i18n/', '.json');
}

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideHttpClient(),
    provideTranslateService({
      loader: {
        provide: TranslateLoader,
        useFactory: HttpLoaderFactory,
        deps: [HttpClient],
      },
      defaultLanguage: 'en',
    }),
  ],
};
```

- [ ] **Step 5: Routes**

```typescript
// frontend/src/app/app.routes.ts
import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'admin/events',
    loadComponent: () => import('./admin/events/events-list.component')
      .then(m => m.EventsListComponent),
  },
  {
    path: 'admin/events/new',
    loadComponent: () => import('./admin/events/event-new.component')
      .then(m => m.EventNewComponent),
  },
  {
    path: 'admin/events/:id',
    loadComponent: () => import('./admin/events/event-detail.component')
      .then(m => m.EventDetailComponent),
  },
  {
    path: 'admin/djs',
    loadComponent: () => import('./admin/djs/djs.component')
      .then(m => m.DjsComponent),
  },
  {
    path: 'events/:id',
    loadComponent: () => import('./schedule/schedule.component')
      .then(m => m.ScheduleComponent),
  },
  { path: '', redirectTo: 'admin/events', pathMatch: 'full' },
];
```

- [ ] **Step 6: Root app component (nav + router outlet)**

```typescript
// frontend/src/app/app.component.ts
import { Component, OnInit } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, TranslateModule, CommonModule],
  template: `
    <nav class="h-14 bg-slate-900 flex items-center justify-between px-6 sticky top-0 z-50">
      <div class="flex items-center gap-8">
        <span class="text-white font-semibold text-lg">EventLineup</span>
        <a routerLink="/admin/events" routerLinkActive="border-b-2 border-indigo-400 text-white"
           class="text-slate-300 hover:text-white pb-1 transition-colors text-sm font-medium">
          {{ 'nav.events' | translate }}
        </a>
        <a routerLink="/admin/djs" routerLinkActive="border-b-2 border-indigo-400 text-white"
           class="text-slate-300 hover:text-white pb-1 transition-colors text-sm font-medium">
          {{ 'nav.djs' | translate }}
        </a>
      </div>
      <div class="flex gap-1">
        <button (click)="setLang('en')"
          class="px-3 py-1 rounded text-xs font-medium transition-colors"
          [class]="currentLang==='en' ? 'bg-white text-slate-900' : 'bg-slate-700 text-slate-300 hover:bg-slate-600'">
          EN
        </button>
        <button (click)="setLang('zh-TW')"
          class="px-3 py-1 rounded text-xs font-medium transition-colors"
          [class]="currentLang==='zh-TW' ? 'bg-white text-slate-900' : 'bg-slate-700 text-slate-300 hover:bg-slate-600'">
          中文
        </button>
      </div>
    </nav>
    <main class="min-h-[calc(100dvh-3.5rem)] bg-gray-50">
      <router-outlet />
    </main>
  `,
})
export class AppComponent implements OnInit {
  currentLang = 'en';
  constructor(private translate: TranslateService) {}
  ngOnInit() {
    const saved = localStorage.getItem('lang') || 'en';
    this.translate.use(saved);
    this.currentLang = saved;
  }
  setLang(lang: string) {
    this.translate.use(lang);
    this.currentLang = lang;
    localStorage.setItem('lang', lang);
  }
}
```

- [ ] **Step 7: i18n files**

```json
// frontend/src/assets/i18n/en.json
{
  "nav.events": "Events",
  "nav.djs": "DJs",
  "events.title": "Events",
  "events.new": "New Event",
  "events.upcoming": "Upcoming",
  "events.past": "Past",
  "events.noUpcoming": "No upcoming events",
  "events.noPast": "No past events",
  "events.view": "View",
  "events.duplicate": "Duplicate",
  "events.deleteConfirm": "Delete this event and all its stages and slots?",
  "eventNew.title": "New Event",
  "eventNew.name": "Event name",
  "eventNew.venue": "Venue name",
  "eventNew.startDate": "Start date",
  "eventNew.endDate": "End date (optional — defaults to start)",
  "eventNew.create": "Create Event",
  "eventDetail.shareViaLine": "Share via LINE",
  "eventDetail.viewPublic": "View Public Schedule",
  "stages.title": "Stages",
  "stages.name": "Stage name",
  "stages.color": "Color",
  "stages.add": "Add Stage",
  "schedule.title": "Schedule",
  "slots.stage": "Stage",
  "slots.date": "Date",
  "slots.start": "Start",
  "slots.end": "End",
  "slots.dj": "DJ",
  "slots.notes": "Notes",
  "slots.noSlots": "No slots yet.",
  "slots.unassigned": "— unassigned —",
  "djs.title": "DJs",
  "djs.name": "Name",
  "djs.genre": "Genre tags (comma-separated)",
  "djs.add": "Add DJ",
  "djs.noDJs": "No DJs yet. Add your first DJ above.",
  "schedule.share": "Share via LINE",
  "schedule.noSlots": "No slots scheduled yet.",
  "actions.add": "Add",
  "actions.delete": "Delete",
  "actions.cancel": "Cancel",
  "actions.save": "Save"
}
```

```json
// frontend/src/assets/i18n/zh-TW.json
{
  "nav.events": "活動",
  "nav.djs": "DJ",
  "events.title": "活動",
  "events.new": "新增活動",
  "events.upcoming": "即將舉行",
  "events.past": "過去活動",
  "events.noUpcoming": "目前沒有即將舉行的活動",
  "events.noPast": "沒有過去的活動",
  "events.view": "查看",
  "events.duplicate": "複製",
  "events.deleteConfirm": "確定要刪除此活動及所有舞台和時段嗎？",
  "eventNew.title": "新增活動",
  "eventNew.name": "活動名稱",
  "eventNew.venue": "場地名稱",
  "eventNew.startDate": "開始日期",
  "eventNew.endDate": "結束日期（選填，預設同開始日期）",
  "eventNew.create": "建立活動",
  "eventDetail.shareViaLine": "分享至 LINE",
  "eventDetail.viewPublic": "查看公開節目表",
  "stages.title": "舞台",
  "stages.name": "舞台名稱",
  "stages.color": "顏色",
  "stages.add": "新增舞台",
  "schedule.title": "時間表",
  "slots.stage": "舞台",
  "slots.date": "日期",
  "slots.start": "開始",
  "slots.end": "結束",
  "slots.dj": "DJ",
  "slots.notes": "備註",
  "slots.noSlots": "尚未新增時段。",
  "slots.unassigned": "— 未指定 —",
  "djs.title": "DJ 名單",
  "djs.name": "名稱",
  "djs.genre": "音樂類型（逗號分隔）",
  "djs.add": "新增 DJ",
  "djs.noDJs": "尚未新增 DJ，請在上方新增。",
  "schedule.share": "分享至 LINE",
  "schedule.noSlots": "尚未安排任何時段。",
  "actions.add": "新增",
  "actions.delete": "刪除",
  "actions.cancel": "取消",
  "actions.save": "儲存"
}
```

- [ ] **Step 8: API service**

```typescript
// frontend/src/app/services/api.service.ts
import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

export interface DJ {
  id: string;
  name: string;
  genre_tags: string[];
}

export interface Event {
  id: string;
  name: string;
  venue_name: string;
  start_date: string; // 'YYYY-MM-DD'
  end_date: string;   // 'YYYY-MM-DD'
}

export interface Stage {
  id: string;
  event_id: string;
  name: string;
  color: string;
  display_order: number;
}

export interface Slot {
  id: string;
  event_id: string;
  stage_id: string;
  stage_name: string;
  dj_id: string;
  dj_name: string;
  slot_date: string;   // 'YYYY-MM-DD'
  start_time: string;  // 'HH:MM'
  end_time: string;    // 'HH:MM'
  notes: string;
}

@Injectable({ providedIn: 'root' })
export class ApiService {
  private base = environment.apiUrl;

  constructor(private http: HttpClient) {}

  // DJs
  getDJs() { return this.http.get<DJ[]>(`${this.base}/api/djs`); }
  createDJ(d: Pick<DJ, 'name' | 'genre_tags'>) { return this.http.post<DJ>(`${this.base}/api/djs`, d); }
  deleteDJ(id: string) { return this.http.delete(`${this.base}/api/djs/${id}`); }

  // Events
  getEvents() { return this.http.get<Event[]>(`${this.base}/api/events`); }
  getEvent(id: string) { return this.http.get<Event>(`${this.base}/api/events/${id}`); }
  createEvent(e: Omit<Event, 'id'>) { return this.http.post<Event>(`${this.base}/api/events`, e); }
  deleteEvent(id: string) { return this.http.delete(`${this.base}/api/events/${id}`); }
  duplicateEvent(id: string) { return this.http.post<Event>(`${this.base}/api/events/${id}/duplicate`, {}); }

  // Stages
  getStages(eventId: string) { return this.http.get<Stage[]>(`${this.base}/api/events/${eventId}/stages`); }
  createStage(eventId: string, s: Pick<Stage, 'name' | 'color'>) {
    return this.http.post<Stage>(`${this.base}/api/events/${eventId}/stages`, s);
  }
  deleteStage(eventId: string, stageId: string) {
    return this.http.delete(`${this.base}/api/events/${eventId}/stages/${stageId}`);
  }

  // Slots
  getSlots(eventId: string) { return this.http.get<Slot[]>(`${this.base}/api/events/${eventId}/slots`); }
  createSlot(eventId: string, s: Pick<Slot, 'stage_id' | 'dj_id' | 'slot_date' | 'start_time' | 'end_time' | 'notes'>) {
    return this.http.post<Slot>(`${this.base}/api/events/${eventId}/slots`, s);
  }
  deleteSlot(eventId: string, slotId: string) {
    return this.http.delete(`${this.base}/api/events/${eventId}/slots/${slotId}`);
  }
}
```

- [ ] **Step 9: Verify build**

```bash
export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"
cd frontend && ng build 2>&1 | tail -5
```
Expected: `Build at: ... - Hash: ...` with no errors

- [ ] **Step 10: Commit**

```bash
git add frontend/
git commit -m "feat(frontend): Angular scaffold, Tailwind, i18n, routing, API service"
```

---

### Task 8: Events list + Create event pages

**Files:**
- Create: `frontend/src/app/admin/events/events-list.component.ts`
- Create: `frontend/src/app/admin/events/event-new.component.ts`

**Interfaces:**
- Consumes: `ApiService.getEvents()`, `ApiService.deleteEvent()`, `ApiService.duplicateEvent()`, `Event` interface
- Produces: `/admin/events` page with upcoming/past tabs
- Produces: `/admin/events/new` form that navigates to `/admin/events/:id` on success

---

- [ ] **Step 1: Write events-list.component.ts**

```typescript
// frontend/src/app/admin/events/events-list.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { ApiService, Event } from '../../services/api.service';

@Component({
  selector: 'app-events-list',
  standalone: true,
  imports: [CommonModule, RouterLink, TranslateModule],
  template: `
    <div class="max-w-4xl mx-auto px-6 py-8">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold text-gray-900">{{ 'events.title' | translate }}</h1>
        <a routerLink="/admin/events/new"
           class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-4 py-2 rounded-md text-sm transition-colors">
          + {{ 'events.new' | translate }}
        </a>
      </div>

      <!-- Tabs -->
      <div class="flex gap-4 border-b border-gray-200 mb-6">
        <button (click)="activeTab='upcoming'"
          class="pb-3 text-sm font-medium transition-colors"
          [class]="activeTab==='upcoming'
            ? 'border-b-2 border-indigo-600 text-indigo-600'
            : 'text-gray-500 hover:text-gray-700'">
          {{ 'events.upcoming' | translate }}
        </button>
        <button (click)="activeTab='past'"
          class="pb-3 text-sm font-medium transition-colors"
          [class]="activeTab==='past'
            ? 'border-b-2 border-indigo-600 text-indigo-600'
            : 'text-gray-500 hover:text-gray-700'">
          {{ 'events.past' | translate }}
        </button>
      </div>

      <!-- Event cards -->
      <div class="space-y-3">
        <ng-container *ngIf="activeTab==='upcoming'">
          <p *ngIf="upcoming.length===0" class="text-gray-500 text-sm py-8 text-center">
            {{ 'events.noUpcoming' | translate }}
          </p>
          <div *ngFor="let e of upcoming"
               class="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
            <div class="flex items-start justify-between">
              <div>
                <p class="font-semibold text-gray-900">{{ e.name }}</p>
                <p class="text-sm text-gray-500 mt-0.5">{{ e.venue_name }} · {{ formatDates(e) }}</p>
              </div>
            </div>
            <div class="flex items-center gap-3 mt-3">
              <a [routerLink]="['/admin/events', e.id]"
                 class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
                {{ 'events.view' | translate }}
              </a>
              <button (click)="duplicate(e.id)"
                class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
                {{ 'events.duplicate' | translate }}
              </button>
              <button (click)="delete(e.id)"
                class="text-red-500 hover:text-red-700 text-sm font-medium transition-colors ml-auto">
                {{ 'actions.delete' | translate }}
              </button>
            </div>
          </div>
        </ng-container>

        <ng-container *ngIf="activeTab==='past'">
          <p *ngIf="past.length===0" class="text-gray-500 text-sm py-8 text-center">
            {{ 'events.noPast' | translate }}
          </p>
          <div *ngFor="let e of past"
               class="bg-white border border-gray-200 rounded-lg p-4 shadow-sm opacity-75">
            <div class="flex items-start justify-between">
              <div>
                <p class="font-semibold text-gray-900">{{ e.name }}</p>
                <p class="text-sm text-gray-500 mt-0.5">{{ e.venue_name }} · {{ formatDates(e) }}</p>
              </div>
            </div>
            <div class="flex items-center gap-3 mt-3">
              <a [routerLink]="['/admin/events', e.id]"
                 class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
                {{ 'events.view' | translate }}
              </a>
              <button (click)="duplicate(e.id)"
                class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors">
                {{ 'events.duplicate' | translate }}
              </button>
            </div>
          </div>
        </ng-container>
      </div>
    </div>
  `,
})
export class EventsListComponent implements OnInit {
  events: Event[] = [];
  activeTab: 'upcoming' | 'past' = 'upcoming';
  today = new Date().toISOString().slice(0, 10);

  constructor(private api: ApiService) {}

  ngOnInit() { this.load(); }

  load() {
    this.api.getEvents().subscribe(events => this.events = events);
  }

  get upcoming() { return this.events.filter(e => e.end_date >= this.today); }
  get past() { return this.events.filter(e => e.end_date < this.today); }

  formatDates(e: Event): string {
    if (e.start_date === e.end_date) return e.start_date;
    return `${e.start_date} – ${e.end_date}`;
  }

  delete(id: string) {
    if (!confirm('Delete this event and all its stages and slots?')) return;
    this.api.deleteEvent(id).subscribe(() => this.load());
  }

  duplicate(id: string) {
    this.api.duplicateEvent(id).subscribe(() => this.load());
  }
}
```

- [ ] **Step 2: Write event-new.component.ts**

```typescript
// frontend/src/app/admin/events/event-new.component.ts
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-event-new',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslateModule],
  template: `
    <div class="max-w-lg mx-auto px-6 py-8">
      <a routerLink="/admin/events" class="text-sm text-indigo-600 hover:underline mb-6 inline-block">
        ← Back to Events
      </a>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-8">
        <h1 class="text-xl font-semibold text-gray-900 mb-6">{{ 'eventNew.title' | translate }}</h1>
        <form (ngSubmit)="submit()" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ 'eventNew.name' | translate }} *
            </label>
            <input [(ngModel)]="form.name" name="name" required
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ 'eventNew.venue' | translate }} *
            </label>
            <input [(ngModel)]="form.venue_name" name="venue" required
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ 'eventNew.startDate' | translate }} *
            </label>
            <input [(ngModel)]="form.start_date" name="start_date" type="date" required
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ 'eventNew.endDate' | translate }}
            </label>
            <input [(ngModel)]="form.end_date" name="end_date" type="date"
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div class="flex gap-3 justify-end pt-2">
            <a routerLink="/admin/events"
               class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-4 py-2 rounded-md text-sm transition-colors">
              {{ 'actions.cancel' | translate }}
            </a>
            <button type="submit"
              class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-4 py-2 rounded-md text-sm transition-colors">
              {{ 'eventNew.create' | translate }} →
            </button>
          </div>
        </form>
      </div>
    </div>
  `,
})
export class EventNewComponent {
  form = { name: '', venue_name: '', start_date: '', end_date: '' };

  constructor(private api: ApiService, private router: Router) {}

  submit() {
    if (!this.form.name || !this.form.venue_name || !this.form.start_date) return;
    const payload = {
      ...this.form,
      end_date: this.form.end_date || this.form.start_date,
    };
    this.api.createEvent(payload).subscribe(e => {
      this.router.navigate(['/admin/events', e.id]);
    });
  }
}
```

- [ ] **Step 3: Verify build**

```bash
export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"
cd frontend && ng build 2>&1 | tail -5
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/admin/events/events-list.component.ts \
        frontend/src/app/admin/events/event-new.component.ts
git commit -m "feat(frontend): events list and create event pages"
```

---

### Task 9: Event detail page (stages + slots)

**Files:**
- Create: `frontend/src/app/admin/events/event-detail.component.ts`

**Interfaces:**
- Consumes: `ApiService.getEvent()`, `getStages()`, `createStage()`, `deleteStage()`, `getSlots()`, `createSlot()`, `deleteSlot()`, `getDJs()`
- Consumes: `Event`, `Stage`, `Slot`, `DJ` interfaces

---

- [ ] **Step 1: Write event-detail.component.ts**

```typescript
// frontend/src/app/admin/events/event-detail.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { ApiService, Event, Stage, Slot, DJ } from '../../services/api.service';

@Component({
  selector: 'app-event-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslateModule],
  template: `
    <div class="max-w-6xl mx-auto px-6 py-8" *ngIf="event">
      <!-- Header -->
      <div class="mb-6">
        <a routerLink="/admin/events" class="text-sm text-indigo-600 hover:underline mb-2 inline-block">
          ← Events
        </a>
        <div class="flex items-start justify-between flex-wrap gap-3">
          <div>
            <h1 class="text-2xl font-semibold text-gray-900">{{ event.name }}</h1>
            <p class="text-sm text-gray-500 mt-1">{{ event.venue_name }} · {{ formatDates(event) }}</p>
          </div>
          <div class="flex gap-3">
            <button (click)="shareViaLine()"
              class="bg-green-500 hover:bg-green-600 text-white font-medium px-4 py-2 rounded-md text-sm transition-colors">
              {{ 'eventDetail.shareViaLine' | translate }}
            </button>
            <a [href]="publicUrl" target="_blank"
               class="border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-4 py-2 rounded-md text-sm transition-colors">
              {{ 'eventDetail.viewPublic' | translate }} →
            </a>
          </div>
        </div>
      </div>

      <!-- Two-column layout -->
      <div class="flex gap-6 flex-col md:flex-row">

        <!-- Stages panel -->
        <div class="w-full md:w-56 shrink-0">
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
            {{ 'stages.title' | translate }}
          </p>
          <ul class="space-y-2 mb-4">
            <li *ngFor="let s of stages"
                class="flex items-center gap-2 bg-white border border-gray-200 rounded-lg px-3 py-2">
              <span class="w-3 h-3 rounded-full shrink-0" [style.background]="s.color"></span>
              <span class="text-sm text-gray-800 flex-1 truncate">{{ s.name }}</span>
              <button (click)="deleteStage(s.id)"
                class="text-gray-400 hover:text-red-500 transition-colors text-sm ml-auto"
                aria-label="Delete stage">✕</button>
            </li>
          </ul>
          <form (ngSubmit)="addStage()" class="space-y-2">
            <input [(ngModel)]="newStage.name" name="stage_name"
              [placeholder]="'stages.name' | translate"
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            <div class="flex gap-2 items-center">
              <input [(ngModel)]="newStage.color" name="stage_color" type="color"
                class="h-9 w-9 rounded border border-gray-300 cursor-pointer" />
              <button type="submit"
                class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 rounded-md text-sm transition-colors">
                {{ 'stages.add' | translate }}
              </button>
            </div>
          </form>
        </div>

        <!-- Schedule panel -->
        <div class="flex-1 min-w-0">
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
            {{ 'schedule.title' | translate }}
          </p>

          <!-- Add slot form -->
          <form (ngSubmit)="addSlot()" class="bg-white border border-gray-200 rounded-lg p-3 mb-4">
            <div class="flex flex-wrap gap-2">
              <select [(ngModel)]="newSlot.stage_id" name="slot_stage"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">{{ 'slots.stage' | translate }}</option>
                <option *ngFor="let s of stages" [value]="s.id">{{ s.name }}</option>
              </select>
              <input [(ngModel)]="newSlot.slot_date" name="slot_date" type="date"
                [min]="event.start_date" [max]="event.end_date"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              <input [(ngModel)]="newSlot.start_time" name="slot_start" type="time"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              <input [(ngModel)]="newSlot.end_time" name="slot_end" type="time"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              <select [(ngModel)]="newSlot.dj_id" name="slot_dj"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">{{ 'slots.unassigned' | translate }}</option>
                <option *ngFor="let d of djs" [value]="d.id">{{ d.name }}</option>
              </select>
              <input [(ngModel)]="newSlot.notes" name="slot_notes"
                [placeholder]="'slots.notes' | translate"
                class="border border-gray-300 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 w-24" />
              <button type="submit"
                class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-3 py-2 rounded-md text-sm transition-colors"
                aria-label="Add slot">+</button>
            </div>
          </form>

          <!-- Slot list -->
          <div class="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <p *ngIf="slots.length===0"
               class="text-gray-400 text-sm text-center py-8">
              {{ 'slots.noSlots' | translate }}
            </p>
            <table *ngIf="slots.length>0" class="w-full text-sm">
              <thead class="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.date' | translate }}</th>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.stage' | translate }}</th>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.start' | translate }}–{{ 'slots.end' | translate }}</th>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.dj' | translate }}</th>
                  <th class="text-left px-4 py-2 text-xs font-medium text-gray-500">{{ 'slots.notes' | translate }}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100">
                <tr *ngFor="let s of slots" class="hover:bg-gray-50">
                  <td class="px-4 py-2 text-gray-700">{{ s.slot_date }}</td>
                  <td class="px-4 py-2">
                    <div class="flex items-center gap-1.5">
                      <span class="w-2.5 h-2.5 rounded-full" [style.background]="stageColor(s.stage_id)"></span>
                      <span class="text-gray-700">{{ s.stage_name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-2 text-gray-700">{{ s.start_time }}–{{ s.end_time }}</td>
                  <td class="px-4 py-2 text-gray-700">{{ s.dj_name || '—' }}</td>
                  <td class="px-4 py-2 text-gray-500">{{ s.notes }}</td>
                  <td class="px-4 py-2 text-right">
                    <button (click)="deleteSlot(s.id)"
                      class="text-gray-400 hover:text-red-500 transition-colors"
                      aria-label="Delete slot">✕</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class EventDetailComponent implements OnInit {
  event: Event | null = null;
  stages: Stage[] = [];
  slots: Slot[] = [];
  djs: DJ[] = [];
  newStage = { name: '', color: '#6366F1' };
  newSlot = { stage_id: '', dj_id: '', slot_date: '', start_time: '', end_time: '', notes: '' };

  private eventId = '';

  constructor(private api: ApiService, private route: ActivatedRoute) {}

  ngOnInit() {
    this.eventId = this.route.snapshot.paramMap.get('id')!;
    this.api.getEvent(this.eventId).subscribe(e => this.event = e);
    this.api.getDJs().subscribe(d => this.djs = d);
    this.loadStages();
    this.loadSlots();
  }

  loadStages() { this.api.getStages(this.eventId).subscribe(s => this.stages = s); }
  loadSlots() { this.api.getSlots(this.eventId).subscribe(s => this.slots = s); }

  stageColor(stageId: string): string {
    return this.stages.find(s => s.id === stageId)?.color ?? '#6366F1';
  }

  formatDates(e: Event): string {
    return e.start_date === e.end_date ? e.start_date : `${e.start_date} – ${e.end_date}`;
  }

  addStage() {
    if (!this.newStage.name.trim()) return;
    this.api.createStage(this.eventId, this.newStage).subscribe(() => {
      this.newStage = { name: '', color: '#6366F1' };
      this.loadStages();
    });
  }

  deleteStage(stageId: string) {
    this.api.deleteStage(this.eventId, stageId).subscribe(() => this.loadStages());
  }

  addSlot() {
    if (!this.newSlot.stage_id || !this.newSlot.slot_date || !this.newSlot.start_time || !this.newSlot.end_time) return;
    this.api.createSlot(this.eventId, this.newSlot).subscribe(() => {
      this.newSlot = { stage_id: '', dj_id: '', slot_date: '', start_time: '', end_time: '', notes: '' };
      this.loadSlots();
    });
  }

  deleteSlot(slotId: string) {
    this.api.deleteSlot(this.eventId, slotId).subscribe(() => this.loadSlots());
  }

  get publicUrl(): string {
    return `${window.location.origin}/events/${this.eventId}`;
  }

  shareViaLine() {
    window.open(`https://line.me/R/msg/text/?${encodeURIComponent(this.publicUrl)}`, '_blank');
  }
}
```

- [ ] **Step 2: Verify build**

```bash
export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"
cd frontend && ng build 2>&1 | tail -5
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/admin/events/event-detail.component.ts
git commit -m "feat(frontend): event detail page — stages + slot management"
```

---

### Task 10: DJ roster page

**Files:**
- Create: `frontend/src/app/admin/djs/djs.component.ts`

**Interfaces:**
- Consumes: `ApiService.getDJs()`, `createDJ()`, `deleteDJ()`, `DJ` interface

---

- [ ] **Step 1: Write djs.component.ts**

```typescript
// frontend/src/app/admin/djs/djs.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { ApiService, DJ } from '../../services/api.service';

@Component({
  selector: 'app-djs',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  template: `
    <div class="max-w-2xl mx-auto px-6 py-8">
      <h1 class="text-2xl font-semibold text-gray-900 mb-6">{{ 'djs.title' | translate }}</h1>

      <!-- Add DJ form -->
      <div class="bg-white border border-gray-200 rounded-lg p-4 shadow-sm mb-6">
        <form (ngSubmit)="add()" class="flex flex-wrap gap-3 items-end">
          <div class="flex-1 min-w-32">
            <label class="block text-xs font-medium text-gray-600 mb-1">{{ 'djs.name' | translate }} *</label>
            <input [(ngModel)]="newName" name="name" required
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          </div>
          <div class="flex-1 min-w-40">
            <label class="block text-xs font-medium text-gray-600 mb-1">{{ 'djs.genre' | translate }}</label>
            <input [(ngModel)]="newGenre" name="genre"
              placeholder="hip hop, r&b, house"
              class="border border-gray-300 rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          </div>
          <button type="submit"
            class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-4 py-2 rounded-md text-sm transition-colors h-[38px]">
            {{ 'djs.add' | translate }}
          </button>
        </form>
      </div>

      <!-- DJ list -->
      <p *ngIf="djs.length===0" class="text-gray-400 text-sm text-center py-8">
        {{ 'djs.noDJs' | translate }}
      </p>
      <ul *ngIf="djs.length>0" class="bg-white border border-gray-200 rounded-lg divide-y divide-gray-100 shadow-sm">
        <li *ngFor="let d of djs"
            class="flex items-center gap-3 px-4 py-3">
          <div class="flex-1 min-w-0">
            <span class="font-medium text-gray-900 text-sm">{{ d.name }}</span>
            <div *ngIf="d.genre_tags?.length" class="flex flex-wrap gap-1.5 mt-1">
              <span *ngFor="let tag of d.genre_tags"
                class="bg-indigo-50 text-indigo-700 text-xs font-medium px-2.5 py-0.5 rounded-full">
                {{ tag }}
              </span>
            </div>
          </div>
          <button (click)="delete(d.id)"
            class="text-gray-400 hover:text-red-500 transition-colors shrink-0"
            [attr.aria-label]="'actions.delete' | translate">✕</button>
        </li>
      </ul>
    </div>
  `,
})
export class DjsComponent implements OnInit {
  djs: DJ[] = [];
  newName = '';
  newGenre = '';

  constructor(private api: ApiService) {}

  ngOnInit() { this.load(); }

  load() { this.api.getDJs().subscribe(d => this.djs = d); }

  add() {
    if (!this.newName.trim()) return;
    const genre_tags = this.newGenre
      .split(',')
      .map(t => t.trim())
      .filter(t => t.length > 0);
    this.api.createDJ({ name: this.newName.trim(), genre_tags }).subscribe(() => {
      this.newName = '';
      this.newGenre = '';
      this.load();
    });
  }

  delete(id: string) {
    this.api.deleteDJ(id).subscribe(() => this.load());
  }
}
```

- [ ] **Step 2: Verify build**

```bash
export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"
cd frontend && ng build 2>&1 | tail -5
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/admin/djs/djs.component.ts
git commit -m "feat(frontend): DJ roster page"
```

---

### Task 11: Public schedule grid

**Files:**
- Create: `frontend/src/app/schedule/schedule.component.ts`

**Interfaces:**
- Consumes: `ApiService.getEvent()`, `getStages()`, `getSlots()`
- Produces: `/events/:id` — mobile-first read-only schedule grid, LINE share button, date tabs, language toggle

---

- [ ] **Step 1: Write schedule.component.ts**

```typescript
// frontend/src/app/schedule/schedule.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ApiService, Event, Stage, Slot } from '../services/api.service';

@Component({
  selector: 'app-schedule',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  template: `
    <div class="min-h-dvh bg-gray-50">
      <!-- Public header (no nav) -->
      <div class="bg-white border-b border-gray-200 px-4 py-4 sticky top-0 z-10">
        <div class="max-w-5xl mx-auto flex items-start justify-between gap-4">
          <div *ngIf="event">
            <h1 class="text-lg font-bold text-gray-900 leading-tight">{{ event.name }}</h1>
            <p class="text-sm text-gray-500 mt-0.5">{{ event.venue_name }}</p>
          </div>
          <div class="flex gap-1 shrink-0">
            <button (click)="setLang('en')"
              class="px-3 py-1 rounded text-xs font-medium transition-colors"
              [class]="currentLang==='en'
                ? 'bg-slate-900 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'">EN</button>
            <button (click)="setLang('zh-TW')"
              class="px-3 py-1 rounded text-xs font-medium transition-colors"
              [class]="currentLang==='zh-TW'
                ? 'bg-slate-900 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'">中文</button>
          </div>
        </div>
      </div>

      <div class="max-w-5xl mx-auto px-4 py-4">
        <!-- LINE share button -->
        <button (click)="shareViaLine()"
          class="w-full sm:w-auto bg-green-500 hover:bg-green-600 text-white font-medium px-6 py-3 rounded-lg text-sm transition-colors mb-4 flex items-center justify-center gap-2">
          {{ 'schedule.share' | translate }}
        </button>

        <!-- Date tabs -->
        <div class="flex gap-2 flex-wrap mb-4" *ngIf="dates.length > 1">
          <button *ngFor="let d of dates" (click)="selectedDate=d"
            class="px-4 py-2 rounded-full text-sm font-medium transition-colors"
            [class]="selectedDate===d
              ? 'bg-indigo-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'">
            {{ d }}
          </button>
        </div>

        <!-- No slots -->
        <p *ngIf="slotsForDate.length===0"
           class="text-gray-400 text-center py-12 text-sm">
          {{ 'schedule.noSlots' | translate }}
        </p>

        <!-- Mobile: grouped by stage -->
        <div *ngIf="slotsForDate.length>0" class="block lg:hidden space-y-4">
          <div *ngFor="let stage of stagesForDate"
               class="bg-white border border-gray-200 rounded-lg overflow-hidden shadow-sm">
            <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-100 bg-gray-50">
              <span class="w-3 h-3 rounded-full shrink-0" [style.background]="stageColor(stage.id)"></span>
              <span class="text-sm font-semibold text-gray-800">{{ stage.name }}</span>
            </div>
            <div class="divide-y divide-gray-100">
              <div *ngFor="let slot of slotsForStage(stage.id)"
                   class="flex items-center gap-3 px-4 py-3">
                <span class="text-xs text-gray-400 w-24 shrink-0">
                  {{ slot.start_time }}–{{ slot.end_time }}
                </span>
                <span class="text-sm font-medium text-gray-900">{{ slot.dj_name || '—' }}</span>
                <span *ngIf="slot.notes" class="text-xs text-gray-400 ml-auto">{{ slot.notes }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Desktop: time grid -->
        <div *ngIf="slotsForDate.length>0" class="hidden lg:block overflow-x-auto">
          <table class="w-full border-collapse text-sm">
            <thead>
              <tr>
                <th class="border border-gray-200 bg-gray-50 px-3 py-2 text-left w-36 text-xs font-medium text-gray-600">
                  Stage
                </th>
                <th *ngFor="let t of timeHeaders"
                    class="border border-gray-200 bg-gray-50 px-2 py-2 text-xs font-medium text-gray-500 w-20 text-center">
                  {{ t }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr *ngFor="let stage of stagesForDate">
                <td class="border border-gray-200 bg-gray-50 px-3 py-2">
                  <div class="flex items-center gap-2">
                    <span class="w-2.5 h-2.5 rounded-full shrink-0" [style.background]="stageColor(stage.id)"></span>
                    <span class="text-xs font-medium text-gray-800">{{ stage.name }}</span>
                  </div>
                </td>
                <td *ngFor="let t of timeHeaders"
                    class="border border-gray-200 px-1 py-1 min-w-[5rem] align-top">
                  <ng-container *ngFor="let slot of slotsAt(stage.id, t)">
                    <div class="rounded px-2 py-1 text-white text-xs font-medium leading-tight"
                         [style.background]="stageColor(stage.id)">
                      {{ slot.dj_name || '—' }}
                    </div>
                  </ng-container>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `,
})
export class ScheduleComponent implements OnInit {
  event: Event | null = null;
  stages: Stage[] = [];
  slots: Slot[] = [];
  dates: string[] = [];
  selectedDate = '';
  timeHeaders: string[] = [];
  currentLang = 'en';

  constructor(
    private api: ApiService,
    private route: ActivatedRoute,
    private translate: TranslateService,
  ) {}

  ngOnInit() {
    const saved = localStorage.getItem('lang') || 'en';
    this.translate.use(saved);
    this.currentLang = saved;

    const eventId = this.route.snapshot.paramMap.get('id')!;
    this.api.getEvent(eventId).subscribe(e => {
      this.event = e;
      this.dates = this.dateRange(e.start_date, e.end_date);
      if (this.dates.length > 0) this.selectedDate = this.dates[0];
    });
    this.api.getStages(eventId).subscribe(s => this.stages = s);
    this.api.getSlots(eventId).subscribe(s => {
      this.slots = s;
      this.buildTimeHeaders();
    });
  }

  get slotsForDate(): Slot[] {
    return this.slots.filter(s => s.slot_date === this.selectedDate);
  }

  get stagesForDate(): Stage[] {
    const ids = new Set(this.slotsForDate.map(s => s.stage_id));
    return this.stages.filter(s => ids.has(s.id));
  }

  slotsForStage(stageId: string): Slot[] {
    return this.slotsForDate
      .filter(s => s.stage_id === stageId)
      .sort((a, b) => a.start_time.localeCompare(b.start_time));
  }

  slotsAt(stageId: string, time: string): Slot[] {
    return this.slotsForDate.filter(s =>
      s.stage_id === stageId && s.start_time <= time && s.end_time > time);
  }

  stageColor(stageId: string): string {
    return this.stages.find(s => s.id === stageId)?.color ?? '#6366F1';
  }

  buildTimeHeaders() {
    if (this.slots.length === 0) {
      this.timeHeaders = this.generateHeaders('14:00', '23:00');
      return;
    }
    const starts = this.slots.map(s => s.start_time).sort();
    const ends = this.slots.map(s => s.end_time).sort();
    const earliest = starts[0];
    const latest = ends[ends.length - 1];
    // Start 30 min before first slot, end at last slot end
    const startMin = this.toMinutes(earliest) - 30;
    const endMin = this.toMinutes(latest);
    this.timeHeaders = this.generateHeadersFromMinutes(
      Math.max(startMin, 0), endMin);
  }

  private toMinutes(t: string): number {
    const [h, m] = t.split(':').map(Number);
    return h * 60 + m;
  }

  private generateHeaders(start: string, end: string): string[] {
    return this.generateHeadersFromMinutes(
      this.toMinutes(start), this.toMinutes(end));
  }

  private generateHeadersFromMinutes(startMin: number, endMin: number): string[] {
    const headers: string[] = [];
    // Round down to nearest 30
    let cur = Math.floor(startMin / 30) * 30;
    while (cur < endMin) {
      const h = Math.floor(cur / 60);
      const m = cur % 60;
      headers.push(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
      cur += 30;
    }
    return headers;
  }

  private dateRange(start: string, end: string): string[] {
    const dates: string[] = [];
    const cur = new Date(start);
    const last = new Date(end);
    while (cur <= last) {
      dates.push(cur.toISOString().slice(0, 10));
      cur.setDate(cur.getDate() + 1);
    }
    return dates;
  }

  setLang(lang: string) {
    this.translate.use(lang);
    this.currentLang = lang;
    localStorage.setItem('lang', lang);
  }

  shareViaLine() {
    const url = encodeURIComponent(window.location.href);
    window.open(`https://line.me/R/msg/text/?${url}`, '_blank');
  }
}
```

- [ ] **Step 2: Verify build**

```bash
export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"
cd frontend && ng build 2>&1 | tail -5
```
Expected: no errors

- [ ] **Step 3: Run dev server and spot-check**

```bash
export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"
cd frontend && ng serve
```
Open `http://localhost:4200` — verify:
- Nav shows Events + DJs links
- Language toggle switches between EN and 中文
- `/admin/events` loads with upcoming/past tabs
- `/admin/events/new` form creates an event and redirects to detail
- Event detail shows stages + slots panels
- `/admin/djs` shows DJ list with genre chips

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/schedule/schedule.component.ts
git commit -m "feat(frontend): public schedule grid with LINE share"
```

---

### Task 12: Netlify config + production environment

**Files:**
- Create: `frontend/netlify.toml`
- Modify: `frontend/src/environments/environment.prod.ts`

**Interfaces:**
- Produces: Netlify build succeeds via `ng build`
- Produces: SPA redirect so `/events/:id` and `/admin/*` routes work on refresh

---

- [ ] **Step 1: Write netlify.toml**

```toml
# frontend/netlify.toml
[build]
  base    = "frontend"
  command = "npm run build"
  publish = "dist/frontend/browser"

[[redirects]]
  from   = "/*"
  to     = "/index.html"
  status = 200
```

- [ ] **Step 2: Update environment.prod.ts with Railway URL placeholder**

```typescript
// frontend/src/environments/environment.prod.ts
// Replace YOUR_RAILWAY_URL after Railway deploy
export const environment = {
  apiUrl: 'https://YOUR_RAILWAY_URL.up.railway.app',
};
```

- [ ] **Step 3: Final build check**

```bash
export PATH="$HOME/.nvm/versions/node/v24.18.0/bin:$PATH"
cd frontend && ng build --configuration production 2>&1 | tail -5
```
Expected: no errors, outputs to `dist/frontend/browser/`

- [ ] **Step 4: Commit**

```bash
git add frontend/netlify.toml frontend/src/environments/environment.prod.ts
git commit -m "feat(deploy): Netlify config + production environment"
```

- [ ] **Step 5: Deploy backend to Railway (manual)**

1. Push repo to GitHub
2. Railway → New Project → Deploy from GitHub → root: `backend/`
3. Set env vars: `DATABASE_URL` (Neon connection string), `FRONTEND_URL` (Netlify URL — set after step 6)
4. Note the Railway URL (e.g. `https://eventlineup-api.up.railway.app`)

- [ ] **Step 6: Update environment.prod.ts with real Railway URL**

```typescript
export const environment = {
  apiUrl: 'https://eventlineup-api.up.railway.app', // your actual URL
};
```

- [ ] **Step 7: Deploy frontend to Netlify (manual)**

1. Netlify → Add new site → Import from GitHub
2. Base directory: `frontend`; build command: `npm run build`; publish: `dist/frontend/browser`
3. Push → Netlify auto-deploys
4. Update `FRONTEND_URL` in Railway to the Netlify URL → redeploy Railway

- [ ] **Step 8: Smoke test production**

```
1. Open Netlify URL → /admin/djs → add "BCXuan" with tags "hip hop, r&b"
2. /admin/events → New Event → "2026 混東區 潮FUN 夏日趴", "B 微醺廣場", 2026-07-25 → 2026-07-26
3. Event detail → add stages: "A 主舞台", "B 微醺", "C 頂好廣場"
4. Add slots matching the original spreadsheet
5. Click "View Public Schedule" → verify grid shows slots
6. Click "Share via LINE" → verify LINE opens with URL
7. Switch language to 中文 → verify labels change
```

- [ ] **Step 9: Final tag**

```bash
git tag v0.1.0
git push origin feature/eventlineup-mvp
git push --tags
```

---

## Done

At completion:
- Organiser adds DJs once to the global roster at `/admin/djs`
- Organiser creates events at `/admin/events` with stages and slots
- Past events archived automatically (end_date < today)
- Public schedule at `/events/:id` — shareable via LINE, no login needed
- EN / 中文 language toggle on all pages
- Docker Compose for local dev; Neon Postgres + Railway for production
