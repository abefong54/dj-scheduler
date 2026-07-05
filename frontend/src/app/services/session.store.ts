import { Injectable, inject, signal } from '@angular/core';
import { ApiService, MeResponse, School, User } from './api.service';

/**
 * Root-provided store for the authenticated user's session (EL-106).
 *
 * Holds the current user and school in signals. `ensure()` fetches /api/me
 * exactly once (idempotent re-visits). `refresh()` re-fetches unconditionally
 * — call it after school creation or any mutation that changes the school.
 *
 * Mirrors the EventsStore pattern but holds a single object rather than a list,
 * so it owns the signal state directly instead of delegating to CachedList.
 */
@Injectable({ providedIn: 'root' })
export class SessionStore {
  private readonly api = inject(ApiService);

  private readonly _user = signal<User | null>(null);
  private readonly _school = signal<School | null>(null);
  private readonly _loading = signal(false);

  /** Whether /api/me has been requested at least once. */
  private _ensured = false;

  /** The authenticated user, or null before the first load. */
  readonly user = this._user.asReadonly();
  /** The user's school, or null if they have no school yet. */
  readonly school = this._school.asReadonly();
  /** True only while a fetch is in flight. */
  readonly loading = this._loading.asReadonly();

  /**
   * Kicks off a fetch of /api/me if one has never been started.
   * Subsequent calls are no-ops — use `refresh()` to force a re-fetch.
   */
  ensure(): void {
    if (this._ensured) return;
    this._fetch();
  }

  /**
   * Forces a re-fetch of /api/me regardless of prior loads.
   * Call after school creation so the school signal updates.
   */
  refresh(): void {
    this._fetch();
  }

  private _fetch(): void {
    this._ensured = true;
    this._loading.set(true);
    this.api.getMe().subscribe({
      next: (me: MeResponse) => {
        this._user.set(me.user);
        this._school.set(me.school);
        this._loading.set(false);
      },
      error: () => {
        this._loading.set(false);
      },
    });
  }
}
