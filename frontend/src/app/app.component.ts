import { Component, ElementRef, HostListener, inject, signal, viewChild } from '@angular/core';
import { NavigationEnd, Router, RouterLink, RouterOutlet } from '@angular/router';
import { toSignal } from '@angular/core/rxjs-interop';
import { filter, map, startWith } from 'rxjs';
import { TranslatePipe } from '@ngx-translate/core';
import { LanguageService } from './services/language.service';
import { AuthService } from './services/auth.service';
import { ThemeService } from './services/theme.service';
import { DialogComponent } from './shared/dialog.component';
import { GlobalLoadingComponent } from './shared/global-loading.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, TranslatePipe, DialogComponent, GlobalLoadingComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {
  langService = inject(LanguageService);
  auth = inject(AuthService);
  // Instantiate the Soundcheck theme controller so the mode attribute stays in
  // sync with the app root (dark "booth" default; EL-078).
  private readonly theme = inject(ThemeService);

  private readonly router = inject(Router);
  private readonly host = inject(ElementRef<HTMLElement>);
  // Below lg the secondary nav items collapse behind this hamburger; focus
  // returns here on close (keyboard/outside-click/Escape).
  private readonly menuToggle = viewChild<ElementRef<HTMLButtonElement>>('menuToggle');

  /** Mobile nav dropdown open state (only meaningful <lg; inline ≥lg). */
  protected readonly menuOpen = signal(false);

  protected toggleMenu(): void {
    this.menuOpen.update((open) => !open);
  }

  protected closeMenu(): void {
    if (!this.menuOpen()) return;
    this.menuOpen.set(false);
    this.menuToggle()?.nativeElement.focus();
  }

  /** Escape closes the open dropdown. */
  @HostListener('document:keydown.escape')
  protected onEscape(): void {
    this.closeMenu();
  }

  /** A click outside the toggle and the menu dismisses the dropdown. */
  @HostListener('document:click', ['$event'])
  protected onDocumentClick(event: MouseEvent): void {
    if (!this.menuOpen()) return;
    const target = event.target as Node;
    const el = this.host.nativeElement as HTMLElement;
    const toggle = el.querySelector('.navbar-menu-toggle');
    const menu = el.querySelector('#navbar-menu');
    if (toggle?.contains(target) || menu?.contains(target)) return;
    this.closeMenu();
  }

  /** Back at desktop width the items are inline again — clear any open dropdown. */
  @HostListener('window:resize')
  protected onResize(): void {
    if (typeof window !== 'undefined' && window.innerWidth >= 1024) {
      this.menuOpen.set(false);
    }
  }

  readonly isLanding = toSignal(
    this.router.events.pipe(
      filter((e) => e instanceof NavigationEnd),
      map(() => this.onLandingRoute()),
      startWith(this.onLandingRoute()),
    ),
    { initialValue: this.onLandingRoute() },
  );
  private onLandingRoute(): boolean {
    return this.router.url.split(/[?#]/)[0] === '/';
  }
}
