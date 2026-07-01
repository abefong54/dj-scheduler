import { Component, input } from '@angular/core';

/**
 * Soundcheck loading indicator (EL-078) — "levels, not fades".
 *
 * A row of VU-meter bars that rise and fall instead of a spinning ring, per the
 * visual-identity direction §6. This establishes the app's loading/progress
 * convention: progress reads as audio levels. `prefers-reduced-motion` is
 * honoured in styles.css — the bars hold at a static filled level rather than
 * animating.
 *
 *   <app-loading-meter [label]="'Loading…'" />
 *
 * Visual styles live in the global components layer of styles.css (.vu-loader*),
 * keeping the per-component anyComponentStyle budget clear (no @apply here).
 */
@Component({
  selector: 'app-loading-meter',
  standalone: true,
  template: `
    <span class="vu-loader" role="status" [attr.aria-label]="label() || null">
      <span class="vu-loader-bar"></span>
      <span class="vu-loader-bar"></span>
      <span class="vu-loader-bar"></span>
      <span class="vu-loader-bar"></span>
      @if (label()) {
        <span class="vu-loader-label">{{ label() }}</span>
      }
    </span>
  `,
})
export class LoadingMeterComponent {
  /** Optional accessible/visible label shown beside the meter. */
  label = input<string>('');
}
