# EventLineup DJ Scheduler

A full-stack application for managing event schedules, DJs, and stages. Features an Angular frontend with signals-based state management and a Go backend with PostgreSQL.

## 🚀 Quick Start

### Prerequisites

**Backend (Go):**
- Go 1.26.4+
- PostgreSQL 13+
- Docker & Docker Compose (optional)

**Frontend (Angular):**
- Node.js 18+
- npm 10+

### Local Development Setup

#### 1. Clone the Repository

```bash
git clone git@github.com:abefong54/dj-scheduler.git
cd dj-scheduler
```

#### 2. Start PostgreSQL (via Docker Compose)

```bash
docker-compose up -d
```

This starts PostgreSQL on localhost:5432 with default credentials:
- **User:** postgres
- **Password:** postgres
- **Database:** eventlineup

#### 3. Run the Backend (Go API)

```bash
cd backend

# Set environment variables
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/eventlineup?sslmode=disable"
export FRONTEND_URL="http://localhost:4200"
export PORT="8080"

# Run the server
go run main.go

# Server runs on http://localhost:8080
```

**API Endpoints:**
- `GET /api/events` — List all events
- `GET /api/events/:id` — Get event details
- `POST /api/events` — Create new event
- `GET /api/events/:id/stages` — List stages for event
- `POST /api/events/:id/stages` — Create stage
- `GET /api/events/:id/slots` — List slots for event
- `POST /api/events/:id/slots` — Create slot
- `GET /api/djs` — List all DJs
- `POST /api/djs` — Create new DJ

#### 4. Run the Frontend (Angular)

In a new terminal:

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm start

# or use ng directly
ng serve --open
```

Frontend runs on `http://localhost:4200`

### Full Project Structure

```
dj-scheduler/
├── backend/                          # Go API
│   ├── cmd/
│   ├── internal/
│   ├── main.go
│   ├── go.mod
│   └── README.md                     # Backend-specific docs
│
├── frontend/                         # Angular UI
│   ├── src/
│   │   ├── app/
│   │   │   ├── services/
│   │   │   │   ├── api.service.ts
│   │   │   │   └── language.service.ts
│   │   │   ├── admin/                # Admin features
│   │   │   ├── schedule/             # Public schedule view
│   │   │   ├── app.component.ts
│   │   │   └── app.routes.ts
│   │   └── assets/i18n/              # Translations (en, zh-TW)
│   ├── angular.json
│   ├── package.json
│   └── README.md                     # Frontend-specific docs
│
├── docker-compose.yml                # PostgreSQL + setup
└── README.md                         # This file
```

## 🏗️ Architecture

### Backend (Go)
- **Framework:** Gin HTTP router
- **Database:** PostgreSQL with pgxpool connection pooling
- **CORS:** Configured for frontend origin
- **Models:** DJ, Event, Stage, Slot

### Frontend (Angular 18+)
- **State Management:** Angular Signals for reactive state
- **Components:** 6 standalone components with signals-based architecture
- **Styling:** Tailwind CSS v4 with responsive design
- **i18n:** ngx-translate with English and Traditional Chinese
- **Routing:** Lazy-loaded components per feature

## 📋 Features

### Admin Section (`/admin`)
- **Events** — Create, view, duplicate, delete events
- **Stages** — Manage stages within events
- **Slots** — Schedule DJs into time slots
- **DJs** — Manage DJ roster with genre tags

### Public Section
- **Schedule Grid** (`/schedule/:eventId`)
  - Desktop: Time grid view (stages × time slots)
  - Mobile: Grouped by stage view
  - Language toggle (EN/中文)
  - LINE share button

## 🔄 Data Flow

1. **Frontend** → Calls API service methods (returns Observables)
2. **API Service** → HTTP requests to backend
3. **Backend** → Queries PostgreSQL, returns JSON
4. **Frontend** → Wraps Observable with signals via `.subscribe()` pattern
5. **Signals** → Auto-update computed values and re-render templates

Example (EventsListComponent):
```typescript
constructor() {
  this.api.getEvents().subscribe(events => {
    this.events.set(events);  // Update signal
  });
}

upcoming = computed(() =>        // Auto-recalculates when events changes
  this.events().filter(e => e.end_date >= today)
);
```

## 🧪 Testing

### Frontend Unit Tests

```bash
cd frontend
npm test
```

Note: Test files (.spec.ts) not yet created. Consider adding unit tests for critical services.

### Manual Testing

1. **Create Event:**
   - Navigate to `/admin/events`
   - Click "+ New Event"
   - Fill in name, venue, start/end dates
   - Submit

2. **Add Stages:**
   - View event details
   - Click "+ New Stage"
   - Enter stage name and color

3. **Schedule DJs:**
   - Click "+ New Slot"
   - Select DJ, date, time range, stage

4. **View Public Schedule:**
   - Share event ID: `/schedule/{eventId}`
   - View in desktop (grid) or mobile (grouped) layout

## 📦 Environment Configuration

### Development

**Frontend (frontend/.env or via npm start):**
```
API_URL=http://localhost:8080
```

**Backend (.env file or shell export):**
```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/eventlineup?sslmode=disable
FRONTEND_URL=http://localhost:4200
PORT=8080
```

### Production

**Frontend (via Netlify):**
- Environment variable: `VITE_API_URL` — set to production API URL
- Build: `npm run build` → `dist/frontend/`
- Deploy: Netlify from GitHub

**Backend (Railway or Docker):**
- Environment variables: `DATABASE_URL`, `FRONTEND_URL`, `PORT`
- Docker: Build with `backend/Dockerfile`

## 🚢 Deployment

### Frontend (Netlify)

1. Push to GitHub
2. Connect repo to Netlify
3. Set environment variable `VITE_API_URL` to production API
4. Deploy automatically on push to main

### Backend (Railway or Docker Hub)

1. Build Docker image: `docker build -f backend/Dockerfile -t dj-scheduler-api .`
2. Push to registry
3. Deploy with environment variables set

## 🛠️ Development Tips

### Adding a New Component

```bash
cd frontend
ng generate component admin/my-feature
```

Then use signals for state:
```typescript
import { signal, computed, inject } from '@angular/core';

export class MyFeatureComponent {
  myState = signal('value');
  derived = computed(() => this.myState().toUpperCase());
}
```

### Adding a Translation

1. Add to `frontend/src/assets/i18n/en.json`:
```json
{
  "myFeature.title": "My Feature Title"
}
```

2. Add to `frontend/src/assets/i18n/zh-TW.json`:
```json
{
  "myFeature.title": "我的功能標題"
}
```

3. Use in template:
```html
<h1>{{ 'myFeature.title' | translate }}</h1>
```

## 📚 Documentation

- **[Backend README](backend/README.md)** — Go API architecture, handlers, database schema
- **[Frontend README](frontend/README.md)** — Angular components, state management, styling

## ❓ Troubleshooting

### "Cannot connect to database"
- Ensure PostgreSQL is running: `docker-compose ps`
- Check DATABASE_URL is correct
- Verify credentials in docker-compose.yml

### "Frontend cannot reach API"
- Ensure backend is running on port 8080
- Check CORS is configured correctly in backend
- Verify API_URL in frontend env matches backend host

### "Translations not loading"
- Check files exist in `frontend/src/assets/i18n/`
- Verify `app.config.ts` has correct i18n provider setup
- Check browser console for 404 errors

### Build failures
- Run `npm install` to ensure dependencies
- Clear cache: `rm -rf frontend/dist node_modules`
- Rebuild: `npm install && npm run build`

## 📄 License

Part of EventLineup project.

## 🔗 Useful Links

- [Angular 18 Guide](https://angular.dev)
- [Angular Signals](https://angular.dev/guide/signals)
- [Tailwind CSS](https://tailwindcss.com)
- [PostgreSQL Docs](https://www.postgresql.org/docs)
- [Go Documentation](https://golang.org/doc)
