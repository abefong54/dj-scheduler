import {
  Component,
  ElementRef,
  HostListener,
  effect,
  inject,
  input,
  model,
  output,
  signal,
  viewChild,
} from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';

export type ShellNav = 'events' | 'djs' | 'performance';

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
 * EL-093 makes it responsive: below `lg` the sidebar collapses into an
 * off-canvas drawer (☰) over a backdrop, and the topbar actions fold into a
 * `⋯` overflow menu; below `md` the search drops to its own row. Desktop
 * (≥lg) is visually unchanged — the responsive rules live in global
 * `styles.css` (`@media (max-width: …)`) and only add behaviour below lg.
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
 *   [shell-actions]    — extra topbar buttons (fold into the ⋯ menu below lg)
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

  /** Off-canvas nav drawer open state (only meaningful <lg; static ≥lg). */
  protected readonly drawerOpen = signal(false);
  /** Topbar overflow (⋯) menu open state (only meaningful <lg). */
  protected readonly actionsOpen = signal(false);

  private readonly host = inject(ElementRef<HTMLElement>);
  private readonly hamburger = viewChild<ElementRef<HTMLButtonElement>>('hamburger');

  constructor() {
    // Lock body scroll while the drawer covers the page, so the content behind
    // it doesn't scroll under the overlay.
    effect(() => {
      if (typeof document !== 'undefined') {
        document.body.classList.toggle('shell-drawer-lock', this.drawerOpen());
      }
    });
  }

  protected readonly navItems: readonly NavItem[] = [
    { id: 'events', labelKey: 'shell.nav.events', link: '/admin/events' },
    { id: 'djs', labelKey: 'shell.nav.djs', link: '/admin/djs' },
    { id: 'performance', labelKey: 'shell.nav.performance', link: '/admin/performance' },
  ];

  protected isActive(id: ShellNav): boolean {
    return this.activeNav() === id;
  }

  protected onSearch(event: Event): void {
    this.searchValue.set((event.target as HTMLInputElement).value);
  }

  protected toggleDrawer(): void {
    this.drawerOpen.update((open) => !open);
  }

  protected closeDrawer(): void {
    if (!this.drawerOpen()) return;
    this.drawerOpen.set(false);
    // Return focus to the trigger for keyboard/backdrop/Escape closes.
    this.hamburger()?.nativeElement.focus();
  }

  protected toggleActions(): void {
    this.actionsOpen.update((open) => !open);
  }

  protected closeActions(): void {
    this.actionsOpen.set(false);
  }

  /** Escape closes whichever overlay is open. */
  @HostListener('document:keydown.escape')
  protected onEscape(): void {
    this.closeActions();
    this.closeDrawer();
  }

  /** A click outside the ⋯ toggle and its menu dismisses the overflow menu. */
  @HostListener('document:click', ['$event'])
  protected onDocumentClick(event: MouseEvent): void {
    if (!this.actionsOpen()) return;
    const target = event.target as Node;
    const el = this.host.nativeElement as HTMLElement;
    const toggle = el.querySelector('.shell-actions-toggle');
    const group = el.querySelector('.shell-actions-group');
    if (toggle?.contains(target) || group?.contains(target)) return;
    this.closeActions();
  }

  /** Back at desktop width the sidebar is static again — clear any open overlay. */
  @HostListener('window:resize')
  protected onResize(): void {
    if (typeof window !== 'undefined' && window.innerWidth >= 1024) {
      this.drawerOpen.set(false);
      this.actionsOpen.set(false);
    }
  }
}
