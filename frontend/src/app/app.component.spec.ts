import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
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
});
