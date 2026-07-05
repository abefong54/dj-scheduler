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
    expect(brand?.getAttribute('aria-label')).toBe('Soundcheck home');
  });

  describe('nav "Book a demo" button (landing-only)', () => {
    async function buildWithRouter(authed: boolean): Promise<{ fixture: ComponentFixture<AppComponent>; router: Router }> {
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
      const link = (fixture.nativeElement as HTMLElement)
        .querySelector<HTMLAnchorElement>('[data-testid="nav-console"]');
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
});
