# DJ Scheduler — Backend

Go REST API for the EventLineup DJ scheduling app. Built with [Gin](https://github.com/gin-gonic/gin) and PostgreSQL via [pgx](https://github.com/jackc/pgx).

---

## Architecture

The backend follows [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html). Dependencies only point **inward** — outer layers know about inner layers, never the reverse.

```
cmd/app/main.go                  Entry point — calls app.Run()
internal/
├── domain/
│   ├── model/                   Entities: DJ, Event, Stage, Slot structs
│   └── repository/              Repository interfaces (ports)
├── usecase/
│   ├── dj/
│   ├── event/
│   ├── stage/
│   └── slot/                    Application logic — orchestrates repo calls
├── interfaces/
│   └── http/                    Gin HTTP handlers + router — depend on use cases
├── infrastructure/
│   ├── config/                  Env var loading (DATABASE_URL, PORT, FRONTEND_URL)
│   └── database/                pgx repository implementations — all SQL lives here
└── app/
    └── app.go                   Bootstrap: wires config → DB → repos → use cases → handlers
```

### How a request flows

```
HTTP request
  → internal/interfaces/http   (handler parses request, calls use case)
  → internal/usecase/{feature} (business logic, calls repository interface)
  → internal/infrastructure/database (executes SQL, returns domain model)
  → handler serialises response as JSON
```

The handler never touches SQL. The repository never knows about HTTP. The domain models (`internal/domain/model`) have no dependencies on anything else.

### Database schema

Four tables, all using UUID primary keys:

| Table    | Description                                    |
|----------|------------------------------------------------|
| `djs`    | DJ roster with name and genre tags             |
| `events` | Events (festival, club night, etc.) with venue and date range |
| `stages` | Stages belonging to an event, with display order and colour |
| `slots`  | Time slots on a stage, optionally assigned to a DJ |

Schema is applied automatically by Docker Compose on first run via `migrations/001_init.sql`.

---

## Running locally

### Option 1 — Docker Compose (easiest)

From the **repo root**:

```bash
docker compose up --build
```

This starts both Postgres and the API. The API is available at `http://localhost:8080`.

### Option 2 — Run the API directly

You need a running Postgres instance first (e.g. via `docker compose up db`).

```bash
# From the backend/ directory
export DATABASE_URL="postgresql://eventlineup:eventlineup@localhost:5432/eventlineup"
export PORT=8080
export FRONTEND_URL="http://localhost:4200"

go run ./cmd/app
```

---

## Environment variables

| Variable       | Required | Default                  | Description                        |
|----------------|----------|--------------------------|------------------------------------|
| `DATABASE_URL` | Yes      | —                        | PostgreSQL connection string        |
| `PORT`         | No       | `8080`                   | Port the API listens on            |
| `FRONTEND_URL` | No       | `http://localhost:4200`  | Allowed CORS origin                |
| `COOKIE_SECURE`| No       | `false`                  | Set to `true` in production so the OAuth state cookie is HTTPS-only |

---

## API routes

All routes are prefixed with `/api`.

| Method | Path                              | Description                  |
|--------|-----------------------------------|------------------------------|
| GET    | `/api/djs`                        | List all DJs                 |
| POST   | `/api/djs`                        | Create a DJ                  |
| DELETE | `/api/djs/:id`                    | Delete a DJ                  |
| GET    | `/api/events`                     | List all events               |
| POST   | `/api/events`                     | Create an event               |
| GET    | `/api/events/:id`                 | Get a single event            |
| DELETE | `/api/events/:id`                 | Delete an event               |
| POST   | `/api/events/:id/duplicate`       | Duplicate an event (+ stages) |
| GET    | `/api/events/:id/stages`          | List stages for an event      |
| POST   | `/api/events/:id/stages`          | Create a stage                |
| DELETE | `/api/events/:id/stages/:stage_id`| Delete a stage                |
| GET    | `/api/events/:id/slots`           | List slots for an event       |
| POST   | `/api/events/:id/slots`           | Create a slot                 |
| DELETE | `/api/events/:id/slots/:slot_id`  | Delete a slot                 |

---

## Running tests

Tests are integration tests and require a real Postgres database.

```bash
export DATABASE_URL="postgresql://eventlineup:eventlineup@localhost:5432/eventlineup"
go test ./...
```

Tests live alongside the code they test in `internal/interfaces/http/`.

---

## Building for production

```bash
go build -o server ./cmd/app
./server
```

Deployed via the `Dockerfile` in this directory. Railway config is in `railway.toml`.
