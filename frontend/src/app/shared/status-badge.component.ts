import { Component, computed, input } from '@angular/core';
import { TranslatePipe } from '@ngx-translate/core';

export type BadgeStatus = 'confirmed' | 'pending' | 'declined' | 'live';

/**
 * Console status pill (EL-061). Mono uppercase label + tinted background + a
 * solid status dot, per design-system §6. Reused across pages for slot/event
 * confirmation state.
 *
 *   <app-status-badge status="confirmed" />
 */
@Component({
  selector: 'app-status-badge',
  standalone: true,
  imports: [TranslatePipe],
  templateUrl: './status-badge.component.html',
  styleUrl: './status-badge.component.css',
})
export class StatusBadgeComponent {
  /** One of CONFIRMED / PENDING / DECLINED / LIVE. */
  status = input.required<BadgeStatus>();

  protected readonly tintClass = computed(() => `status-${this.status()}`);
  protected readonly labelKey = computed(() => `badge.${this.status()}`);
}
