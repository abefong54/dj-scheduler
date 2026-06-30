import { Component, input, model, output } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

export type ShellNav = 'events' | 'djs' | 'schedule';

interface NavItem {
  readonly id: ShellNav;
  readonly labelKey: string;
  readonly link: string;
}

/**
 * The Console app shell (EL-061) — identical chrome on every admin page, per
 * design-system §5. A 240px dark glowing sidebar (nav + footer) framing a calm
 * light workspace, with a white topbar (page title, breadcrumb, search,
 * `+ New`, avatar).
 *
 * Usage:
 *   <app-admin-shell activeNav="events" title="Events" breadcrumb="WORKSPACE">
 *     <button shell-actions class="...">extra action</button>
 *     <!-- default-projected page content goes here -->
 *   </app-admin-shell>
 *
 * Slots:
 *   default            — page content (rendered in the 1200px light surface)
 *   [shell-title]      — rich page-title node (overrides the `title` input)
 *   [shell-actions]    — extra topbar buttons (left of the avatar)
 */
@Component({
  selector: 'app-admin-shell',
  standalone: true,
  imports: [RouterLink, TranslatePipe],
  templateUrl: './admin-shell.component.html',
  styleUrl: './admin-shell.component.css',
})
export class AdminShellComponent {
  /** Which nav item renders as active. `null` = none active. */
  activeNav = input<ShellNav | null>(null);
  /** Page title (Space Grotesk). Ignored if a `[shell-title]` node is projected. */
  title = input('');
  /** Mono muted breadcrumb shown under the title. */
  breadcrumb = input('');
  /** Organisation name shown in the sidebar footer (mono). */
  orgName = input('');
  /**
   * When non-null, renders the violet `+ New` topbar button. Empty string uses
   * the default translated "New" label; any other string is used verbatim.
   */
  newLabel = input<string | null>(null);

  /** Two-way bound topbar search text. */
  searchValue = model('');
  /** Emitted when the `+ New` button is clicked. */
  newClicked = output<void>();

  protected readonly navItems: readonly NavItem[] = [
    { id: 'events', labelKey: 'shell.nav.events', link: '/admin/events' },
    { id: 'djs', labelKey: 'shell.nav.djs', link: '/admin/djs' },
    { id: 'schedule', labelKey: 'shell.nav.schedule', link: '/admin/schedule' },
  ];

  protected isActive(id: ShellNav): boolean {
    return this.activeNav() === id;
  }

  protected onSearch(event: Event): void {
    this.searchValue.set((event.target as HTMLInputElement).value);
  }
}
