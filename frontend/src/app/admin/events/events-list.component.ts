import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService, Event } from '../../services/api.service';
import { DialogService } from '../../shared/dialog.service';

@Component({
  selector: 'app-events-list',
  standalone: true,
  imports: [RouterLink, TranslatePipe],
  templateUrl: './events-list.component.html',
  styleUrl: './events-list.component.css',
})
export class EventsListComponent {
  private api = inject(ApiService);
  private translate = inject(TranslateService);
  private dialog = inject(DialogService);

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

  duplicate(id: string) {
    this.api.duplicateEvent(id).subscribe(() => this.loadEvents());
  }
}
