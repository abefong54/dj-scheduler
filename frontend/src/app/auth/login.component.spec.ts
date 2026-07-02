import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideTranslateService } from '@ngx-translate/core';

import { LoginComponent } from './login.component';
import { AuthService } from '../services/auth.service';

describe('LoginComponent (EL-079 Soundcheck reskin)', () => {
  let signInWithGoogle: ReturnType<typeof vi.fn>;

  async function build(): Promise<ComponentFixture<LoginComponent>> {
    signInWithGoogle = vi.fn();
    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        provideTranslateService(),
        { provide: AuthService, useValue: { signInWithGoogle } },
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

  it('renders the magic-link control as visual-only / disabled (no backend)', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const email = el.querySelector<HTMLInputElement>('[data-testid="magic-link-email"]');
    const submit = el.querySelector<HTMLButtonElement>('[data-testid="magic-link-submit"]');
    expect(email?.disabled).toBe(true);
    expect(submit?.disabled).toBe(true);
  });
});
