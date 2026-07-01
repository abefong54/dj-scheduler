import { Component, computed, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService, Event } from '../../services/api.service';
import { AdminShellComponent } from '../../shared/admin-shell.component';
import { StatusBadgeComponent } from '../../shared/status-badge.component';
import { ButtonComponent } from '../../shared/button.component';
import { DialogService } from '../../shared/dialog.service';

/** Lifecycle state derived purely from an event's start/end dates. */
type EventLifecycle = 'live' | 'upcoming' | 'past';

@Component({
  selector: 'app-events-list',
  standalone: true,
  imports: [RouterLink, TranslatePipe, AdminShellComponent, StatusBadgeComponent, ButtonComponent],
  templateUrl: './events-list.component.html',
  styleUrl: './events-list.component.css',
})
export class EventsListComponent {
  private api = inject(ApiService);
  private translate = inject(TranslateService);
  private dialog = inject(DialogService);
  private router = inject(Router);

  activeTab = signal<'upcoming' | 'past'>('upcoming');
  today = new Date().toISOString().slice(0, 10);

  events = signal<Event[]>([]);

  upcoming = computed(() => this.events().filter(e => e.end_date >= this.today));
  past = computed(() => this.events().filter(e => e.end_date < this.today));

  constructor() {
    this.loadEvents();
  }

  private loadEvents() {
    this.api.getEvents().subscribe(events => this.events.set(events));
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
    this.api.deleteEvent(id).subscribe(() => this.loadEvents());
  }

  async clone(id: string) {
    const ok = await this.dialog.confirm({
      title: this.translate.instant('events.clone'),
      message: this.translate.instant('events.clone.confirm'),
      confirmLabel: this.translate.instant('events.clone'),
    });
    if (!ok) return;
    this.api.cloneEvent(id).subscribe(newEvent =>
      this.router.navigate(['/admin/events', newEvent.id]),
    );
  }
}
