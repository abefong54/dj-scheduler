import { Signal, computed, signal } from '@angular/core';
import { Observable } from 'rxjs';

/**
 * A stale-while-revalidate cache for a single list resource (EL-091).
 *
 * Holds the last fetched list in a signal so a re-visit paints instantly, while
 * a background revalidation keeps it fresh. The loading spinner shows only on the
 * very first load (nothing cached yet); background refreshes are silent, and a
 * failed background refresh keeps the stale data rather than blanking the screen.
 *
 * Root-provide the store that owns one of these so it outlives component
 * instances and survives navigation.
 */
export class CachedList<T> {
  private readonly data = signal<T[] | null>(null); // null = never loaded
  private readonly revalidating = signal(false);
  private readonly failed = signal(false);

  /** The cached list, or null before the first successful load. */
  readonly value: Signal<T[] | null> = this.data.asReadonly();
  /** True only while the first-ever load is in flight (no cache to show). */
  readonly loading = computed(() => this.data() === null && this.revalidating());
  /** True only when the first-ever load failed (no cache to fall back on). */
  readonly error = computed(() => this.data() === null && this.failed());

  constructor(private fetcher: () => Observable<T[]>) {}

  /** Serve the cache immediately if present; always kick a background revalidate. */
  ensure(): void {
    this.revalidating.set(true);
    this.failed.set(false);
    this.fetcher().subscribe({
      next: (v) => {
        this.data.set(v);
        this.revalidating.set(false);
      },
      error: () => {
        // Keep any stale data; only a first-load failure surfaces via error().
        this.failed.set(true);
        this.revalidating.set(false);
      },
    });
  }

  /** Explicit re-fetch after a mutation. */
  refresh(): void {
    this.ensure();
  }

  /** Drop the cache so the next ensure() shows the spinner again. */
  invalidate(): void {
    this.data.set(null);
  }
}
