import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';
import { signal } from '@angular/core';

import { AppComponent } from './app.component';
import { AuthService } from './services/auth.service';

describe('AppComponent top bar', () => {
  let isAuthenticated: ReturnType<typeof signal<boolean>>;

  async function build(authed: boolean): Promise<ComponentFixture<AppComponent>> {
    isAuthenticated = signal(authed);
    await TestBed.configureTestingModule({
      imports: [AppComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        {
          provide: AuthService,
          useValue: {
            isAuthenticated,
            signOut: vi.fn(),
            signInWithGoogle: vi.fn(),
            token: signal<string | null>(null),
          },
        },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(AppComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('shows a Sign in control and no Sign out when logged out', async () => {
    const el = (await build(false)).nativeElement as HTMLElement;
    expect(el.querySelector('[data-testid="signin"]')).toBeTruthy();
    expect(el.querySelector('[data-testid="signout"]')).toBeNull();
  });

  it('shows Sign out and no Sign in when logged in', async () => {
    const el = (await build(true)).nativeElement as HTMLElement;
    expect(el.querySelector('[data-testid="signout"]')).toBeTruthy();
    expect(el.querySelector('[data-testid="signin"]')).toBeNull();
  });

  it('routes the Sign in control to /login', async () => {
    const el = (await build(false)).nativeElement as HTMLElement;
    const signin = el.querySelector<HTMLAnchorElement>('[data-testid="signin"]');
    expect(signin?.getAttribute('href')).toBe('/login');
  });

  it('shows the Soundcheck wordmark with its amber VU-meter detail (EL-078)', async () => {
    const el = (await build(false)).nativeElement as HTMLElement;
    const brand = el.querySelector<HTMLElement>('[data-testid="brand-wordmark"]');
    expect(brand?.textContent).toContain('soundcheck');
    // The single amber VU detail is an inline SVG, not a dependency on the
    // finalized logo asset (EL-071).
    expect(brand?.querySelector('svg.shell-vu')).toBeTruthy();
    expect(brand?.querySelectorAll('.shell-vu-bar').length).toBe(3);
  });

  it('links the Soundcheck wordmark to the landing route', async () => {
    const el = (await build(false)).nativeElement as HTMLElement;
    const brand = el.querySelector<HTMLAnchorElement>('[data-testid="brand-wordmark"]');
    expect(brand?.tagName).toBe('A');
    expect(brand?.getAttribute('href')).toBe('/');
    // Routed through the translate pipe (i18n); with no translations loaded in
    // this TestBed, the pipe renders the raw key — which still proves the
    // aria-label binding is wired up correctly.
    expect(brand?.getAttribute('aria-label')).toBe('nav.home');
  });

  describe('nav "Book a demo" button (landing-only)', () => {
    async function buildWithRouter(
      authed: boolean,
    ): Promise<{ fixture: ComponentFixture<AppComponent>; router: Router }> {
      isAuthenticated = signal(authed);
      await TestBed.configureTestingModule({
        imports: [AppComponent],
        providers: [
          provideRouter([
            { path: '', children: [] },
            { path: 'login', children: [] },
          ]),
          provideTranslateService(),
          {
            provide: AuthService,
            useValue: {
              isAuthenticated,
              signOut: vi.fn(),
              signInWithGoogle: vi.fn(),
              token: signal<string | null>(null),
            },
          },
        ],
      }).compileComponents();
      const fixture = TestBed.createComponent(AppComponent);
      const router = TestBed.inject(Router);
      fixture.detectChanges();
      return { fixture, router };
    }

    it('shows the Book a demo button on the landing route (/)', async () => {
      const { fixture, router } = await buildWithRouter(false);
      await router.navigate(['/']);
      fixture.detectChanges();
      const el = fixture.nativeElement as HTMLElement;
      expect(el.querySelector('[data-testid="nav-book-demo"]')).toBeTruthy();
    });

    it('does not show the Book a demo button on a non-landing route (/login)', async () => {
      const { fixture, router } = await buildWithRouter(false);
      await router.navigate(['/login']);
      fixture.detectChanges();
      const el = fixture.nativeElement as HTMLElement;
      expect(el.querySelector('[data-testid="nav-book-demo"]')).toBeNull();
    });

    it('Sign in link still renders when unauthenticated on any route', async () => {
      const { fixture, router } = await buildWithRouter(false);
      await router.navigate(['/login']);
      fixture.detectChanges();
      const el = fixture.nativeElement as HTMLElement;
      expect(el.querySelector('[data-testid="signin"]')).toBeTruthy();
    });

    // A logged-in visitor who lands on the marketing page needs a way into their
    // console. On the landing route, the primary CTA becomes "Go to console".
    it('shows a Go-to-console link to /admin/events for an authenticated landing visitor', async () => {
      const { fixture, router } = await buildWithRouter(true);
      await router.navigate(['/']);
      fixture.detectChanges();
      const link = (fixture.nativeElement as HTMLElement).querySelector<HTMLAnchorElement>(
        '[data-testid="nav-console"]',
      );
      expect(link).toBeTruthy();
      expect(link?.getAttribute('href')).toBe('/admin/events');
    });

    it('replaces Book-a-demo with Go-to-console when authenticated on the landing route', async () => {
      const { fixture, router } = await buildWithRouter(true);
      await router.navigate(['/']);
      fixture.detectChanges();
      const el = fixture.nativeElement as HTMLElement;
      expect(el.querySelector('[data-testid="nav-console"]')).toBeTruthy();
      expect(el.querySelector('[data-testid="nav-book-demo"]')).toBeNull();
    });

    it('does not show Go-to-console for an unauthenticated landing visitor', async () => {
      const { fixture, router } = await buildWithRouter(false);
      await router.navigate(['/']);
      fixture.detectChanges();
      const el = fixture.nativeElement as HTMLElement;
      expect(el.querySelector('[data-testid="nav-console"]')).toBeNull();
      expect(el.querySelector('[data-testid="nav-book-demo"]')).toBeTruthy();
    });

    it('does not show Go-to-console off the landing route even when authenticated', async () => {
      const { fixture, router } = await buildWithRouter(true);
      await router.navigate(['/login']);
      fixture.detectChanges();
      const el = fixture.nativeElement as HTMLElement;
      expect(el.querySelector('[data-testid="nav-console"]')).toBeNull();
    });
  });

  describe('mobile hamburger menu', () => {
    it('starts closed — the menu has no open class', async () => {
      const el = (await build(false)).nativeElement as HTMLElement;
      const menu = el.querySelector('[data-testid="nav-menu"]');
      expect(menu).toBeTruthy();
      expect(menu?.classList.contains('navbar-menu-open')).toBe(false);
    });

    it('the hamburger toggle opens and closes the menu', async () => {
      const fixture = await build(false);
      const el = fixture.nativeElement as HTMLElement;
      const toggle = el.querySelector<HTMLButtonElement>('[data-testid="nav-menu-toggle"]');
      const menu = el.querySelector('[data-testid="nav-menu"]');

      toggle!.click();
      fixture.detectChanges();
      expect(menu?.classList.contains('navbar-menu-open')).toBe(true);
      expect(toggle?.getAttribute('aria-expanded')).toBe('true');

      toggle!.click();
      fixture.detectChanges();
      expect(menu?.classList.contains('navbar-menu-open')).toBe(false);
      expect(toggle?.getAttribute('aria-expanded')).toBe('false');
    });

    it('Escape closes an open menu', async () => {
      const fixture = await build(false);
      const el = fixture.nativeElement as HTMLElement;
      el.querySelector<HTMLButtonElement>('[data-testid="nav-menu-toggle"]')!.click();
      fixture.detectChanges();

      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
      fixture.detectChanges();
      expect(el.querySelector('[data-testid="nav-menu"]')?.classList.contains('navbar-menu-open')).toBe(false);
    });

    it('resizing back to desktop width clears an open menu', async () => {
      const fixture = await build(false);
      const el = fixture.nativeElement as HTMLElement;
      el.querySelector<HTMLButtonElement>('[data-testid="nav-menu-toggle"]')!.click();
      fixture.detectChanges();

      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1280 });
      window.dispatchEvent(new Event('resize'));
      fixture.detectChanges();
      expect(el.querySelector('[data-testid="nav-menu"]')?.classList.contains('navbar-menu-open')).toBe(false);
    });

    it('language toggles and the sign-in control live inside the menu; the CTA does not', async () => {
      const el = (await build(false)).nativeElement as HTMLElement;
      const menu = el.querySelector('[data-testid="nav-menu"]')!;
      // Sign-in and both language buttons are inside the dropdown.
      expect(menu.querySelector('[data-testid="signin"]')).toBeTruthy();
      expect(menu.textContent).toContain('EN');
      expect(menu.textContent).toContain('中文');
      // The primary CTA (landing-only, shown here because url is '/') stays in the bar.
      const cta = el.querySelector('[data-testid="nav-book-demo"]');
      expect(cta).toBeTruthy();
      expect(cta?.closest('[data-testid="nav-menu"]')).toBeNull();
    });

    it('picking a language closes the mobile dropdown', async () => {
      const fixture = await build(false);
      const el = fixture.nativeElement as HTMLElement;
      const toggle = el.querySelector<HTMLButtonElement>('[data-testid="nav-menu-toggle"]');
      const menu = el.querySelector('[data-testid="nav-menu"]');

      toggle!.click();
      fixture.detectChanges();
      expect(menu?.classList.contains('navbar-menu-open')).toBe(true);

      const zhButton = Array.from(menu!.querySelectorAll('button')).find((b) => b.textContent?.includes('中文'));
      zhButton!.click();
      fixture.detectChanges();
      expect(menu?.classList.contains('navbar-menu-open')).toBe(false);
    });

    it('the hamburger toggle aria-label is wired through the translate pipe', async () => {
      const el = (await build(false)).nativeElement as HTMLElement;
      const toggle = el.querySelector<HTMLButtonElement>('[data-testid="nav-menu-toggle"]');
      // With no translations loaded in this TestBed, the pipe renders the raw
      // key — which still proves the aria-label binding is wired up correctly.
      expect(toggle?.getAttribute('aria-label')).toBe('nav.menu');
    });
  });
});
