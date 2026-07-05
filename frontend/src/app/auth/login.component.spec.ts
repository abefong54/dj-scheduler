import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, provideRouter } from '@angular/router';
import { provideTranslateService } from '@ngx-translate/core';

import { LoginComponent } from './login.component';
import { AuthService } from '../services/auth.service';

describe('LoginComponent (EL-079 Soundcheck reskin)', () => {
  let signInWithGoogle: ReturnType<typeof vi.fn>;
  let signInWithLine: ReturnType<typeof vi.fn>;

  async function build(error?: string): Promise<ComponentFixture<LoginComponent>> {
    signInWithGoogle = vi.fn();
    signInWithLine = vi.fn();
    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        provideRouter([]),
        provideTranslateService(),
        { provide: AuthService, useValue: { signInWithGoogle, signInWithLine } },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParamMap: convertToParamMap(error ? { error } : {}) } },
        },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(LoginComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('renders the dark-booth card with the Soundcheck wordmark on token surfaces', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const wordmark = el.querySelector('h1');
    expect(wordmark?.textContent).toContain('Soundcheck');
    expect(wordmark?.className).toContain('font-display');
    // Wordmark uses the semantic ink token (no hard-coded white).
    expect(wordmark?.getAttribute('style')).toContain('var(--ink)');

    // Full-height field sits on the Soundcheck Booth Black canvas token.
    const section = el.querySelector('section');
    expect(section?.getAttribute('style')).toContain('var(--surface)');

    // Card is the raised Console Slate surface, bordered with the token hairline.
    const card = section?.querySelector('div');
    expect(card?.getAttribute('style')).toContain('var(--card)');
    expect(card?.getAttribute('style')).toContain('var(--border)');
  });

  it('styles the Google action with the shared Cue-Amber .btn-primary', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const googleBtn = el.querySelector<HTMLButtonElement>('[data-testid="google-signin"]');
    expect(googleBtn?.className).toContain('btn');
    expect(googleBtn?.className).toContain('btn-primary');
  });

  it('keeps Google OAuth as the working primary action', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const googleBtn = el.querySelector<HTMLButtonElement>('[data-testid="google-signin"]');
    expect(googleBtn).toBeTruthy();
    expect(googleBtn?.disabled).toBe(false);

    googleBtn!.click();
    expect(signInWithGoogle).toHaveBeenCalledTimes(1);
  });

  it('offers a working LINE sign-in that calls signInWithLine', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const lineBtn = el.querySelector<HTMLButtonElement>('[data-testid="line-signin"]');
    expect(lineBtn).toBeTruthy();
    expect(lineBtn?.disabled).toBe(false);

    lineBtn!.click();
    expect(signInWithLine).toHaveBeenCalledTimes(1);
  });

  it('shows a friendly message when LINE returned no email (?error=no_email)', async () => {
    const fixture = await build('no_email');
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-testid="login-error"]')).toBeTruthy();
  });

  it('shows an email-consent notice that links the privacy policy', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const consent = el.querySelector('[data-testid="login-consent"]');
    expect(consent).toBeTruthy();
    expect(consent!.querySelector('a')?.getAttribute('href')).toBe('/privacy');
  });

  it('shows no error banner when there is no error param', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-testid="login-error"]')).toBeNull();
  });

  it('renders the magic-link control as visual-only / disabled (no backend)', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const email = el.querySelector<HTMLInputElement>('[data-testid="magic-link-email"]');
    const submit = el.querySelector<HTMLButtonElement>('[data-testid="magic-link-submit"]');
    expect(email?.disabled).toBe(true);
    expect(submit?.disabled).toBe(true);
  });
});
