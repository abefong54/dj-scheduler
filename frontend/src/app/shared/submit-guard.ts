import { signal } from '@angular/core';
import { finalize, Observable, Observer } from 'rxjs';

/**
 * Guards a mutation against double-submission (BUG-008): a double-clicked
 * Create/Save button firing two requests before the first returns, which
 * duplicates records. `run` ignores a call while a previous one is still in
 * flight; `pending` drives `[disabled]` on the triggering button.
 *
 * Standardises the ad-hoc guard already used in landing/lead-form.
 */
export class SubmitGuard {
  readonly pending = signal(false);

  run<T>(source: Observable<T>, observer?: Partial<Observer<T>>): void {
    if (this.pending()) return;
    this.pending.set(true);
    source.pipe(finalize(() => this.pending.set(false))).subscribe(observer);
  }
}
