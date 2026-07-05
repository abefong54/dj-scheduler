import { DestroyRef, Injectable, Signal, inject, signal } from '@angular/core';

/** Tailwind `md` — below this we treat the viewport as a phone. */
const COMPACT_QUERY = '(max-width: 767px)';

/**
 * Root-provided viewport signal (EL-094). `compact()` is true on phone-sized
 * screens, driven by `matchMedia`, so components can switch between desktop and
 * mobile presentations reactively (e.g. the planner's table vs. stacked cards).
 * SSR-safe: with no `window`/`matchMedia` it stays `false`.
 */
@Injectable({ providedIn: 'root' })
export class BreakpointService {
  private readonly _compact = signal(this.query(COMPACT_QUERY));
  /** True below the `md` breakpoint (≤767px). */
  readonly compact: Signal<boolean> = this._compact.asReadonly();

  constructor() {
    if (typeof window === 'undefined' || !window.matchMedia) return;
    const mql = window.matchMedia(COMPACT_QUERY);
    const onChange = (e: MediaQueryListEvent) => this._compact.set(e.matches);
    mql.addEventListener('change', onChange);
    inject(DestroyRef).onDestroy(() => mql.removeEventListener('change', onChange));
  }

  private query(q: string): boolean {
    return typeof window !== 'undefined' && !!window.matchMedia && window.matchMedia(q).matches;
  }
}
