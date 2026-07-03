import { Injectable, computed, signal } from '@angular/core';

/**
 * Tracks how many HTTP requests are in flight so the UI can show a global
 * progress indicator. A counter (not a boolean) so concurrent requests — e.g.
 * event-detail fires four at once — don't clear the flag prematurely.
 */
@Injectable({ providedIn: 'root' })
export class LoadingService {
  private readonly active = signal(0);

  /** True whenever at least one tracked request is in flight. */
  readonly isLoading = computed(() => this.active() > 0);

  start(): void {
    this.active.update((n) => n + 1);
  }

  stop(): void {
    // Guard against underflow so a stray stop() can't wedge isLoading at false.
    this.active.update((n) => Math.max(0, n - 1));
  }
}
