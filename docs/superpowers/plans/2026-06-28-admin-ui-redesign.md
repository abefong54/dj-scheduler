# Admin UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the EventLineup admin with an artistic visual identity, a reusable `DataTableComponent` with search/sort used across all tables, an inline quick-add row replacing the slot modal, and client-side Excel export.

**Architecture:** Six sequential tasks, each independently testable. A shared `DataTableComponent` (plus `ColumnDefDirective` for custom cells) provides uniform search/sort for flat tables. The event detail page gets a hero header, compact stage chip strip, date tabs, an inline add row with smart pre-fills, and an Export .xlsx button. Excel is generated client-side using the `xlsx` npm package.

**Tech Stack:** Angular 18+ standalone components, Angular Signals, Tailwind CSS v4, ngx-translate, xlsx (npm)

## Global Constraints

- Angular standalone components only — no NgModule
- Signals for all reactive state — no RxJS BehaviorSubject
- Tailwind utility classes via `@layer components` in `styles.css` — no inline style objects except dynamic `[style.background]` color bindings
- All user-visible strings use `translate` pipe or `TranslateService.instant()`; add new keys to both `en.json` and `zh-TW.json`
- Primary accent color: `violet-600` (`#7c3aed`) — replaces amber throughout
- No backend changes in this plan

---

## File Map

**New:**
- `frontend/src/app/shared/column-def.directive.ts`
- `frontend/src/app/shared/data-table.component.ts`
- `frontend/src/app/shared/data-table.component.html`
- `frontend/src/app/shared/data-table.component.css`
- `frontend/src/app/shared/data-table.component.spec.ts`

**Modified:**
- `frontend/src/styles.css` — violet accent, updated button/focus/back-link/genre-pill/tab classes
- `frontend/src/app/app.component.css` — charcoal navbar background
- `frontend/src/assets/i18n/en.json` — add `table.*`, `export.*`, `eventDetail.export` keys
- `frontend/src/assets/i18n/zh-TW.json` — same keys in Chinese
- `frontend/src/app/admin/djs/djs.component.ts` — add `djColumns`, `djsAsRows`, helpers; import `DataTableComponent`
- `frontend/src/app/admin/djs/djs.component.html` — replace manual table with `<app-data-table>`
- `frontend/src/app/admin/djs/djs.component.css` — remove `.djs-table-wrap`; update `.genre-tag` to violet
- `frontend/src/app/admin/events/event-detail.component.ts` — add `selectedDate`, `dates`, `slotsForSelectedDate`, sort/search/add-row signals, `defaultStartTime`, `defaultStageId`, `exportXlsx`; inject `LanguageService`; remove modal state
- `frontend/src/app/admin/events/event-detail.component.html` — hero header, stage chips, date tabs, search/sort headers, inline add row; remove Add Slot modal
- `frontend/src/app/admin/events/event-detail.component.css` — hero, stage chip, date tab, add row, violet edit row
- `frontend/package.json` — add `xlsx` dependency

**Not touched:** `slot-new.component.*` (its route stays as fallback; can be removed in a follow-up PR)

---

## Task 1: i18n Keys + Visual Identity

**Files:**
- Modify: `frontend/src/assets/i18n/en.json`
- Modify: `frontend/src/assets/i18n/zh-TW.json`
- Modify: `frontend/src/styles.css`
- Modify: `frontend/src/app/app.component.css`
- Modify: `frontend/src/app/admin/events/event-detail.component.ts` (stageColors + default only)

**Interfaces:**
- Produces: updated CSS utility classes and i18n keys consumed by all later tasks

- [ ] **Step 1: Add new i18n keys to `frontend/src/assets/i18n/en.json`**

After the `"eventDetail.viewPublic"` line, add:
```json
  "table.search": "Search...",
  "table.empty": "No results.",
  "eventDetail.export": "Export .xlsx",
  "export.timeSlot": "Time Slot",
  "export.genre": "Genre"
```

- [ ] **Step 2: Add same keys to `frontend/src/assets/i18n/zh-TW.json`**

After the `"eventDetail.viewPublic"` line, add:
```json
  "table.search": "搜尋...",
  "table.empty": "無結果。",
  "eventDetail.export": "匯出 .xlsx",
  "export.timeSlot": "時段",
  "export.genre": "風格"
```

- [ ] **Step 3: Replace `frontend/src/styles.css` with violet accent**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  body {
    font-family: 'Inter', system-ui, sans-serif;
  }

  h1, h2, h3, h4, h5, h6 {
    font-family: 'Space Grotesk', system-ui, sans-serif;
  }
}

@layer components {
  /* Layout */
  .page { @apply max-w-4xl mx-auto px-6 py-8; }
  .page-sm { @apply max-w-2xl mx-auto px-6 py-8; }
  .page-header { @apply flex items-center justify-between mb-6; }
  .page-title { @apply text-2xl font-semibold text-gray-900; }
  .section-title { @apply text-lg font-semibold text-gray-900; }
  .back-link { @apply text-violet-600 hover:text-violet-700 text-sm font-medium; }

  /* Card */
  .card { @apply bg-white border border-gray-200 rounded-lg; }
  .card-body { @apply p-6; }
  .card-header { @apply flex items-center justify-between mb-4; }

  /* Forms */
  .form-group { @apply mb-4; }
  .form-label { @apply block text-sm font-medium text-gray-700 mb-1; }
  .form-input { @apply w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-violet-400; }
  .form-select { @apply w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-violet-400 disabled:bg-gray-50 disabled:text-gray-400; }
  .form-textarea { @apply w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-violet-400 resize-none; }
  .form-actions { @apply flex gap-3; }
  .form-grid-2 { @apply grid grid-cols-2 gap-4; }

  /* Buttons */
  .btn-primary { @apply bg-violet-600 hover:bg-violet-700 text-white font-medium px-4 py-2 rounded-lg transition-colors; }
  .btn-secondary { @apply border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-4 py-2 rounded-lg transition-colors; }
  .btn-sm { @apply bg-violet-600 hover:bg-violet-700 text-white font-medium px-3 py-1.5 rounded-md text-sm transition-colors; }
  .btn-sm-outline { @apply border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium px-3 py-1.5 rounded-md text-sm transition-colors; }
  .btn-green { @apply bg-green-500 hover:bg-green-600 text-white font-medium px-3 py-1.5 rounded-md text-sm transition-colors; }
  .btn-danger { @apply text-red-500 hover:text-red-700 text-sm font-medium transition-colors; }

  /* Modals */
  .modal-overlay { @apply fixed inset-0 z-50 flex items-center justify-center bg-black/40; }
  .modal-box-md { @apply bg-white rounded-xl shadow-xl p-6 w-full max-w-md mx-4; }
  .modal-box-sm { @apply bg-white rounded-xl shadow-xl p-6 w-full max-w-sm mx-4; }
  .modal-title { @apply text-lg font-semibold text-gray-900 mb-4; }
  .modal-actions { @apply flex gap-3 mt-5; }

  /* Tabs */
  .tabs { @apply flex gap-1 border-b border-gray-200; }
  .tab-btn { @apply pb-2 px-1 text-sm font-medium border-b-2 transition-colors; }
  .tab-active { @apply border-violet-600 text-violet-600; }
  .tab-inactive { @apply border-transparent text-gray-500 hover:text-gray-700; }

  /* Empty state */
  .empty-state { @apply text-gray-500 text-sm py-8 text-center; }
}
```

- [ ] **Step 4: Replace `frontend/src/app/app.component.css`**

```css
.navbar {
  @apply h-14 flex items-center justify-between px-6 sticky top-0 z-50;
  background-color: #1a1a2e;
}

.navbar-brand { @apply text-white font-bold text-lg; }
.navbar-links { @apply flex items-center gap-8; }
.navbar-link { @apply text-slate-300 hover:text-white pb-1 transition-colors text-sm font-medium; }
.lang-switcher { @apply flex gap-1; }
.lang-btn { @apply px-3 py-1 rounded text-xs font-medium transition-colors; }
.lang-btn-active { @apply bg-white text-slate-900; }
.lang-btn-inactive { @apply text-slate-300 hover:text-white; }
.app-main { @apply min-h-[calc(100dvh-3.5rem)] bg-gray-50; }
```

- [ ] **Step 5: Update stage color swatches in `frontend/src/app/admin/events/event-detail.component.ts`**

Replace lines 114–118 (the `newStageColor` default and `stageColors` array):
```typescript
newStageColor = '#06b6d4';
readonly stageColors = [
  '#06b6d4', // cyan-500
  '#e879f9', // fuchsia-400
  '#84cc16', // lime-500
  '#f97316', // orange-500
  '#0ea5e9', // sky-500
  '#ec4899', // pink-500
  '#facc15', // yellow-400
  '#14b8a6', // teal-500
];
```

- [ ] **Step 6: Verify visually**

Run `cd frontend && npm start`, open `http://localhost:4200`. Confirm:
- Navbar is deep charcoal (not slate-900)
- "+ New Event" button is violet
- Back links are violet
- Focus ring on form inputs is violet (click any input to check)
- DJs page "Add" button is violet

- [ ] **Step 7: Commit**

```bash
git add frontend/src/styles.css frontend/src/app/app.component.css frontend/src/assets/i18n/en.json frontend/src/assets/i18n/zh-TW.json frontend/src/app/admin/events/event-detail.component.ts
git commit -m "feat(ui): violet accent, charcoal navbar, vivid stage colors, new i18n keys"
```

---

## Task 2: Shared DataTableComponent

**Files:**
- Create: `frontend/src/app/shared/column-def.directive.ts`
- Create: `frontend/src/app/shared/data-table.component.ts`
- Create: `frontend/src/app/shared/data-table.component.html`
- Create: `frontend/src/app/shared/data-table.component.css`
- Create: `frontend/src/app/shared/data-table.component.spec.ts`

**Interfaces:**
- Produces: exported `TableColumn` interface, `DataTableComponent`, `ColumnDefDirective` — consumed by Tasks 3 and 4

- [ ] **Step 1: Write the failing test at `frontend/src/app/shared/data-table.component.spec.ts`**

```typescript
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { DataTableComponent, TableColumn } from './data-table.component';
import { TranslateModule } from '@ngx-translate/core';

const COLS: TableColumn[] = [
  { key: 'name', label: 'Name' },
  { key: 'genre', label: 'Genre' },
];
const DATA: Record<string, unknown>[] = [
  { id: '1', name: 'DJ Alpha', genre: 'House' },
  { id: '2', name: 'DJ Beta', genre: 'Techno' },
  { id: '3', name: 'MC Gamma', genre: 'House' },
];

describe('DataTableComponent', () => {
  let fixture: ComponentFixture<DataTableComponent>;
  let component: DataTableComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DataTableComponent, TranslateModule.forRoot()],
    }).compileComponents();
    fixture = TestBed.createComponent(DataTableComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput('columns', COLS);
    fixture.componentRef.setInput('data', DATA);
    fixture.detectChanges();
  });

  it('shows all rows by default', () => {
    expect(component.filteredData().length).toBe(3);
  });

  it('filters by search query case-insensitively', () => {
    component.searchQuery.set('dj');
    expect(component.filteredData().length).toBe(2);
    expect(component.filteredData().map(r => r['name'])).toEqual(['DJ Alpha', 'DJ Beta']);
  });

  it('filters across all searchable columns', () => {
    component.searchQuery.set('house');
    expect(component.filteredData().length).toBe(2);
  });

  it('sorts ascending on first header click', () => {
    component.sort(COLS[0]);
    expect(component.filteredData()[0]['name']).toBe('DJ Alpha');
  });

  it('sorts descending on second click of same column', () => {
    component.sort(COLS[0]);
    component.sort(COLS[0]);
    expect(component.filteredData()[0]['name']).toBe('MC Gamma');
  });

  it('resets to ascending when a different column is clicked', () => {
    component.sort(COLS[0]);
    component.sort(COLS[0]); // now descending
    component.sort(COLS[1]); // new column — should reset to asc
    expect(component.sortDir()).toBe('asc');
  });

  it('excludes searchable:false columns from search', () => {
    const cols: TableColumn[] = [
      { key: 'name', label: 'Name', searchable: false },
      { key: 'genre', label: 'Genre' },
    ];
    fixture.componentRef.setInput('columns', cols);
    component.searchQuery.set('dj'); // only matches 'name', which is non-searchable
    expect(component.filteredData().length).toBe(0);
  });

  it('does not sort when sortable is false', () => {
    const col: TableColumn = { key: 'name', label: 'Name', sortable: false };
    component.sort(col);
    expect(component.sortKey()).toBe('');
  });
});
```

- [ ] **Step 2: Run test — verify FAIL**

```bash
cd frontend && npx ng test --include="src/app/shared/data-table.component.spec.ts" --watch=false
```
Expected: FAIL — `DataTableComponent` not found.

- [ ] **Step 3: Create `frontend/src/app/shared/column-def.directive.ts`**

```typescript
import { Directive, TemplateRef, input } from '@angular/core';

@Directive({ selector: '[appColumnDef]', standalone: true })
export class ColumnDefDirective {
  columnKey = input.required<string>({ alias: 'appColumnDef' });
  constructor(public template: TemplateRef<{ row: Record<string, unknown> }>) {}
}
```

- [ ] **Step 4: Create `frontend/src/app/shared/data-table.component.ts`**

```typescript
import { Component, contentChildren, input, signal, computed, TemplateRef } from '@angular/core';
import { NgTemplateOutlet } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';
import { ColumnDefDirective } from './column-def.directive';

export interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;   // default: true
  searchable?: boolean; // default: true — participates in text filter
}

@Component({
  selector: 'app-data-table',
  standalone: true,
  imports: [FormsModule, NgTemplateOutlet, TranslatePipe, ColumnDefDirective],
  templateUrl: './data-table.component.html',
  styleUrl: './data-table.component.css',
})
export class DataTableComponent {
  columns = input<TableColumn[]>([]);
  data = input<Record<string, unknown>[]>([]);

  private defs = contentChildren(ColumnDefDirective);

  searchQuery = signal('');
  sortKey = signal('');
  sortDir = signal<'asc' | 'desc'>('asc');

  filteredData = computed(() => {
    const q = this.searchQuery().toLowerCase().trim();
    const key = this.sortKey();
    const dir = this.sortDir();
    const searchableCols = this.columns()
      .filter(c => c.searchable !== false)
      .map(c => c.key);

    let rows = this.data();

    if (q) {
      rows = rows.filter(row =>
        searchableCols.some(k => String(row[k] ?? '').toLowerCase().includes(q))
      );
    }

    if (key) {
      rows = [...rows].sort((a, b) => {
        const av = String(a[key] ?? '');
        const bv = String(b[key] ?? '');
        return dir === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av);
      });
    }

    return rows;
  });

  sort(col: TableColumn) {
    if (col.sortable === false) return;
    if (this.sortKey() === col.key) {
      this.sortDir.set(this.sortDir() === 'asc' ? 'desc' : 'asc');
    } else {
      this.sortKey.set(col.key);
      this.sortDir.set('asc');
    }
  }

  getTemplate(key: string): TemplateRef<{ row: Record<string, unknown> }> | null {
    return this.defs().find(d => d.columnKey() === key)?.template ?? null;
  }
}
```

- [ ] **Step 5: Create `frontend/src/app/shared/data-table.component.html`**

```html
<div class="dt-wrap">
  <div class="dt-toolbar">
    <input
      type="text"
      [ngModel]="searchQuery()"
      (ngModelChange)="searchQuery.set($event)"
      class="dt-search"
      [placeholder]="'table.search' | translate"
    />
  </div>
  <div class="dt-table-scroll">
    <table class="dt-table">
      <thead>
        <tr class="dt-head-row">
          @for (col of columns(); track col.key) {
            <th
              class="dt-th"
              [class.dt-th-sortable]="col.sortable !== false"
              (click)="sort(col)"
            >
              {{ col.label }}
              @if (sortKey() === col.key) {
                <span class="dt-sort-icon">{{ sortDir() === 'asc' ? '↑' : '↓' }}</span>
              }
            </th>
          }
        </tr>
      </thead>
      <tbody>
        @for (row of filteredData(); track row['id']) {
          <tr class="dt-row">
            @for (col of columns(); track col.key) {
              <td class="dt-td">
                @if (getTemplate(col.key); as tpl) {
                  <ng-container *ngTemplateOutlet="tpl; context: { row: row }"></ng-container>
                } @else {
                  {{ row[col.key] }}
                }
              </td>
            }
          </tr>
        }
        @if (filteredData().length === 0) {
          <tr>
            <td [attr.colspan]="columns().length" class="dt-empty">
              {{ 'table.empty' | translate }}
            </td>
          </tr>
        }
      </tbody>
    </table>
  </div>
</div>
```

- [ ] **Step 6: Create `frontend/src/app/shared/data-table.component.css`**

```css
.dt-wrap { @apply w-full; }
.dt-toolbar { @apply mb-3; }
.dt-search { @apply w-full sm:w-72 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-violet-400; }
.dt-table-scroll { @apply overflow-x-auto rounded-lg border border-gray-200; }
.dt-table { @apply w-full text-sm bg-white; }
.dt-head-row { @apply border-b border-gray-200 bg-gray-50; }
.dt-th { @apply text-left px-4 py-3 font-medium text-gray-700 select-none whitespace-nowrap; }
.dt-th-sortable { @apply cursor-pointer hover:bg-gray-100; }
.dt-sort-icon { @apply ml-1 text-violet-600; }
.dt-row { @apply border-b border-gray-100 last:border-0 hover:bg-gray-50; }
.dt-td { @apply px-4 py-3 text-gray-700; }
.dt-empty { @apply px-4 py-8 text-center text-gray-400 text-sm; }
```

- [ ] **Step 7: Run tests — verify PASS**

```bash
cd frontend && npx ng test --include="src/app/shared/data-table.component.spec.ts" --watch=false
```
Expected: 8 tests PASS.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/app/shared/column-def.directive.ts frontend/src/app/shared/data-table.component.ts frontend/src/app/shared/data-table.component.html frontend/src/app/shared/data-table.component.css frontend/src/app/shared/data-table.component.spec.ts
git commit -m "feat(shared): DataTableComponent with search, sort, and custom cell templates"
```

---

## Task 3: Migrate DJs Table to DataTableComponent

**Files:**
- Modify: `frontend/src/app/admin/djs/djs.component.ts`
- Modify: `frontend/src/app/admin/djs/djs.component.html`
- Modify: `frontend/src/app/admin/djs/djs.component.css`

**Interfaces:**
- Consumes: `DataTableComponent`, `ColumnDefDirective`, `TableColumn` from Task 2
- Consumes: `LanguageService` at `frontend/src/app/services/language.service.ts` — already in the codebase

- [ ] **Step 1: Replace `frontend/src/app/admin/djs/djs.component.ts`**

```typescript
import { Component, computed, inject, signal, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, DJ } from '../../services/api.service';
import { DialogService } from '../../shared/dialog.service';
import { LanguageService } from '../../services/language.service';
import { DataTableComponent, TableColumn } from '../../shared/data-table.component';
import { ColumnDefDirective } from '../../shared/column-def.directive';

@Component({
  selector: 'app-djs',
  standalone: true,
  imports: [FormsModule, TranslatePipe, DataTableComponent, ColumnDefDirective],
  templateUrl: './djs.component.html',
  styleUrl: './djs.component.css',
})
export class DjsComponent implements OnDestroy {
  private api = inject(ApiService);
  private translate = inject(TranslateService);
  private dialog = inject(DialogService);
  private langService = inject(LanguageService);

  djs = signal<DJ[]>([]);
  newDJ = signal({ name: '' });
  newDJGenres = signal('');

  private subscriptions: Subscription[] = [];

  constructor() {
    this.loadDJs();
  }

  private loadDJs() {
    this.subscriptions.push(
      this.api.getDJs().subscribe(djs => this.djs.set(djs))
    );
  }

  // Convert DJ[] to Record<string, unknown>[] for DataTableComponent
  djsAsRows = computed((): Record<string, unknown>[] =>
    this.djs().map(d => ({ id: d.id, name: d.name, genre_tags: d.genre_tags }) as Record<string, unknown>)
  );

  // Reactive column labels — re-evaluates when language changes
  djColumns = computed((): TableColumn[] => {
    this.langService.currentLang(); // register dependency so labels update on language switch
    return [
      { key: 'name', label: this.translate.instant('djs.name') },
      { key: 'genre_tags', label: this.translate.instant('djs.genres'), sortable: false },
      { key: 'actions', label: this.translate.instant('actions.title'), sortable: false, searchable: false },
    ];
  });

  genreTags(row: Record<string, unknown>): string[] {
    return Array.isArray(row['genre_tags']) ? (row['genre_tags'] as string[]) : [];
  }

  rowId(row: Record<string, unknown>): string {
    return String(row['id']);
  }

  async addDJ() {
    const name = this.newDJ().name;
    const genres = this.newDJGenres()
      .split(',')
      .map(g => g.trim())
      .filter(g => g.length > 0);

    if (!name || genres.length === 0) {
      await this.dialog.alert({
        title: this.translate.instant('dialog.validationTitle'),
        message: this.translate.instant('djs.fillRequired'),
      });
      return;
    }

    this.api.createDJ({ name, genre_tags: genres }).subscribe(() => {
      this.newDJ.set({ name: '' });
      this.newDJGenres.set('');
      this.loadDJs();
    });
  }

  async deleteDJ(id: string) {
    const ok = await this.dialog.confirm({
      title: this.translate.instant('dialog.deleteTitle'),
      message: this.translate.instant('djs.deleteConfirm'),
      confirmLabel: this.translate.instant('actions.delete'),
      variant: 'danger',
    });
    if (!ok) return;
    this.api.deleteDJ(id).subscribe(() => this.loadDJs());
  }

  ngOnDestroy() {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }
}
```

- [ ] **Step 2: Replace `frontend/src/app/admin/djs/djs.component.html`**

```html
<div class="page">
  <h1 class="page-title mb-6">{{ 'djs.title' | translate }}</h1>

  <div class="card card-body mb-6">
    <h2 class="section-title mb-4">{{ 'djs.new' | translate }}</h2>
    <div class="form-group">
      <label class="form-label">{{ 'djs.name' | translate }}</label>
      <input [ngModel]="newDJ().name" (ngModelChange)="newDJ.set({name: $event})"
        name="name" type="text" class="form-input" placeholder="DJ Name" />
    </div>
    <div class="form-group">
      <label class="form-label">{{ 'djs.genres' | translate }}</label>
      <input [ngModel]="newDJGenres()" (ngModelChange)="newDJGenres.set($event)"
        name="genres" type="text" class="form-input" placeholder="genre1, genre2, genre3" />
    </div>
    <button (click)="addDJ()" class="btn-primary">+ {{ 'actions.add' | translate }}</button>
  </div>

  <app-data-table [columns]="djColumns()" [data]="djsAsRows()">
    <ng-template appColumnDef="genre_tags" let-row="row">
      @for (g of genreTags(row); track g) {
        <span class="genre-tag">{{ g }}</span>
      }
    </ng-template>
    <ng-template appColumnDef="actions" let-row="row">
      <button (click)="deleteDJ(rowId(row))" class="btn-danger">
        {{ 'actions.delete' | translate }}
      </button>
    </ng-template>
  </app-data-table>
</div>
```

- [ ] **Step 3: Replace `frontend/src/app/admin/djs/djs.component.css`**

```css
.genre-tag { @apply inline-block bg-violet-50 text-violet-700 px-2 py-0.5 rounded-full text-xs mr-1 mb-1; }
```

- [ ] **Step 4: Verify in browser**

Navigate to `/admin/djs`. Confirm:
- Search input filters DJs by name or genre live
- Clicking "Name" header sorts alphabetically; clicking again reverses
- Genre tags render as violet pills
- Delete button works (triggers confirm dialog)
- Switching language (EN ↔ 中文) re-labels the column headers

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/admin/djs/djs.component.ts frontend/src/app/admin/djs/djs.component.html frontend/src/app/admin/djs/djs.component.css
git commit -m "feat(djs): migrate to shared DataTableComponent with search and sort"
```

---

## Task 4: Event Detail — Hero Header, Stage Chips, Date Tabs

**Files:**
- Modify: `frontend/src/app/admin/events/event-detail.component.ts`
- Modify: `frontend/src/app/admin/events/event-detail.component.html`
- Modify: `frontend/src/app/admin/events/event-detail.component.css`

**Interfaces:**
- Produces: `selectedDate` signal, `dates` computed, `slotsForSelectedDate` computed, `tabDate()` — consumed by Task 5

- [ ] **Step 1: Add `selectedDate`, `dates`, `slotsForSelectedDate`, `tabDate` to `event-detail.component.ts`**

After the existing signal declarations (after `djs = signal...` on line 25), add:
```typescript
selectedDate = signal('');

dates = computed(() => {
  const slotDates = [...new Set(this.slots().map(s => s.slot_date))].sort();
  const start = this.event()?.start_date;
  if (start && !slotDates.includes(start)) slotDates.unshift(start);
  return slotDates.sort();
});

slotsForSelectedDate = computed(() =>
  this.slots()
    .filter(s => s.slot_date === this.selectedDate())
    .sort((a, b) => a.start_time.localeCompare(b.start_time))
);

tabDate(d: string): string {
  const [, month, day] = d.split('-');
  return `${parseInt(month)}/${parseInt(day)}`;
}
```

Update the event subscription inside `loadData()` to initialize `selectedDate` when the event loads. Replace:
```typescript
this.api.getEvent(eventId).subscribe(e => this.event.set(e)),
```
With:
```typescript
this.api.getEvent(eventId).subscribe(e => {
  this.event.set(e);
  if (this.selectedDate() === '') this.selectedDate.set(e.start_date);
}),
```

Also add this stub (implemented fully in Task 6):
```typescript
exportXlsx() { /* implemented in Task 6 */ }
```

- [ ] **Step 2: Replace `frontend/src/app/admin/events/event-detail.component.html`**

```html
@if (event()) {
  <div class="page">

    <!-- Hero Header -->
    <div class="detail-hero">
      <div class="detail-hero-info">
        <h1 class="detail-event-name">{{ event()!.name }}</h1>
        <p class="detail-event-meta">{{ dateRange() }} · {{ event()!.venue_name }}</p>
        @if (event()!.genres?.length) {
          <div class="flex gap-1 flex-wrap mt-2">
            @for (g of event()!.genres; track g) {
              <span class="genre-pill">{{ g }}</span>
            }
          </div>
        }
      </div>
      <div class="detail-hero-actions">
        <a routerLink="/admin/events" class="back-link">← {{ 'actions.back' | translate }}</a>
        <button (click)="viewPublic()" class="btn-sm-outline">{{ 'eventDetail.viewPublic' | translate }}</button>
        <button (click)="shareViaLine()" class="btn-green">{{ 'schedule.share' | translate }}</button>
        <button (click)="exportXlsx()" class="btn-sm-outline">{{ 'eventDetail.export' | translate }}</button>
      </div>
    </div>

    <!-- Stage Chips Strip -->
    <div class="stage-strip">
      @for (s of stages(); track s.id) {
        <div class="stage-chip group">
          <span class="stage-chip-dot" [style.background]="s.color"></span>
          <span class="stage-chip-name">{{ s.name }}</span>
          <button (click)="deleteStage(s.id)" class="stage-chip-delete" title="Delete">×</button>
        </div>
      }
      <button (click)="addStage()" class="stage-chip-add">
        + {{ 'stages.new' | translate }}
      </button>
    </div>

    <!-- Slots Section -->
    <div class="mt-4">
      <h2 class="section-title mb-3">{{ 'slots.title' | translate }}</h2>

      <!-- Date Tabs (only shown for multi-day events) -->
      @if (dates().length > 1) {
        <div class="tabs mb-4">
          @for (d of dates(); track d) {
            <button (click)="selectedDate.set(d)"
              class="tab-btn"
              [class]="selectedDate() === d ? 'tab-active' : 'tab-inactive'">
              {{ tabDate(d) }}
            </button>
          }
        </div>
      }

      <div class="card overflow-hidden">
        <table class="slot-table">
          <thead>
            <tr class="border-b border-gray-200 bg-gray-50">
              <th class="slot-th">{{ 'slots.stage' | translate }}</th>
              <th class="slot-th">{{ 'slots.dj' | translate }}</th>
              <th class="slot-th">{{ 'slots.genre' | translate }}</th>
              <th class="slot-th">{{ 'slots.start' | translate }}</th>
              <th class="slot-th">{{ 'slots.duration' | translate }}</th>
              <th class="slot-th"></th>
            </tr>
          </thead>
          <tbody>
            @for (slot of slotsForSelectedDate(); track slot.id) {
              @if (editingSlotId() === slot.id) {
                <tr class="edit-row">
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotStageId" [name]="'es_stage_' + slot.id" class="edit-select">
                      @for (s of stages(); track s.id) {
                        <option [value]="s.id">{{ s.name }}</option>
                      }
                    </select>
                  </td>
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotDjId" [name]="'es_dj_' + slot.id"
                      (ngModelChange)="editSlotGenre = ''" class="edit-select">
                      <option value="">—</option>
                      @for (d of djs(); track d.id) {
                        <option [value]="d.id">{{ d.name }}</option>
                      }
                    </select>
                  </td>
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotGenre" [name]="'es_genre_' + slot.id"
                      [disabled]="!editSlotDjId || djGenresForEdit().length === 0"
                      class="edit-select">
                      <option value="">{{ djGenresForEdit().length ? '— select —' : '—' }}</option>
                      @for (g of djGenresForEdit(); track g) {
                        <option [value]="g">{{ g }}</option>
                      }
                    </select>
                  </td>
                  <td class="edit-cell">
                    <input [(ngModel)]="editSlotStart" [name]="'es_start_' + slot.id"
                      type="time" class="edit-time-input" />
                  </td>
                  <td class="edit-cell">
                    <input [(ngModel)]="editSlotDuration" [name]="'es_dur_' + slot.id"
                      type="number" min="5" step="5" list="durationPresets"
                      class="edit-dur-input" />
                    <datalist id="durationPresets">
                      @for (d of durationOptions; track d.mins) {
                        <option [value]="d.mins">{{ d.label }}</option>
                      }
                    </datalist>
                  </td>
                  <td class="edit-actions-cell">
                    <button (click)="saveEdit(slot.id)" class="btn-save">Save</button>
                    <button (click)="cancelEdit()" class="btn-cancel-edit">Cancel</button>
                  </td>
                </tr>
              } @else {
                <tr class="slot-row">
                  <td class="slot-stage-cell">
                    <span class="slot-stage-inner">
                      <span class="slot-stage-dot" [style.background]="stageColor(slot.stage_id)"></span>
                      {{ slot.stage_name }}
                    </span>
                  </td>
                  <td class="slot-dj-cell">{{ slot.dj_name || '—' }}</td>
                  <td class="slot-genre-cell">{{ slot.genre || '—' }}</td>
                  <td class="slot-time-cell">{{ slot.start_time }}</td>
                  <td class="slot-time-cell">{{ duration(slot.start_time, slot.end_time) }}</td>
                  <td class="slot-actions-cell">
                    <button (click)="startEdit(slot)" class="slot-edit-btn">Edit</button>
                    <button (click)="deleteSlot(slot.id)" class="btn-danger">
                      {{ 'actions.delete' | translate }}
                    </button>
                  </td>
                </tr>
              }
            }
            <!-- Add row placeholder — replaced in Task 5 -->
            <tr class="add-trigger-row">
              <td colspan="6" class="add-trigger-cell">+ {{ 'slots.new' | translate }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
}

<!-- Add Stage Modal -->
@if (showStageModal()) {
  <div class="modal-overlay" (click)="cancelStage()">
    <div class="modal-box-sm" (click)="$event.stopPropagation()">
      <h2 class="modal-title">{{ 'stages.new' | translate }}</h2>
      <div class="form-group">
        <label class="form-label">{{ 'stages.name' | translate }}</label>
        <input [(ngModel)]="newStageName" name="newStageName" type="text" autofocus
          class="form-input"
          (keydown.enter)="submitNewStage()"
          (keydown.escape)="cancelStage()" />
      </div>
      <div class="mb-5">
        <label class="form-label">Color</label>
        <div class="color-swatches">
          @for (c of stageColors; track c) {
            <button type="button" (click)="newStageColor = c"
              class="color-swatch"
              [style.background]="c"
              [class]="newStageColor === c ? 'swatch-active' : 'swatch-inactive'">
            </button>
          }
        </div>
      </div>
      <div class="form-actions">
        <button type="button" (click)="submitNewStage()" class="btn-primary">
          {{ 'actions.create' | translate }}
        </button>
        <button type="button" (click)="cancelStage()" class="btn-secondary">
          {{ 'actions.cancel' | translate }}
        </button>
      </div>
    </div>
  </div>
}
```

- [ ] **Step 3: Replace `frontend/src/app/admin/events/event-detail.component.css`**

```css
/* Hero header */
.detail-hero { @apply flex items-start justify-between mb-4 gap-4 flex-wrap; }
.detail-hero-info { @apply flex-1 min-w-0; }
.detail-event-name { @apply text-3xl font-bold text-gray-900 leading-tight; }
.detail-event-meta { @apply text-sm text-gray-500 mt-1; }
.detail-hero-actions { @apply flex items-center gap-2 flex-shrink-0 flex-wrap; }

/* Genre pills */
.genre-pill { @apply px-2 py-0.5 bg-violet-50 text-violet-700 rounded-full text-xs font-medium; }

/* Stage chip strip */
.stage-strip { @apply flex items-center gap-2 flex-wrap py-2 mb-2; }
.stage-chip { @apply flex items-center gap-1.5 px-3 py-1.5 bg-white border border-gray-200 rounded-full text-sm; }
.stage-chip-dot { @apply w-2.5 h-2.5 rounded-full flex-shrink-0; }
.stage-chip-name { @apply font-medium text-gray-800; }
.stage-chip-delete { @apply ml-1 text-gray-300 hover:text-red-400 text-base leading-none hidden group-hover:inline transition-colors; }
.stage-chip-add { @apply px-3 py-1.5 border border-dashed border-gray-300 rounded-full text-sm text-gray-500 hover:border-violet-400 hover:text-violet-600 transition-colors; }

/* Slot table */
.slot-table { @apply w-full text-sm; }
.slot-th { @apply text-left px-3 py-2 font-medium text-gray-700; }
.slot-row { @apply border-b border-gray-100 last:border-0 hover:bg-gray-50; }
.slot-stage-cell { @apply px-3 py-3; }
.slot-stage-inner { @apply flex items-center gap-2; }
.slot-stage-dot { @apply w-2.5 h-2.5 rounded-full flex-shrink-0; }
.slot-dj-cell { @apply px-3 py-3 font-medium text-gray-900; }
.slot-genre-cell { @apply px-3 py-3 text-gray-500; }
.slot-time-cell { @apply px-3 py-3 text-gray-700 whitespace-nowrap; }
.slot-actions-cell { @apply px-3 py-3 text-right whitespace-nowrap; }
.slot-edit-btn { @apply text-gray-500 hover:text-gray-800 text-sm font-medium mr-3; }

/* Inline edit row */
.edit-row { @apply border-b border-gray-100 bg-violet-50; }
.edit-cell { @apply px-2 py-2; }
.edit-select { @apply w-full text-sm px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-violet-400 disabled:bg-gray-50 disabled:text-gray-400; }
.edit-time-input { @apply text-sm px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-violet-400 w-24; }
.edit-dur-input { @apply text-sm px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-violet-400 w-16; }
.edit-actions-cell { @apply px-2 py-2 text-right whitespace-nowrap; }
.btn-save { @apply bg-violet-600 hover:bg-violet-700 text-white text-xs font-medium px-3 py-1 rounded mr-1 transition-colors; }
.btn-cancel-edit { @apply border border-gray-300 text-gray-600 hover:bg-gray-50 text-xs font-medium px-3 py-1 rounded transition-colors; }

/* Add row (placeholder — wired up in Task 5) */
.add-trigger-row { @apply hover:bg-gray-50 cursor-pointer; }
.add-trigger-cell { @apply px-3 py-3 text-sm text-violet-600 font-medium; }

/* Stage modal color swatches */
.color-swatches { @apply flex gap-2 flex-wrap; }
.color-swatch { @apply w-7 h-7 rounded-full border-2 transition-all; }
.swatch-active { @apply border-gray-900; }
.swatch-inactive { @apply border-transparent; }
```

- [ ] **Step 4: Verify in browser**

Open an event detail page. Confirm:
- Event name renders large (`text-3xl bold`)
- Genre pills are violet
- Stage chips appear as a horizontal strip with colored dots; hovering reveals `×`
- "+ New Stage" ghost chip opens the stage modal
- Date tabs appear for multi-day events; active tab is violet underlined
- Clicking a tab updates the slot list to show only that date's slots

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/admin/events/event-detail.component.ts frontend/src/app/admin/events/event-detail.component.html frontend/src/app/admin/events/event-detail.component.css
git commit -m "feat(event-detail): hero header, stage chips, date tabs"
```

---

## Task 5: Inline Quick-Add Row + Slot Search/Sort

**Files:**
- Modify: `frontend/src/app/admin/events/event-detail.component.ts`
- Modify: `frontend/src/app/admin/events/event-detail.component.html`
- Modify: `frontend/src/app/admin/events/event-detail.component.css`

**Interfaces:**
- Consumes: `selectedDate`, `slotsForSelectedDate` from Task 4

- [ ] **Step 1: Add add-row state, pre-fill computeds, slot search/sort to `event-detail.component.ts`**

Remove the following properties and methods (they belong to the old modal):
- `showSlotModal = signal(false)`
- `newSlotStageId`, `newSlotDjId`, `newSlotGenre`, `newSlotDate`, `newSlotStart`, `newSlotDuration`, `newSlotNotes`
- `addSlot()`, `submitNewSlot()`, `cancelSlot()`

Add after `slotsForSelectedDate` computed (from Task 4):

```typescript
// ── Slot search + sort ──────────────────────────────────────────────
slotSearch = signal('');
slotSortKey = signal<'stage_name' | 'start_time' | 'dj_name'>('start_time');
slotSortDir = signal<'asc' | 'desc'>('asc');

sortedFilteredSlots = computed(() => {
  const q = this.slotSearch().toLowerCase().trim();
  const key = this.slotSortKey();
  const dir = this.slotSortDir();

  let rows = this.slotsForSelectedDate();
  if (q) {
    rows = rows.filter(s =>
      s.dj_name?.toLowerCase().includes(q) ||
      s.stage_name?.toLowerCase().includes(q) ||
      s.genre?.toLowerCase().includes(q)
    );
  }
  return [...rows].sort((a, b) => {
    const av = String(a[key] ?? '');
    const bv = String(b[key] ?? '');
    return dir === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av);
  });
});

sortSlots(key: 'stage_name' | 'start_time' | 'dj_name') {
  if (this.slotSortKey() === key) {
    this.slotSortDir.set(this.slotSortDir() === 'asc' ? 'desc' : 'asc');
  } else {
    this.slotSortKey.set(key);
    this.slotSortDir.set('asc');
  }
}

sortIndicator(key: 'stage_name' | 'start_time' | 'dj_name'): string {
  if (this.slotSortKey() !== key) return '';
  return this.slotSortDir() === 'asc' ? ' ↑' : ' ↓';
}

// ── Inline add row ──────────────────────────────────────────────────
addRowActive = signal(false);
addStageId = '';
addDjId = '';
addGenre = '';
addDate = '';
addStart = '';
addDuration = 60;
addNotes = '';

defaultStageId = computed(() => this.stages().length === 1 ? this.stages()[0].id : '');

defaultStartTime = computed(() => {
  const slots = this.slotsForSelectedDate();
  return slots.length > 0 ? slots[slots.length - 1].end_time : '18:00';
});

djGenresForAdd(): string[] {
  return this.djs().find(d => d.id === this.addDjId)?.genre_tags ?? [];
}

activateAddRow() {
  this.addStageId = this.defaultStageId();
  this.addDjId = '';
  this.addGenre = '';
  this.addDate = this.selectedDate();
  this.addStart = this.defaultStartTime();
  this.addDuration = 60;
  this.addNotes = '';
  this.addRowActive.set(true);
}

cancelAddRow() {
  this.addRowActive.set(false);
}

submitAddRow() {
  if (!this.addStageId || !this.addDate || !this.addStart) return;
  const eventId = this.event()!.id;
  this.api.createSlot(eventId, {
    stage_id: this.addStageId,
    dj_id: this.addDjId,
    genre: this.addGenre,
    slot_date: this.addDate,
    start_time: this.addStart,
    end_time: this.addMinutes(this.addStart, this.addDuration),
    notes: this.addNotes,
  }).subscribe(() => {
    this.api.getSlots(eventId).subscribe(s => {
      this.slots.set(s);
      this.activateAddRow(); // reset with fresh pre-fills for the next slot
    });
  });
}
```

- [ ] **Step 2: Update the slots section in `event-detail.component.html`**

Replace the entire `<!-- Slots Section -->` div (from Task 4) with:

```html
    <!-- Slots Section -->
    <div class="mt-4">
      <div class="flex items-center justify-between mb-3">
        <h2 class="section-title">{{ 'slots.title' | translate }}</h2>
        <input
          type="text"
          [ngModel]="slotSearch()"
          (ngModelChange)="slotSearch.set($event)"
          class="slot-search-input"
          [placeholder]="'table.search' | translate"
        />
      </div>

      @if (dates().length > 1) {
        <div class="tabs mb-4">
          @for (d of dates(); track d) {
            <button (click)="selectedDate.set(d)"
              class="tab-btn"
              [class]="selectedDate() === d ? 'tab-active' : 'tab-inactive'">
              {{ tabDate(d) }}
            </button>
          }
        </div>
      }

      <div class="card overflow-hidden">
        <table class="slot-table">
          <thead>
            <tr class="border-b border-gray-200 bg-gray-50">
              <th class="slot-th slot-th-sort" (click)="sortSlots('stage_name')">
                {{ 'slots.stage' | translate }}{{ sortIndicator('stage_name') }}
              </th>
              <th class="slot-th slot-th-sort" (click)="sortSlots('dj_name')">
                {{ 'slots.dj' | translate }}{{ sortIndicator('dj_name') }}
              </th>
              <th class="slot-th">{{ 'slots.genre' | translate }}</th>
              <th class="slot-th slot-th-sort" (click)="sortSlots('start_time')">
                {{ 'slots.start' | translate }}{{ sortIndicator('start_time') }}
              </th>
              <th class="slot-th">{{ 'slots.duration' | translate }}</th>
              <th class="slot-th"></th>
            </tr>
          </thead>
          <tbody>
            @for (slot of sortedFilteredSlots(); track slot.id) {
              @if (editingSlotId() === slot.id) {
                <tr class="edit-row">
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotStageId" [name]="'es_stage_' + slot.id" class="edit-select"
                      (keydown.escape)="cancelEdit()">
                      @for (s of stages(); track s.id) {
                        <option [value]="s.id">{{ s.name }}</option>
                      }
                    </select>
                  </td>
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotDjId" [name]="'es_dj_' + slot.id"
                      (ngModelChange)="editSlotGenre = ''" class="edit-select"
                      (keydown.escape)="cancelEdit()">
                      <option value="">—</option>
                      @for (d of djs(); track d.id) {
                        <option [value]="d.id">{{ d.name }}</option>
                      }
                    </select>
                  </td>
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotGenre" [name]="'es_genre_' + slot.id"
                      [disabled]="!editSlotDjId || djGenresForEdit().length === 0"
                      class="edit-select"
                      (keydown.escape)="cancelEdit()">
                      <option value="">{{ djGenresForEdit().length ? '— select —' : '—' }}</option>
                      @for (g of djGenresForEdit(); track g) {
                        <option [value]="g">{{ g }}</option>
                      }
                    </select>
                  </td>
                  <td class="edit-cell">
                    <input [(ngModel)]="editSlotStart" [name]="'es_start_' + slot.id"
                      type="time" class="edit-time-input"
                      (keydown.enter)="saveEdit(slot.id)"
                      (keydown.escape)="cancelEdit()" />
                  </td>
                  <td class="edit-cell">
                    <input [(ngModel)]="editSlotDuration" [name]="'es_dur_' + slot.id"
                      type="number" min="5" step="5" list="durationPresets"
                      class="edit-dur-input"
                      (keydown.enter)="saveEdit(slot.id)"
                      (keydown.escape)="cancelEdit()" />
                    <datalist id="durationPresets">
                      @for (d of durationOptions; track d.mins) {
                        <option [value]="d.mins">{{ d.label }}</option>
                      }
                    </datalist>
                  </td>
                  <td class="edit-actions-cell">
                    <button (click)="saveEdit(slot.id)" class="btn-save">Save</button>
                    <button (click)="cancelEdit()" class="btn-cancel-edit">Cancel</button>
                  </td>
                </tr>
              } @else {
                <tr class="slot-row">
                  <td class="slot-stage-cell">
                    <span class="slot-stage-inner">
                      <span class="slot-stage-dot" [style.background]="stageColor(slot.stage_id)"></span>
                      {{ slot.stage_name }}
                    </span>
                  </td>
                  <td class="slot-dj-cell">{{ slot.dj_name || '—' }}</td>
                  <td class="slot-genre-cell">{{ slot.genre || '—' }}</td>
                  <td class="slot-time-cell">{{ slot.start_time }}</td>
                  <td class="slot-time-cell">{{ duration(slot.start_time, slot.end_time) }}</td>
                  <td class="slot-actions-cell">
                    <button (click)="startEdit(slot)" class="slot-edit-btn">Edit</button>
                    <button (click)="deleteSlot(slot.id)" class="btn-danger">
                      {{ 'actions.delete' | translate }}
                    </button>
                  </td>
                </tr>
              }
            }

            <!-- Inline Add Row -->
            @if (addRowActive()) {
              <tr class="add-row">
                <td class="add-cell">
                  <select [(ngModel)]="addStageId" name="addStageId" class="edit-select"
                    (keydown.escape)="cancelAddRow()">
                    <option value="">— stage —</option>
                    @for (s of stages(); track s.id) {
                      <option [value]="s.id">{{ s.name }}</option>
                    }
                  </select>
                </td>
                <td class="add-cell">
                  <select [(ngModel)]="addDjId" name="addDjId"
                    (ngModelChange)="addGenre = ''" class="edit-select"
                    (keydown.escape)="cancelAddRow()">
                    <option value="">{{ 'slots.unassigned' | translate }}</option>
                    @for (d of djs(); track d.id) {
                      <option [value]="d.id">{{ d.name }}</option>
                    }
                  </select>
                </td>
                <td class="add-cell">
                  <select [(ngModel)]="addGenre" name="addGenre"
                    [disabled]="!addDjId || djGenresForAdd().length === 0"
                    class="edit-select"
                    (keydown.escape)="cancelAddRow()">
                    <option value="">
                      {{ addDjId ? (djGenresForAdd().length ? '— select —' : '—') : '— DJ first —' }}
                    </option>
                    @for (g of djGenresForAdd(); track g) {
                      <option [value]="g">{{ g }}</option>
                    }
                  </select>
                </td>
                <td class="add-cell">
                  <input [(ngModel)]="addStart" name="addStart" type="time" class="edit-time-input"
                    (keydown.enter)="submitAddRow()"
                    (keydown.escape)="cancelAddRow()" />
                </td>
                <td class="add-cell">
                  <input [(ngModel)]="addDuration" name="addDuration"
                    type="number" min="5" step="5" list="durationPresets"
                    class="edit-dur-input"
                    (keydown.enter)="submitAddRow()"
                    (keydown.escape)="cancelAddRow()" />
                </td>
                <td class="add-cell text-right whitespace-nowrap">
                  <button (click)="submitAddRow()" class="btn-save">Save</button>
                  <button (click)="cancelAddRow()" class="btn-cancel-edit">Cancel</button>
                </td>
              </tr>
            } @else {
              <tr class="add-trigger-row" (click)="activateAddRow()">
                <td colspan="6" class="add-trigger-cell">
                  + {{ 'slots.new' | translate }}
                </td>
              </tr>
            }
          </tbody>
        </table>
      </div>
    </div>
```

- [ ] **Step 3: Add add-row and search CSS to `event-detail.component.css`**

Append to the existing CSS file:
```css
/* Sortable slot headers */
.slot-th-sort { @apply cursor-pointer hover:bg-gray-100 select-none; }

/* Slot search */
.slot-search-input { @apply w-52 px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-violet-400; }

/* Inline add row */
.add-row { @apply border-b border-violet-100 bg-violet-50; }
.add-cell { @apply px-2 py-2; }
.add-trigger-row { @apply hover:bg-violet-50 cursor-pointer transition-colors; }
.add-trigger-cell { @apply px-3 py-3 text-sm text-violet-600 font-medium; }
```

- [ ] **Step 4: Verify in browser**

Open an event detail page with at least two slots. Confirm:
- Search input filters slot rows live by DJ name, stage name, or genre
- Clicking Stage/DJ/Start column headers sorts; arrow indicator appears; clicking again reverses
- Clicking the "+ New Slot" row activates the inline add row
- Add row pre-fills: Stage (if only one stage exists), Start time (end time of last slot, or `18:00`)
- Filling Stage + Start and clicking Save creates the slot; add row resets for the next entry
- Enter key in time/duration fields saves; Escape cancels
- Genre dropdown disables until a DJ is selected
- Edit row still works for existing slots with Escape to cancel

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/admin/events/event-detail.component.ts frontend/src/app/admin/events/event-detail.component.html frontend/src/app/admin/events/event-detail.component.css
git commit -m "feat(event-detail): inline quick-add row with smart pre-fills, slot search and sort"
```

---

## Task 6: Excel Export

**Files:**
- Modify: `frontend/package.json` (via npm install)
- Modify: `frontend/src/app/admin/events/event-detail.component.ts`

**Interfaces:**
- Consumes: `event()`, `slots()`, `LanguageService.currentLang()`, `TranslateService.instant()`, `toMins()` (already on the class)

- [ ] **Step 1: Install xlsx**

```bash
cd frontend && npm install xlsx
```
Expected: `"xlsx"` appears under `"dependencies"` in `package.json`.

- [ ] **Step 2: Add `LanguageService` injection to `event-detail.component.ts`**

Add import at the top:
```typescript
import { LanguageService } from '../../services/language.service';
```

Add injection in the class body (after `private dialog = inject(DialogService)`):
```typescript
private langService = inject(LanguageService);
```

- [ ] **Step 3: Add `exportXlsx()` to `event-detail.component.ts`**

Add import at the top of the file:
```typescript
import * as XLSX from 'xlsx';
```

Add method in the class body:
```typescript
exportXlsx() {
  const event = this.event()!;
  const lang = this.langService.currentLang();
  const t = (key: string) => this.translate.instant(key);

  const headers = lang === 'zh-TW'
    ? [t('slots.date'), t('slots.stage'), t('export.timeSlot'), 'DJ', t('export.genre'), t('slots.notes')]
    : [t('slots.date'), t('slots.stage'), t('export.timeSlot'), 'DJ', t('export.genre'), t('slots.notes')];

  const sorted = [...this.slots()].sort((a, b) => {
    if (a.slot_date !== b.slot_date) return a.slot_date.localeCompare(b.slot_date);
    if ((a.stage_name ?? '') !== (b.stage_name ?? '')) {
      return (a.stage_name ?? '').localeCompare(b.stage_name ?? '');
    }
    return a.start_time.localeCompare(b.start_time);
  });

  const rows: (string | number)[][] = [];
  let lastDate = '';
  for (const slot of sorted) {
    if (lastDate && slot.slot_date !== lastDate) rows.push([]); // blank row between dates
    lastDate = slot.slot_date;

    const [, month, day] = slot.slot_date.split('-');
    const dateStr = `${parseInt(month)}/${parseInt(day)}`;
    const mins = this.toMins(slot.end_time) - this.toMins(slot.start_time);
    const hrs = mins / 60;
    const durStr = Number.isInteger(hrs) ? `${hrs}hr` : `${hrs}hr`;
    const timeSlot = `${slot.start_time} - ${slot.end_time} (${durStr})`;

    rows.push([
      dateStr,
      slot.stage_name ?? '',
      timeSlot,
      slot.dj_name ?? '',
      slot.genre ?? '',
      slot.notes ?? '',
    ]);
  }

  const wsData: (string | number)[][] = [[event.name], headers, ...rows];
  const ws = XLSX.utils.aoa_to_sheet(wsData);
  if (!ws['!merges']) ws['!merges'] = [];
  ws['!merges'].push({ s: { r: 0, c: 0 }, e: { r: 0, c: 5 } });

  const wb = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(wb, ws, 'Schedule');
  XLSX.writeFile(wb, `${event.name.replace(/\s+/g, '-')}-schedule.xlsx`);
}
```

Note: `toMins` is already a `private` method on the class. Change its access modifier from `private` to `protected` or remove the `private` keyword so `exportXlsx` can call it from within the same class. (Both methods are on the same class, so `private` is already accessible — no change needed.)

- [ ] **Step 4: Verify the export**

1. Open an event detail page with slots on at least two different dates
2. Click "Export .xlsx"
3. Open the downloaded file in Excel or Google Sheets. Confirm:
   - Row 1: event name spanning all 6 columns
   - Row 2: `Date | Stage | Time Slot | DJ | Genre | Notes`
   - Data sorted by date → stage name → start time
   - Blank row between each date group
   - Time Slot format: `HH:MM - HH:MM (Xhr)` e.g. `21:00 - 22:30 (1.5hr)`
4. Switch app language to 中文, click Export again. Confirm headers are: `日期 | 舞台 | 時段 | DJ | 風格 | 備註`

- [ ] **Step 5: Commit**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/app/admin/events/event-detail.component.ts
git commit -m "feat(event-detail): client-side Excel export with language-aware headers"
```

---

## Task 7: Slot Notes Field UI (EL-031)

Adds a Notes column to the slots table and a notes text input to both the inline add row and the inline edit row. The DB column (`slots.notes TEXT`), Go model (`Notes string`), API service `Slot.notes` field, and `addNotes`/`editSlotNotes` component state already exist (from earlier work and Task 5). The Excel export (Task 6) already writes a Notes column — this task populates the UI so notes can be entered.

**Files:**
- Modify: `frontend/src/app/admin/events/event-detail.component.html`
- Modify: `frontend/src/app/admin/events/event-detail.component.css`

**Interfaces:**
- Consumes (from Task 5): `addNotes` (string), `editSlotNotes` (string), `submitAddRow()`, `cancelAddRow()`, `saveEdit(id)`, `cancelEdit()`, `sortedFilteredSlots()`, `editingSlotId()`, `addRowActive()`
- Consumes (existing): `Slot.notes` field on the API model, `'slots.notes'` i18n key (already in en.json + zh-TW.json)
- Produces: a 7-column slots table (Stage | DJ | Genre | Start | Duration | Notes | Actions) consumed by Task 8

- [ ] **Step 1: Add the Notes column header**

In `event-detail.component.html`, find the slots table `<thead>` row (the one with sortable headers from Task 5) and add a Notes `<th>` immediately before the empty actions `<th>`:

Old:
```html
              <th class="slot-th">{{ 'slots.duration' | translate }}</th>
              <th class="slot-th"></th>
            </tr>
```

New:
```html
              <th class="slot-th">{{ 'slots.duration' | translate }}</th>
              <th class="slot-th">{{ 'slots.notes' | translate }}</th>
              <th class="slot-th"></th>
            </tr>
```

- [ ] **Step 2: Add the notes cell to the read-only slot row**

Find the `@else` read-only `<tr class="slot-row">` block. Add a notes cell immediately before `<td class="slot-actions-cell">`:

Old:
```html
                  <td class="slot-time-cell">{{ duration(slot.start_time, slot.end_time) }}</td>
                  <td class="slot-actions-cell">
```

New:
```html
                  <td class="slot-time-cell">{{ duration(slot.start_time, slot.end_time) }}</td>
                  <td class="slot-notes-cell">
                    @if (slot.notes) {
                      <span class="slot-notes-text" [title]="slot.notes">{{ slot.notes }}</span>
                    }
                  </td>
                  <td class="slot-actions-cell">
```

- [ ] **Step 3: Add the notes input to the inline edit row**

Find the edit row (`<tr class="edit-row">`). Add a notes cell immediately before `<td class="edit-actions-cell">`:

Old:
```html
                    <datalist id="durationPresets">
                      @for (d of durationOptions; track d.mins) {
                        <option [value]="d.mins">{{ d.label }}</option>
                      }
                    </datalist>
                  </td>
                  <td class="edit-actions-cell">
                    <button (click)="saveEdit(slot.id)" class="btn-save">Save</button>
```

New:
```html
                    <datalist id="durationPresets">
                      @for (d of durationOptions; track d.mins) {
                        <option [value]="d.mins">{{ d.label }}</option>
                      }
                    </datalist>
                  </td>
                  <td class="edit-cell">
                    <input [(ngModel)]="editSlotNotes" [name]="'es_notes_' + slot.id"
                      type="text" class="edit-notes-input"
                      [placeholder]="'slots.notes' | translate"
                      (keydown.enter)="saveEdit(slot.id)"
                      (keydown.escape)="cancelEdit()" />
                  </td>
                  <td class="edit-actions-cell">
                    <button (click)="saveEdit(slot.id)" class="btn-save">Save</button>
```

- [ ] **Step 4: Add the notes input to the inline add row**

Find the add row (`<tr class="add-row">`). Add a notes cell immediately before the final actions cell (`<td class="add-cell text-right whitespace-nowrap">`):

Old:
```html
                  <input [(ngModel)]="addDuration" name="addDuration"
                    type="number" min="5" step="5" list="durationPresets"
                    class="edit-dur-input"
                    (keydown.enter)="submitAddRow()"
                    (keydown.escape)="cancelAddRow()" />
                </td>
                <td class="add-cell text-right whitespace-nowrap">
                  <button (click)="submitAddRow()" class="btn-save">Save</button>
```

New:
```html
                  <input [(ngModel)]="addDuration" name="addDuration"
                    type="number" min="5" step="5" list="durationPresets"
                    class="edit-dur-input"
                    (keydown.enter)="submitAddRow()"
                    (keydown.escape)="cancelAddRow()" />
                </td>
                <td class="add-cell">
                  <input [(ngModel)]="addNotes" name="addNotes"
                    type="text" class="edit-notes-input"
                    [placeholder]="'slots.notes' | translate"
                    (keydown.enter)="submitAddRow()"
                    (keydown.escape)="cancelAddRow()" />
                </td>
                <td class="add-cell text-right whitespace-nowrap">
                  <button (click)="submitAddRow()" class="btn-save">Save</button>
```

- [ ] **Step 5: Update the add-trigger row colspan from 6 to 7**

Find the `@else` add-trigger row:

Old:
```html
              <tr class="add-trigger-row" (click)="activateAddRow()">
                <td colspan="6" class="add-trigger-cell">
                  + {{ 'slots.new' | translate }}
                </td>
              </tr>
```

New:
```html
              <tr class="add-trigger-row" (click)="activateAddRow()">
                <td colspan="7" class="add-trigger-cell">
                  + {{ 'slots.new' | translate }}
                </td>
              </tr>
```

- [ ] **Step 6: Add notes CSS**

Append to `event-detail.component.css`:
```css
/* Slot notes column */
.slot-notes-cell { @apply px-3 py-3 max-w-36; }
.slot-notes-text { @apply block text-xs text-gray-500 truncate cursor-help; }
.edit-notes-input { @apply text-sm px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-violet-400 w-28; }
```

- [ ] **Step 7: Verify the app compiles**

Run: `cd frontend && npx ng build --configuration=development 2>&1 | tail -20`
Expected: builds with no errors

- [ ] **Step 8: Verify in browser**

Open an event detail page. Confirm:
- The slots table has a Notes column header between Duration and Actions
- The add row has a Notes text input; typing a note and saving persists it
- A saved note shows truncated text in the read row; hovering shows the full note via tooltip
- Clicking Edit on a slot with a note pre-fills the notes input with the current value
- Editing the note and saving updates the displayed note

- [ ] **Step 9: Commit**

```bash
git add frontend/src/app/admin/events/event-detail.component.html frontend/src/app/admin/events/event-detail.component.css
git commit -m "feat(slots): notes column with inline add/edit inputs and truncated indicator (EL-031)"
```

---

## Task 8: Genre-First Bidirectional DJ Filtering (EL-030)

Makes the Genre and DJ dropdowns in the add row and edit row filter each other in both directions. Currently (after Task 5) the genre dropdown is disabled until a DJ is selected, and the DJ dropdown always lists all DJs. EL-030 requires: selecting a genre first narrows the DJ dropdown to DJs who have that genre; clearing the genre restores the full DJ list; selecting a DJ first still auto-fills the genre dropdown with that DJ's genres (the existing Task 5 behavior).

**Files:**
- Modify: `frontend/src/app/admin/events/event-detail.component.ts`
- Modify: `frontend/src/app/admin/events/event-detail.component.html`

**Interfaces:**
- Consumes (from Task 5): `addDjId`, `addGenre`, `editSlotDjId`, `editSlotGenre`, `djs()` signal, `djGenresForAdd()`, `djGenresForEdit()`
- Produces: `genreOptionsForAdd()`, `filteredDjsForAdd()`, `onAddDjChange()`, `onAddGenreChange()`, `genreOptionsForEdit()`, `filteredDjsForEdit()`, `onEditDjChange()`, `onEditGenreChange()`

- [ ] **Step 1: Add the bidirectional helper methods to `event-detail.component.ts`**

Replace the existing `djGenresForAdd()` method (added in Task 5):

Old:
```typescript
  djGenresForAdd(): string[] {
    return this.djs().find(d => d.id === this.addDjId)?.genre_tags ?? [];
  }
```

New:
```typescript
  // ── Genre ↔ DJ bidirectional filtering (add row) ──────────────────
  private allGenres(): string[] {
    const all = new Set<string>();
    this.djs().forEach(d => d.genre_tags.forEach(g => all.add(g)));
    return [...all].sort();
  }

  genreOptionsForAdd(): string[] {
    return this.addDjId
      ? (this.djs().find(d => d.id === this.addDjId)?.genre_tags ?? [])
      : this.allGenres();
  }

  filteredDjsForAdd(): DJ[] {
    return this.addGenre
      ? this.djs().filter(d => d.genre_tags.includes(this.addGenre))
      : this.djs();
  }

  onAddDjChange() {
    const djGenres = this.djs().find(d => d.id === this.addDjId)?.genre_tags ?? [];
    if (this.addGenre && this.addDjId && !djGenres.includes(this.addGenre)) {
      this.addGenre = '';
    }
  }

  onAddGenreChange() {
    const djGenres = this.djs().find(d => d.id === this.addDjId)?.genre_tags ?? [];
    if (this.addDjId && !djGenres.includes(this.addGenre)) {
      this.addDjId = '';
    }
  }
```

- [ ] **Step 2: Replace the edit-row helper method in `event-detail.component.ts`**

Replace the existing `djGenresForEdit()` method:

Old:
```typescript
  djGenresForEdit(): string[] {
    return this.djs().find(d => d.id === this.editSlotDjId)?.genre_tags ?? [];
  }
```

New:
```typescript
  // ── Genre ↔ DJ bidirectional filtering (edit row) ─────────────────
  genreOptionsForEdit(): string[] {
    return this.editSlotDjId
      ? (this.djs().find(d => d.id === this.editSlotDjId)?.genre_tags ?? [])
      : this.allGenres();
  }

  filteredDjsForEdit(): DJ[] {
    return this.editSlotGenre
      ? this.djs().filter(d => d.genre_tags.includes(this.editSlotGenre))
      : this.djs();
  }

  onEditDjChange() {
    const djGenres = this.djs().find(d => d.id === this.editSlotDjId)?.genre_tags ?? [];
    if (this.editSlotGenre && this.editSlotDjId && !djGenres.includes(this.editSlotGenre)) {
      this.editSlotGenre = '';
    }
  }

  onEditGenreChange() {
    const djGenres = this.djs().find(d => d.id === this.editSlotDjId)?.genre_tags ?? [];
    if (this.editSlotDjId && !djGenres.includes(this.editSlotGenre)) {
      this.editSlotDjId = '';
    }
  }
```

- [ ] **Step 3: Rewire the add-row DJ and Genre dropdowns in `event-detail.component.html`**

Find the add row DJ `<select>` and Genre `<select>`. Replace both:

Old:
```html
                <td class="add-cell">
                  <select [(ngModel)]="addDjId" name="addDjId"
                    (ngModelChange)="addGenre = ''" class="edit-select"
                    (keydown.escape)="cancelAddRow()">
                    <option value="">{{ 'slots.unassigned' | translate }}</option>
                    @for (d of djs(); track d.id) {
                      <option [value]="d.id">{{ d.name }}</option>
                    }
                  </select>
                </td>
                <td class="add-cell">
                  <select [(ngModel)]="addGenre" name="addGenre"
                    [disabled]="!addDjId || djGenresForAdd().length === 0"
                    class="edit-select"
                    (keydown.escape)="cancelAddRow()">
                    <option value="">
                      {{ addDjId ? (djGenresForAdd().length ? '— select —' : '—') : '— DJ first —' }}
                    </option>
                    @for (g of djGenresForAdd(); track g) {
                      <option [value]="g">{{ g }}</option>
                    }
                  </select>
                </td>
```

New:
```html
                <td class="add-cell">
                  <select [(ngModel)]="addDjId" name="addDjId"
                    (ngModelChange)="onAddDjChange()" class="edit-select"
                    (keydown.escape)="cancelAddRow()">
                    <option value="">{{ 'slots.unassigned' | translate }}</option>
                    @for (d of filteredDjsForAdd(); track d.id) {
                      <option [value]="d.id">{{ d.name }}</option>
                    }
                  </select>
                </td>
                <td class="add-cell">
                  <select [(ngModel)]="addGenre" name="addGenre"
                    (ngModelChange)="onAddGenreChange()"
                    class="edit-select"
                    (keydown.escape)="cancelAddRow()">
                    <option value="">—</option>
                    @for (g of genreOptionsForAdd(); track g) {
                      <option [value]="g">{{ g }}</option>
                    }
                  </select>
                </td>
```

- [ ] **Step 4: Rewire the edit-row DJ and Genre dropdowns in `event-detail.component.html`**

Find the edit row DJ `<select>` and Genre `<select>`. Replace both:

Old:
```html
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotDjId" [name]="'es_dj_' + slot.id"
                      (ngModelChange)="editSlotGenre = ''" class="edit-select"
                      (keydown.escape)="cancelEdit()">
                      <option value="">—</option>
                      @for (d of djs(); track d.id) {
                        <option [value]="d.id">{{ d.name }}</option>
                      }
                    </select>
                  </td>
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotGenre" [name]="'es_genre_' + slot.id"
                      [disabled]="!editSlotDjId || djGenresForEdit().length === 0"
                      class="edit-select"
                      (keydown.escape)="cancelEdit()">
                      <option value="">{{ djGenresForEdit().length ? '— select —' : '—' }}</option>
                      @for (g of djGenresForEdit(); track g) {
                        <option [value]="g">{{ g }}</option>
                      }
                    </select>
                  </td>
```

New:
```html
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotDjId" [name]="'es_dj_' + slot.id"
                      (ngModelChange)="onEditDjChange()" class="edit-select"
                      (keydown.escape)="cancelEdit()">
                      <option value="">—</option>
                      @for (d of filteredDjsForEdit(); track d.id) {
                        <option [value]="d.id">{{ d.name }}</option>
                      }
                    </select>
                  </td>
                  <td class="edit-cell">
                    <select [(ngModel)]="editSlotGenre" [name]="'es_genre_' + slot.id"
                      (ngModelChange)="onEditGenreChange()"
                      class="edit-select"
                      (keydown.escape)="cancelEdit()">
                      <option value="">—</option>
                      @for (g of genreOptionsForEdit(); track g) {
                        <option [value]="g">{{ g }}</option>
                      }
                    </select>
                  </td>
```

- [ ] **Step 5: Verify the app compiles**

Run: `cd frontend && npx ng build --configuration=development 2>&1 | tail -20`
Expected: builds with no errors

- [ ] **Step 6: Verify in browser**

Open an event detail page (DJs must have genre tags). In the add row:
- Select a genre with no DJ chosen → DJ dropdown narrows to DJs with that genre
- Clear the genre → DJ dropdown shows all DJs again
- Select a DJ first → genre dropdown shows only that DJ's genres
- Select a DJ, then pick an incompatible genre → DJ resets (bidirectional)
Repeat the same checks in the edit row (click Edit on an existing slot).

- [ ] **Step 7: Commit**

```bash
git add frontend/src/app/admin/events/event-detail.component.ts frontend/src/app/admin/events/event-detail.component.html
git commit -m "feat(slots): genre-first bidirectional DJ filtering in add and edit rows (EL-030)"
```
