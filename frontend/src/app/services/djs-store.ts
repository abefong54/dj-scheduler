import { Injectable, computed, inject } from '@angular/core';
import { ApiService, DJ } from './api.service';
import { CachedList } from '../shared/cached-list';

/**
 * Root-provided stale-while-revalidate cache for the DJ roster (EL-092, reusing
 * the EL-091 primitive). Backs `/admin/djs` so a re-visit paints instantly
 * instead of re-fetching behind a spinner. Every roster write must call
 * `refresh()` — create/update/delete is the complete set; generating a portal
 * token does not change the roster and must not refresh.
 */
@Injectable({ providedIn: 'root' })
export class DjsStore {
  private readonly api = inject(ApiService);
  private readonly list = new CachedList<DJ>(() => this.api.getDJs());

  /** The cached roster (empty until the first load resolves). */
  readonly djs = computed(() => this.list.value() ?? []);
  /** True only while the first-ever load is in flight. */
  readonly loading = this.list.loading;
  /** True only when the first-ever load failed. */
  readonly error = this.list.error;

  /** Paint from cache and revalidate in the background. Call on roster entry. */
  ensure(): void {
    this.list.ensure();
  }

  /** Force a re-fetch after a create/update/delete. */
  refresh(): void {
    this.list.refresh();
  }
}
