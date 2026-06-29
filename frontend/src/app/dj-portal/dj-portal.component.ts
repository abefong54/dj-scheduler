import { Component, computed, inject, signal } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { ApiService, DJPortalResponse, DJPortalSlot } from '../services/api.service';

interface EventGroup {
  eventName: string;
  slots: DJPortalSlot[];
}

@Component({
  selector: 'app-dj-portal',
  standalone: true,
  imports: [TranslatePipe],
  templateUrl: './dj-portal.component.html',
  styleUrl: './dj-portal.component.css',
})
export class DJPortalComponent {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private token = '';

  dj = signal<DJPortalResponse['dj'] | null>(null);
  slots = signal<DJPortalSlot[]>([]);
  error = signal<'expired' | 'unknown' | null>(null);
  loading = signal(true);

  // Slots grouped by event, each group sorted by date then start time ascending.
  slotsByEvent = computed<EventGroup[]>(() => {
    const groups = new Map<string, DJPortalSlot[]>();
    for (const slot of this.slots()) {
      const list = groups.get(slot.event_name) ?? [];
      list.push(slot);
      groups.set(slot.event_name, list);
    }
    return [...groups.entries()].map(([eventName, slots]) => ({
      eventName,
      slots: slots.sort((a, b) =>
        (a.slot_date + a.start_time).localeCompare(b.slot_date + b.start_time)),
    }));
  });

  constructor() {
    this.token = this.route.snapshot.queryParams['token'] ?? '';
    if (this.token) {
      // EL-038: drop the token from the URL once read, so it doesn't linger in
      // the address bar, browser history, or any Referer header.
      history.replaceState(history.state, '', window.location.pathname);
    }
    if (!this.token) {
      this.error.set('expired');
      this.loading.set(false);
      return;
    }
    this.api.getDJPortal(this.token).subscribe({
      next: (data) => {
        this.dj.set(data.dj);
        this.slots.set(data.slots);
        this.loading.set(false);
      },
      error: (err) => {
        this.error.set(err.status === 401 ? 'expired' : 'unknown');
        this.loading.set(false);
      },
    });
  }

  confirm(slotId: string, action: 'confirmed' | 'flagged') {
    this.api.confirmSlot(slotId, action, this.token).subscribe((updated) => {
      this.slots.update((slots) =>
        slots.map((s) => (s.id === slotId ? { ...s, dj_confirmation: updated.dj_confirmation } : s)));
    });
  }
}
