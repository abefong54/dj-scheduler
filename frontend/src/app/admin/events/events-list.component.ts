import { Component, computed, inject, linkedSignal, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService, Event } from '../../services/api.service';
import { EventsStore } from '../../services/events-store';
import { AdminShellComponent } from '../../shared/admin-shell.component';
import { StatusBadgeComponent } from '../../shared/status-badge.component';
import { ButtonComponent } from '../../shared/button.component';
import { LoadingMeterComponent } from '../../shared/loading-meter.component';
import { DataTableComponent, TableColumn } from '../../shared/data-table.component';
import { ColumnDefDirective } from '../../shared/column-def.directive';
import { EditCellComponent } from '../../shared/edit-cell.component';
import { DialogService } from '../../shared/dialog.service';
import { SubmitGuard } from '../../shared/submit-guard';

/** Lifecycle state derived purely from an event's start/end dates. */
type EventLifecycle = 'live' | 'upcoming' | 'past';

@Component({
  selector: 'app-events-list',
  standalone: true,
  imports: [RouterLink, TranslatePipe, AdminShellComponent, StatusBadgeComponent, ButtonComponent, LoadingMeterComponent, DataTableComponent, ColumnDefDirective, EditCellComponent],
  templateUrl: './events-list.component.html',
  styleUrl: './events-list.component.css',
})
export class EventsListComponent {
  private api = inject(ApiService);
  private translate = inject(TranslateService);
  private dialog = inject(DialogService);
  private router = inject(Router);
  private store = inject(EventsStore);

  activeTab = signal<'upcoming' | 'past'>('upcoming');
  today = new Date().toISOString().slice(0, 10);
  readonly cloneGuard = new SubmitGuard();

  // Backed by the root EventsStore (EL-091): a re-visit paints from cache with no
  // spinner while the list revalidates in the background.
  // linkedSignal keeps this writable so inline edits can update it optimistically
  // without a full re-fetch, while still resetting whenever the store refreshes.
  events = linkedSignal(() => this.store.events());
  loading = this.store.loading;
  loadError = this.store.error;

  upcoming = computed(() => this.events().filter(e => e.end_date >= this.today));
  past = computed(() => this.events().filter(e => e.end_date < this.today));

  constructor() {
    this.store.ensure();
  }

  retry() {
    this.store.ensure();
  }

  formatDates(e: Event): string {
    if (e.start_date === e.end_date) return e.start_date;
    return `${e.start_date} – ${e.end_date}`;
  }

  /** Date-derived lifecycle: live while running, otherwise upcoming or past. */
  lifecycle(e: Event): EventLifecycle {
    if (e.start_date <= this.today && this.today <= e.end_date) return 'live';
    return e.end_date < this.today ? 'past' : 'upcoming';
  }

  /** Only currently-running events carry a real status badge (no fabricated state). */
  isLive(e: Event): boolean {
    return this.lifecycle(e) === 'live';
  }

  goToNew() {
    this.router.navigate(['/admin/events/new']);
  }

  edit(id: string) {
    this.router.navigate(['/admin/events', id]);
  }

  async delete(id: string) {
    const ok = await this.dialog.confirm({
      title: this.translate.instant('dialog.deleteTitle'),
      message: this.translate.instant('events.deleteConfirm'),
      confirmLabel: this.translate.instant('actions.delete'),
      variant: 'danger',
    });
    if (!ok) return;
    this.api.deleteEvent(id).subscribe(() => this.store.refresh());
  }

  async clone(id: string) {
    const ok = await this.dialog.confirm({
      title: this.translate.instant('events.clone'),
      message: this.translate.instant('events.clone.confirm'),
      confirmLabel: this.translate.instant('events.clone'),
    });
    if (!ok) return;
    this.cloneGuard.run(this.api.cloneEvent(id), {
      next: newEvent => {
        this.store.refresh(); // keep the cached list current for the return visit
        this.router.navigate(['/admin/events', newEvent.id]);
      },
    });
  }

  // ── Inline-edit table (Task 3) ────────────────────────────────────────────

  activeTabEvents = computed(() => (this.activeTab() === 'upcoming' ? this.upcoming() : this.past()));

  rowsForActiveTab = computed((): Record<string, unknown>[] =>
    this.activeTabEvents().map(e => ({
      id: e.id,
      name: e.name,
      venue_name: e.venue_name,
      start_date: e.start_date,
      end_date: e.end_date,
      genres: e.genres,
      formattedDates: this.formatDates(e),
      isLive: this.isLive(e),
      lifecycle: this.lifecycle(e),
    }) as Record<string, unknown>));

  eventColumns = computed((): TableColumn[] => [
    { key: 'name', label: this.translate.instant('events.col.event') },
    { key: 'venue_name', label: this.translate.instant('events.col.venue') },
    { key: 'start_date', label: this.translate.instant('events.col.dates') },
    { key: 'genres', label: this.translate.instant('events.col.genres') },
    { key: 'status', label: this.translate.instant('events.col.status'), sortable: false, searchable: false },
    { key: 'actions', label: this.translate.instant('events.col.actions'), sortable: false, searchable: false },
  ]);

  asString(v: unknown): string { return v == null ? '' : String(v); }
  genresString(row: Record<string, unknown>): string {
    return Array.isArray(row['genres']) ? (row['genres'] as string[]).join(', ') : '';
  }

  commitName(row: Record<string, unknown>, e: { value: string; endValue: string }) {
    this.save(this.asString(row['id']), { name: e.value.trim() });
  }
  commitVenue(row: Record<string, unknown>, e: { value: string; endValue: string }) {
    this.save(this.asString(row['id']), { venue_name: e.value.trim() });
  }
  commitDates(row: Record<string, unknown>, e: { value: string; endValue: string }) {
    this.save(this.asString(row['id']), { start_date: e.value, end_date: e.endValue || e.value });
  }
  commitGenres(row: Record<string, unknown>, e: { value: string; endValue: string }) {
    const genres = e.value.split(',').map(g => g.trim()).filter(g => g.length > 0);
    this.save(this.asString(row['id']), { genres });
  }

  private save(id: string, patch: Partial<Omit<Event, 'id'>>) {
    const ev = this.events().find(e => e.id === id);
    if (!ev) return;
    const body: Omit<Event, 'id'> = {
      name: ev.name, venue_name: ev.venue_name, start_date: ev.start_date,
      end_date: ev.end_date, genres: ev.genres, ...patch,
    };
    this.api.updateEvent(id, body).subscribe({
      next: updated => {
        this.events.update(list => list.map(e => e.id === id ? updated : e));
        this.store.refresh(); // keep the cached list current for the return visit
      },
      error: () => this.dialog.alert({
        title: this.translate.instant('common.loadError'),
        message: this.translate.instant('common.loadError'),
      }),
    });
  }
}
