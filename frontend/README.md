# Soundcheck Frontend

A modern Angular application for managing event schedules, DJs, and stages using Angular Signals for reactive state management.

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ and npm 10+
- Angular CLI 18+ (installed globally or via npx)

### Installation & Development

1. **Install dependencies:**
   ```bash
   cd frontend
   npm install
   ```

2. **Start the dev server:**
   ```bash
   npm start
   # or
   ng serve --open
   ```

   The app opens at `http://localhost:4200` and auto-reloads on changes.

3. **Build for production:**
   ```bash
   npm run build
   # or
   ng build --configuration production
   ```

4. **Run tests:**
   ```bash
   npm test
   # or
   ng test
   ```

## 📂 Repository Structure

```
frontend/
├── src/
│   ├── app/
│   │   ├── services/
│   │   │   ├── api.service.ts          # HTTP API calls to backend
│   │   │   └── language.service.ts     # Language/i18n state management
│   │   │
│   │   ├── admin/                      # Admin features (events, DJs)
│   │   │   ├── events/
│   │   │   │   ├── events-list.component.ts      # List all events
│   │   │   │   ├── event-detail.component.ts     # View/manage single event
│   │   │   │   └── event-new.component.ts        # Create new event
│   │   │   │
│   │   │   └── djs/
│   │   │       └── djs.component.ts              # Manage DJs
│   │   │
│   │   ├── schedule/
│   │   │   └── schedule.component.ts   # Public event schedule view
│   │   │
│   │   ├── app.component.ts            # Root component (navigation + outlet)
│   │   ├── app.routes.ts               # Route definitions
│   │   ├── app.config.ts               # App providers (services, translations)
│   │   └── main.ts                     # Bootstrap entry point
│   │
│   ├── environments/
│   │   ├── environment.ts              # Dev environment config
│   │   └── environment.prod.ts         # Production environment config
│   │
│   ├── assets/
│   │   └── i18n/
│   │       ├── en.json                 # English translations
│   │       └── zh-TW.json              # Traditional Chinese translations
│   │
│   └── index.html                      # HTML entry point
│
├── package.json
├── angular.json                        # Angular CLI configuration
├── tsconfig.json                       # TypeScript configuration
├── tailwind.config.js                  # Tailwind CSS configuration
├── netlify.toml                        # Netlify deployment config
└── README.md                           # This file
```

## 🏗️ Architecture & Patterns

### State Management: Angular Signals

All components use **Angular Signals** for state management:

- **`signal(value)`** — Create writable state
- **`computed(() => ...)`** — Derive state from signals (auto-memoized)
- **`effect(() => ...)`** — Run side effects when signals change
- **`resource()`** — Async data loading with built-in lifecycle management

**Example:**

```typescript
export class EventsListComponent {
  private api = inject(ApiService);

  events = signal<Event[]>([]);
  activeTab = signal<'upcoming' | 'past'>('upcoming');

  upcoming = computed(() =>
    this.events().filter(e => e.end_date >= today)
  );

  constructor() {
    this.api.getEvents().subscribe(events => {
      this.events.set(events);
    });
  }
}
```

### Services

#### **ApiService** (`services/api.service.ts`)

Wraps HTTP calls to the backend. Returns RxJS Observables (components wrap these with signals as needed).

**Methods:**
- `getEvents()`, `getEvent(id)`, `createEvent()`, `deleteEvent()`, `duplicateEvent()`
- `getDJs()`, `createDJ()`, `deleteDJ()`
- `getStages(eventId)`, `createStage()`, `deleteStage()`
- `getSlots(eventId)`, `createSlot()`, `deleteSlot()`

#### **LanguageService** (`services/language.service.ts`)

Manages app language state and i18n switching. Automatically syncs with `localStorage` and ngx-translate.

**Interface:**
```typescript
currentLang: WritableSignal<string>   // Current language ('en', 'zh-TW')
setLanguage(lang: string): void       // Change language
```

**Usage:**
```typescript
export class MyComponent {
  langService = inject(LanguageService);

  // In template:
  // {{ langService.currentLang() }}
  // (click)="langService.setLanguage('en')"
}
```

### Components

#### **Admin Section** (`admin/`)

- **EventsListComponent** — List all events, filter by upcoming/past, create/duplicate/delete
- **EventDetailComponent** — Manage stages and slots for an event
- **EventNewComponent** — Form to create a new event
- **DjsComponent** — Manage DJ database with genres

#### **Public Section** (`schedule/`)

- **ScheduleComponent** — Display event schedule in grid (desktop) or grouped (mobile)
  - URL: `/schedule/:id`
  - Shows stages as rows, time slots as columns
  - Mobile view groups by stage
  - Share via LINE button

#### **Root** (`app.component.ts`)

- Navigation bar with links to admin sections
- Language toggle (EN / 中文)
- Router outlet for child components

### Routing

Defined in `app.routes.ts`:

```typescript
export const routes: Routes = [
  { path: 'admin/events', component: EventsListComponent },
  { path: 'admin/events/new', component: EventNewComponent },
  { path: 'admin/events/:id', component: EventDetailComponent },
  { path: 'admin/djs', component: DjsComponent },
  { path: 'schedule/:id', component: ScheduleComponent },
  { path: '', redirectTo: '/admin/events', pathMatch: 'full' },
];
```

### Styling

- **Tailwind CSS v4** — All styling done with utility classes
- **Component-scoped** — Each component manages its own styles (no CSS files)
- **Responsive** — Mobile-first design with `sm:`, `lg:` breakpoints
- **Dark mode ready** — Tailwind classes support dark mode

### Translations (i18n)

- **ngx-translate** library
- Files in `src/assets/i18n/` (en.json, zh-TW.json)
- Usage in templates: `{{ 'key.path' | translate }}`
- Usage in components: `this.translate.instant('key.path')`
- Language persists in `localStorage` with key `'lang'`

## 🔄 Data Flow

1. **Component initializes** → Injects `ApiService` and calls API method
2. **API returns Observable** → Component subscribes and sets signal
3. **Signal updates** → Computed values recalculate automatically
4. **Template re-renders** → Angular detects signal changes, updates DOM
5. **User interacts** → Calls component method (e.g., `deleteDJ()`)
6. **Component calls API** → Reloads data and updates signals

**Example (EventsListComponent):**

```typescript
constructor() {
  // 1. Call API
  this.api.getEvents().subscribe(events => {
    // 2. Update signal
    this.events.set(events);
  });
}

// 3. Computed value auto-updates
upcoming = computed(() =>
  this.events().filter(e => e.end_date >= today)
);

// 4. Template renders {{ upcoming() }}
// 5. User clicks delete
delete(id: string) {
  this.api.deleteEvent(id).subscribe(() => {
    // 6. Reload data
    this.loadEvents();
  });
}
```

## 🧪 Testing

Tests use **Vitest** and **@angular/core/testing** utilities.

Run tests:
```bash
npm test
```

Example test:
```typescript
describe('EventsListComponent', () => {
  it('should filter upcoming events', () => {
    const component = new EventsListComponent(apiService);
    component.events.set([
      { id: '1', end_date: '2099-12-31' },  // upcoming
      { id: '2', end_date: '2020-01-01' },  // past
    ]);
    expect(component.upcoming().length).toBe(1);
  });
});
```

## 🌐 Environment Configuration

- **Dev:** Uses `environment.ts` (API URL typically `http://localhost:8080`)
- **Production:** Uses `environment.prod.ts` (API URL from Netlify env var `VITE_API_URL`)

Set env vars in `netlify.toml` or Netlify dashboard:
```toml
[context.production]
  environment = { VITE_API_URL = "https://api.example.com" }
```

## 📦 Key Dependencies

- **Angular 18+** — Framework
- **ngx-translate** — i18n
- **Tailwind CSS 4** — Styling
- **RxJS** — Reactive programming
- **TypeScript 5+** — Language

## 🚀 Deployment

### Netlify

1. **Build:** `npm run build`
2. **Output:** `dist/`
3. **Environment variables:** Set `VITE_API_URL` in Netlify dashboard
4. **Redirects:** `netlify.toml` handles SPA routing

### Local Production Build

```bash
npm run build
npx http-server dist/eventlineup/browser
# Open http://localhost:8080
```

## 💡 Development Tips

### Adding a New Component

1. Generate scaffold:
   ```bash
   ng generate component admin/events/event-edit
   ```

2. Use signals for state:
   ```typescript
   import { signal, computed, inject } from '@angular/core';
   
   export class MyComponent {
     myState = signal('value');
     derived = computed(() => this.myState().toUpperCase());
   }
   ```

3. Inject services:
   ```typescript
   private api = inject(ApiService);
   private lang = inject(LanguageService);
   ```

4. Call API and update signals:
   ```typescript
   this.api.getEvents().subscribe(events => {
     this.events.set(events);
   });
   ```

### Adding a Translation

1. Add key to `src/assets/i18n/en.json`:
   ```json
   {
     "myFeature.title": "My Feature Title"
   }
   ```

2. Add to `src/assets/i18n/zh-TW.json`:
   ```json
   {
     "myFeature.title": "我的功能標題"
   }
   ```

3. Use in template:
   ```html
   <h1>{{ 'myFeature.title' | translate }}</h1>
   ```

### Debugging Signals

```typescript
// Log signal value
effect(() => {
  console.log('Events changed:', this.events());
});

// Log computed value
effect(() => {
  console.log('Upcoming count:', this.upcoming().length);
});
```

## 🐛 Common Issues

**"Cannot find module" errors after refactoring:**
- Run `npm install` to ensure all dependencies are installed
- Restart dev server: `Ctrl+C` then `npm start`

**Translations not loading:**
- Check file exists at `src/assets/i18n/en.json`
- Verify `app.config.ts` has correct provider setup
- Check browser console for HTTP 404 errors

**Signals not updating UI:**
- Make sure to call `.set()` or `.update()` on the signal
- Templates must call signal as function: `{{ mySignal() }}` not `{{ mySignal }}`

## 📚 Resources

- [Angular Signals Guide](https://angular.io/guide/signals)
- [Tailwind CSS Docs](https://tailwindcss.com)
- [ngx-translate Docs](https://github.com/ngx-translate/core)
- [Angular Routing](https://angular.dev/guide/routing)

## 📝 License

Part of the Soundcheck project.
