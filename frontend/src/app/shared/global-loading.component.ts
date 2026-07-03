import { Component, inject } from '@angular/core';
import { toObservable, toSignal } from '@angular/core/rxjs-interop';
import { map, of, switchMap, timer } from 'rxjs';
import { LoadingService } from '../services/loading.service';
import { LoadingMeterComponent } from './loading-meter.component';

/**
 * App-wide loading indicator (EL-087): the branded VU-meter ("levels, not fades")
 * pinned to a fixed corner, shown while any HTTP request is in flight — so reads
 * and mutations (save/delete) both get feedback. Rendered once at the app root.
 *
 * Visibility is debounced 150ms so fast (co-located DB) requests don't flash it;
 * going idle hides it immediately (switchMap cancels the pending timer).
 */
@Component({
  selector: 'app-global-loading',
  standalone: true,
  imports: [LoadingMeterComponent],
  template: `
    @if (visible()) {
      <div
        class="fixed right-4 top-4 z-50 rounded-lg px-3 py-2 shadow-lg"
        style="background: var(--card); border: 1px solid var(--border);"
        data-testid="global-loading"
      >
        <app-loading-meter />
      </div>
    }
  `,
})
export class GlobalLoadingComponent {
  private readonly loading = inject(LoadingService);

  readonly visible = toSignal(
    toObservable(this.loading.isLoading).pipe(
      switchMap((busy) => (busy ? timer(150).pipe(map(() => true)) : of(false))),
    ),
    { initialValue: false },
  );
}
