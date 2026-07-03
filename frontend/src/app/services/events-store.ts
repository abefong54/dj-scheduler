import { Injectable, computed, inject } from '@angular/core';
import { ApiService, Event } from './api.service';
import { CachedList } from '../shared/cached-list';

/**
 * Root-provided stale-while-revalidate cache for the organizer's events list
 * (EL-091). Backs `/admin/events` so a re-visit paints instantly instead of
 * re-fetching behind a spinner. Every write path must call `refresh()` to keep
 * the cache honest — create/delete/clone and inline updateEvent edits (Task 3)
 * are all invalidation points.
 */
@Injectable({ providedIn: 'root' })
export class EventsStore {
  private readonly api = inject(ApiService);
  private readonly list = new CachedList<Event>(() => this.api.getEvents());

  /** The cached events (empty until the first load resolves). */
  readonly events = computed(() => this.list.value() ?? []);
  /** True only while the first-ever load is in flight. */
  readonly loading = this.list.loading;
  /** True only when the first-ever load failed. */
  readonly error = this.list.error;

  /** Paint from cache and revalidate in the background. Call on list entry. */
  ensure(): void {
    this.list.ensure();
  }

  /** Force a re-fetch after a create/delete/clone. */
  refresh(): void {
    this.list.refresh();
  }
}
